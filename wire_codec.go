package coakka_v2_connector

import "github.com/phuong-tran/coakka-runtime-go/internal/wirecodec"

func encodeEnvelope(envelope *Envelope) ([]byte, error) {
	return wirecodec.EncodeEnvelope(envelopeToWire(envelope))
}

func decodeEnvelope(data []byte) (*Envelope, error) {
	wireEnvelope, err := wirecodec.DecodeEnvelope(data)
	if err != nil {
		return nil, err
	}
	return envelopeFromWire(wireEnvelope), nil
}

func decodeDeadletter(data []byte) (*Deadletter, error) {
	wireDeadletter, err := wirecodec.DecodeDeadletter(data)
	if err != nil {
		return nil, err
	}
	return &Deadletter{
		OriginalEnvelope: envelopeFromWire(wireDeadletter.OriginalEnvelope),
		Reason:           DeadletterReason(wireDeadletter.Reason),
		Detail:           wireDeadletter.Detail,
		ActiveGeneration: wireDeadletter.ActiveGeneration,
		ResolvedHost:     wireDeadletter.ResolvedHost,
		ResolvedPort:     wireDeadletter.ResolvedPort,
	}, nil
}

func encodeRouteSnapshotControl(
	generation uint64,
	routes []RouteSpec,
	sourceConnector string,
	seq uint64,
	overloadPolicy *RuntimeOverloadPolicy,
) ([]byte, error) {
	wireRoutes := make([]wirecodec.Route, 0, len(routes))
	for _, route := range routes {
		wireEndpoints := make([]wirecodec.Endpoint, 0, len(route.Endpoints))
		for _, endpoint := range route.Endpoints {
			weight := endpoint.Weight
			if weight == 0 {
				weight = 1
			}
			wireEndpoints = append(wireEndpoints, wirecodec.Endpoint{
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
		wireRoutes = append(wireRoutes, wirecodec.Route{
			Target:       route.Target,
			Endpoints:    wireEndpoints,
			Strategy:     int32(strategy),
			RouteKeyHint: route.RouteKeyHint,
			Flags:        route.Flags,
		})
	}

	var wirePolicy *wirecodec.OverloadPolicy
	if overloadPolicy != nil {
		wirePolicy = &wirecodec.OverloadPolicy{
			IngressMode:                     int32(overloadPolicy.IngressMode),
			LocalDeliveryMode:               int32(overloadPolicy.LocalDeliveryMode),
			RemoteOutboundMode:              int32(overloadPolicy.RemoteOutboundMode),
			RemoteOutboundReplyReserveSlots: overloadPolicy.RemoteOutboundReplyReserveSlots,
		}
	}
	return wirecodec.EncodeRouteSnapshotControl(generation, wireRoutes, sourceConnector, seq, wirePolicy)
}

func envelopeToWire(envelope *Envelope) *wirecodec.Envelope {
	if envelope == nil {
		return nil
	}
	return &wirecodec.Envelope{
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

func envelopeFromWire(envelope *wirecodec.Envelope) *Envelope {
	if envelope == nil {
		return nil
	}
	return &Envelope{
		MessageId:            envelope.MessageID,
		CorrelationId:        envelope.CorrelationID,
		Source:               envelope.Source,
		Target:               envelope.Target,
		ReplyTo:              envelope.ReplyTo,
		Kind:                 MessageKind(envelope.Kind),
		OneWay:               envelope.OneWay,
		TimeoutMs:            envelope.TimeoutMillis,
		Payload:              envelope.Payload,
		Headers:              envelope.Headers,
		Status:               BusinessStatus(envelope.Status),
		ErrorCode:            envelope.ErrorCode,
		ErrorMessage:         envelope.ErrorMessage,
		DeliveryHint:         DeliveryHint(envelope.DeliveryHint),
		MessageType:          envelope.MessageType,
		PayloadSchemaVersion: envelope.PayloadSchemaVersion,
		PayloadFormat:        PayloadFormat(envelope.PayloadFormat),
	}
}

func cloneEnvelope(envelope *Envelope) *Envelope {
	if envelope == nil {
		return nil
	}
	clone := *envelope
	clone.Payload = append([]byte(nil), envelope.Payload...)
	if envelope.Headers != nil {
		clone.Headers = make(map[string]string, len(envelope.Headers))
		for key, value := range envelope.Headers {
			clone.Headers[key] = value
		}
	}
	return &clone
}
