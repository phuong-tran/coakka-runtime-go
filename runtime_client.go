package coakka_v2_connector

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	coakkav2 "github.com/phuong-tran/coakka-runtime-go/coakka/v2"
	"google.golang.org/protobuf/proto"
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
	LateResponses                  int `json:"lateResponses"`
	UnhandledDeadletters           int `json:"unhandledDeadletters"`
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
	RuntimeInfo   RuntimeInfoSnapshot      `json:"runtimeInfo"`
	RuntimeConfig RuntimeConfigSnapshot    `json:"runtimeConfig"`
	Connector     RuntimeConnectorSnapshot `json:"connector"`
	Health        RuntimeHealthSnapshot    `json:"health"`
	Stats         RuntimeStatsSnapshot     `json:"stats"`
}

type MonitorSnapshot struct {
	SignalCount uint64                `json:"signalCount"`
	Health      RuntimeHealthSnapshot `json:"health"`
	Stats       RuntimeStatsSnapshot  `json:"stats"`
}

type HandlerFn func(request *Envelope) *Envelope

type registeredHandler struct {
	handler      HandlerFn
	typedReplies bool
}

type pendingResult struct {
	response *Envelope
	err      error
}

type pendingRequest struct {
	ch chan pendingResult
}

type trackedRequest struct {
	messageID     string
	correlationID string
}

type GoRuntimeClient struct {
	config                       ConnectorConfig
	bindings                     *nativeBindings
	runtime                      nativeRuntime
	hostHandles                  HostHandlesSnapshot
	separateDeliveredRequestLane bool

	pendingMu sync.Mutex
	pending   map[string]*pendingRequest

	trackedMu                        sync.Mutex
	trackedRequestsByMessageID       map[string]trackedRequest
	trackedRequestMessageIDsByCorrID map[string]string

	handlersMu sync.RWMutex
	handlers   map[string]registeredHandler

	terminalSubscribersMu    sync.Mutex
	terminalSubscribers      map[uint64]chan RequestTerminalEvent
	nextTerminalSubscriberID uint64

	deadletterSubscribersMu    sync.Mutex
	deadletterSubscribers      map[uint64]chan ObservedDeadletter
	nextDeadletterSubscriberID uint64

	submitMu                sync.Mutex
	statsMu                 sync.Mutex
	stats                   RuntimeClientStats
	transportMu             sync.Mutex
	startupConnectionResult *RuntimeTCPConnectionApplyResult
	startupSecurityResult   *RuntimeTCPSecurityApplyResult

	monitorSignalCh chan uint64
	closedCh        chan struct{}
	closeOnce       sync.Once
	wg              sync.WaitGroup
}

func NewGoRuntimeClient(runtimeLibPath string, config ConnectorConfig) (*GoRuntimeClient, error) {
	if err := config.RequireValid(); err != nil {
		return nil, err
	}
	bindings, err := openNativeBindings(runtimeLibPath)
	if err != nil {
		return nil, err
	}
	if bindings.getAbiVersion() != COAKKAABIVersion {
		bindings.close()
		return nil, fmt.Errorf("unexpected ABI version: %d", bindings.getAbiVersion())
	}
	runtime, err := bindings.createRuntime(config)
	if err != nil {
		bindings.close()
		return nil, err
	}
	client := &GoRuntimeClient{
		config:                           config,
		bindings:                         bindings,
		runtime:                          runtime,
		pending:                          make(map[string]*pendingRequest),
		trackedRequestsByMessageID:       make(map[string]trackedRequest),
		trackedRequestMessageIDsByCorrID: make(map[string]string),
		handlers:                         make(map[string]registeredHandler),
		terminalSubscribers:              make(map[uint64]chan RequestTerminalEvent),
		deadletterSubscribers:            make(map[uint64]chan ObservedDeadletter),
		monitorSignalCh:                  make(chan uint64, 64),
		closedCh:                         make(chan struct{}),
	}
	if config.ConnectionStrategy != nil {
		result, applyErr := bindings.applyTCPConnectionStrategy(runtime, *config.ConnectionStrategy)
		if applyErr != nil {
			bindings.destroyRuntime(runtime)
			bindings.close()
			return nil, applyErr
		}
		client.startupConnectionResult = &result
		if !result.Applied() {
			bindings.destroyRuntime(runtime)
			bindings.close()
			return nil, &RuntimeTCPConnectionApplyError{Result: result}
		}
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
	hostHandles, err := bindings.getHostHandles(runtime, client.hostHandleFlags())
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
	deliveredRequestFD := hostHandles.DeliveredRequestReadFD
	responseFD := hostHandles.ResponseReadFD
	client.separateDeliveredRequestLane = deliveredRequestFD >= 0 && deliveredRequestFD != responseFD
	requestReadFD := responseFD
	if deliveredRequestFD >= 0 {
		requestReadFD = deliveredRequestFD
	}
	client.startEnvelopeReader(requestReadFD, "request", client.onRequestFrame)
	if client.separateDeliveredRequestLane {
		client.startEnvelopeReader(responseFD, "response", client.onResponseFrame)
	}
	client.startEnvelopeReader(hostHandles.DeadletterReadFD, "deadletter", client.onDeadletterFrame)
	if hostHandles.MonitorReadFD >= 0 {
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
		Kind:                 coakkav2.MessageKind_MESSAGE_KIND_REQUEST,
		OneWay:               oneWay,
		TimeoutMs:            int32(timeout / time.Millisecond),
		Payload:              payload,
		Headers:              outHeaders,
		Status:               coakkav2.BusinessStatus_BUSINESS_STATUS_OK,
		ErrorCode:            "",
		ErrorMessage:         "",
		DeliveryHint:         coakkav2.DeliveryHint(deliveryHint),
		MessageType:          payloadIdentity.MessageType,
		PayloadSchemaVersion: payloadIdentity.PayloadSchemaVersion,
		PayloadFormat:        coakkav2.PayloadFormat(payloadIdentity.PayloadFormat),
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
		Kind:                 coakkav2.MessageKind_MESSAGE_KIND_RESPONSE,
		OneWay:               false,
		TimeoutMs:            request.GetTimeoutMs(),
		Payload:              payload,
		Headers:              map[string]string{},
		Status:               coakkav2.BusinessStatus_BUSINESS_STATUS_OK,
		ErrorCode:            "",
		ErrorMessage:         "",
		DeliveryHint:         coakkav2.DeliveryHint_DELIVERY_HINT_ROUTER_DEFAULT,
		MessageType:          payloadIdentity.MessageType,
		PayloadSchemaVersion: payloadIdentity.PayloadSchemaVersion,
		PayloadFormat:        coakkav2.PayloadFormat(payloadIdentity.PayloadFormat),
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

func MakeRawRequest(messageID string, source string, target string, replyTo string, payload []byte, timeout time.Duration, deliveryHint DeliveryHint) *Envelope {
	if messageID == "" {
		messageID = NextMessageID(source)
	}
	return &Envelope{
		MessageId:     messageID,
		CorrelationId: messageID,
		Source:        source,
		Target:        target,
		ReplyTo:       replyTo,
		Kind:          coakkav2.MessageKind_MESSAGE_KIND_REQUEST,
		OneWay:        false,
		TimeoutMs:     int32(timeout / time.Millisecond),
		Payload:       payload,
		Headers:       map[string]string{},
		Status:        coakkav2.BusinessStatus_BUSINESS_STATUS_OK,
		DeliveryHint:  coakkav2.DeliveryHint(deliveryHint),
	}
}

func MakeRawOneWay(messageID string, source string, target string, payload []byte, timeout time.Duration, deliveryHint DeliveryHint) *Envelope {
	out := MakeRawRequest(messageID, source, target, "", payload, timeout, deliveryHint)
	out.CorrelationId = ""
	out.ReplyTo = ""
	out.OneWay = true
	return out
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
	return c.submitPendingRequest(envelope, timeout, func(normalized *Envelope) error {
		return c.SubmitTypedEnvelope(normalized)
	})
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

func (c *GoRuntimeClient) AskRaw(request *Envelope, timeout time.Duration) (*Envelope, error) {
	if request.GetKind() != coakkav2.MessageKind_MESSAGE_KIND_REQUEST {
		return nil, fmt.Errorf("AskRaw requires MESSAGE_KIND_REQUEST")
	}
	if request.GetOneWay() {
		return nil, fmt.Errorf("AskRaw requires oneWay=false")
	}
	if request.GetMessageId() == "" {
		return nil, fmt.Errorf("AskRaw requires messageId")
	}
	normalized := protoCloneEnvelope(request)
	if normalized.CorrelationId == "" {
		normalized.CorrelationId = normalized.MessageId
	}
	if timeout <= 0 {
		timeout = time.Duration(normalized.GetTimeoutMs()) * time.Millisecond
	}
	if timeout <= 0 {
		timeout = time.Second
	}
	return c.submitPendingRequest(normalized, timeout, func(envelope *Envelope) error {
		return c.SubmitRawEnvelope(envelope)
	})
}

func (c *GoRuntimeClient) SubmitRequestTyped(source string, target string, payload []byte, payloadIdentity PayloadIdentity, timeout time.Duration, operation string, deliveryHint DeliveryHint, headers map[string]string) (*SubmittedRequest, error) {
	envelope, err := MakeTypedRequest(source, target, payload, payloadIdentity, timeout, operation, deliveryHint, headers, NextMessageID(source), false)
	if err != nil {
		return nil, err
	}
	return c.submitTrackedRequest(envelope, func(normalized *Envelope) error {
		return c.SubmitTypedEnvelope(normalized)
	})
}

func (c *GoRuntimeClient) SubmitRequestJSON(source string, target string, payload any, payloadIdentity PayloadIdentity, timeout time.Duration, operation string, deliveryHint DeliveryHint, headers map[string]string) (*SubmittedRequest, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return c.SubmitRequestTyped(source, target, bytes, payloadIdentity, timeout, operation, deliveryHint, headers)
}

func (c *GoRuntimeClient) SubmitRequestRaw(request *Envelope) (*SubmittedRequest, error) {
	return c.submitTrackedRequest(request, func(normalized *Envelope) error {
		return c.SubmitRawEnvelope(normalized)
	})
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
	bytes, err := encodeEnvelope(envelope)
	if err != nil {
		return err
	}
	c.submitMu.Lock()
	defer c.submitMu.Unlock()
	return c.bindings.submitEnvelope(c.runtime, bytes)
}

func (c *GoRuntimeClient) ApplySnapshot(generation uint64, routes []RouteSpec, sourceConnector string, seq uint64) error {
	return c.ApplySnapshotWithPolicy(generation, routes, sourceConnector, seq, c.config.OverloadPolicy)
}

func (c *GoRuntimeClient) ApplySnapshotWithPolicy(
	generation uint64,
	routes []RouteSpec,
	sourceConnector string,
	seq uint64,
	overloadPolicy *RuntimeOverloadPolicy,
) error {
	if overloadPolicy != nil {
		if err := overloadPolicy.RequireValid(); err != nil {
			return err
		}
	}
	bytes, err := c.buildControlEnvelopeBytes(generation, routes, sourceConnector, seq, overloadPolicy)
	if err != nil {
		return err
	}
	c.submitMu.Lock()
	defer c.submitMu.Unlock()
	return c.bindings.applyControlEnvelope(c.runtime, bytes)
}

func (c *GoRuntimeClient) RegisterHandler(target string, handler HandlerFn, typedReplies bool) error {
	c.handlersMu.Lock()
	defer c.handlersMu.Unlock()
	if _, exists := c.handlers[target]; exists {
		return fmt.Errorf("handler already registered for target=%s", target)
	}
	c.handlers[target] = registeredHandler{handler: handler, typedReplies: typedReplies}
	return nil
}

func (c *GoRuntimeClient) RegisterRawHandler(target string, handler HandlerFn) error {
	return c.RegisterHandler(target, handler, false)
}

func (c *GoRuntimeClient) TerminalEvents(ctx context.Context, buffer int) <-chan RequestTerminalEvent {
	if ctx == nil {
		ctx = context.Background()
	}
	if buffer <= 0 {
		buffer = 16
	}
	events := make(chan RequestTerminalEvent, buffer)
	subscriberID := c.addTerminalSubscriber(events)
	go func() {
		select {
		case <-ctx.Done():
		case <-c.closedCh:
		}
		c.removeTerminalSubscriber(subscriberID, events)
	}()
	return events
}

func (c *GoRuntimeClient) Deadletters(ctx context.Context, buffer int) <-chan ObservedDeadletter {
	if ctx == nil {
		ctx = context.Background()
	}
	if buffer <= 0 {
		buffer = 128
	}
	events := make(chan ObservedDeadletter, buffer)
	subscriberID := c.addDeadletterSubscriber(events)
	go func() {
		select {
		case <-ctx.Done():
		case <-c.closedCh:
		}
		c.removeDeadletterSubscriber(subscriberID, events)
	}()
	return events
}

func (c *GoRuntimeClient) SnapshotStats() RuntimeClientStats {
	c.statsMu.Lock()
	defer c.statsMu.Unlock()
	out := c.stats
	c.pendingMu.Lock()
	out.PendingRequests = len(c.pending)
	c.pendingMu.Unlock()
	return out
}

func (c *GoRuntimeClient) HealthSnapshot() RuntimeHealthSnapshot {
	health, err := c.bindings.getHealth(c.runtime)
	if err != nil {
		return RuntimeHealthSnapshot{}
	}
	return health
}

func (c *GoRuntimeClient) StatsSnapshot() RuntimeStatsSnapshot {
	stats, err := c.bindings.getStats(c.runtime)
	if err != nil {
		return RuntimeStatsSnapshot{}
	}
	return stats
}

func (c *GoRuntimeClient) RuntimeInfoSnapshot() RuntimeInfoSnapshot {
	info, err := c.bindings.readRuntimeInfo()
	if err != nil {
		return RuntimeInfoSnapshot{}
	}
	return info
}

func (c *GoRuntimeClient) RuntimeConfigSnapshot() RuntimeConfigSnapshot {
	config, err := c.bindings.getConfig(c.runtime)
	if err != nil {
		return RuntimeConfigSnapshot{}
	}
	return config
}

func (c *GoRuntimeClient) MonitorIsEnabled() bool {
	return c.hostHandles.MonitorReadFD >= 0
}

func (c *GoRuntimeClient) MonitorSnapshot(signalCount uint64) MonitorSnapshot {
	return MonitorSnapshot{
		SignalCount: signalCount,
		Health:      c.HealthSnapshot(),
		Stats:       c.StatsSnapshot(),
	}
}

func (c *GoRuntimeClient) AwaitNextMonitor(timeout time.Duration) (*MonitorSnapshot, error) {
	if !c.MonitorIsEnabled() {
		return nil, fmt.Errorf("runtime monitor is not enabled")
	}
	if timeout <= 0 {
		timeout = time.Second
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-c.closedCh:
		return nil, fmt.Errorf("runtime client closed")
	case signalCount := <-c.monitorSignalCh:
		snapshot := c.MonitorSnapshot(signalCount)
		return &snapshot, nil
	case <-timer.C:
		return nil, nil
	}
}

func (c *GoRuntimeClient) AwaitAppliedGenerationAtLeast(generation uint64, timeout time.Duration) (*MonitorSnapshot, error) {
	if current := c.MonitorSnapshot(0); current.Health.AppliedGeneration >= generation {
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

func (c *GoRuntimeClient) RuntimeSnapshot() RuntimeSnapshot {
	return RuntimeSnapshot{
		RuntimeInfo:   c.RuntimeInfoSnapshot(),
		RuntimeConfig: c.RuntimeConfigSnapshot(),
		Connector: RuntimeConnectorSnapshot{
			RuntimeClientStats:           c.SnapshotStats(),
			SeparateDeliveredRequestLane: c.separateDeliveredRequestLane,
			SystemName:                   c.config.SystemName,
			NodeID:                       c.config.NodeID,
		},
		Health: c.HealthSnapshot(),
		Stats:  c.StatsSnapshot(),
	}
}

func (c *GoRuntimeClient) Close() error {
	var closeErr error
	c.closeOnce.Do(func() {
		close(c.closedCh)
		c.pendingMu.Lock()
		for key, pending := range c.pending {
			pending.ch <- pendingResult{err: fmt.Errorf("runtime client closed")}
			close(pending.ch)
			delete(c.pending, key)
		}
		c.pendingMu.Unlock()
		c.trackedMu.Lock()
		clear(c.trackedRequestsByMessageID)
		clear(c.trackedRequestMessageIDsByCorrID)
		c.trackedMu.Unlock()
		c.terminalSubscribersMu.Lock()
		for subscriberID, events := range c.terminalSubscribers {
			delete(c.terminalSubscribers, subscriberID)
			close(events)
		}
		c.terminalSubscribersMu.Unlock()
		c.deadletterSubscribersMu.Lock()
		for subscriberID, events := range c.deadletterSubscribers {
			delete(c.deadletterSubscribers, subscriberID)
			close(events)
		}
		c.deadletterSubscribersMu.Unlock()
		c.wg.Wait()
		c.transportMu.Lock()
		closeErr = c.bindings.stopRuntime(c.runtime)
		c.cleanupNative()
		c.transportMu.Unlock()
	})
	return closeErr
}

func (c *GoRuntimeClient) cleanupNative() {
	for _, fd := range []int{
		c.hostHandles.ResponseReadFD,
		c.hostHandles.DeadletterReadFD,
		c.hostHandles.RequestWriteFD,
		c.hostHandles.ControlWriteFD,
		c.hostHandles.MonitorReadFD,
		c.hostHandles.DeliveredRequestReadFD,
	} {
		safeCloseFD(fd)
	}
	c.bindings.destroyRuntime(c.runtime)
	c.bindings.close()
}

func (c *GoRuntimeClient) hostHandleFlags() uint32 {
	var flags uint32
	if c.config.EnableMonitor {
		flags |= HostHandlesEnableMonitor
	}
	if c.config.SeparateDeliveredRequestLane {
		flags |= HostHandlesSeparateDeliveredLane
	}
	return flags
}

func (c *GoRuntimeClient) buildControlEnvelopeBytes(
	generation uint64,
	routes []RouteSpec,
	sourceConnector string,
	seq uint64,
	overloadPolicy *RuntimeOverloadPolicy,
) ([]byte, error) {
	if generation == 0 {
		generation = c.config.Generation
	}
	if seq == 0 {
		seq = 1
	}
	if routes == nil {
		routes = c.config.Routes
	}
	if sourceConnector == "" {
		sourceConnector = c.config.SystemName
	}
	snapshot := &RouteSnapshotPayload{
		Generation: generation,
		Routes:     make([]*coakkav2.Route, 0, len(routes)),
	}
	if overloadPolicy != nil {
		snapshot.OverloadPolicy = &coakkav2.OverloadPolicy{
			IngressMode:                     coakkav2.OverloadMode(overloadPolicy.IngressMode),
			LocalDeliveryMode:               coakkav2.OverloadMode(overloadPolicy.LocalDeliveryMode),
			RemoteOutboundMode:              coakkav2.OverloadMode(overloadPolicy.RemoteOutboundMode),
			RemoteOutboundReplyReserveSlots: overloadPolicy.RemoteOutboundReplyReserveSlots,
		}
	}
	for _, route := range routes {
		endpoints := make([]*coakkav2.Endpoint, 0, len(route.Endpoints))
		for _, endpoint := range route.Endpoints {
			weight := endpoint.Weight
			if weight == 0 {
				weight = 1
			}
			endpoints = append(endpoints, &coakkav2.Endpoint{
				Host:   endpoint.Host,
				Port:   endpoint.Port,
				Weight: weight,
				Flags:  endpoint.Flags,
			})
		}
		strategy := route.Strategy
		if strategy == RouteStrategyUnspecified {
			strategy = RouteStrategySingleOwner
		}
		snapshot.Routes = append(snapshot.Routes, &coakkav2.Route{
			Target:       route.Target,
			Strategy:     coakkav2.RouteResolutionStrategy(strategy),
			RouteKeyHint: route.RouteKeyHint,
			Flags:        route.Flags,
			Endpoints:    endpoints,
		})
	}
	payloadBytes, err := encodeRouteSnapshotPayload(snapshot)
	if err != nil {
		return nil, err
	}
	envelopeBytes, err := encodeControlEnvelope(&ControlEnvelope{
		Seq:           seq,
		Generation:    generation,
		Kind:          coakkav2.ControlKind_CONTROL_KIND_APPLY_SNAPSHOT,
		PayloadFormat: coakkav2.ConfigFormat_CONFIG_FORMAT_PROTOBUF,
		PayloadType:   coakkav2.ControlPayloadType_CONTROL_PAYLOAD_TYPE_ROUTE_SNAPSHOT,
		SchemaVersion: 1,
		Payload:       payloadBytes,
		Metadata: map[string]string{
			"source_connector": sourceConnector,
		},
	})
	if err != nil {
		return nil, err
	}
	return envelopeBytes, nil
}

func (c *GoRuntimeClient) submitPendingRequest(request *Envelope, timeout time.Duration, submit func(*Envelope) error) (*Envelope, error) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	normalized, tracked, err := c.trackRequest(request, "request")
	if err != nil {
		return nil, err
	}
	waiter := &pendingRequest{ch: make(chan pendingResult, 1)}
	c.pendingMu.Lock()
	if _, exists := c.pending[tracked.messageID]; exists {
		c.pendingMu.Unlock()
		c.forgetTrackedRequest(tracked.messageID)
		return nil, fmt.Errorf("duplicate pending messageId=%s", tracked.messageID)
	}
	c.pending[tracked.messageID] = waiter
	c.pendingMu.Unlock()

	if err := submit(normalized); err != nil {
		c.pendingMu.Lock()
		delete(c.pending, tracked.messageID)
		c.pendingMu.Unlock()
		c.forgetTrackedRequest(tracked.messageID)
		return nil, err
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case result := <-waiter.ch:
		return result.response, result.err
	case <-timer.C:
		c.pendingMu.Lock()
		delete(c.pending, tracked.messageID)
		c.pendingMu.Unlock()
		c.forgetTrackedRequest(tracked.messageID)
		select {
		case result := <-waiter.ch:
			return result.response, result.err
		default:
		}
		return nil, fmt.Errorf("timed out waiting for %s", tracked.messageID)
	case <-c.closedCh:
		return nil, fmt.Errorf("runtime client closed")
	}
}

func (c *GoRuntimeClient) submitTrackedRequest(request *Envelope, submit func(*Envelope) error) (*SubmittedRequest, error) {
	normalized, tracked, err := c.trackRequest(request, "submitRequest")
	if err != nil {
		return nil, err
	}
	if err := submit(normalized); err != nil {
		c.forgetTrackedRequest(tracked.messageID)
		return nil, err
	}
	return &SubmittedRequest{
		MessageID:     tracked.messageID,
		CorrelationID: tracked.correlationID,
	}, nil
}

func (c *GoRuntimeClient) startEnvelopeReader(fd int, lane string, onFrame func([]byte) error) {
	if fd < 0 {
		return
	}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		buffer := make([]byte, 0, 64*1024)
		chunk := make([]byte, 64*1024)
		for {
			select {
			case <-c.closedCh:
				return
			default:
			}
			ready, err := waitReadable(fd, 100*time.Millisecond)
			if err != nil {
				if !c.isClosed() {
					log.Printf("[coakka-go] %s reader poll failed: %v", lane, err)
				}
				continue
			}
			if !ready {
				continue
			}
			for {
				n, readErr := nativeReadFD(fd, chunk)
				if n > 0 {
					buffer = append(buffer, chunk[:n]...)
					for {
						if len(buffer) < 8 {
							break
						}
						frameLength := binary.LittleEndian.Uint64(buffer[:8])
						if frameLength > uint64(len(buffer)-8) {
							break
						}
						frameSize := int(frameLength)
						frame := append([]byte(nil), buffer[8:8+frameSize]...)
						buffer = buffer[8+frameSize:]
						if err := onFrame(frame); err != nil && !c.isClosed() {
							log.Printf("[coakka-go] %s frame handler failed: %v", lane, err)
						}
					}
				}
				if readErr == nil {
					if n < len(chunk) {
						break
					}
					continue
				}
				if isRetryableReadError(readErr) {
					break
				}
				if isTerminalReadError(readErr) {
					return
				}
				if !c.isClosed() {
					log.Printf("[coakka-go] %s reader failed: %v", lane, readErr)
				}
				break
			}
		}
	}()
}

func (c *GoRuntimeClient) startMonitorReader(fd int) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-c.closedCh:
				return
			default:
			}
			ready, err := waitReadable(fd, 100*time.Millisecond)
			if err != nil {
				if !c.isClosed() {
					log.Printf("[coakka-go] monitor reader poll failed: %v", err)
				}
				continue
			}
			if !ready {
				continue
			}
			signalCount, err := c.bindings.monitorConsume(fd)
			if err != nil {
				if !c.isClosed() {
					log.Printf("[coakka-go] monitor reader failed: %v", err)
				}
				continue
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

func (c *GoRuntimeClient) onRequestFrame(frame []byte) error {
	envelope, err := decodeEnvelope(frame)
	if err != nil {
		return err
	}
	if c.separateDeliveredRequestLane {
		if envelope.GetKind() != coakkav2.MessageKind_MESSAGE_KIND_REQUEST {
			return fmt.Errorf("unexpected envelope kind=%d on delivered-request lane", envelope.GetKind())
		}
		c.dispatchRequest(envelope)
		return nil
	}
	switch envelope.GetKind() {
	case coakkav2.MessageKind_MESSAGE_KIND_REQUEST:
		c.dispatchRequest(envelope)
	case coakkav2.MessageKind_MESSAGE_KIND_RESPONSE:
		c.dispatchResponse(envelope)
	default:
		return fmt.Errorf("unexpected envelope kind=%d", envelope.GetKind())
	}
	return nil
}

func (c *GoRuntimeClient) onResponseFrame(frame []byte) error {
	envelope, err := decodeEnvelope(frame)
	if err != nil {
		return err
	}
	switch envelope.GetKind() {
	case coakkav2.MessageKind_MESSAGE_KIND_RESPONSE:
		c.dispatchResponse(envelope)
	case coakkav2.MessageKind_MESSAGE_KIND_REQUEST:
		return fmt.Errorf("unexpected request on response-only lane")
	default:
		return fmt.Errorf("unexpected envelope kind=%d", envelope.GetKind())
	}
	return nil
}

func (c *GoRuntimeClient) onDeadletterFrame(frame []byte) error {
	deadletter, err := decodeDeadletter(frame)
	if err != nil {
		return err
	}
	c.dispatchDeadletter(deadletter)
	return nil
}

func (c *GoRuntimeClient) dispatchRequest(request *Envelope) {
	c.statsMu.Lock()
	c.stats.DeliveredRequests++
	c.statsMu.Unlock()

	c.handlersMu.RLock()
	registered, ok := c.handlers[request.GetTarget()]
	c.handlersMu.RUnlock()
	if !ok {
		return
	}
	go func() {
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
		if registered.typedReplies {
			err = c.SubmitTypedEnvelope(reply)
		} else {
			err = c.SubmitRawEnvelope(reply)
		}
		if err != nil {
			log.Printf("[coakka-go] request handler reply submit failed target=%s err=%v", request.GetTarget(), err)
		}
	}()
}

func (c *GoRuntimeClient) dispatchResponse(response *Envelope) {
	tracked, ok := c.forgetTrackedRequestByCorrelationID(response.GetCorrelationId())
	if !ok {
		c.statsMu.Lock()
		c.stats.LateResponses++
		c.statsMu.Unlock()
		return
	}
	c.publishTerminalEvent(RequestTerminalEvent{
		Kind:             RequestTerminalEventResponse,
		RequestMessageID: tracked.messageID,
		CorrelationID:    tracked.correlationID,
		Response:         response,
	})
	c.pendingMu.Lock()
	pending, ok := c.pending[tracked.messageID]
	if ok {
		delete(c.pending, tracked.messageID)
	}
	c.pendingMu.Unlock()
	c.statsMu.Lock()
	c.stats.MatchedResponses++
	c.statsMu.Unlock()
	if ok {
		pending.ch <- pendingResult{response: response}
		close(pending.ch)
	}
}

func (c *GoRuntimeClient) dispatchDeadletter(deadletter *Deadletter) {
	if deadletter == nil || deadletter.OriginalEnvelope == nil {
		return
	}
	tracked, ok := c.forgetTrackedRequest(deadletter.OriginalEnvelope.GetMessageId())
	c.publishDeadletterObservation(deadletter, tracked, ok)
	if !ok {
		c.statsMu.Lock()
		c.stats.UnhandledDeadletters++
		c.statsMu.Unlock()
		return
	}
	c.publishTerminalEvent(RequestTerminalEvent{
		Kind:             RequestTerminalEventDeadletter,
		RequestMessageID: tracked.messageID,
		CorrelationID:    tracked.correlationID,
		Deadletter:       deadletter,
	})
	c.pendingMu.Lock()
	pending, waiterExists := c.pending[tracked.messageID]
	if waiterExists {
		delete(c.pending, tracked.messageID)
	}
	c.pendingMu.Unlock()
	c.statsMu.Lock()
	c.stats.MatchedDeadletters++
	c.statsMu.Unlock()
	if waiterExists {
		pending.ch <- pendingResult{err: &DeadletterError{Deadletter: deadletter}}
		close(pending.ch)
	}
}

func (c *GoRuntimeClient) publishDeadletterObservation(deadletter *Deadletter, tracked trackedRequest, matched bool) {
	observed := ObservedDeadletter{
		Deadletter:            deadletter,
		MatchedPendingRequest: matched,
	}
	if matched {
		observed.RequestMessageID = tracked.messageID
		observed.CorrelationID = tracked.correlationID
	} else if deadletter != nil && deadletter.OriginalEnvelope != nil {
		observed.RequestMessageID = deadletter.OriginalEnvelope.GetMessageId()
		observed.CorrelationID = deadletter.OriginalEnvelope.GetCorrelationId()
	}

	c.deadletterSubscribersMu.Lock()
	defer c.deadletterSubscribersMu.Unlock()
	for _, events := range c.deadletterSubscribers {
		select {
		case events <- observed:
		default:
			c.statsMu.Lock()
			c.stats.DeadletterObservationDropCount++
			c.statsMu.Unlock()
		}
	}
}

func (c *GoRuntimeClient) isClosed() bool {
	select {
	case <-c.closedCh:
		return true
	default:
		return false
	}
}

func (c *GoRuntimeClient) trackRequest(request *Envelope, caller string) (*Envelope, trackedRequest, error) {
	if request.GetKind() != coakkav2.MessageKind_MESSAGE_KIND_REQUEST {
		return nil, trackedRequest{}, fmt.Errorf("%s requires MESSAGE_KIND_REQUEST", caller)
	}
	if request.GetOneWay() {
		return nil, trackedRequest{}, fmt.Errorf("%s requires oneWay=false", caller)
	}
	if request.GetMessageId() == "" {
		return nil, trackedRequest{}, fmt.Errorf("%s requires messageId", caller)
	}
	normalized := protoCloneEnvelope(request)
	if normalized.CorrelationId == "" {
		normalized.CorrelationId = normalized.MessageId
	}
	tracked := trackedRequest{
		messageID:     normalized.GetMessageId(),
		correlationID: normalized.GetCorrelationId(),
	}

	c.trackedMu.Lock()
	defer c.trackedMu.Unlock()
	if _, exists := c.trackedRequestsByMessageID[tracked.messageID]; exists {
		return nil, trackedRequest{}, fmt.Errorf("duplicate tracked messageId=%s", tracked.messageID)
	}
	if existingMessageID, exists := c.trackedRequestMessageIDsByCorrID[tracked.correlationID]; exists {
		return nil, trackedRequest{}, fmt.Errorf("duplicate tracked correlationId=%s existingMessageId=%s", tracked.correlationID, existingMessageID)
	}
	c.trackedRequestsByMessageID[tracked.messageID] = tracked
	c.trackedRequestMessageIDsByCorrID[tracked.correlationID] = tracked.messageID
	return normalized, tracked, nil
}

func (c *GoRuntimeClient) forgetTrackedRequest(messageID string) (trackedRequest, bool) {
	c.trackedMu.Lock()
	defer c.trackedMu.Unlock()
	tracked, ok := c.trackedRequestsByMessageID[messageID]
	if !ok {
		return trackedRequest{}, false
	}
	delete(c.trackedRequestsByMessageID, messageID)
	delete(c.trackedRequestMessageIDsByCorrID, tracked.correlationID)
	return tracked, true
}

func (c *GoRuntimeClient) forgetTrackedRequestByCorrelationID(correlationID string) (trackedRequest, bool) {
	c.trackedMu.Lock()
	defer c.trackedMu.Unlock()
	messageID, ok := c.trackedRequestMessageIDsByCorrID[correlationID]
	if !ok {
		return trackedRequest{}, false
	}
	tracked, trackedExists := c.trackedRequestsByMessageID[messageID]
	if !trackedExists {
		delete(c.trackedRequestMessageIDsByCorrID, correlationID)
		return trackedRequest{}, false
	}
	delete(c.trackedRequestMessageIDsByCorrID, correlationID)
	delete(c.trackedRequestsByMessageID, messageID)
	return tracked, true
}

func (c *GoRuntimeClient) addTerminalSubscriber(events chan RequestTerminalEvent) uint64 {
	c.terminalSubscribersMu.Lock()
	defer c.terminalSubscribersMu.Unlock()
	c.nextTerminalSubscriberID++
	subscriberID := c.nextTerminalSubscriberID
	c.terminalSubscribers[subscriberID] = events
	return subscriberID
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

func (c *GoRuntimeClient) addDeadletterSubscriber(events chan ObservedDeadletter) uint64 {
	c.deadletterSubscribersMu.Lock()
	defer c.deadletterSubscribersMu.Unlock()
	c.nextDeadletterSubscriberID++
	subscriberID := c.nextDeadletterSubscriberID
	c.deadletterSubscribers[subscriberID] = events
	return subscriberID
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
			c.statsMu.Lock()
			c.stats.TerminalEventDropCount++
			c.statsMu.Unlock()
		}
	}
}

func protoCloneEnvelope(envelope *Envelope) *Envelope {
	if envelope == nil {
		return nil
	}
	return proto.Clone(envelope).(*Envelope)
}
