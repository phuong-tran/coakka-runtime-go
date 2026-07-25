package coakka_v2_connector

import (
	"fmt"

	coakkav2 "github.com/phuong-tran/coakka-runtime-go/coakka/v2"
	"google.golang.org/protobuf/proto"
)

type Envelope = coakkav2.Envelope
type Deadletter = coakkav2.Deadletter
type ControlEnvelope = coakkav2.ControlEnvelope
type RouteSnapshotPayload = coakkav2.RouteSnapshotPayload

func encodeEnvelope(envelope *Envelope) ([]byte, error) {
	if envelope == nil {
		return nil, fmt.Errorf("envelope is nil")
	}
	return proto.Marshal(envelope)
}

func decodeEnvelope(data []byte) (*Envelope, error) {
	envelope := &coakkav2.Envelope{}
	if err := proto.Unmarshal(data, envelope); err != nil {
		return nil, err
	}
	return envelope, nil
}

func decodeDeadletter(data []byte) (*Deadletter, error) {
	deadletter := &coakkav2.Deadletter{}
	if err := proto.Unmarshal(data, deadletter); err != nil {
		return nil, err
	}
	return deadletter, nil
}

func encodeControlEnvelope(envelope *ControlEnvelope) ([]byte, error) {
	if envelope == nil {
		return nil, fmt.Errorf("control envelope is nil")
	}
	return proto.Marshal(envelope)
}

func encodeRouteSnapshotPayload(payload *RouteSnapshotPayload) ([]byte, error) {
	if payload == nil {
		return nil, fmt.Errorf("route snapshot payload is nil")
	}
	return proto.Marshal(payload)
}
