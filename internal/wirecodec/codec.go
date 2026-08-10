// Package wirecodec owns the private serialization boundary between the Go
// connector and the CoAkka core runtime.
package wirecodec

import (
	"fmt"

	"github.com/phuong-tran/coakka-runtime-go/internal/wirepb"
	"google.golang.org/protobuf/proto"
)

// Envelope is the connector-neutral projection consumed by the private codec.
type Envelope struct {
	MessageID            string
	CorrelationID        string
	Source               string
	Target               string
	ReplyTo              string
	Kind                 int32
	OneWay               bool
	TimeoutMillis        int32
	Payload              []byte
	Headers              map[string]string
	Status               int32
	ErrorCode            string
	ErrorMessage         string
	DeliveryHint         int32
	MessageType          string
	PayloadSchemaVersion uint32
	PayloadFormat        int32
}

// Deadletter is the decoded terminal-delivery projection returned to the connector.
type Deadletter struct {
	OriginalEnvelope *Envelope
	Reason           int32
	Detail           string
	ActiveGeneration uint64
	ResolvedHost     string
	ResolvedPort     uint32
}

// Endpoint contains only the stable fields required to encode a route endpoint.
type Endpoint struct {
	Host   string
	Port   uint32
	Weight uint32
	Flags  uint32
}

// Route contains one connector-normalized route snapshot entry.
type Route struct {
	Target       string
	Endpoints    []Endpoint
	Strategy     int32
	RouteKeyHint string
	Flags        uint32
}

// OverloadPolicy contains the numeric policy values defined by the core contract.
type OverloadPolicy struct {
	IngressMode                     int32
	LocalDeliveryMode               int32
	RemoteOutboundMode              int32
	RemoteOutboundReplyReserveSlots uint32
}

// EncodeEnvelope serializes one validated connector envelope for the core runtime.
func EncodeEnvelope(envelope *Envelope) ([]byte, error) {
	if envelope == nil {
		return nil, fmt.Errorf("envelope is nil")
	}
	return proto.Marshal(envelopeToGenerated(envelope))
}

// DecodeEnvelope parses one complete runtime frame into connector-neutral data.
func DecodeEnvelope(data []byte) (*Envelope, error) {
	generated := &wirepb.Envelope{}
	if err := proto.Unmarshal(data, generated); err != nil {
		return nil, err
	}
	return envelopeFromGenerated(generated), nil
}

// DecodeDeadletter parses one complete terminal-delivery frame.
func DecodeDeadletter(data []byte) (*Deadletter, error) {
	generated := &wirepb.Deadletter{}
	if err := proto.Unmarshal(data, generated); err != nil {
		return nil, err
	}
	return &Deadletter{
		OriginalEnvelope: envelopeFromGenerated(generated.OriginalEnvelope),
		Reason:           int32(generated.Reason),
		Detail:           generated.Detail,
		ActiveGeneration: generated.ActiveGeneration,
		ResolvedHost:     generated.ResolvedHost,
		ResolvedPort:     generated.ResolvedPort,
	}, nil
}

// EncodeRouteSnapshotControl serializes one normalized route generation update.
func EncodeRouteSnapshotControl(
	generation uint64,
	routes []Route,
	sourceConnector string,
	seq uint64,
	overloadPolicy *OverloadPolicy,
) ([]byte, error) {
	snapshot := &wirepb.RouteSnapshotPayload{
		Generation: generation,
		Routes:     make([]*wirepb.Route, 0, len(routes)),
	}
	if overloadPolicy != nil {
		snapshot.OverloadPolicy = &wirepb.OverloadPolicy{
			IngressMode:                     wirepb.OverloadMode(overloadPolicy.IngressMode),
			LocalDeliveryMode:               wirepb.OverloadMode(overloadPolicy.LocalDeliveryMode),
			RemoteOutboundMode:              wirepb.OverloadMode(overloadPolicy.RemoteOutboundMode),
			RemoteOutboundReplyReserveSlots: overloadPolicy.RemoteOutboundReplyReserveSlots,
		}
	}
	for _, route := range routes {
		endpoints := make([]*wirepb.Endpoint, 0, len(route.Endpoints))
		for _, endpoint := range route.Endpoints {
			endpoints = append(endpoints, &wirepb.Endpoint{
				Host:   endpoint.Host,
				Port:   endpoint.Port,
				Weight: endpoint.Weight,
				Flags:  endpoint.Flags,
			})
		}
		snapshot.Routes = append(snapshot.Routes, &wirepb.Route{
			Target:       route.Target,
			Strategy:     wirepb.RouteResolutionStrategy(route.Strategy),
			RouteKeyHint: route.RouteKeyHint,
			Flags:        route.Flags,
			Endpoints:    endpoints,
		})
	}

	payloadBytes, err := proto.Marshal(snapshot)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(&wirepb.ControlEnvelope{
		Seq:           seq,
		Generation:    generation,
		Kind:          wirepb.ControlKind_CONTROL_KIND_APPLY_SNAPSHOT,
		PayloadFormat: wirepb.ConfigFormat_CONFIG_FORMAT_PROTOBUF,
		PayloadType:   wirepb.ControlPayloadType_CONTROL_PAYLOAD_TYPE_ROUTE_SNAPSHOT,
		SchemaVersion: 1,
		Payload:       payloadBytes,
		Metadata: map[string]string{
			"source_connector": sourceConnector,
		},
	})
}

func envelopeToGenerated(envelope *Envelope) *wirepb.Envelope {
	if envelope == nil {
		return nil
	}
	return &wirepb.Envelope{
		MessageId:            envelope.MessageID,
		CorrelationId:        envelope.CorrelationID,
		Source:               envelope.Source,
		Target:               envelope.Target,
		ReplyTo:              envelope.ReplyTo,
		Kind:                 wirepb.MessageKind(envelope.Kind),
		OneWay:               envelope.OneWay,
		TimeoutMs:            envelope.TimeoutMillis,
		Payload:              envelope.Payload,
		Headers:              envelope.Headers,
		Status:               wirepb.BusinessStatus(envelope.Status),
		ErrorCode:            envelope.ErrorCode,
		ErrorMessage:         envelope.ErrorMessage,
		DeliveryHint:         wirepb.DeliveryHint(envelope.DeliveryHint),
		MessageType:          envelope.MessageType,
		PayloadSchemaVersion: envelope.PayloadSchemaVersion,
		PayloadFormat:        wirepb.PayloadFormat(envelope.PayloadFormat),
	}
}

func envelopeFromGenerated(envelope *wirepb.Envelope) *Envelope {
	if envelope == nil {
		return nil
	}
	return &Envelope{
		MessageID:            envelope.MessageId,
		CorrelationID:        envelope.CorrelationId,
		Source:               envelope.Source,
		Target:               envelope.Target,
		ReplyTo:              envelope.ReplyTo,
		Kind:                 int32(envelope.Kind),
		OneWay:               envelope.OneWay,
		TimeoutMillis:        envelope.TimeoutMs,
		Payload:              envelope.Payload,
		Headers:              envelope.Headers,
		Status:               int32(envelope.Status),
		ErrorCode:            envelope.ErrorCode,
		ErrorMessage:         envelope.ErrorMessage,
		DeliveryHint:         int32(envelope.DeliveryHint),
		MessageType:          envelope.MessageType,
		PayloadSchemaVersion: envelope.PayloadSchemaVersion,
		PayloadFormat:        int32(envelope.PayloadFormat),
	}
}
