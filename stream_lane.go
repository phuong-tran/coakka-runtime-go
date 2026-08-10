package coakka_v2_connector

/*
#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>

typedef struct coakka_v2_stream_lane_t coakka_v2_stream_lane_t;
typedef struct coakka_v2_go_bindings_t coakka_v2_go_bindings_t;

typedef struct {
  size_t struct_size;
  uint64_t sequence;
  uint64_t captured_mono_ns;
  uint64_t dropped_before;
  uint32_t flags;
  size_t size;
} coakka_v2_stream_frame_t;

typedef struct {
  size_t struct_size;
  uint32_t mode;
  uint32_t reserved;
  uint64_t credential_generation;
  const char *credential_id;
  const char *ca_certificate_file;
  const char *identity_certificate_file;
  const char *private_key_file;
} coakka_v2_stream_lane_security_config_t;

typedef struct {
  size_t struct_size;
  uint32_t flags;
  const char *bind_host;
  uint16_t bind_port;
  size_t capacity;
  uint32_t max_frame_bytes;
  uint32_t max_window_bytes;
  uint32_t io_timeout_ms;
  uint32_t source_retry_ms;
  uint32_t progress_frames;
  uint32_t progress_interval_ms;
  uint32_t publisher_worker_count;
  uint32_t subscriber_worker_count;
  const coakka_v2_stream_lane_security_config_t *security;
  uint32_t pressure_after_ms;
  uint32_t stalled_after_ms;
  uint32_t recovery_after_ms;
  uint32_t pressure_observation_ms;
} coakka_v2_stream_lane_config_t;

typedef struct {
  size_t struct_size;
  uint32_t direction;
  uint32_t state;
  uint32_t result;
  uint64_t format_id;
  uint64_t frames;
  uint64_t bytes;
  uint64_t dropped_frames;
  uint64_t last_sequence;
  uint32_t negotiated_max_frame_bytes;
  uint32_t window_bytes;
  uint32_t cancel_requested;
  uint64_t update_sequence;
  uint64_t submitted_mono_ns;
  uint64_t started_mono_ns;
  uint64_t updated_mono_ns;
  uint64_t terminal_mono_ns;
  char detail[160];
} coakka_v2_stream_session_snapshot_t;

typedef struct {
  size_t struct_size;
  uint32_t direction;
  uint32_t state;
  uint32_t reason_bits;
  uint32_t available_credit_bytes;
  uint32_t window_capacity_bytes;
  uint64_t update_sequence;
  uint64_t transition_count;
  uint64_t observed_mono_ns;
  uint64_t state_started_mono_ns;
  uint64_t pressure_started_mono_ns;
  uint64_t last_progress_mono_ns;
  uint64_t observed_delivery_bps;
  uint64_t current_operation_ns;
  uint64_t last_operation_ns;
  uint64_t total_pressured_ns;
  uint64_t max_pressured_ns;
} coakka_v2_stream_pressure_snapshot_t;

typedef struct {
  size_t struct_size;
  size_t capacity;
  size_t queued_subscribers;
  size_t prepared_publishers;
  size_t active_publishers;
  size_t active_subscribers;
  size_t retained_records;
  uint64_t submitted_subscribers;
  uint64_t prepared_publisher_count;
  uint64_t ended_publishers;
  uint64_t ended_subscribers;
  uint64_t failed_publishers;
  uint64_t failed_subscribers;
  uint64_t canceled_sessions;
  uint64_t published_frames;
  uint64_t published_bytes;
  uint64_t consumed_frames;
  uint64_t consumed_bytes;
  uint64_t source_reported_drops;
} coakka_v2_stream_lane_stats_t;

int coakka_v2_go_stream_lane_available(coakka_v2_go_bindings_t *);
int coakka_v2_go_stream_lane_create_ex(coakka_v2_go_bindings_t *, const coakka_v2_stream_lane_config_t *, coakka_v2_stream_lane_t **);
void coakka_v2_go_stream_lane_destroy(coakka_v2_go_bindings_t *, coakka_v2_stream_lane_t *);
int coakka_v2_go_stream_lane_start(coakka_v2_go_bindings_t *, coakka_v2_stream_lane_t *);
int coakka_v2_go_stream_lane_stop(coakka_v2_go_bindings_t *, coakka_v2_stream_lane_t *);
int coakka_v2_go_stream_lane_get_bound_port(coakka_v2_go_bindings_t *, coakka_v2_stream_lane_t *, uint16_t *);
int coakka_v2_go_stream_lane_prepare_publish(coakka_v2_go_bindings_t *, coakka_v2_stream_lane_t *, const char *, const char *, uint64_t, uint32_t, uintptr_t);
int coakka_v2_go_stream_lane_subscribe(coakka_v2_go_bindings_t *, coakka_v2_stream_lane_t *, const char *, const char *, const char *, uint16_t, uint64_t, uint32_t, uint32_t, uint32_t, uintptr_t);
int coakka_v2_go_stream_lane_get_session(coakka_v2_go_bindings_t *, coakka_v2_stream_lane_t *, const char *, uint32_t, coakka_v2_stream_session_snapshot_t *);
int coakka_v2_go_stream_lane_wait_session(coakka_v2_go_bindings_t *, coakka_v2_stream_lane_t *, const char *, uint32_t, uint64_t, uint32_t, coakka_v2_stream_session_snapshot_t *);
int coakka_v2_go_stream_lane_get_pressure(coakka_v2_go_bindings_t *, coakka_v2_stream_lane_t *, const char *, uint32_t, coakka_v2_stream_pressure_snapshot_t *);
int coakka_v2_go_stream_lane_wait_pressure(coakka_v2_go_bindings_t *, coakka_v2_stream_lane_t *, const char *, uint32_t, uint64_t, uint32_t, coakka_v2_stream_pressure_snapshot_t *);
int coakka_v2_go_stream_lane_cancel_session(coakka_v2_go_bindings_t *, coakka_v2_stream_lane_t *, const char *, uint32_t);
int coakka_v2_go_stream_lane_forget_session(coakka_v2_go_bindings_t *, coakka_v2_stream_lane_t *, const char *, uint32_t);
int coakka_v2_go_stream_lane_get_stats(coakka_v2_go_bindings_t *, coakka_v2_stream_lane_t *, coakka_v2_stream_lane_stats_t *);
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime/cgo"
	"sync"
	"unsafe"
)

const (
	// StreamLanePublisher enables prepared publishers and the lane listener.
	StreamLanePublisher uint32 = 1
	// StreamLaneSubscriber enables queued outbound subscribers.
	StreamLaneSubscriber uint32 = 1 << 1
	// StreamFrameKeyframe marks an application-defined key frame.
	StreamFrameKeyframe uint32 = 1
	// StreamFrameDiscontinuity marks an application-defined discontinuity.
	StreamFrameDiscontinuity uint32 = 1 << 1
	// StreamFrameEndOfSegment marks an application-defined segment boundary.
	StreamFrameEndOfSegment uint32 = 1 << 2
	// StreamPressureCreditWait means the publisher is waiting for receive credit.
	StreamPressureCreditWait uint32 = 1
	// StreamPressureTransportWrite means a transport write exceeded the threshold.
	StreamPressureTransportWrite uint32 = 1 << 1
	// StreamPressureConsumerBusy means the consumer exceeded the threshold.
	StreamPressureConsumerBusy uint32 = 1 << 2
	// StreamPressureTransportRead means a transport read exceeded the threshold.
	StreamPressureTransportRead uint32 = 1 << 3
)

// StreamLaneSecurityMode selects transport protection. Session authorization remains mandatory.
type StreamLaneSecurityMode uint32

const (
	StreamLaneDirect StreamLaneSecurityMode = iota
	StreamLaneTLS
	StreamLaneMutualTLS
)

// StreamDirection identifies one retained side of a stream session.
type StreamDirection uint32

const (
	StreamPublish   StreamDirection = 1
	StreamSubscribe StreamDirection = 2
)

// StreamState is an observable lifecycle state for one stream-session side.
type StreamState uint32

const (
	StreamPrepared StreamState = iota + 1
	StreamQueued
	StreamConnecting
	StreamActive
	StreamStopping
	StreamEnded
	StreamRejected
	StreamFailed
	StreamCanceled
)

// StreamResult is a stable terminal outcome reported independently by each peer.
type StreamResult uint32

const (
	StreamResultNone StreamResult = iota
	StreamResultOK
	StreamResultNotPrepared
	StreamResultTokenMismatch
	StreamResultFormatMismatch
	StreamResultFrameLimit
	StreamResultNetworkIO
	StreamResultTimeout
	StreamResultQueueFull
	StreamResultProtocolError
	StreamResultSourceError
	StreamResultConsumerError
	StreamResultInternalError
	StreamResultCanceledByHost
	StreamResultTLSConfigInvalid
	StreamResultTLSHandshakeFailed
	StreamResultPeerCertUntrusted
	StreamResultPeerCertExpired
	StreamResultPeerIdentityMismatch
	StreamResultClientCertRequired
)

// StreamPressureState is a coalesced neutral transport-pressure state.
type StreamPressureState uint32

const (
	StreamPressureInactive StreamPressureState = iota
	StreamPressureFlowing
	StreamPressurePressured
	StreamPressureStalled
	StreamPressureRecovering
)

// StreamSourceStatus describes one bounded source callback outcome.
type StreamSourceStatus uint32

const (
	// StreamSourceFrame says Size bytes were written into the borrowed destination.
	StreamSourceFrame StreamSourceStatus = iota
	// StreamSourceWouldBlock asks native code to retry after the configured delay.
	StreamSourceWouldBlock
	// StreamSourceEnd ends the publisher normally.
	StreamSourceEnd
)

// StreamConsumerDecision tells native code whether to continue after consuming a frame.
type StreamConsumerDecision uint32

const (
	StreamConsumerContinue StreamConsumerDecision = iota
	StreamConsumerStop
)

// StreamSourceResult is returned by StreamSource.
// Size must be in [1, len(destination)] when Status is StreamSourceFrame.
type StreamSourceResult struct {
	Status         StreamSourceStatus
	Size           int
	CapturedMonoNS uint64
	DroppedBefore  uint64
	Flags          uint32
}

// StreamFrameMetadata is immutable metadata paired with one borrowed frame.
type StreamFrameMetadata struct {
	Sequence       uint64
	CapturedMonoNS uint64
	DroppedBefore  uint64
	Flags          uint32
}

// StreamSource fills the borrowed destination and returns promptly.
// The slice is valid only during the callback and must not be retained.
type StreamSource func(destination []byte) StreamSourceResult

// StreamConsumer consumes borrowed frame bytes and returns promptly.
// The slice is valid only during the callback and must not be retained.
type StreamConsumer func(frame []byte, metadata StreamFrameMetadata) StreamConsumerDecision

// StreamLaneSecurityConfig names TLS material copied when a lane starts.
// Credential paths and per-session authorization tokens must never be logged.
type StreamLaneSecurityConfig struct {
	Mode                    StreamLaneSecurityMode
	CredentialGeneration    uint64
	CredentialID            string
	CACertificateFile       string
	IdentityCertificateFile string
	PrivateKeyFile          string
}

// StreamLaneConfig controls bounded workers, frames, flow control, and pressure observation.
// Size fields are bytes, time fields are milliseconds, and zero tuning fields
// select conservative core-runtime defaults.
type StreamLaneConfig struct {
	Flags                     uint32
	BindHost                  string
	BindPort                  uint16
	Capacity                  uint64
	MaxFrameBytes             uint32
	MaxWindowBytes            uint32
	IOTimeoutMillis           uint32
	SourceRetryMillis         uint32
	ProgressFrames            uint32
	ProgressIntervalMillis    uint32
	PublisherWorkerCount      uint32
	SubscriberWorkerCount     uint32
	Security                  *StreamLaneSecurityConfig
	PressureAfterMillis       uint32
	StalledAfterMillis        uint32
	RecoveryAfterMillis       uint32
	PressureObservationMillis uint32
}

// DefaultStreamLaneConfig returns a loopback lane with publisher and subscriber enabled.
func DefaultStreamLaneConfig() StreamLaneConfig {
	return StreamLaneConfig{Flags: StreamLanePublisher | StreamLaneSubscriber, BindHost: "127.0.0.1"}
}

// StreamPublishSpec authorizes a publisher and provides its bounded frame source.
type StreamPublishSpec struct {
	SessionID          string
	AuthorizationToken string
	FormatID           uint64
	MaxFrameBytes      uint32
	Source             StreamSource
}

// StreamSubscribeSpec names the publisher endpoint and provides the frame consumer.
type StreamSubscribeSpec struct {
	SessionID          string
	AuthorizationToken string
	RemoteHost         string
	RemotePort         uint16
	FormatID           uint64
	MaxFrameBytes      uint32
	InitialWindowBytes uint32
	TimeoutMillis      uint32
	Consumer           StreamConsumer
}

// StreamSessionSnapshot is a copied session view with process-local monotonic timestamps.
// Frame and byte counters are cumulative for this side; UpdateSequence advances
// whenever retained state changes.
type StreamSessionSnapshot struct {
	Direction                                    StreamDirection
	State                                        StreamState
	Result                                       StreamResult
	FormatID, Frames, Bytes, DroppedFrames       uint64
	LastSequence                                 uint64
	NegotiatedMaxFrameBytes, WindowBytes         uint32
	CancelRequested                              bool
	UpdateSequence, SubmittedMonoNS              uint64
	StartedMonoNS, UpdatedMonoNS, TerminalMonoNS uint64
	Detail                                       string
}

// Terminal reports whether this side reached a terminal state.
func (s StreamSessionSnapshot) Terminal() bool { return s.State >= StreamEnded }

// Succeeded reports whether this side ended normally.
func (s StreamSessionSnapshot) Succeeded() bool {
	return s.State == StreamEnded && s.Result == StreamResultOK
}

// StreamPressureSnapshot is a copied, policy-neutral transport-pressure observation.
// Durations and timestamps are nanoseconds; ObservedDeliveryBPS is bytes per second.
type StreamPressureSnapshot struct {
	Direction                                         StreamDirection
	State                                             StreamPressureState
	ReasonBits                                        uint32
	AvailableCreditBytes, WindowCapacityBytes         uint32
	UpdateSequence, TransitionCount                   uint64
	ObservedMonoNS, StateStartedMonoNS                uint64
	PressureStartedMonoNS, LastProgressMonoNS         uint64
	ObservedDeliveryBPS, CurrentOperationNS           uint64
	LastOperationNS, TotalPressuredNS, MaxPressuredNS uint64
}

// StreamLaneStats contains bounded queue, session, frame, byte, and drop counters.
type StreamLaneStats struct {
	Capacity, QueuedSubscribers, PreparedPublishers                        uint64
	ActivePublishers, ActiveSubscribers, RetainedRecords                   uint64
	SubmittedSubscribers, PreparedPublisherCount                           uint64
	EndedPublishers, EndedSubscribers, FailedPublishers, FailedSubscribers uint64
	CanceledSessions, PublishedFrames, PublishedBytes                      uint64
	ConsumedFrames, ConsumedBytes, SourceReportedDrops                     uint64
}

type streamCallbackKey struct {
	id        string
	direction StreamDirection
}

type streamSourceHolder struct{ source StreamSource }
type streamConsumerHolder struct{ consumer StreamConsumer }

// StreamLane owns one independent stream transport lane backed by the CoAkka core-runtime.
// Callback handles remain rooted until Forget succeeds or Close joins all workers.
type StreamLane struct {
	bindings  *nativeBindings
	lane      nativeStreamLane
	mu        sync.Mutex
	drained   *sync.Cond
	closing   bool
	active    int
	callbacks map[streamCallbackKey]cgo.Handle
}

type nativeStreamLane unsafe.Pointer

//export coakka_v2_go_stream_source_next
func coakka_v2_go_stream_source_next(context C.uintptr_t, destination *C.uint8_t, capacity C.size_t, out *C.coakka_v2_stream_frame_t) (status C.int) {
	status = -5
	defer func() { _ = recover() }()
	if context == 0 || destination == nil || out == nil || capacity == 0 {
		return -1
	}
	holder, ok := cgo.Handle(context).Value().(*streamSourceHolder)
	if !ok || holder.source == nil {
		return -5
	}
	result := holder.source(unsafe.Slice((*byte)(unsafe.Pointer(destination)), int(capacity)))
	switch result.Status {
	case StreamSourceWouldBlock:
		return -6
	case StreamSourceEnd:
		return -7
	case StreamSourceFrame:
		if result.Size <= 0 || result.Size > int(capacity) {
			return -1
		}
		out.captured_mono_ns = C.uint64_t(result.CapturedMonoNS)
		out.dropped_before = C.uint64_t(result.DroppedBefore)
		out.flags = C.uint32_t(result.Flags)
		out.size = C.size_t(result.Size)
		return 0
	default:
		return -1
	}
}

//export coakka_v2_go_stream_consume
func coakka_v2_go_stream_consume(context C.uintptr_t, data *C.uint8_t, frame *C.coakka_v2_stream_frame_t) (status C.int) {
	status = -5
	defer func() { _ = recover() }()
	if context == 0 || data == nil || frame == nil || frame.size == 0 {
		return -1
	}
	holder, ok := cgo.Handle(context).Value().(*streamConsumerHolder)
	if !ok || holder.consumer == nil {
		return -5
	}
	decision := holder.consumer(unsafe.Slice((*byte)(unsafe.Pointer(data)), int(frame.size)), StreamFrameMetadata{
		Sequence: uint64(frame.sequence), CapturedMonoNS: uint64(frame.captured_mono_ns),
		DroppedBefore: uint64(frame.dropped_before), Flags: uint32(frame.flags),
	})
	if decision == StreamConsumerStop {
		return -7
	}
	if decision != StreamConsumerContinue {
		return -1
	}
	return 0
}

// OpenStreamLane resolves, creates, and starts a Stream Lane.
// runtimeLibPath selects an explicit core-runtime; an empty value uses package resolution.
func OpenStreamLane(config StreamLaneConfig, runtimeLibPath string) (*StreamLane, error) {
	if config.Flags == 0 {
		config.Flags = StreamLanePublisher | StreamLaneSubscriber
	}
	if config.BindHost == "" {
		config.BindHost = "127.0.0.1"
	}
	if config.Flags&^(StreamLanePublisher|StreamLaneSubscriber) != 0 || config.Capacity > 64 ||
		config.MaxFrameBytes > 4*1024*1024 || config.MaxWindowBytes > 16*1024*1024 ||
		config.SourceRetryMillis > 1000 || config.PublisherWorkerCount > 4 || config.SubscriberWorkerCount > 4 ||
		config.PressureAfterMillis > 60000 || config.StalledAfterMillis > 60000 ||
		config.RecoveryAfterMillis > 60000 || config.PressureObservationMillis > 60000 ||
		(config.MaxFrameBytes != 0 && config.MaxWindowBytes != 0 && config.MaxWindowBytes < config.MaxFrameBytes) {
		return nil, errors.New("invalid stream-lane bounds, flags, or worker count")
	}
	path, err := (RuntimeLibraryResolver{}).Resolve(runtimeLibPath, nil)
	if err != nil {
		return nil, err
	}
	bindings, err := openNativeBindings(path)
	if err != nil {
		return nil, err
	}
	if C.coakka_v2_go_stream_lane_available(bindings.ptr) == 0 {
		bindings.close()
		return nil, errors.New("native runtime does not export the complete stream-lane ABI")
	}
	lane, err := bindings.createStreamLane(config)
	if err != nil {
		bindings.close()
		return nil, err
	}
	if err := requireStatus(C.coakka_v2_go_stream_lane_start(bindings.ptr, (*C.coakka_v2_stream_lane_t)(lane)), "stream_lane_start"); err != nil {
		C.coakka_v2_go_stream_lane_destroy(bindings.ptr, (*C.coakka_v2_stream_lane_t)(lane))
		bindings.close()
		return nil, err
	}
	result := &StreamLane{bindings: bindings, lane: lane, callbacks: make(map[streamCallbackKey]cgo.Handle)}
	result.drained = sync.NewCond(&result.mu)
	return result, nil
}

func (b *nativeBindings) createStreamLane(config StreamLaneConfig) (nativeStreamLane, error) {
	bind := C.CString(config.BindHost)
	defer C.free(unsafe.Pointer(bind))
	var security *C.coakka_v2_stream_lane_security_config_t
	var nativeSecurity C.coakka_v2_stream_lane_security_config_t
	var values []*C.char
	if config.Security != nil {
		for _, value := range []string{config.Security.CredentialID, config.Security.CACertificateFile, config.Security.IdentityCertificateFile, config.Security.PrivateKeyFile} {
			values = append(values, C.CString(value))
		}
		defer func() {
			for _, value := range values {
				C.free(unsafe.Pointer(value))
			}
		}()
		nativeSecurity = C.coakka_v2_stream_lane_security_config_t{
			struct_size: C.size_t(C.sizeof_coakka_v2_stream_lane_security_config_t), mode: C.uint32_t(config.Security.Mode),
			credential_generation: C.uint64_t(config.Security.CredentialGeneration), credential_id: values[0],
			ca_certificate_file: values[1], identity_certificate_file: values[2], private_key_file: values[3],
		}
		security = &nativeSecurity
	}
	native := C.coakka_v2_stream_lane_config_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_stream_lane_config_t), flags: C.uint32_t(config.Flags), bind_host: bind,
		bind_port: C.uint16_t(config.BindPort), capacity: C.size_t(config.Capacity), max_frame_bytes: C.uint32_t(config.MaxFrameBytes),
		max_window_bytes: C.uint32_t(config.MaxWindowBytes), io_timeout_ms: C.uint32_t(config.IOTimeoutMillis),
		source_retry_ms: C.uint32_t(config.SourceRetryMillis), progress_frames: C.uint32_t(config.ProgressFrames),
		progress_interval_ms: C.uint32_t(config.ProgressIntervalMillis), publisher_worker_count: C.uint32_t(config.PublisherWorkerCount),
		subscriber_worker_count: C.uint32_t(config.SubscriberWorkerCount), security: security,
		pressure_after_ms: C.uint32_t(config.PressureAfterMillis), stalled_after_ms: C.uint32_t(config.StalledAfterMillis),
		recovery_after_ms: C.uint32_t(config.RecoveryAfterMillis), pressure_observation_ms: C.uint32_t(config.PressureObservationMillis),
	}
	var lane *C.coakka_v2_stream_lane_t
	if err := requireStatus(C.coakka_v2_go_stream_lane_create_ex(b.ptr, &native, &lane), "stream_lane_create"); err != nil {
		return nil, err
	}
	return nativeStreamLane(unsafe.Pointer(lane)), nil
}

func (l *StreamLane) acquire() (nativeStreamLane, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closing || l.lane == nil {
		return nil, errors.New("stream lane is closed")
	}
	l.active++
	return l.lane, nil
}

func (l *StreamLane) release() {
	l.mu.Lock()
	l.active--
	if l.active == 0 {
		l.drained.Broadcast()
	}
	l.mu.Unlock()
}

// BoundPort returns the publisher port selected when the lane started.
func (l *StreamLane) BoundPort() (uint16, error) {
	lane, err := l.acquire()
	if err != nil {
		return 0, err
	}
	defer l.release()
	var port C.uint16_t
	err = requireStatus(C.coakka_v2_go_stream_lane_get_bound_port(l.bindings.ptr, (*C.coakka_v2_stream_lane_t)(lane), &port), "stream_lane_get_bound_port")
	return uint16(port), err
}

// PreparePublish authorizes one session and roots source until Forget or Close.
func (l *StreamLane) PreparePublish(spec StreamPublishSpec) error {
	if err := validateStreamSpec(spec.SessionID, spec.AuthorizationToken, spec.MaxFrameBytes); err != nil {
		return err
	}
	if spec.Source == nil {
		return errors.New("stream source is required")
	}
	lane, err := l.acquire()
	if err != nil {
		return err
	}
	defer l.release()
	handle := cgo.NewHandle(&streamSourceHolder{source: spec.Source})
	id, token := C.CString(spec.SessionID), C.CString(spec.AuthorizationToken)
	defer C.free(unsafe.Pointer(id))
	defer C.free(unsafe.Pointer(token))
	err = requireStatus(C.coakka_v2_go_stream_lane_prepare_publish(l.bindings.ptr, (*C.coakka_v2_stream_lane_t)(lane), id, token, C.uint64_t(spec.FormatID), C.uint32_t(spec.MaxFrameBytes), C.uintptr_t(handle)), "stream_lane_prepare_publish")
	if err != nil {
		handle.Delete()
		return err
	}
	l.mu.Lock()
	l.callbacks[streamCallbackKey{spec.SessionID, StreamPublish}] = handle
	l.mu.Unlock()
	return nil
}

// Subscribe queues a subscriber and roots consumer until Forget or Close.
func (l *StreamLane) Subscribe(spec StreamSubscribeSpec) error {
	if err := validateStreamSpec(spec.SessionID, spec.AuthorizationToken, spec.MaxFrameBytes); err != nil {
		return err
	}
	if spec.RemoteHost == "" || spec.RemotePort == 0 {
		return errors.New("remote host and port are required")
	}
	if spec.Consumer == nil {
		return errors.New("stream consumer is required")
	}
	lane, err := l.acquire()
	if err != nil {
		return err
	}
	defer l.release()
	handle := cgo.NewHandle(&streamConsumerHolder{consumer: spec.Consumer})
	id, token, host := C.CString(spec.SessionID), C.CString(spec.AuthorizationToken), C.CString(spec.RemoteHost)
	defer C.free(unsafe.Pointer(id))
	defer C.free(unsafe.Pointer(token))
	defer C.free(unsafe.Pointer(host))
	err = requireStatus(C.coakka_v2_go_stream_lane_subscribe(l.bindings.ptr, (*C.coakka_v2_stream_lane_t)(lane), id, token, host, C.uint16_t(spec.RemotePort), C.uint64_t(spec.FormatID), C.uint32_t(spec.MaxFrameBytes), C.uint32_t(spec.InitialWindowBytes), C.uint32_t(spec.TimeoutMillis), C.uintptr_t(handle)), "stream_lane_subscribe")
	if err != nil {
		handle.Delete()
		return err
	}
	l.mu.Lock()
	l.callbacks[streamCallbackKey{spec.SessionID, StreamSubscribe}] = handle
	l.mu.Unlock()
	return nil
}

func validateStreamSpec(sessionID, token string, maxFrame uint32) error {
	if sessionID == "" || len(sessionID) > 64 || token == "" || len(token) > 128 {
		return errors.New("session ID or authorization token is empty or exceeds its byte limit")
	}
	if maxFrame == 0 || maxFrame > 4*1024*1024 {
		return errors.New("max frame bytes must be in [1, 4194304]")
	}
	return nil
}

// Session returns the current copied session snapshot without waiting.
func (l *StreamLane) Session(sessionID string, direction StreamDirection) (StreamSessionSnapshot, error) {
	return l.session(sessionID, direction, 0, 0, false)
}

// WaitSession waits for a session update newer than afterSequence or until timeoutMillis expires.
func (l *StreamLane) WaitSession(sessionID string, direction StreamDirection, afterSequence uint64, timeoutMillis uint32) (StreamSessionSnapshot, error) {
	return l.session(sessionID, direction, afterSequence, timeoutMillis, true)
}

func (l *StreamLane) session(sessionID string, direction StreamDirection, sequence uint64, timeout uint32, wait bool) (StreamSessionSnapshot, error) {
	lane, err := l.acquire()
	if err != nil {
		return StreamSessionSnapshot{}, err
	}
	defer l.release()
	cID := C.CString(sessionID)
	defer C.free(unsafe.Pointer(cID))
	native := C.coakka_v2_stream_session_snapshot_t{struct_size: C.size_t(C.sizeof_coakka_v2_stream_session_snapshot_t)}
	var status C.int
	if wait {
		status = C.coakka_v2_go_stream_lane_wait_session(l.bindings.ptr, (*C.coakka_v2_stream_lane_t)(lane), cID, C.uint32_t(direction), C.uint64_t(sequence), C.uint32_t(timeout), &native)
	} else {
		status = C.coakka_v2_go_stream_lane_get_session(l.bindings.ptr, (*C.coakka_v2_stream_lane_t)(lane), cID, C.uint32_t(direction), &native)
	}
	if err := requireStatus(status, "stream_lane_session"); err != nil {
		return StreamSessionSnapshot{}, err
	}
	return StreamSessionSnapshot{Direction: StreamDirection(native.direction), State: StreamState(native.state), Result: StreamResult(native.result), FormatID: uint64(native.format_id), Frames: uint64(native.frames), Bytes: uint64(native.bytes), DroppedFrames: uint64(native.dropped_frames), LastSequence: uint64(native.last_sequence), NegotiatedMaxFrameBytes: uint32(native.negotiated_max_frame_bytes), WindowBytes: uint32(native.window_bytes), CancelRequested: native.cancel_requested != 0, UpdateSequence: uint64(native.update_sequence), SubmittedMonoNS: uint64(native.submitted_mono_ns), StartedMonoNS: uint64(native.started_mono_ns), UpdatedMonoNS: uint64(native.updated_mono_ns), TerminalMonoNS: uint64(native.terminal_mono_ns), Detail: C.GoString(&native.detail[0])}, nil
}

// Pressure returns the current policy-neutral pressure snapshot without waiting.
func (l *StreamLane) Pressure(sessionID string, direction StreamDirection) (StreamPressureSnapshot, error) {
	return l.pressure(sessionID, direction, 0, 0, false)
}

// WaitPressure waits for a pressure update newer than afterSequence or until timeoutMillis expires.
func (l *StreamLane) WaitPressure(sessionID string, direction StreamDirection, afterSequence uint64, timeoutMillis uint32) (StreamPressureSnapshot, error) {
	return l.pressure(sessionID, direction, afterSequence, timeoutMillis, true)
}

func (l *StreamLane) pressure(sessionID string, direction StreamDirection, sequence uint64, timeout uint32, wait bool) (StreamPressureSnapshot, error) {
	lane, err := l.acquire()
	if err != nil {
		return StreamPressureSnapshot{}, err
	}
	defer l.release()
	cID := C.CString(sessionID)
	defer C.free(unsafe.Pointer(cID))
	n := C.coakka_v2_stream_pressure_snapshot_t{struct_size: C.size_t(C.sizeof_coakka_v2_stream_pressure_snapshot_t)}
	var status C.int
	if wait {
		status = C.coakka_v2_go_stream_lane_wait_pressure(l.bindings.ptr, (*C.coakka_v2_stream_lane_t)(lane), cID, C.uint32_t(direction), C.uint64_t(sequence), C.uint32_t(timeout), &n)
	} else {
		status = C.coakka_v2_go_stream_lane_get_pressure(l.bindings.ptr, (*C.coakka_v2_stream_lane_t)(lane), cID, C.uint32_t(direction), &n)
	}
	if err := requireStatus(status, "stream_lane_pressure"); err != nil {
		return StreamPressureSnapshot{}, err
	}
	return StreamPressureSnapshot{Direction: StreamDirection(n.direction), State: StreamPressureState(n.state), ReasonBits: uint32(n.reason_bits), AvailableCreditBytes: uint32(n.available_credit_bytes), WindowCapacityBytes: uint32(n.window_capacity_bytes), UpdateSequence: uint64(n.update_sequence), TransitionCount: uint64(n.transition_count), ObservedMonoNS: uint64(n.observed_mono_ns), StateStartedMonoNS: uint64(n.state_started_mono_ns), PressureStartedMonoNS: uint64(n.pressure_started_mono_ns), LastProgressMonoNS: uint64(n.last_progress_mono_ns), ObservedDeliveryBPS: uint64(n.observed_delivery_bps), CurrentOperationNS: uint64(n.current_operation_ns), LastOperationNS: uint64(n.last_operation_ns), TotalPressuredNS: uint64(n.total_pressured_ns), MaxPressuredNS: uint64(n.max_pressured_ns)}, nil
}

// Cancel requests cooperative cancellation; observe terminal state before Forget.
func (l *StreamLane) Cancel(sessionID string, direction StreamDirection) error {
	return l.control(sessionID, direction, false)
}

// Forget releases a terminal record and its callback handle.
func (l *StreamLane) Forget(sessionID string, direction StreamDirection) error {
	return l.control(sessionID, direction, true)
}

func (l *StreamLane) control(sessionID string, direction StreamDirection, forget bool) error {
	lane, err := l.acquire()
	if err != nil {
		return err
	}
	defer l.release()
	cID := C.CString(sessionID)
	defer C.free(unsafe.Pointer(cID))
	var status C.int
	if forget {
		status = C.coakka_v2_go_stream_lane_forget_session(l.bindings.ptr, (*C.coakka_v2_stream_lane_t)(lane), cID, C.uint32_t(direction))
	} else {
		status = C.coakka_v2_go_stream_lane_cancel_session(l.bindings.ptr, (*C.coakka_v2_stream_lane_t)(lane), cID, C.uint32_t(direction))
	}
	if err := requireStatus(status, "stream_lane_session_control"); err != nil {
		return err
	}
	if forget {
		l.mu.Lock()
		key := streamCallbackKey{sessionID, direction}
		if handle, ok := l.callbacks[key]; ok {
			handle.Delete()
			delete(l.callbacks, key)
		}
		l.mu.Unlock()
	}
	return nil
}

// Stats returns a copied lane-level observability snapshot.
func (l *StreamLane) Stats() (StreamLaneStats, error) {
	lane, err := l.acquire()
	if err != nil {
		return StreamLaneStats{}, err
	}
	defer l.release()
	n := C.coakka_v2_stream_lane_stats_t{struct_size: C.size_t(C.sizeof_coakka_v2_stream_lane_stats_t)}
	if err := requireStatus(C.coakka_v2_go_stream_lane_get_stats(l.bindings.ptr, (*C.coakka_v2_stream_lane_t)(lane), &n), "stream_lane_get_stats"); err != nil {
		return StreamLaneStats{}, err
	}
	return StreamLaneStats{uint64(n.capacity), uint64(n.queued_subscribers), uint64(n.prepared_publishers), uint64(n.active_publishers), uint64(n.active_subscribers), uint64(n.retained_records), uint64(n.submitted_subscribers), uint64(n.prepared_publisher_count), uint64(n.ended_publishers), uint64(n.ended_subscribers), uint64(n.failed_publishers), uint64(n.failed_subscribers), uint64(n.canceled_sessions), uint64(n.published_frames), uint64(n.published_bytes), uint64(n.consumed_frames), uint64(n.consumed_bytes), uint64(n.source_reported_drops)}, nil
}

// Close stops the lane, joins native workers, drains host calls, and releases callback handles.
func (l *StreamLane) Close() error {
	l.mu.Lock()
	if l.lane == nil {
		l.mu.Unlock()
		return nil
	}
	if l.closing {
		for l.lane != nil {
			l.drained.Wait()
		}
		l.mu.Unlock()
		return nil
	}
	l.closing = true
	lane := l.lane
	l.mu.Unlock()
	status := C.coakka_v2_go_stream_lane_stop(l.bindings.ptr, (*C.coakka_v2_stream_lane_t)(lane))
	l.mu.Lock()
	for l.active != 0 {
		l.drained.Wait()
	}
	C.coakka_v2_go_stream_lane_destroy(l.bindings.ptr, (*C.coakka_v2_stream_lane_t)(lane))
	for key, handle := range l.callbacks {
		handle.Delete()
		delete(l.callbacks, key)
	}
	l.bindings.close()
	l.lane = nil
	l.drained.Broadcast()
	l.mu.Unlock()
	if status != 0 && status != -7 {
		return fmt.Errorf("stream_lane_stop failed: native status %d", int(status))
	}
	return nil
}
