package coakka_v2_connector

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type DeadletterError struct {
	Deadletter *Deadletter
}

func (e *DeadletterError) Error() string {
	if e == nil || e.Deadletter == nil {
		return "deadletter"
	}
	return fmt.Sprintf("deadletter reason=%d detail=%s", e.Deadletter.Reason, e.Deadletter.Detail)
}

type RuntimeClientStats struct {
	PendingRequests                int `json:"pendingRequests"`
	DeliveredRequests              int `json:"deliveredRequests"`
	MatchedResponses               int `json:"matchedResponses"`
	MatchedDeadletters             int `json:"matchedDeadletters"`
	TerminalEventDropCount         int `json:"terminalEventDropCount"`
	DeadletterObservationDropCount int `json:"deadletterObservationDropCount"`
}

type RuntimeConnectorSnapshot struct {
	RuntimeClientStats
	SeparateDeliveredRequestLane bool   `json:"separateDeliveredRequestLane"`
	SystemName                   string `json:"systemName"`
	NodeID                       string `json:"nodeId"`
}

type RuntimeSnapshot struct {
	RuntimeInfo   RuntimeInfoSnapshot          `json:"runtimeInfo"`
	StartupConfig RuntimeStartupConfigSnapshot `json:"startupConfig"`
	RuntimeConfig RuntimeConfigSnapshot        `json:"runtimeConfig"`
	NetworkConfig RuntimeNetworkSnapshot       `json:"networkConfig"`
	Connector     RuntimeConnectorSnapshot     `json:"connector"`
	Health        RuntimeHealthSnapshot        `json:"health"`
	Stats         RuntimeStatsSnapshot         `json:"stats"`
}

type MonitorSnapshot struct {
	SignalCount uint64                `json:"signalCount"`
	Health      RuntimeHealthSnapshot `json:"health"`
	Stats       RuntimeStatsSnapshot  `json:"stats"`
}

// HandlerFn runs bounded application work. Do not synchronously Close this host
// from its handler: Close joins the caller's handler/reader lifetime.
type HandlerFn func(request *Envelope) *Envelope

// PayloadHandlerFn borrows the synchronous reader turn. Return promptly and
// request host shutdown from outside that turn, never synchronously Close it.
type PayloadHandlerFn func(payload []byte) []byte

type registeredHandler struct {
	handler        HandlerFn
	payloadHandler PayloadHandlerFn
	typedReplies   bool
}

type deliveredRequestJob struct {
	request    *nativeDeliveredRequest
	registered registeredHandler
}

type handlerScheduling struct {
	capacity int
	workers  int
	queue    *handlerDispatchQueue
}

type encodedPayloadSource struct {
	source  nativeCString
	replyTo nativeCString
}

type runtimeClientCounters struct {
	corePending                atomic.Int64
	deliveredRequests          atomic.Int64
	matchedResponses           atomic.Int64
	matchedDeadletters         atomic.Int64
	terminalEventDrops         atomic.Int64
	deadletterObservationDrops atomic.Int64
}

const (
	payloadEncodingCacheCapacity = 64
	payloadEncodingCacheBytes    = 256 * 1024
	handlerRegistrationCapacity  = 256
	handlerRegistrationBytes     = 64 * 1024
	observerCapacity             = 64
	observerBufferCapacity       = 1024
)

type GoRuntimeClient struct {
	bindings               *nativeBindings
	runtime                nativeRuntime
	hostHandles            HostHandlesSnapshot
	requestReader          nativeRequestReader
	askClient              nativeAskClient
	payloadMessagePrefix   []byte
	nextPayloadMessageID   atomic.Uint64
	handlerSchedule        handlerScheduling
	lastDispatchTarget     []byte
	lastDispatchHandler    registeredHandler
	lastDispatchHandlerSet bool

	payloadEncodingMu    sync.Mutex
	payloadSources       sync.Map
	payloadTargets       sync.Map
	payloadSourceCount   int
	payloadTargetCount   int
	payloadRetainedBytes int

	handlersMu         sync.RWMutex
	handlers           map[string]registeredHandler
	handlerTargetBytes int

	terminalSubscribersMu    sync.Mutex
	terminalSubscribers      map[uint64]chan RequestTerminalEvent
	nextTerminalSubscriberID uint64

	deadletterSubscribersMu    sync.Mutex
	deadletterSubscribers      map[uint64]chan ObservedDeadletter
	nextDeadletterSubscriberID uint64

	counters              runtimeClientCounters
	transportMu           sync.Mutex
	startupSecurityResult *RuntimeTCPSecurityApplyResult

	monitorSignalCh chan uint64
	lifetime        runtimeLifetime
	backgroundWG    sync.WaitGroup
}

func NewGoRuntimeClient(runtimeLibPath string, config ConnectorConfig) (*GoRuntimeClient, error) {
	if err := config.RequireValid(); err != nil {
		return nil, err
	}
	bindings, err := openNativeBindings(runtimeLibPath)
	if err != nil {
		return nil, err
	}
	if abi := bindings.getAbiVersion(); abi != COAKKAABIVersion {
		bindings.close()
		return nil, fmt.Errorf("unexpected ABI version: %d", abi)
	}
	runtime, err := bindings.createRuntime(config)
	if err != nil {
		bindings.close()
		return nil, err
	}
	client := &GoRuntimeClient{
		bindings:              bindings,
		runtime:               runtime,
		handlers:              make(map[string]registeredHandler),
		terminalSubscribers:   make(map[uint64]chan RequestTerminalEvent),
		deadletterSubscribers: make(map[uint64]chan ObservedDeadletter),
		monitorSignalCh:       make(chan uint64, 64),
		lifetime:              runtimeLifetime{closed: make(chan struct{})},
	}
	if err := bindings.applyNetworkOptions(runtime, config.Network); err != nil {
		bindings.destroyRuntime(runtime)
		bindings.close()
		return nil, err
	}
	if config.Security != nil {
		result, applyErr := bindings.applyTCPSecurity(runtime, *config.Security)
		if applyErr != nil {
			bindings.destroyRuntime(runtime)
			bindings.close()
			return nil, applyErr
		}
		client.startupSecurityResult = &result
		if !result.Applied() {
			bindings.destroyRuntime(runtime)
			bindings.close()
			return nil, &RuntimeTCPSecurityApplyError{Result: result}
		}
	}
	initial, err := bindings.getStartupConfig(runtime)
	if err != nil {
		// No host handles have been exported yet; there is no descriptor owner
		// to close here (a zero-valued handle record is not an empty owner).
		bindings.destroyRuntime(runtime)
		bindings.close()
		return nil, err
	}
	var handleFlags uint32
	if initial.Monitor.Enabled {
		handleFlags = HostHandlesEnableMonitor
	}
	hostHandles, err := bindings.getHostHandles(runtime, handleFlags)
	if err != nil {
		bindings.destroyRuntime(runtime)
		bindings.close()
		return nil, err
	}
	client.hostHandles = hostHandles
	controlBytes, err := client.buildControlEnvelopeBytes(
		config.Generation,
		config.Routes,
		config.SystemName,
		1,
		config.OverloadPolicy,
	)
	if err != nil {
		client.cleanupNative()
		return nil, err
	}
	if err := bindings.applyControlEnvelope(runtime, controlBytes); err != nil {
		client.cleanupNative()
		return nil, err
	}
	if err := bindings.startRuntime(runtime); err != nil {
		client.cleanupNative()
		return nil, err
	}
	requestReader, err := bindings.createRequestReader(runtime, hostHandles)
	if err != nil {
		_ = bindings.stopRuntime(runtime)
		client.cleanupNative()
		return nil, err
	}
	client.requestReader = requestReader
	askClient, err := bindings.createAskClient(runtime, hostHandles)
	if err != nil {
		bindings.destroyRequestReader(requestReader)
		client.requestReader = nil
		_ = bindings.stopRuntime(runtime)
		client.cleanupNative()
		return nil, err
	}
	client.askClient = askClient
	client.payloadMessagePrefix = append([]byte(config.NodeID), ".go."...)
	startup, err := bindings.getStartupConfig(runtime)
	if err != nil {
		bindings.destroyAskClient(askClient)
		client.askClient = nil
		bindings.destroyRequestReader(requestReader)
		client.requestReader = nil
		_ = bindings.stopRuntime(runtime)
		client.cleanupNative()
		return nil, err
	}
	handlerWorkers := int(startup.Execution.SelectedCPUCount)
	if handlerWorkers < 1 {
		handlerWorkers = 1
	}
	client.handlerSchedule = handlerScheduling{capacity: int(startup.QueueCapacity), workers: handlerWorkers}
	client.startRequestReader()
	if hostHandles.Flags&HostHandlesEnableMonitor != 0 {
		client.startMonitorReader(hostHandles.MonitorReadFD)
	}
	return client, nil
}

func NextMessageID(source string) string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s-%d", source, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s-%s", source, hex.EncodeToString(buf))
}

func MakeTypedRequest(source string, target string, payload []byte, payloadIdentity PayloadIdentity, timeout time.Duration, operation string, deliveryHint DeliveryHint, headers map[string]string, messageID string, oneWay bool) (*Envelope, error) {
	timeoutMillis, err := envelopeTimeoutMillis(timeout)
	if err != nil {
		return nil, err
	}
	if err := payloadIdentity.RequireTyped("typed request"); err != nil {
		return nil, err
	}
	if messageID == "" {
		messageID = NextMessageID(source)
	}
	outHeaders := map[string]string{
		"method":    "POST",
		"operation": operation,
	}
	for key, value := range headers {
		outHeaders[key] = value
	}
	envelope := &Envelope{
		MessageId:            messageID,
		CorrelationId:        "",
		Source:               source,
		Target:               target,
		ReplyTo:              "",
		Kind:                 MessageKindRequest,
		OneWay:               oneWay,
		TimeoutMs:            timeoutMillis,
		Payload:              payload,
		Headers:              outHeaders,
		Status:               BusinessStatusOK,
		ErrorCode:            "",
		ErrorMessage:         "",
		DeliveryHint:         deliveryHint,
		MessageType:          payloadIdentity.MessageType,
		PayloadSchemaVersion: payloadIdentity.PayloadSchemaVersion,
		PayloadFormat:        payloadIdentity.PayloadFormat,
	}
	if !oneWay {
		envelope.CorrelationId = messageID
		envelope.ReplyTo = fmt.Sprintf("%s/replies", source)
	}
	return envelope, nil
}

func MakeJSONRequest(source string, target string, payload any, payloadIdentity PayloadIdentity, timeout time.Duration, operation string, deliveryHint DeliveryHint, headers map[string]string, messageID string, oneWay bool) (*Envelope, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return MakeTypedRequest(source, target, bytes, payloadIdentity, timeout, operation, deliveryHint, headers, messageID, oneWay)
}

func MakeTypedReply(request *Envelope, source string, payload []byte, payloadIdentity PayloadIdentity) (*Envelope, error) {
	if err := payloadIdentity.RequireTyped("typed reply"); err != nil {
		return nil, err
	}
	correlationID := request.GetCorrelationId()
	if correlationID == "" {
		correlationID = request.GetMessageId()
	}
	return &Envelope{
		MessageId:            fmt.Sprintf("%s.reply.%s", request.GetMessageId(), source),
		CorrelationId:        correlationID,
		Source:               source,
		Target:               request.GetSource(),
		ReplyTo:              "",
		Kind:                 MessageKindResponse,
		OneWay:               false,
		TimeoutMs:            request.GetTimeoutMs(),
		Payload:              payload,
		Headers:              map[string]string{},
		Status:               BusinessStatusOK,
		ErrorCode:            "",
		ErrorMessage:         "",
		DeliveryHint:         DeliveryHintRouterDefault,
		MessageType:          payloadIdentity.MessageType,
		PayloadSchemaVersion: payloadIdentity.PayloadSchemaVersion,
		PayloadFormat:        payloadIdentity.PayloadFormat,
	}, nil
}

func MakeJSONReply(request *Envelope, source string, payload any, payloadIdentity PayloadIdentity) (*Envelope, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return MakeTypedReply(request, source, bytes, payloadIdentity)
}

func MakeReplyFromRequestIdentity(request *Envelope, source string, payload []byte) (*Envelope, error) {
	identity, err := RequireTypedPayloadIdentity(request, "request")
	if err != nil {
		return nil, err
	}
	return MakeTypedReply(request, source, payload, identity)
}

func MakeJSONReplyFromRequestIdentity(request *Envelope, source string, payload any) (*Envelope, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return MakeReplyFromRequestIdentity(request, source, bytes)
}

func MakeRawRequest(messageID string, source string, target string, replyTo string, payload []byte, timeout time.Duration, deliveryHint DeliveryHint) (*Envelope, error) {
	timeoutMillis, err := envelopeTimeoutMillis(timeout)
	if err != nil {
		return nil, err
	}
	if messageID == "" {
		messageID = NextMessageID(source)
	}
	return &Envelope{
		MessageId:     messageID,
		CorrelationId: messageID,
		Source:        source,
		Target:        target,
		ReplyTo:       replyTo,
		Kind:          MessageKindRequest,
		OneWay:        false,
		TimeoutMs:     timeoutMillis,
		Payload:       payload,
		Headers:       map[string]string{},
		Status:        BusinessStatusOK,
		DeliveryHint:  deliveryHint,
	}, nil
}

func MakeRawOneWay(messageID string, source string, target string, payload []byte, timeout time.Duration, deliveryHint DeliveryHint) (*Envelope, error) {
	out, err := MakeRawRequest(messageID, source, target, "", payload, timeout, deliveryHint)
	if err != nil {
		return nil, err
	}
	out.CorrelationId = ""
	out.ReplyTo = ""
	out.OneWay = true
	return out, nil
}

func RequireTypedPayloadIdentity(envelope *Envelope, what string) (PayloadIdentity, error) {
	identity := PayloadIdentity{
		MessageType:          envelope.GetMessageType(),
		PayloadSchemaVersion: envelope.GetPayloadSchemaVersion(),
		PayloadFormat:        PayloadFormat(envelope.GetPayloadFormat()),
	}
	return identity, identity.RequireTyped(what)
}

func (c *GoRuntimeClient) AskTyped(source string, target string, payload []byte, payloadIdentity PayloadIdentity, timeout time.Duration, operation string, deliveryHint DeliveryHint, headers map[string]string) (*Envelope, error) {
	envelope, err := MakeTypedRequest(source, target, payload, payloadIdentity, timeout, operation, deliveryHint, headers, NextMessageID(source), false)
	if err != nil {
		return nil, err
	}
	return c.submitPendingRequest(envelope, timeout)
}

func (c *GoRuntimeClient) AskJSON(source string, target string, payload any, payloadIdentity PayloadIdentity, timeout time.Duration, operation string, deliveryHint DeliveryHint, headers map[string]string) (map[string]any, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	response, err := c.AskTyped(source, target, bytes, payloadIdentity, timeout, operation, deliveryHint, headers)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(response.GetPayload(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// AskPayload performs one raw request/reply without constructing connector
// Envelope values. Core builds and validates the request, matches the terminal
// result and projects the response payload into Go-owned bytes.
func (c *GoRuntimeClient) AskPayload(source string, target string, payload []byte, timeout time.Duration, deliveryHint DeliveryHint) ([]byte, error) {
	if !c.beginCall() {
		return nil, ErrClosed
	}
	defer c.endCall()
	timeoutMillis, err := coreTimeoutMillis(timeout)
	if err != nil {
		return nil, err
	}
	if timeoutMillis > uint32(^uint32(0)>>1) {
		return nil, fmt.Errorf("request timeout exceeds Core int32 millisecond range")
	}
	sequence := c.nextPayloadMessageID.Add(1)
	messageID := make(nativeCString, 0, len(c.payloadMessagePrefix)+20+1)
	messageID = append(messageID, c.payloadMessagePrefix...)
	messageID = strconv.AppendUint(messageID, sequence, 10)
	messageID = append(messageID, 0)
	encodedSource, err := c.encodedPayloadSource(source)
	if err != nil {
		return nil, err
	}
	encodedTarget, err := c.encodedPayloadTarget(target)
	if err != nil {
		return nil, err
	}
	c.counters.corePending.Add(1)
	defer func() {
		c.counters.corePending.Add(-1)
	}()
	kind, resultBytes, status, err := c.bindings.askPayload(
		c.askClient,
		messageID,
		encodedSource.source,
		encodedTarget,
		encodedSource.replyTo,
		payload,
		timeoutMillis,
		timeoutMillis,
		deliveryHint,
	)
	if err != nil {
		return nil, err
	}
	if status == COAKKAErrWouldBlock {
		return nil, nativeStatusError("ask_payload", status)
	}
	if status == COAKKAErrClosed {
		return nil, nativeStatusError("ask_payload", status)
	}
	if kind == nativeClientResultResponse {
		c.counters.matchedResponses.Add(1)
		return resultBytes, nil
	}
	if kind == nativeClientResultDeadletter {
		deadletter, decodeErr := c.bindings.decodeDeadletter(resultBytes)
		if decodeErr != nil {
			return nil, decodeErr
		}
		c.counters.matchedDeadletters.Add(1)
		return nil, &DeadletterError{Deadletter: deadletter}
	}
	return nil, fmt.Errorf("Core returned unknown ask result kind=%d", kind)
}

func (c *GoRuntimeClient) encodedPayloadSource(value string) (encodedPayloadSource, error) {
	if cached, ok := c.payloadSources.Load(value); ok {
		return cached.(encodedPayloadSource), nil
	}
	source, err := makeNativeCString("source", value)
	if err != nil {
		return encodedPayloadSource{}, err
	}
	replyTo := make(nativeCString, 0, len(value)+len("/replies")+1)
	replyTo = append(replyTo, value...)
	replyTo = append(replyTo, "/replies"...)
	replyTo = append(replyTo, 0)
	encoded := encodedPayloadSource{source: source, replyTo: replyTo}
	c.payloadEncodingMu.Lock()
	if existing, exists := c.payloadSources.Load(value); exists {
		encoded = existing.(encodedPayloadSource)
	} else if cost := 3*len(value) + len("/replies") + 64; c.payloadSourceCount < payloadEncodingCacheCapacity && cost <= payloadEncodingCacheBytes-c.payloadRetainedBytes {
		// Own the key: a tiny substring must not retain the caller's huge backing.
		c.payloadSources.Store(strings.Clone(value), encoded)
		c.payloadSourceCount++
		c.payloadRetainedBytes += cost
	}
	c.payloadEncodingMu.Unlock()
	return encoded, nil
}

func (c *GoRuntimeClient) encodedPayloadTarget(value string) (nativeCString, error) {
	if cached, ok := c.payloadTargets.Load(value); ok {
		return cached.(nativeCString), nil
	}
	encoded, err := makeNativeCString("target", value)
	if err != nil {
		return nil, err
	}
	c.payloadEncodingMu.Lock()
	if existing, exists := c.payloadTargets.Load(value); exists {
		encoded = existing.(nativeCString)
	} else if cost := 2*len(value) + 32; c.payloadTargetCount < payloadEncodingCacheCapacity && cost <= payloadEncodingCacheBytes-c.payloadRetainedBytes {
		c.payloadTargets.Store(strings.Clone(value), encoded)
		c.payloadTargetCount++
		c.payloadRetainedBytes += cost
	}
	c.payloadEncodingMu.Unlock()
	return encoded, nil
}

func (c *GoRuntimeClient) AskRaw(request *Envelope, timeout time.Duration) (*Envelope, error) {
	// Core validates request kind/identity and normalizes missing correlation.
	// The explicit Go wait budget is never reconstructed from envelope input.
	return c.submitPendingRequest(request, timeout)
}

func (c *GoRuntimeClient) SubmitRequestTyped(source string, target string, payload []byte, payloadIdentity PayloadIdentity, timeout time.Duration, operation string, deliveryHint DeliveryHint, headers map[string]string) (*SubmittedRequest, error) {
	envelope, err := MakeTypedRequest(source, target, payload, payloadIdentity, timeout, operation, deliveryHint, headers, NextMessageID(source), false)
	if err != nil {
		return nil, err
	}
	return c.submitTicketRequest(envelope)
}

func (c *GoRuntimeClient) SubmitRequestJSON(source string, target string, payload any, payloadIdentity PayloadIdentity, timeout time.Duration, operation string, deliveryHint DeliveryHint, headers map[string]string) (*SubmittedRequest, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return c.SubmitRequestTyped(source, target, bytes, payloadIdentity, timeout, operation, deliveryHint, headers)
}

func (c *GoRuntimeClient) SubmitRequestRaw(request *Envelope) (*SubmittedRequest, error) {
	return c.submitTicketRequest(request)
}

func (c *GoRuntimeClient) SendOneWayTyped(source string, target string, payload []byte, payloadIdentity PayloadIdentity, deliveryHint DeliveryHint, headers map[string]string) error {
	envelope, err := MakeTypedRequest(source, target, payload, payloadIdentity, 0, "one_way", deliveryHint, headers, NextMessageID(source), true)
	if err != nil {
		return err
	}
	return c.SubmitTypedEnvelope(envelope)
}

func (c *GoRuntimeClient) SendOneWayJSON(source string, target string, payload any, payloadIdentity PayloadIdentity, deliveryHint DeliveryHint, headers map[string]string) error {
	envelope, err := MakeJSONRequest(source, target, payload, payloadIdentity, 0, "one_way", deliveryHint, headers, NextMessageID(source), true)
	if err != nil {
		return err
	}
	return c.SubmitTypedEnvelope(envelope)
}

func (c *GoRuntimeClient) SubmitEnvelope(envelope *Envelope) error {
	return c.SubmitRawEnvelope(envelope)
}

func (c *GoRuntimeClient) SubmitTypedEnvelope(envelope *Envelope) error {
	if _, err := RequireTypedPayloadIdentity(envelope, "SubmitTypedEnvelope"); err != nil {
		return err
	}
	return c.SubmitRawEnvelope(envelope)
}

func (c *GoRuntimeClient) SubmitRawEnvelope(envelope *Envelope) error {
	if !c.beginCall() {
		return ErrClosed
	}
	defer c.endCall()
	bytes, err := c.bindings.encodeEnvelope(envelope)
	if err != nil {
		return err
	}
	// Core owns concurrent ingress admission and its serialization. A Go-side
	// mutex here would duplicate Core policy and unnecessarily couple callers.
	return c.bindings.submitEnvelope(c.runtime, bytes)
}

func (c *GoRuntimeClient) ApplySnapshot(generation uint64, routes []RouteSpec, sourceConnector string, seq uint64) error {
	return c.ApplySnapshotWithPolicy(generation, routes, sourceConnector, seq, nil)
}

func (c *GoRuntimeClient) ApplySnapshotWithPolicy(
	generation uint64,
	routes []RouteSpec,
	sourceConnector string,
	seq uint64,
	overloadPolicy *RuntimeOverloadPolicy,
) error {
	if !c.beginCall() {
		return ErrClosed
	}
	defer c.endCall()
	bytes, err := c.buildControlEnvelopeBytes(generation, routes, sourceConnector, seq, overloadPolicy)
	if err != nil {
		return err
	}
	return c.bindings.applyControlEnvelope(c.runtime, bytes)
}

func (c *GoRuntimeClient) RegisterHandler(target string, handler HandlerFn, typedReplies bool) error {
	if !c.beginCall() {
		return ErrClosed
	}
	defer c.endCall()
	if handler == nil {
		return fmt.Errorf("handler must not be nil")
	}
	if err := validateNativeText([2]string{"handler target", target}); err != nil {
		return err
	}
	c.handlersMu.Lock()
	defer c.handlersMu.Unlock()
	if _, exists := c.handlers[target]; exists {
		return fmt.Errorf("handler already registered for target=%s", target)
	}
	if len(c.handlers) >= handlerRegistrationCapacity || len(target) > handlerRegistrationBytes-c.handlerTargetBytes {
		return fmt.Errorf("Go handler registration capacity exhausted")
	}
	if c.handlerSchedule.queue == nil {
		queue := newHandlerDispatchQueue(c.handlerSchedule.capacity)
		c.handlerSchedule.queue = queue
		c.startHandlerWorkers(queue, c.handlerSchedule.workers)
	}
	c.handlers[strings.Clone(target)] = registeredHandler{handler: handler, typedReplies: typedReplies}
	c.handlerTargetBytes += len(target)
	return nil
}

func (c *GoRuntimeClient) RegisterRawHandler(target string, handler HandlerFn) error {
	return c.RegisterHandler(target, handler, false)
}

// RegisterPayloadHandler installs the allocation-lean raw payload path. Core
// remains responsible for request validation, target/one-way projection and
// reply identity. The synchronous handler runs on the sole request-reader turn
// and must return promptly; use RegisterHandler for work that needs dispatch to
// the bounded handler workers.
func (c *GoRuntimeClient) RegisterPayloadHandler(target string, handler PayloadHandlerFn) error {
	if !c.beginCall() {
		return ErrClosed
	}
	defer c.endCall()
	if err := validateNativeText([2]string{"handler target", target}); err != nil {
		return err
	}
	if handler == nil {
		return fmt.Errorf("payload handler must not be nil")
	}
	c.handlersMu.Lock()
	defer c.handlersMu.Unlock()
	if _, exists := c.handlers[target]; exists {
		return fmt.Errorf("handler already registered for target=%s", target)
	}
	if len(c.handlers) >= handlerRegistrationCapacity || len(target) > handlerRegistrationBytes-c.handlerTargetBytes {
		return fmt.Errorf("Go handler registration capacity exhausted")
	}
	c.handlers[strings.Clone(target)] = registeredHandler{payloadHandler: handler}
	c.handlerTargetBytes += len(target)
	return nil
}

func (c *GoRuntimeClient) TerminalEvents(ctx context.Context, buffer int) (<-chan RequestTerminalEvent, error) {
	if !c.beginCall() {
		return nil, ErrClosed
	}
	defer c.endCall()
	if ctx == nil {
		ctx = context.Background()
	}
	if buffer <= 0 {
		buffer = 16
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if buffer > observerBufferCapacity {
		return nil, fmt.Errorf("Go observer buffer exceeds %d slots", observerBufferCapacity)
	}
	events := make(chan RequestTerminalEvent, buffer)
	subscriberID, err := c.addTerminalSubscriber(events)
	if err != nil {
		return nil, err
	}
	go func() {
		select {
		case <-ctx.Done():
		case <-c.lifetime.closed:
		}
		c.removeTerminalSubscriber(subscriberID, events)
	}()
	return events, nil
}

func (c *GoRuntimeClient) Deadletters(ctx context.Context, buffer int) (<-chan ObservedDeadletter, error) {
	if !c.beginCall() {
		return nil, ErrClosed
	}
	defer c.endCall()
	if ctx == nil {
		ctx = context.Background()
	}
	if buffer <= 0 {
		buffer = 128
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if buffer > observerBufferCapacity {
		return nil, fmt.Errorf("Go observer buffer exceeds %d slots", observerBufferCapacity)
	}
	events := make(chan ObservedDeadletter, buffer)
	subscriberID, err := c.addDeadletterSubscriber(events)
	if err != nil {
		return nil, err
	}
	go func() {
		select {
		case <-ctx.Done():
		case <-c.lifetime.closed:
		}
		c.removeDeadletterSubscriber(subscriberID, events)
	}()
	return events, nil
}

func (c *GoRuntimeClient) SnapshotStats() RuntimeClientStats {
	return RuntimeClientStats{
		PendingRequests:                int(c.counters.corePending.Load()),
		DeliveredRequests:              int(c.counters.deliveredRequests.Load()),
		MatchedResponses:               int(c.counters.matchedResponses.Load()),
		MatchedDeadletters:             int(c.counters.matchedDeadletters.Load()),
		TerminalEventDropCount:         int(c.counters.terminalEventDrops.Load()),
		DeadletterObservationDropCount: int(c.counters.deadletterObservationDrops.Load()),
	}
}

func (c *GoRuntimeClient) HealthSnapshot() (RuntimeHealthSnapshot, error) {
	if !c.beginCall() {
		return RuntimeHealthSnapshot{}, ErrClosed
	}
	defer c.endCall()
	return c.bindings.getHealth(c.runtime)
}

func (c *GoRuntimeClient) StatsSnapshot() (RuntimeStatsSnapshot, error) {
	if !c.beginCall() {
		return RuntimeStatsSnapshot{}, ErrClosed
	}
	defer c.endCall()
	return c.bindings.getStats(c.runtime)
}

func (c *GoRuntimeClient) RuntimeInfoSnapshot() (RuntimeInfoSnapshot, error) {
	if !c.beginCall() {
		return RuntimeInfoSnapshot{}, ErrClosed
	}
	defer c.endCall()
	return c.bindings.readRuntimeInfo()
}

func (c *GoRuntimeClient) RuntimeConfigSnapshot() (RuntimeConfigSnapshot, error) {
	if !c.beginCall() {
		return RuntimeConfigSnapshot{}, ErrClosed
	}
	defer c.endCall()
	c.transportMu.Lock()
	defer c.transportMu.Unlock()
	return c.bindings.getConfig(c.runtime)
}

func (c *GoRuntimeClient) RuntimeStartupConfigSnapshot() (RuntimeStartupConfigSnapshot, error) {
	if !c.beginCall() {
		return RuntimeStartupConfigSnapshot{}, ErrClosed
	}
	defer c.endCall()
	return c.bindings.getStartupConfig(c.runtime)
}

func (c *GoRuntimeClient) MonitorIsEnabled() (bool, error) {
	startup, err := c.RuntimeStartupConfigSnapshot()
	return startup.Monitor.Enabled, err
}

func (c *GoRuntimeClient) MonitorSnapshot(signalCount uint64) (MonitorSnapshot, error) {
	if !c.beginCall() {
		return MonitorSnapshot{}, ErrClosed
	}
	defer c.endCall()
	health, err := c.bindings.getHealth(c.runtime)
	if err != nil {
		return MonitorSnapshot{}, err
	}
	stats, err := c.bindings.getStats(c.runtime)
	if err != nil {
		return MonitorSnapshot{}, err
	}
	return MonitorSnapshot{SignalCount: signalCount, Health: health, Stats: stats}, nil
}

func (c *GoRuntimeClient) AwaitNextMonitor(timeout time.Duration) (*MonitorSnapshot, error) {
	enabled, err := c.MonitorIsEnabled()
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, &FeatureUnavailableError{Feature: "runtime monitor collection"}
	}
	if timeout <= 0 {
		timeout = time.Second
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-c.lifetime.closed:
		return nil, ErrClosed
	case signalCount := <-c.monitorSignalCh:
		snapshot, err := c.MonitorSnapshot(signalCount)
		if err != nil {
			return nil, err
		}
		return &snapshot, nil
	case <-timer.C:
		return nil, nil
	}
}

func (c *GoRuntimeClient) AwaitAppliedGenerationAtLeast(generation uint64, timeout time.Duration) (*MonitorSnapshot, error) {
	current, err := c.MonitorSnapshot(0)
	if err != nil {
		return nil, err
	}
	if current.Health.AppliedGeneration >= generation {
		return &current, nil
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		snapshot, err := c.AwaitNextMonitor(remaining)
		if err != nil || snapshot == nil {
			return snapshot, err
		}
		if snapshot.Health.AppliedGeneration >= generation {
			return snapshot, nil
		}
	}
	return nil, nil
}

// RuntimeSnapshot combines successive Core snapshots, not one atomic transaction.
func (c *GoRuntimeClient) RuntimeSnapshot() (RuntimeSnapshot, error) {
	if !c.beginCall() {
		return RuntimeSnapshot{}, ErrClosed
	}
	defer c.endCall()
	c.transportMu.Lock()
	defer c.transportMu.Unlock()
	info, err := c.bindings.readRuntimeInfo()
	if err != nil {
		return RuntimeSnapshot{}, err
	}
	startup, err := c.bindings.getStartupConfig(c.runtime)
	if err != nil {
		return RuntimeSnapshot{}, err
	}
	config, err := c.bindings.getConfig(c.runtime)
	if err != nil {
		return RuntimeSnapshot{}, err
	}
	health, err := c.bindings.getHealth(c.runtime)
	if err != nil {
		return RuntimeSnapshot{}, err
	}
	stats, err := c.bindings.getStats(c.runtime)
	if err != nil {
		return RuntimeSnapshot{}, err
	}
	network, err := c.bindings.getNetworkConfig(c.runtime)
	if err != nil {
		return RuntimeSnapshot{}, err
	}
	return RuntimeSnapshot{RuntimeInfo: info, StartupConfig: startup, RuntimeConfig: config, NetworkConfig: network,
		Connector: RuntimeConnectorSnapshot{RuntimeClientStats: c.SnapshotStats(), SeparateDeliveredRequestLane: c.hostHandles.Flags&HostHandlesSeparateDeliveredLane != 0, SystemName: config.SystemName, NodeID: config.NodeID},
		Health:    health, Stats: stats}, nil
}

func (c *GoRuntimeClient) Close() error {
	return c.lifetime.close(func() error {
		c.handlersMu.RLock()
		if c.handlerSchedule.queue != nil {
			c.handlerSchedule.queue.stop()
		}
		c.handlersMu.RUnlock()
		c.transportMu.Lock()
		defer c.transportMu.Unlock()
		return c.bindings.stopRuntime(c.runtime)
	}, c.backgroundWG.Wait, func() {
		c.bindings.destroyRequestReader(c.requestReader)
		c.requestReader = nil
		c.bindings.destroyAskClient(c.askClient)
		c.askClient = nil
		c.cleanupNative()
		c.retireGoState()
		c.terminalSubscribersMu.Lock()
		for id, events := range c.terminalSubscribers {
			delete(c.terminalSubscribers, id)
			close(events)
		}
		c.terminalSubscribersMu.Unlock()
		c.deadletterSubscribersMu.Lock()
		for id, events := range c.deadletterSubscribers {
			delete(c.deadletterSubscribers, id)
			close(events)
		}
		c.deadletterSubscribersMu.Unlock()
	})
}

// Called only after reader/handler and acquired-call drains. Release language
// callback roots and queued value copies even when the app retains a closed host.
func (c *GoRuntimeClient) retireGoState() {
	c.handlersMu.Lock()
	clear(c.handlers)
	c.handlerTargetBytes = 0
	c.handlersMu.Unlock()
	c.handlerSchedule.queue = nil
	c.lastDispatchTarget = nil
	c.lastDispatchHandler = registeredHandler{}
	c.lastDispatchHandlerSet = false
	c.payloadMessagePrefix = nil
	c.payloadEncodingMu.Lock()
	c.payloadSources.Range(func(key, value any) bool { c.payloadSources.Delete(key); return true })
	c.payloadTargets.Range(func(key, value any) bool { c.payloadTargets.Delete(key); return true })
	c.payloadSourceCount, c.payloadTargetCount, c.payloadRetainedBytes = 0, 0, 0
	c.payloadEncodingMu.Unlock()
}

func (c *GoRuntimeClient) beginCall() bool { return c.lifetime.acquire() }

func (c *GoRuntimeClient) endCall() { c.lifetime.release() }

func (c *GoRuntimeClient) cleanupNative() {
	c.bindings.closeHostHandles(c.hostHandles)
	c.bindings.destroyRuntime(c.runtime)
	c.bindings.close()
}

func (c *GoRuntimeClient) buildControlEnvelopeBytes(
	generation uint64,
	routes []RouteSpec,
	sourceConnector string,
	seq uint64,
	overloadPolicy *RuntimeOverloadPolicy,
) ([]byte, error) {
	// This is a full new caller snapshot, not a replay/default table retained
	// from construction. Core validates generation and routes; seq is diagnostic.
	return c.bindings.encodeRouteSnapshotControl(generation, routes, sourceConnector, seq, overloadPolicy)
}

func (c *GoRuntimeClient) submitPendingRequest(request *Envelope, timeout time.Duration) (*Envelope, error) {
	if !c.beginCall() {
		return nil, ErrClosed
	}
	defer c.endCall()
	timeoutMillis, err := coreTimeoutMillis(timeout)
	if err != nil {
		return nil, err
	}
	frame, err := c.bindings.encodeEnvelope(request)
	if err != nil {
		return nil, err
	}
	if c.isClosed() {
		return nil, ErrClosed
	}
	ticket, err := c.bindings.beginAsk(c.askClient, frame)
	if err != nil {
		return nil, err
	}
	defer c.bindings.destroyAskTicket(ticket)
	c.counters.corePending.Add(1)
	defer func() {
		c.counters.corePending.Add(-1)
	}()

	kind, resultFrame, status, err := c.bindings.awaitAsk(ticket, timeoutMillis)
	if err != nil {
		return nil, err
	}
	if status == COAKKAErrWouldBlock {
		return nil, nativeStatusError("ask_ticket_await", status)
	}
	if status == COAKKAErrClosed {
		return nil, nativeStatusError("ask_ticket_await", status)
	}
	switch kind {
	case nativeClientResultResponse:
		response, decodeErr := c.bindings.decodeEnvelope(resultFrame)
		if decodeErr != nil {
			return nil, decodeErr
		}
		c.counters.matchedResponses.Add(1)
		c.publishTerminalEvent(RequestTerminalEvent{
			Kind:             RequestTerminalEventResponse,
			RequestMessageID: response.GetCorrelationId(),
			CorrelationID:    response.GetCorrelationId(),
			Response:         response,
		})
		return response, nil
	case nativeClientResultDeadletter:
		deadletter, decodeErr := c.bindings.decodeDeadletter(resultFrame)
		if decodeErr != nil {
			return nil, decodeErr
		}
		c.counters.matchedDeadletters.Add(1)
		c.publishDeadletterObservation(deadletter)
		c.publishTerminalEvent(RequestTerminalEvent{
			Kind:             RequestTerminalEventDeadletter,
			RequestMessageID: deadletter.GetOriginalEnvelope().GetMessageId(),
			CorrelationID:    deadletter.GetOriginalEnvelope().GetCorrelationId(),
			Deadletter:       deadletter,
		})
		return nil, &DeadletterError{Deadletter: deadletter}
	default:
		return nil, fmt.Errorf("Core returned unknown ask result kind=%d", kind)
	}
}

func coreTimeoutMillis(timeout time.Duration) (uint32, error) {
	if timeout <= 0 {
		return 0, fmt.Errorf("timeout must be positive")
	}
	milliseconds := timeout / time.Millisecond
	if timeout%time.Millisecond != 0 {
		milliseconds++
	}
	if milliseconds > time.Duration(^uint32(0)) {
		return 0, fmt.Errorf("timeout exceeds Core uint32 millisecond range")
	}
	return uint32(milliseconds), nil
}

// Envelope timeout zero stays unspecified. A positive sub-millisecond value
// rounds up; negative or out-of-range intent must never wrap the int32 wire.
func envelopeTimeoutMillis(timeout time.Duration) (int32, error) {
	if timeout == 0 {
		return 0, nil
	}
	milliseconds, err := coreTimeoutMillis(timeout)
	if err != nil {
		return 0, err
	}
	if milliseconds > 2147483647 {
		return 0, fmt.Errorf("envelope timeout cannot cross the int32 millisecond boundary")
	}
	return int32(milliseconds), nil
}

func (c *GoRuntimeClient) submitTicketRequest(request *Envelope) (*SubmittedRequest, error) {
	if !c.beginCall() {
		return nil, ErrClosed
	}
	frame, err := c.bindings.encodeEnvelope(request)
	if err != nil {
		c.endCall()
		return nil, err
	}
	messageID := request.GetMessageId()
	ticket, err := c.bindings.beginAsk(c.askClient, frame)
	if err != nil {
		c.endCall()
		return nil, err
	}
	c.counters.corePending.Add(1)
	go func() {
		defer c.endCall()
		defer c.counters.corePending.Add(-1)
		defer c.bindings.destroyAskTicket(ticket)
		for {
			select {
			case <-c.lifetime.closed:
				c.bindings.cancelAsk(ticket)
				return
			default:
			}
			kind, resultFrame, status, awaitErr := c.bindings.awaitAsk(ticket, 100)
			if awaitErr != nil {
				if !c.isClosed() {
					log.Printf("[coakka-go] Core ask failed: %v", awaitErr)
				}
				return
			}
			if status == COAKKAErrWouldBlock {
				continue
			}
			if status == COAKKAErrClosed {
				return
			}
			switch kind {
			case nativeClientResultResponse:
				response, err := c.bindings.decodeEnvelope(resultFrame)
				if err == nil {
					c.observeCoreResponse(response)
				}
			case nativeClientResultDeadletter:
				deadletter, err := c.bindings.decodeDeadletter(resultFrame)
				if err == nil {
					c.observeCoreDeadletter(deadletter)
				}
			}
			return
		}
	}()
	return &SubmittedRequest{MessageID: messageID}, nil
}

func (c *GoRuntimeClient) startRequestReader() {
	c.backgroundWG.Add(1)
	go func() {
		defer c.backgroundWG.Done()
		staging := make([]byte, 64*1024)
		for {
			select {
			case <-c.lifetime.closed:
				return
			default:
			}
			request, status, err := c.bindings.readDeliveredRequest(c.requestReader, 100, staging)
			if err != nil {
				if !c.isClosed() {
					log.Printf("[coakka-go] Core request reader failed: %v", err)
				}
				return
			}
			if status == COAKKAErrWouldBlock {
				continue
			}
			if status == COAKKAErrClosed {
				return
			}
			c.dispatchDeliveredRequest(request)
		}
	}()
}

func (c *GoRuntimeClient) startHandlerWorkers(queue *handlerDispatchQueue, count int) {
	for worker := 0; worker < count; worker++ {
		c.backgroundWG.Add(1)
		go func() {
			defer c.backgroundWG.Done()
			for {
				select {
				case <-c.lifetime.closed:
					return
				case job := <-queue.jobs:
					c.handleDeliveredRequest(job)
					queue.release(job)
				}
			}
		}()
	}
}

func (c *GoRuntimeClient) startMonitorReader(fd int) {
	c.backgroundWG.Add(1)
	go func() {
		defer c.backgroundWG.Done()
		for {
			select {
			case <-c.lifetime.closed:
				return
			default:
			}
			signalCount, status, err := c.bindings.monitorWaitConsume(fd, 100)
			if err != nil {
				if !c.isClosed() {
					log.Printf("[coakka-go] Core monitor wait failed: %v", err)
				}
				return
			}
			if status == COAKKAErrWouldBlock {
				continue
			}
			if status == COAKKAErrClosed {
				return
			}
			select {
			case c.monitorSignalCh <- signalCount:
			default:
				select {
				case <-c.monitorSignalCh:
				default:
				}
				c.monitorSignalCh <- signalCount
			}
		}
	}()
}

func (c *GoRuntimeClient) dispatchDeliveredRequest(request nativeDeliveredRequest) {
	c.counters.deliveredRequests.Add(1)
	registered, ok := c.deliveredHandler(request.Target)
	if !ok {
		return
	}
	if registered.payloadHandler != nil {
		c.handlePayloadDeliveredRequest(request, registered)
		return
	}
	// Typed registration publishes its queue before the immutable handler.
	// The sole reader's cache preserves that handoff on later exact hits.
	c.enqueueTypedDeliveredRequest(request, registered)
}

// Only the asynchronous typed path needs a heap-owned record. Keep this
// boxing behind the branch so raw synchronous turns retain their Go stack value.
func (c *GoRuntimeClient) enqueueTypedDeliveredRequest(request nativeDeliveredRequest, registered registeredHandler) {
	c.handlerSchedule.queue.enqueue(deliveredRequestJob{request: &request, registered: registered}, c.lifetime.closed)
}

func (c *GoRuntimeClient) deliveredHandler(target []byte) (registeredHandler, bool) {
	if c.lastDispatchHandlerSet && bytes.Equal(c.lastDispatchTarget, target) {
		return c.lastDispatchHandler, true
	}
	c.handlersMu.RLock()
	registered, ok := c.handlers[string(target)]
	c.handlersMu.RUnlock()
	if !ok {
		return registeredHandler{}, false
	}
	// The sole reader owns this exact-byte cache. Registrations cannot be
	// removed or replaced, so a byte-equal Core-issued target keeps its owner.
	c.lastDispatchTarget = append(c.lastDispatchTarget[:0], target...)
	c.lastDispatchHandler = registered
	c.lastDispatchHandlerSet = true
	return registered, true
}

func (c *GoRuntimeClient) handlePayloadDeliveredRequest(request nativeDeliveredRequest, registered registeredHandler) {
	defer func() {
		if recover() != nil {
			log.Printf("[coakka-go] application handler panicked; no reply submitted")
		}
	}()
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("[coakka-go] payload handler panic target=%s err=%v", request.Target, recovered)
		}
	}()
	replyPayload := registered.payloadHandler(request.Payload)
	if request.OneWay || replyPayload == nil {
		return
	}
	err := c.bindings.submitRawReplyCString(c.runtime, request.Frame, request.TargetCString, replyPayload)
	if err != nil && !c.isClosed() {
		log.Printf("[coakka-go] payload reply submit failed target=%s err=%v", request.Target, err)
	}
	return
}

func (c *GoRuntimeClient) handleDeliveredRequest(job deliveredRequestJob) {
	defer func() {
		if recover() != nil {
			log.Printf("[coakka-go] application handler panicked; no reply submitted")
		}
	}()
	request := job.request
	registered := job.registered

	envelope, err := c.bindings.decodeEnvelope(request.Frame)
	if err != nil {
		if !c.isClosed() {
			log.Printf("[coakka-go] request value projection failed: %v", err)
		}
		return
	}
	// Routing and reply semantics use Core-issued dispatch facts. The decoder
	// fills only the remaining language value fields.
	envelope.Kind = MessageKindRequest
	envelope.Target = string(request.Target)
	envelope.Payload = request.Payload
	envelope.OneWay = request.OneWay
	c.handleRegisteredRequest(envelope, request.Frame, registered)
}

func (c *GoRuntimeClient) handleRegisteredRequest(request *Envelope, requestFrame []byte, registered registeredHandler) {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("[coakka-go] request handler panic target=%s err=%v", request.GetTarget(), recovered)
		}
	}()
	reply := registered.handler(request)
	if request.GetOneWay() || reply == nil {
		return
	}
	var err error
	if registered.typedReplies || len(requestFrame) == 0 {
		err = c.SubmitTypedEnvelope(reply)
	} else {
		err = c.bindings.submitRawReply(c.runtime, requestFrame, reply.GetSource(), reply.GetPayload())
	}
	if err != nil && !c.isClosed() {
		log.Printf("[coakka-go] request handler reply submit failed target=%s err=%v", request.GetTarget(), err)
	}
}

func (c *GoRuntimeClient) observeCoreResponse(response *Envelope) {
	// This frame belongs to the completed Core ticket; never match it again.
	c.counters.matchedResponses.Add(1)
	c.publishTerminalEvent(RequestTerminalEvent{Kind: RequestTerminalEventResponse,
		RequestMessageID: response.GetCorrelationId(), CorrelationID: response.GetCorrelationId(), Response: response})
}

func (c *GoRuntimeClient) observeCoreDeadletter(deadletter *Deadletter) {
	if deadletter == nil {
		return
	}
	c.counters.matchedDeadletters.Add(1)
	c.publishDeadletterObservation(deadletter)
	original := deadletter.GetOriginalEnvelope()
	c.publishTerminalEvent(RequestTerminalEvent{Kind: RequestTerminalEventDeadletter,
		RequestMessageID: original.GetMessageId(), CorrelationID: original.GetCorrelationId(), Deadletter: deadletter})
}

func (c *GoRuntimeClient) publishDeadletterObservation(deadletter *Deadletter) {
	original := deadletter.GetOriginalEnvelope()
	observed := ObservedDeadletter{Deadletter: deadletter, MatchedPendingRequest: true,
		RequestMessageID: original.GetMessageId(), CorrelationID: original.GetCorrelationId()}
	c.deadletterSubscribersMu.Lock()
	defer c.deadletterSubscribersMu.Unlock()
	for _, events := range c.deadletterSubscribers {
		select {
		case events <- observed:
		default:
			c.counters.deadletterObservationDrops.Add(1)
		}
	}
}

func (c *GoRuntimeClient) isClosed() bool {
	select {
	case <-c.lifetime.closed:
		return true
	default:
		return false
	}
}

func (c *GoRuntimeClient) addTerminalSubscriber(events chan RequestTerminalEvent) (uint64, error) {
	c.terminalSubscribersMu.Lock()
	defer c.terminalSubscribersMu.Unlock()
	if len(c.terminalSubscribers) >= observerCapacity {
		return 0, fmt.Errorf("Go terminal observer capacity exhausted")
	}
	c.nextTerminalSubscriberID++
	subscriberID := c.nextTerminalSubscriberID
	c.terminalSubscribers[subscriberID] = events
	return subscriberID, nil
}

func (c *GoRuntimeClient) removeTerminalSubscriber(subscriberID uint64, events chan RequestTerminalEvent) {
	c.terminalSubscribersMu.Lock()
	defer c.terminalSubscribersMu.Unlock()
	current, ok := c.terminalSubscribers[subscriberID]
	if !ok || current != events {
		return
	}
	delete(c.terminalSubscribers, subscriberID)
	close(events)
}

func (c *GoRuntimeClient) addDeadletterSubscriber(events chan ObservedDeadletter) (uint64, error) {
	c.deadletterSubscribersMu.Lock()
	defer c.deadletterSubscribersMu.Unlock()
	if len(c.deadletterSubscribers) >= observerCapacity {
		return 0, fmt.Errorf("Go deadletter observer capacity exhausted")
	}
	c.nextDeadletterSubscriberID++
	subscriberID := c.nextDeadletterSubscriberID
	c.deadletterSubscribers[subscriberID] = events
	return subscriberID, nil
}

func (c *GoRuntimeClient) removeDeadletterSubscriber(subscriberID uint64, events chan ObservedDeadletter) {
	c.deadletterSubscribersMu.Lock()
	defer c.deadletterSubscribersMu.Unlock()
	current, ok := c.deadletterSubscribers[subscriberID]
	if !ok || current != events {
		return
	}
	delete(c.deadletterSubscribers, subscriberID)
	close(events)
}

func (c *GoRuntimeClient) publishTerminalEvent(event RequestTerminalEvent) {
	c.terminalSubscribersMu.Lock()
	defer c.terminalSubscribersMu.Unlock()
	for _, events := range c.terminalSubscribers {
		select {
		case events <- event:
		default:
			c.counters.terminalEventDrops.Add(1)
		}
	}
}
