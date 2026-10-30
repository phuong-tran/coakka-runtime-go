package coakka_v2_connector

/*
#include "runtime_bridge.h"
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

// These are synchronous borrows. Pin every referenced Go allocation, including
// the records containing pointers; Core retains none after encode returns.
func pinValueText(pin *runtime.Pinner, text string) C.coakka_v2_message_span_t {
	if len(text) == 0 {
		return C.coakka_v2_message_span_t{}
	}
	data := unsafe.StringData(text)
	pin.Pin(data)
	return C.coakka_v2_message_span_t{data: (*C.uint8_t)(unsafe.Pointer(data)), size: C.size_t(len(text))}
}

func pinValueBytes(pin *runtime.Pinner, bytes []byte) C.coakka_v2_message_span_t {
	if len(bytes) == 0 {
		return C.coakka_v2_message_span_t{}
	}
	pin.Pin(&bytes[0])
	return C.coakka_v2_message_span_t{data: (*C.uint8_t)(unsafe.Pointer(&bytes[0])), size: C.size_t(len(bytes))}
}

func pinValueArray[T any](pin *runtime.Pinner, values []T) *T {
	if len(values) == 0 {
		return nil
	}
	pin.Pin(&values[0])
	return &values[0]
}

// Bound only duplicate language staging, using the owning SDK ceiling. Text and
// payload borrow their existing storage; Core still validates its total charge.
func chargeValueTable(remaining *uint64, count int, width uintptr) error {
	if width == 0 || uint64(count) > *remaining/uint64(width) {
		return fmt.Errorf("value table exceeds the SDK staging bound")
	}
	*remaining -= uint64(count) * uint64(width)
	return nil
}

func (b *nativeBindings) encodedValue(status C.int, bytes *C.uint8_t, size C.size_t, operation string) ([]byte, error) {
	if err := requireStatus(status, operation); err != nil {
		return nil, err
	}
	defer C.coakka_v2_go_client_bytes_release(b.ptr, bytes)
	if err := checkValueSpan(C.coakka_v2_message_span_t{data: bytes, size: size}); err != nil {
		return nil, err
	}
	return C.GoBytes(unsafe.Pointer(bytes), C.int(size)), nil
}

func (b *nativeBindings) encodeEnvelope(envelope *Envelope) ([]byte, error) {
	if envelope == nil {
		return nil, fmt.Errorf("envelope is nil")
	}
	remaining := uint64(C.COAKKA_V2_MESSAGE_VALUE_MAX_METADATA_BYTES)
	if err := chargeValueTable(&remaining, len(envelope.Headers), C.sizeof_coakka_v2_message_property_t); err != nil {
		return nil, err
	}
	var pin runtime.Pinner
	defer pin.Unpin()
	properties := make([]C.coakka_v2_message_property_t, len(envelope.Headers))
	i := 0
	for key, value := range envelope.Headers {
		properties[i] = C.coakka_v2_message_property_t{key: pinValueText(&pin, key), value: pinValueText(&pin, value)}
		i++
	}
	identity := C.coakka_v2_message_payload_identity_t{
		message_type:   pinValueText(&pin, envelope.MessageType),
		schema_version: C.uint32_t(envelope.PayloadSchemaVersion), format: C.int32_t(envelope.PayloadFormat),
	}
	business := C.coakka_v2_message_business_result_t{status: C.int32_t(envelope.Status),
		error_code: pinValueText(&pin, envelope.ErrorCode), error_message: pinValueText(&pin, envelope.ErrorMessage)}
	pin.Pin(&identity)
	pin.Pin(&business)
	value := C.coakka_v2_message_value_t{
		struct_size: C.sizeof_coakka_v2_message_value_t,
		address: C.coakka_v2_message_address_t{
			message_id: pinValueText(&pin, envelope.MessageId), correlation_id: pinValueText(&pin, envelope.CorrelationId),
			source: pinValueText(&pin, envelope.Source), target: pinValueText(&pin, envelope.Target), reply_to: pinValueText(&pin, envelope.ReplyTo)},
		delivery: C.coakka_v2_message_delivery_t{kind: C.int32_t(envelope.Kind), timeout_ms: C.int32_t(envelope.TimeoutMs), delivery_hint: C.int32_t(envelope.DeliveryHint)},
		payload:  pinValueBytes(&pin, envelope.Payload), payload_identity: &identity, business: &business,
		properties: pinValueArray(&pin, properties), property_count: C.size_t(len(properties)),
	}
	if envelope.OneWay {
		value.delivery.one_way = 1
	}
	var bytes *C.uint8_t
	var size C.size_t
	status := C.coakka_v2_go_encode_message_value(b.ptr, &value, &bytes, &size)
	return b.encodedValue(status, bytes, size, "client_encode_message_value")
}

func checkValueSpan(span C.coakka_v2_message_span_t) error {
	if uint64(span.size) > maxNativeProjectionBytes || (span.size != 0 && span.data == nil) {
		return &CoreContractError{Operation: "value projection", Detail: "invalid pointer or length"}
	}
	return nil
}

func valueText(span C.coakka_v2_message_span_t) (string, error) {
	if err := checkValueSpan(span); err != nil {
		return "", err
	}
	if span.size == 0 {
		return "", nil
	}
	return C.GoStringN((*C.char)(unsafe.Pointer(span.data)), C.int(span.size)), nil
}

func valueProperties(data *C.coakka_v2_message_property_t, count C.size_t) (map[string]string, error) {
	if uint64(count) > uint64(C.COAKKA_V2_MESSAGE_VALUE_MAX_METADATA_BYTES)/C.sizeof_coakka_v2_message_property_t || (count != 0 && data == nil) {
		return nil, &CoreContractError{Operation: "value projection", Detail: "invalid property array"}
	}
	if count == 0 {
		return nil, nil
	}
	result := make(map[string]string, int(count))
	for _, property := range unsafe.Slice(data, int(count)) {
		key, err := valueText(property.key)
		if err != nil {
			return nil, err
		}
		value, err := valueText(property.value)
		if err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}

func envelopeValue(value *C.coakka_v2_message_value_t) (*Envelope, error) {
	if value == nil {
		return nil, nil
	}
	if value.struct_size != C.sizeof_coakka_v2_message_value_t || value.delivery.one_way > 1 {
		return nil, &CoreContractError{Operation: "value projection", Detail: "invalid record size or boolean"}
	}
	result := &Envelope{Kind: MessageKind(value.delivery.kind), OneWay: value.delivery.one_way != 0,
		TimeoutMs: int32(value.delivery.timeout_ms), DeliveryHint: DeliveryHint(value.delivery.delivery_hint)}
	for _, field := range []struct {
		span C.coakka_v2_message_span_t
		out  *string
	}{
		{value.address.message_id, &result.MessageId}, {value.address.correlation_id, &result.CorrelationId},
		{value.address.source, &result.Source}, {value.address.target, &result.Target}, {value.address.reply_to, &result.ReplyTo},
	} {
		text, err := valueText(field.span)
		if err != nil {
			return nil, err
		}
		*field.out = text
	}
	if err := checkValueSpan(value.payload); err != nil {
		return nil, err
	}
	if value.payload.size != 0 {
		result.Payload = C.GoBytes(unsafe.Pointer(value.payload.data), C.int(value.payload.size))
	}
	var err error
	result.Headers, err = valueProperties(value.properties, value.property_count)
	if err != nil {
		return nil, err
	}
	if identity := value.payload_identity; identity != nil {
		result.MessageType, err = valueText(identity.message_type)
		if err != nil {
			return nil, err
		}
		result.PayloadSchemaVersion, result.PayloadFormat = uint32(identity.schema_version), PayloadFormat(identity.format)
	}
	if business := value.business; business != nil {
		result.Status = BusinessStatus(business.status)
		result.ErrorCode, err = valueText(business.error_code)
		if err != nil {
			return nil, err
		}
		result.ErrorMessage, err = valueText(business.error_message)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (b *nativeBindings) decodeEnvelope(data []byte) (*Envelope, error) {
	var owner *C.coakka_v2_message_value_owner_t
	value := C.coakka_v2_message_value_t{struct_size: C.sizeof_coakka_v2_message_value_t}
	status := C.coakka_v2_go_decode_message_value(b.ptr, (*C.uint8_t)(unsafe.Pointer(unsafe.SliceData(data))), C.size_t(len(data)), &owner, &value)
	if err := requireStatus(status, "client_decode_message_value"); err != nil {
		return nil, err
	}
	if owner == nil {
		return nil, &CoreContractError{Operation: "client_decode_message_value", Detail: "success without owner"}
	}
	defer C.coakka_v2_go_message_value_release(b.ptr, owner)
	return envelopeValue(&value)
}

func (b *nativeBindings) decodeDeadletter(data []byte) (*Deadletter, error) {
	var owner *C.coakka_v2_message_value_owner_t
	value := C.coakka_v2_deadletter_value_t{struct_size: C.sizeof_coakka_v2_deadletter_value_t}
	status := C.coakka_v2_go_decode_deadletter_value(b.ptr, (*C.uint8_t)(unsafe.Pointer(unsafe.SliceData(data))), C.size_t(len(data)), &owner, &value)
	if err := requireStatus(status, "client_decode_deadletter_value"); err != nil {
		return nil, err
	}
	if owner == nil {
		return nil, &CoreContractError{Operation: "client_decode_deadletter_value", Detail: "success without owner"}
	}
	defer C.coakka_v2_go_message_value_release(b.ptr, owner)
	if value.struct_size != C.sizeof_coakka_v2_deadletter_value_t {
		return nil, &CoreContractError{Operation: "client_decode_deadletter_value", Detail: "invalid record size"}
	}
	result := &Deadletter{Reason: DeadletterReason(value.reason), ActiveGeneration: uint64(value.active_generation),
		ResolvedPort: uint32(value.resolved_port), EndpointAttemptCount: uint32(value.endpoint_attempt_count)}
	var err error
	result.OriginalEnvelope, err = envelopeValue(value.original)
	if err != nil {
		return nil, err
	}
	result.Detail, err = valueText(value.detail)
	if err != nil {
		return nil, err
	}
	result.ResolvedHost, err = valueText(value.resolved_host)
	if err != nil {
		return nil, err
	}
	result.FailureDetails, err = valueProperties(value.failure_details, value.failure_detail_count)
	if err != nil {
		return nil, err
	}
	if failure := value.failure; failure != nil {
		result.NativeFailure = &RuntimeFailureSnapshot{Domain: uint32(failure.domain), Phase: uint32(failure.phase),
			Code: uint32(failure.code), Certainty: uint32(failure.certainty), Scope: uint32(failure.scope)}
	}
	return result, nil
}

func (b *nativeBindings) encodeRouteSnapshotControl(generation uint64, routes []RouteSpec, sourceConnector string, seq uint64, policy *RuntimeOverloadPolicy) ([]byte, error) {
	remaining := uint64(C.COAKKA_V2_MESSAGE_VALUE_MAX_METADATA_BYTES)
	if err := chargeValueTable(&remaining, len(routes), C.sizeof_coakka_v2_control_route_value_t); err != nil {
		return nil, err
	}
	for _, route := range routes {
		if err := chargeValueTable(&remaining, len(route.Endpoints), C.sizeof_coakka_v2_control_endpoint_value_t); err != nil {
			return nil, err
		}
	}
	var pin runtime.Pinner
	defer pin.Unpin()
	values := make([]C.coakka_v2_control_route_value_t, len(routes))
	for i, route := range routes {
		endpoints := make([]C.coakka_v2_control_endpoint_value_t, len(route.Endpoints))
		for j, endpoint := range route.Endpoints {
			endpoints[j] = C.coakka_v2_control_endpoint_value_t{host: pinValueText(&pin, endpoint.Host),
				port: C.uint32_t(endpoint.Port), weight: C.uint32_t(endpoint.Weight), flags: C.uint32_t(endpoint.Flags)}
		}
		values[i] = C.coakka_v2_control_route_value_t{target: pinValueText(&pin, route.Target),
			route_key_hint: pinValueText(&pin, route.RouteKeyHint), strategy: C.int32_t(route.Strategy), flags: C.uint32_t(route.Flags),
			endpoints: pinValueArray(&pin, endpoints), endpoint_count: C.size_t(len(endpoints))}
	}
	snapshot := C.coakka_v2_control_snapshot_value_t{generation: C.uint64_t(generation), routes: pinValueArray(&pin, values), route_count: C.size_t(len(values))}
	metadata := C.coakka_v2_control_value_metadata_t{seq: C.uint64_t(seq), source_connector: pinValueText(&pin, sourceConnector)}
	pin.Pin(&snapshot)
	pin.Pin(&metadata)
	value := C.coakka_v2_control_value_t{struct_size: C.sizeof_coakka_v2_control_value_t, snapshot: &snapshot, metadata: &metadata}
	if policy != nil {
		native := &C.coakka_v2_control_overload_value_t{ingress_mode: C.int32_t(policy.IngressMode),
			local_delivery_mode: C.int32_t(policy.LocalDeliveryMode), remote_outbound_mode: C.int32_t(policy.RemoteOutboundMode),
			remote_outbound_reply_reserve_slots: C.uint32_t(policy.RemoteOutboundReplyReserveSlots)}
		pin.Pin(native)
		value.policy = native
	}
	var bytes *C.uint8_t
	var size C.size_t
	status := C.coakka_v2_go_encode_control_value(b.ptr, &value, &bytes, &size)
	return b.encodedValue(status, bytes, size, "client_encode_control_value")
}
