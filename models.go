package coakka_v2_connector

import (
	"fmt"
	"strings"
)

type MessageKind int32

const (
	MessageKindUnspecified MessageKind = 0
	MessageKindRequest     MessageKind = 1
	MessageKindResponse    MessageKind = 2
	MessageKindEvent       MessageKind = 3
)

type BusinessStatus int32

const (
	BusinessStatusUnspecified BusinessStatus = 0
	BusinessStatusOK          BusinessStatus = 1
	BusinessStatusError       BusinessStatus = 2
)

type DeliveryHint int32

const (
	DeliveryHintUnspecified   DeliveryHint = 0
	DeliveryHintRouterDefault DeliveryHint = 1
	DeliveryHintPreferLocal   DeliveryHint = 2
	DeliveryHintRequireLocal  DeliveryHint = 3
	DeliveryHintRequireRemote DeliveryHint = 4
)

type PayloadFormat int32

const (
	PayloadFormatUnspecified PayloadFormat = 0
	PayloadFormatJSON        PayloadFormat = 1
	PayloadFormatProtobuf    PayloadFormat = 2
	PayloadFormatThrift      PayloadFormat = 3
	PayloadFormatMsgpack     PayloadFormat = 4
	PayloadFormatText        PayloadFormat = 5
	PayloadFormatPlainText   PayloadFormat = PayloadFormatText
	PayloadFormatBinary      PayloadFormat = 6
)

func (f PayloadFormat) String() string {
	switch f {
	case PayloadFormatJSON:
		return "JSON"
	case PayloadFormatProtobuf:
		return "PROTOBUF"
	case PayloadFormatThrift:
		return "THRIFT"
	case PayloadFormatMsgpack:
		return "MSGPACK"
	case PayloadFormatText:
		return "TEXT"
	case PayloadFormatBinary:
		return "BINARY"
	default:
		return "UNSPECIFIED"
	}
}

type RouteStrategy int32

const (
	RouteStrategyUnspecified        RouteStrategy = 0
	RouteStrategySingleOwner        RouteStrategy = 1
	RouteStrategyWeightedRoundRobin RouteStrategy = 2
	RouteStrategyRendezvousHash     RouteStrategy = 3
)

type OverloadMode int32

const (
	OverloadModeReject           OverloadMode = 0
	OverloadModeDropExpiredFirst OverloadMode = 1
	OverloadModeDropOneWayFirst  OverloadMode = 2
)

type EndpointFlag uint32

const (
	EndpointFlagNone        EndpointFlag = 0
	EndpointFlagLocal       EndpointFlag = 1
	EndpointFlagUnavailable EndpointFlag = 1 << 1
)

type PayloadIdentity struct {
	MessageType          string
	PayloadSchemaVersion uint32
	PayloadFormat        PayloadFormat
}

func NewPayloadIdentity(messageType string, payloadSchemaVersion uint32, payloadFormat PayloadFormat) PayloadIdentity {
	return PayloadIdentity{
		MessageType:          messageType,
		PayloadSchemaVersion: payloadSchemaVersion,
		PayloadFormat:        payloadFormat,
	}
}

func NewTextPayloadIdentity(messageType string) PayloadIdentity {
	return NewPayloadIdentity(messageType, 1, PayloadFormatText)
}

func (p PayloadIdentity) IsTyped() bool {
	return strings.TrimSpace(p.MessageType) != "" &&
		p.PayloadSchemaVersion >= 1 &&
		p.PayloadFormat != PayloadFormatUnspecified
}

func (p PayloadIdentity) RequireTyped(what string) error {
	if strings.TrimSpace(p.MessageType) == "" {
		return fmt.Errorf("%s requires messageType", what)
	}
	if p.PayloadSchemaVersion < 1 {
		return fmt.Errorf("%s requires payloadSchemaVersion >= 1", what)
	}
	if p.PayloadFormat == PayloadFormatUnspecified {
		return fmt.Errorf("%s requires declared payloadFormat", what)
	}
	return nil
}

type EndpointSpec struct {
	Host   string
	Port   uint32
	Weight uint32
	Flags  uint32
}

type RouteSpec struct {
	Target       string
	Endpoints    []EndpointSpec
	Strategy     RouteStrategy
	RouteKeyHint string
	Flags        uint32
}

func LocalRoute(target string, port uint32) RouteSpec {
	return RouteSpec{
		Target: target,
		Endpoints: []EndpointSpec{{
			Host:   "127.0.0.1",
			Port:   port,
			Weight: 1,
			Flags:  uint32(EndpointFlagLocal),
		}},
		Strategy: RouteStrategySingleOwner,
	}
}

func LocalRouteDefault(target string) RouteSpec {
	return LocalRoute(target, 9001)
}

func (r RouteSpec) RequireValid() error {
	if strings.TrimSpace(r.Target) == "" {
		return fmt.Errorf("route requires target")
	}
	if len(r.Endpoints) == 0 {
		return fmt.Errorf("route target=%s requires at least one endpoint", r.Target)
	}
	return nil
}

type RuntimeOverloadPolicy struct {
	IngressMode                     OverloadMode
	LocalDeliveryMode               OverloadMode
	RemoteOutboundMode              OverloadMode
	RemoteOutboundReplyReserveSlots uint32
}

func (p RuntimeOverloadPolicy) RequireValid() error {
	if p.IngressMode != OverloadModeReject {
		return fmt.Errorf("overload policy ingress mode currently supports only reject")
	}
	if p.LocalDeliveryMode != OverloadModeReject {
		return fmt.Errorf("overload policy local delivery mode currently supports only reject")
	}
	return nil
}

type ConnectorConfig struct {
	SystemName                   string
	NodeID                       string
	Routes                       []RouteSpec
	StrictNoDrop                 bool
	QueueCapacity                int
	EnableMonitor                bool
	SeparateDeliveredRequestLane bool
	Generation                   uint64
	OverloadPolicy               *RuntimeOverloadPolicy
	ConnectionStrategy           *RuntimeTCPConnectionStrategySpec
	Security                     *RuntimeTCPSecuritySpec
}

func (c ConnectorConfig) RequireValid() error {
	if strings.TrimSpace(c.SystemName) == "" {
		return fmt.Errorf("connector config requires systemName")
	}
	if strings.TrimSpace(c.NodeID) == "" {
		return fmt.Errorf("connector config requires nodeId")
	}
	if c.QueueCapacity < 1 {
		return fmt.Errorf("connector config requires queueCapacity >= 1")
	}
	if c.Generation < 1 {
		return fmt.Errorf("connector config requires generation >= 1")
	}
	if c.OverloadPolicy != nil {
		if err := c.OverloadPolicy.RequireValid(); err != nil {
			return err
		}
	}
	if len(c.Routes) == 0 {
		return fmt.Errorf("connector config requires at least one route")
	}
	for _, route := range c.Routes {
		if err := route.RequireValid(); err != nil {
			return err
		}
	}
	return nil
}

type ConnectorStartSpec struct {
	SystemName                          string
	NodeID                              string
	Routes                              []RouteSpec
	StrictNoDrop                        bool
	QueueCapacity                       int
	EnableMonitor                       bool
	SeparateDeliveredRequestLane        bool
	DisableSeparateDeliveredRequestLane bool
	Generation                          uint64
	OverloadPolicy                      *RuntimeOverloadPolicy
	ConnectionStrategy                  *RuntimeTCPConnectionStrategySpec
	Security                            *RuntimeTCPSecuritySpec
}

func (s ConnectorStartSpec) ToConnectorConfig() ConnectorConfig {
	normalized := s.Normalized()
	return ConnectorConfig{
		SystemName:                   normalized.SystemName,
		NodeID:                       normalized.NodeID,
		Routes:                       normalized.Routes,
		StrictNoDrop:                 normalized.StrictNoDrop,
		QueueCapacity:                normalized.QueueCapacity,
		EnableMonitor:                normalized.EnableMonitor,
		SeparateDeliveredRequestLane: normalized.SeparateDeliveredRequestLane,
		Generation:                   normalized.Generation,
		OverloadPolicy:               normalized.OverloadPolicy,
		ConnectionStrategy:           normalized.ConnectionStrategy,
		Security:                     normalized.Security,
	}
}

func (s ConnectorStartSpec) Normalized() ConnectorStartSpec {
	out := s
	if out.QueueCapacity == 0 {
		out.QueueCapacity = 128
	}
	if out.Generation == 0 {
		out.Generation = 1
	}
	if !out.EnableMonitor {
		out.EnableMonitor = true
	}
	if !out.DisableSeparateDeliveredRequestLane {
		out.SeparateDeliveredRequestLane = true
	}
	if !out.StrictNoDrop {
		out.StrictNoDrop = true
	}
	return out
}

func (s ConnectorStartSpec) RequireValid() error {
	return s.ToConnectorConfig().RequireValid()
}

type SubmittedRequest struct {
	MessageID     string
	CorrelationID string
}

type RequestTerminalEventKind string

const (
	RequestTerminalEventResponse   RequestTerminalEventKind = "response"
	RequestTerminalEventDeadletter RequestTerminalEventKind = "deadletter"
)

type RequestTerminalEvent struct {
	Kind             RequestTerminalEventKind
	RequestMessageID string
	CorrelationID    string
	Response         *Envelope
	Deadletter       *Deadletter
}

type ObservedDeadletter struct {
	Deadletter            *Deadletter `json:"deadletter"`
	RequestMessageID      string      `json:"requestMessageId,omitempty"`
	CorrelationID         string      `json:"correlationId,omitempty"`
	MatchedPendingRequest bool        `json:"matchedPendingRequest"`
}
