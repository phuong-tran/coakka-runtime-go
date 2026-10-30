package coakka_v2_connector

/*
#cgo CFLAGS: -std=c11 -Wall -Werror -I${SRCDIR}/internal/coreabi/include
#cgo linux LDFLAGS: -ldl
#include "runtime_bridge.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

const (
	COAKKAOK                                = 0
	COAKKAErrInvalidArg                     = -1
	COAKKAErrNoMem                          = -2
	COAKKAErrBadState                       = -3
	COAKKAErrIO                             = -5
	COAKKAErrWouldBlock                     = -6
	COAKKAErrClosed                         = -7
	COAKKAABIVersion                        = 1
	HostHandlesEnableMonitor         uint32 = C.COAKKA_V2_HOST_HANDLES_FLAG_ENABLE_MONITOR
	HostHandlesSeparateDeliveredLane uint32 = C.COAKKA_V2_HOST_HANDLES_FLAG_SEPARATE_DELIVERED_REQUEST_LANE
	// C.GoBytes uses C int, independently of Core's configured frame ceiling.
	maxNativeProjectionBytes      = 1<<31 - 1
	runtimeFeatureFileLane        = uint32(C.COAKKA_V2_RUNTIME_FEATURE_FILE_LANE)
	runtimeFeatureStreamLane      = uint32(C.COAKKA_V2_RUNTIME_FEATURE_STREAM_LANE)
	runtimeFeatureLaneOwnerGrants = uint32(C.COAKKA_V2_RUNTIME_FEATURE_LANE_OWNER_GRANTS)
)

type RuntimeInfoSnapshot struct {
	AbiVersion        uint32 `json:"abiVersion"`
	FeatureFlags      uint32 `json:"featureFlags"`
	RuntimeVersion    string `json:"runtimeVersion"`
	GitCommit         string `json:"gitCommit"`
	TransportProvider string `json:"transportProvider"`
	AllocatorProvider string `json:"allocatorProvider"`
	DocsHint          string `json:"docsHint"`
}

type RuntimeConfigSnapshot struct {
	SystemName                               string `json:"systemName"`
	NodeID                                   string `json:"nodeId"`
	RuntimeState                             int32  `json:"runtimeState"`
	SnapshotPresent                          bool   `json:"snapshotPresent"`
	AppliedGeneration                        uint64 `json:"appliedGeneration"`
	RouteCount                               uint64 `json:"routeCount"`
	TransportBindHost                        string `json:"transportBindHost"`
	TransportBindPort                        uint16 `json:"transportBindPort"`
	EffectiveIngressOverloadMode             uint32 `json:"effectiveIngressOverloadMode"`
	EffectiveLocalDeliveryOverloadMode       uint32 `json:"effectiveLocalDeliveryOverloadMode"`
	EffectiveRemoteOutboundOverloadMode      uint32 `json:"effectiveRemoteOutboundOverloadMode"`
	EffectiveRemoteOutboundReplyReserveSlots uint64 `json:"effectiveRemoteOutboundReplyReserveSlots"`
}

type RuntimeStartupCPUInfo struct {
	RequestedPolicy   uint32    `json:"requestedPolicy"`
	Supported         bool      `json:"supported"`
	AvailableCPUCount uint32    `json:"availableCpuCount"`
	SelectedCPUCount  uint32    `json:"selectedCpuCount"`
	SelectedCPUIds    [2]uint32 `json:"selectedCpuIds"`
	StartupApplied    bool      `json:"startupApplied"`
}

type RuntimeStartupIOInfo struct {
	Requested  bool  `json:"requested"`
	Compiled   bool  `json:"compiled"`
	Probed     bool  `json:"probed"`
	Supported  bool  `json:"supported"`
	Effective  bool  `json:"effective"`
	ProbeError int32 `json:"probeError"`
}

type RuntimeStartupMonitorInfo struct {
	Enabled bool `json:"enabled"`
}

type RuntimeStartupConfigSnapshot struct {
	DefaultsRevision    uint32                             `json:"defaultsRevision"`
	DefaultedGroups     uint32                             `json:"defaultedGroups"`
	StrictNoDrop        bool                               `json:"strictNoDrop"`
	QueueCapacity       uint32                             `json:"queueCapacity"`
	RequestMaxFrameSize uint64                             `json:"requestMaxFrameSize"`
	RuntimeState        uint32                             `json:"runtimeState"`
	Execution           RuntimeStartupCPUInfo              `json:"execution"`
	Connection          RuntimeTCPConnectionConfigSnapshot `json:"connection"`
	IO                  RuntimeStartupIOInfo               `json:"io"`
	Monitor             RuntimeStartupMonitorInfo          `json:"monitor"`
}

type RuntimeHealthSnapshot struct {
	RuntimeState      int32  `json:"runtimeState"`
	Flags             uint32 `json:"flags"`
	AppliedGeneration uint64 `json:"appliedGeneration"`
}

type RuntimeStatsSnapshot struct {
	AppliedGeneration                         uint64 `json:"appliedGeneration"`
	RouteCount                                uint64 `json:"routeCount"`
	RuntimeState                              int32  `json:"runtimeState"`
	QueueRejectedCount                        uint64 `json:"queueRejectedCount"`
	RouteMissCount                            uint64 `json:"routeMissCount"`
	DeadletterCount                           uint64 `json:"deadletterCount"`
	DeliveryFailedCount                       uint64 `json:"deliveryFailedCount"`
	ControlRejectedCount                      uint64 `json:"controlRejectedCount"`
	MonitorEventEmittedCount                  uint64 `json:"monitorEventEmittedCount"`
	MonitorEventDroppedCount                  uint64 `json:"monitorEventDroppedCount"`
	MonitorEventEmittedLifetimeCount          uint64 `json:"monitorEventEmittedLifetimeCount"`
	MonitorEventDroppedLifetimeCount          uint64 `json:"monitorEventDroppedLifetimeCount"`
	LocalWorkQueueCapacity                    uint64 `json:"localWorkQueueCapacity"`
	LocalWorkQueueDepth                       uint64 `json:"localWorkQueueDepth"`
	LocalWorkQueueHighWatermark               uint64 `json:"localWorkQueueHighWatermark"`
	DeliveredRequestOutboundDirectWriteCount  uint64 `json:"deliveredRequestOutboundDirectWriteCount"`
	ResponseOutboundDirectWriteCount          uint64 `json:"responseOutboundDirectWriteCount"`
	DeadletterOutboundDirectWriteCount        uint64 `json:"deadletterOutboundDirectWriteCount"`
	RemoteOutboundQueueCapacity               uint64 `json:"remoteOutboundQueueCapacity"`
	RemoteOutboundQueueDepth                  uint64 `json:"remoteOutboundQueueDepth"`
	RemoteOutboundQueueHighWatermark          uint64 `json:"remoteOutboundQueueHighWatermark"`
	RemoteOutboundQueueRejectedCount          uint64 `json:"remoteOutboundQueueRejectedCount"`
	RemoteOutboundExpiredDropCount            uint64 `json:"remoteOutboundExpiredDropCount"`
	RemoteOutboundReplyReserveSlots           uint64 `json:"remoteOutboundReplyReserveSlots"`
	RemoteOutboundReplyReservationRejectCount uint64 `json:"remoteOutboundReplyReservationRejectCount"`
	RemoteOutboundOneWayDropCount             uint64 `json:"remoteOutboundOneWayDropCount"`
}

type HostHandlesSnapshot struct {
	Flags                  uint32
	RequestWriteFD         int
	ResponseReadFD         int
	DeadletterReadFD       int
	ControlWriteFD         int
	MonitorReadFD          int
	DeliveredRequestReadFD int
}

type nativeBindings struct {
	ptr *C.coakka_v2_go_bindings_t
}

type nativeRuntime unsafe.Pointer
type nativeFileLane unsafe.Pointer
type nativeRequestReader unsafe.Pointer
type nativeAskClient unsafe.Pointer
type nativeAskTicket unsafe.Pointer

const (
	nativeClientResultResponse   = 1
	nativeClientResultDeadletter = 2
)

type nativeDeliveredRequest struct {
	Frame         []byte
	Target        []byte
	TargetCString []byte
	Payload       []byte
	OneWay        bool
}

type nativeCString []byte

func makeNativeCString(name string, value string) (nativeCString, error) {
	if err := validateNativeText([2]string{name, value}); err != nil {
		return nil, err
	}
	encoded := make(nativeCString, len(value)+1)
	copy(encoded, value)
	return encoded, nil
}

var nativeLibraryMu sync.Mutex

func validateNativeLibraryPath(path string) error {
	if err := validateNativeText([2]string{"runtime library path", path}); err != nil {
		return err
	}
	// This is the connector-owned loader representation, not Core policy.
	// Refuse before C.CString allocates; the C owner also checks the wide size.
	if runtime.GOOS == "windows" && len(path) > C.COAKKA_GO_WINDOWS_LIBRARY_PATH_MAX_BYTES {
		return fmt.Errorf("runtime library path exceeds Windows loader representation bound")
	}
	return nil
}

func openNativeBindings(path string) (*nativeBindings, error) {
	if err := validateNativeLibraryPath(path); err != nil {
		return nil, err
	}
	nativeLibraryMu.Lock()
	defer nativeLibraryMu.Unlock()
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	var cError *C.char
	ptr := C.coakka_v2_go_open_library(cPath, &cError)
	if cError != nil {
		defer C.free(unsafe.Pointer(cError))
	}
	if ptr == nil {
		if cError != nil {
			return nil, fmt.Errorf("load runtime library: %s", C.GoString(cError))
		}
		return nil, fmt.Errorf("load runtime library: unknown error")
	}
	return &nativeBindings{ptr: ptr}, nil
}

func (b *nativeBindings) close() {
	if b == nil || b.ptr == nil {
		return
	}
	C.coakka_v2_go_close_library(b.ptr)
	b.ptr = nil
}

func (b *nativeBindings) requireCoreFeature(feature uint32, name string) error {
	if b == nil || b.ptr == nil {
		return &CoreContractError{Operation: "runtime_get_info", Detail: "binding is closed"}
	}
	var supported C.uint32_t
	if err := requireStatus(C.coakka_v2_go_query_feature(b.ptr, C.uint32_t(feature), &supported), "runtime_get_info"); err != nil {
		return err
	}
	if supported == 0 {
		return &FeatureUnavailableError{Feature: name}
	}
	if supported != 1 {
		return &CoreContractError{Operation: "runtime_get_info", Detail: "invalid feature boolean"}
	}
	return nil
}

func (b *nativeBindings) loadFileOwnerGrants() error { return requireLaneOwnerGrants(b) }

func (b *nativeBindings) createFileLane(config FileLaneConfig) (nativeFileLane, error) {
	if err := config.validateRepresentation(); err != nil {
		return nil, err
	}
	bindHost := C.CString(config.BindHost)
	defer C.free(unsafe.Pointer(bindHost))
	var security *C.coakka_v2_file_lane_security_config_t
	var securityValue C.coakka_v2_file_lane_security_config_t
	var securityStrings []*C.char
	if config.Security != nil {
		values := []string{config.Security.CredentialID, config.Security.CACertificateFile, config.Security.IdentityCertificateFile, config.Security.PrivateKeyFile}
		for _, value := range values {
			securityStrings = append(securityStrings, C.CString(value))
		}
		defer func() {
			for _, value := range securityStrings {
				C.free(unsafe.Pointer(value))
			}
		}()
		securityValue = C.coakka_v2_file_lane_security_config_t{
			struct_size: C.size_t(C.sizeof_coakka_v2_file_lane_security_config_t), mode: C.uint32_t(config.Security.Mode),
			credential_generation: C.uint64_t(config.Security.CredentialGeneration), credential_id: securityStrings[0],
			ca_certificate_file: securityStrings[1], identity_certificate_file: securityStrings[2], private_key_file: securityStrings[3],
		}
		security = &securityValue
	}
	native := C.coakka_v2_file_lane_config_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_file_lane_config_t), flags: C.uint32_t(config.Flags), bind_host: bindHost,
		bind_port: C.uint16_t(config.BindPort), queue_capacity: C.size_t(config.QueueCapacity), max_file_size: C.uint64_t(config.MaxFileSize),
		io_timeout_ms: C.uint32_t(config.IOTimeoutMillis), checkpoint_bytes: C.uint64_t(config.CheckpointBytes),
		progress_bytes: C.uint64_t(config.ProgressBytes), progress_interval_ms: C.uint32_t(config.ProgressIntervalMillis),
		sender_worker_count: C.uint32_t(config.SenderWorkerCount), receiver_worker_count: C.uint32_t(config.ReceiverWorkerCount), security: security,
	}
	var lane *C.coakka_v2_file_lane_t
	if err := requireStatus(C.coakka_v2_go_file_lane_create_ex(b.ptr, &native, &lane), "file_lane_create"); err != nil {
		return nil, err
	}
	if lane == nil {
		return nil, &CoreContractError{Operation: "file_lane_create", Detail: "success without owner"}
	}
	return nativeFileLane(unsafe.Pointer(lane)), nil
}

func (b *nativeBindings) createOwnedFileLane(config FileLaneConfig, owner LaneOwnerConfig) (nativeFileLane, error) {
	if err := config.validateRepresentation(); err != nil {
		return nil, err
	}
	if err := validateNativeText([2]string{"owner.OwnerInstanceID", owner.OwnerInstanceID}, [2]string{"owner.AdvertisedHost", owner.AdvertisedHost}); err != nil {
		return nil, err
	}
	bindHost, ownerID, advertisedHost := C.CString(config.BindHost), C.CString(owner.OwnerInstanceID), C.CString(owner.AdvertisedHost)
	defer C.free(unsafe.Pointer(bindHost))
	defer C.free(unsafe.Pointer(ownerID))
	defer C.free(unsafe.Pointer(advertisedHost))
	var security *C.coakka_v2_file_lane_security_config_t
	var nativeSecurity C.coakka_v2_file_lane_security_config_t
	var securityStrings []*C.char
	if config.Security != nil {
		for _, value := range []string{config.Security.CredentialID, config.Security.CACertificateFile, config.Security.IdentityCertificateFile, config.Security.PrivateKeyFile} {
			securityStrings = append(securityStrings, C.CString(value))
		}
		defer func() {
			for _, value := range securityStrings {
				C.free(unsafe.Pointer(value))
			}
		}()
		nativeSecurity = C.coakka_v2_file_lane_security_config_t{
			struct_size: C.size_t(C.sizeof_coakka_v2_file_lane_security_config_t), mode: C.uint32_t(config.Security.Mode),
			credential_generation: C.uint64_t(config.Security.CredentialGeneration), credential_id: securityStrings[0],
			ca_certificate_file: securityStrings[1], identity_certificate_file: securityStrings[2], private_key_file: securityStrings[3],
		}
		security = &nativeSecurity
	}
	nativeLane := C.coakka_v2_file_lane_config_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_file_lane_config_t), flags: C.uint32_t(config.Flags), bind_host: bindHost,
		bind_port: C.uint16_t(config.BindPort), queue_capacity: C.size_t(config.QueueCapacity), max_file_size: C.uint64_t(config.MaxFileSize),
		io_timeout_ms: C.uint32_t(config.IOTimeoutMillis), checkpoint_bytes: C.uint64_t(config.CheckpointBytes),
		progress_bytes: C.uint64_t(config.ProgressBytes), progress_interval_ms: C.uint32_t(config.ProgressIntervalMillis),
		sender_worker_count: C.uint32_t(config.SenderWorkerCount), receiver_worker_count: C.uint32_t(config.ReceiverWorkerCount), security: security,
	}
	nativeOwner := C.coakka_v2_lane_owner_config_t{struct_size: C.size_t(C.sizeof_coakka_v2_lane_owner_config_t), owner_instance_id: ownerID, advertised_host: advertisedHost}
	owned := C.coakka_v2_file_lane_owned_config_t{struct_size: C.size_t(C.sizeof_coakka_v2_file_lane_owned_config_t), lane: nativeLane, owner: nativeOwner}
	var lane *C.coakka_v2_file_lane_t
	if err := requireStatus(C.coakka_v2_go_file_lane_create_owned_ex(b.ptr, &owned, &lane), "file_lane_create_owned"); err != nil {
		return nil, err
	}
	if lane == nil {
		return nil, &CoreContractError{Operation: "file_lane_create", Detail: "success without owner"}
	}
	return nativeFileLane(unsafe.Pointer(lane)), nil
}

func (b *nativeBindings) startFileLane(lane nativeFileLane) error {
	return requireStatus(C.coakka_v2_go_file_lane_start(b.ptr, (*C.coakka_v2_file_lane_t)(lane)), "file_lane_start")
}
func (b *nativeBindings) stopFileLane(lane nativeFileLane) int {
	return int(C.coakka_v2_go_file_lane_stop(b.ptr, (*C.coakka_v2_file_lane_t)(lane)))
}
func (b *nativeBindings) destroyFileLane(lane nativeFileLane) {
	C.coakka_v2_go_file_lane_destroy(b.ptr, (*C.coakka_v2_file_lane_t)(lane))
}
func (b *nativeBindings) fileLaneBoundPort(lane nativeFileLane) (uint16, error) {
	var port C.uint16_t
	err := requireStatus(C.coakka_v2_go_file_lane_get_bound_port(b.ptr, (*C.coakka_v2_file_lane_t)(lane), &port), "file_lane_get_bound_port")
	return uint16(port), err
}

func copyFileDigest(out *[32]C.uint8_t, digest [32]byte) {
	// Fixed-width language-boundary copy; no foreign call or interior-field
	// pointer needs to cross cgo just to populate the caller-owned C record.
	for index, value := range digest {
		out[index] = C.uint8_t(value)
	}
}

func (b *nativeBindings) prepareFileReceive(lane nativeFileLane, spec FileReceiveSpec) error {
	if err := validateNativeText([2]string{"spec.TransferID", spec.TransferID}, [2]string{"spec.AuthorizationToken", spec.AuthorizationToken}, [2]string{"spec.DestinationPath", spec.DestinationPath}); err != nil {
		return err
	}
	id, token, path := C.CString(spec.TransferID), C.CString(spec.AuthorizationToken), C.CString(spec.DestinationPath)
	defer C.free(unsafe.Pointer(id))
	defer C.free(unsafe.Pointer(token))
	defer C.free(unsafe.Pointer(path))
	native := C.coakka_v2_file_receive_spec_t{struct_size: C.size_t(C.sizeof_coakka_v2_file_receive_spec_t), transfer_id: id, authorization_token: token, destination_path: path, expected_size: C.uint64_t(spec.ExpectedSize)}
	copyFileDigest((*[32]C.uint8_t)(&native.expected_sha256), spec.ExpectedSHA256)
	return requireStatus(C.coakka_v2_go_file_lane_prepare_receive(b.ptr, (*C.coakka_v2_file_lane_t)(lane), &native), "file_lane_prepare_receive")
}

func (b *nativeBindings) prepareFileReceiveGrant(lane nativeFileLane, spec FileReceiveSpec) (FileReceiveGrant, error) {
	if err := validateNativeText([2]string{"spec.TransferID", spec.TransferID}, [2]string{"spec.AuthorizationToken", spec.AuthorizationToken}, [2]string{"spec.DestinationPath", spec.DestinationPath}); err != nil {
		return FileReceiveGrant{}, err
	}
	id, token, path := C.CString(spec.TransferID), C.CString(spec.AuthorizationToken), C.CString(spec.DestinationPath)
	defer C.free(unsafe.Pointer(id))
	defer C.free(unsafe.Pointer(token))
	defer C.free(unsafe.Pointer(path))
	nativeSpec := C.coakka_v2_file_receive_spec_t{struct_size: C.size_t(C.sizeof_coakka_v2_file_receive_spec_t), transfer_id: id, authorization_token: token, destination_path: path, expected_size: C.uint64_t(spec.ExpectedSize)}
	copyFileDigest((*[32]C.uint8_t)(&nativeSpec.expected_sha256), spec.ExpectedSHA256)
	grant := C.coakka_v2_file_receive_grant_t{struct_size: C.size_t(C.sizeof_coakka_v2_file_receive_grant_t)}
	if err := requireStatus(C.coakka_v2_go_file_lane_prepare_receive_grant(b.ptr, (*C.coakka_v2_file_lane_t)(lane), &nativeSpec, &grant), "file_lane_prepare_receive_grant"); err != nil {
		return FileReceiveGrant{}, err
	}
	var digest [32]byte
	C.memcpy(unsafe.Pointer(&digest[0]), unsafe.Pointer(&grant.expected_sha256[0]), 32)
	return FileReceiveGrant{
		Owner:      LaneOwnerEndpoint{OwnerInstanceID: C.GoString(&grant.owner.owner_instance_id[0]), AdvertisedHost: C.GoString(&grant.owner.advertised_host[0]), Port: uint16(grant.owner.port)},
		TransferID: C.GoString(&grant.transfer_id[0]), AuthorizationToken: C.GoString(&grant.authorization_token[0]), ExpectedSize: uint64(grant.expected_size), ExpectedSHA256: digest,
	}, nil
}

func (b *nativeBindings) submitFileSend(lane nativeFileLane, spec FileSendSpec) error {
	if err := validateNativeText([2]string{"spec.TransferID", spec.TransferID}, [2]string{"spec.AuthorizationToken", spec.AuthorizationToken}, [2]string{"spec.RemoteHost", spec.RemoteHost}, [2]string{"spec.SourcePath", spec.SourcePath}); err != nil {
		return err
	}
	id, token, host, path := C.CString(spec.TransferID), C.CString(spec.AuthorizationToken), C.CString(spec.RemoteHost), C.CString(spec.SourcePath)
	defer C.free(unsafe.Pointer(id))
	defer C.free(unsafe.Pointer(token))
	defer C.free(unsafe.Pointer(host))
	defer C.free(unsafe.Pointer(path))
	native := C.coakka_v2_file_send_spec_t{struct_size: C.size_t(C.sizeof_coakka_v2_file_send_spec_t), transfer_id: id, authorization_token: token, remote_host: host, remote_port: C.uint16_t(spec.RemotePort), source_path: path, expected_size: C.uint64_t(spec.ExpectedSize), timeout_ms: C.uint32_t(spec.TimeoutMillis)}
	copyFileDigest((*[32]C.uint8_t)(&native.expected_sha256), spec.ExpectedSHA256)
	return requireStatus(C.coakka_v2_go_file_lane_submit_send(b.ptr, (*C.coakka_v2_file_lane_t)(lane), &native), "file_lane_submit_send")
}

func (b *nativeBindings) fileTransfer(lane nativeFileLane, transferID string, direction FileTransferDirection, sequence uint64, timeoutMillis uint32, wait bool) (FileTransferSnapshot, error) {
	if err := validateNativeText([2]string{"transferID", transferID}); err != nil {
		return FileTransferSnapshot{}, err
	}
	id := C.CString(transferID)
	defer C.free(unsafe.Pointer(id))
	native := C.coakka_v2_file_transfer_snapshot_t{struct_size: C.size_t(C.sizeof_coakka_v2_file_transfer_snapshot_t)}
	var status C.int
	if wait {
		status = C.coakka_v2_go_file_lane_wait_transfer(b.ptr, (*C.coakka_v2_file_lane_t)(lane), id, C.uint32_t(direction), C.uint64_t(sequence), C.uint32_t(timeoutMillis), &native)
	} else {
		status = C.coakka_v2_go_file_lane_get_transfer(b.ptr, (*C.coakka_v2_file_lane_t)(lane), id, C.uint32_t(direction), &native)
	}
	if err := requireStatus(status, "file_lane_transfer"); err != nil {
		return FileTransferSnapshot{}, err
	}
	return FileTransferSnapshot{Direction: FileTransferDirection(native.direction), State: FileTransferState(native.state), Result: FileTransferResult(native.result), ExpectedSize: uint64(native.expected_size), TransferredBytes: uint64(native.transferred_bytes), CommittedOffset: uint64(native.committed_offset), ProgressMilli: uint32(native.progress_milli), CancelRequested: native.cancel_requested != 0, UpdateSequence: uint64(native.update_sequence), SubmittedMonoNS: uint64(native.submitted_mono_ns), StartedMonoNS: uint64(native.started_mono_ns), UpdatedMonoNS: uint64(native.updated_mono_ns), TerminalMonoNS: uint64(native.terminal_mono_ns), Detail: C.GoString(&native.detail[0])}, nil
}

func (b *nativeBindings) cancelFileTransfer(lane nativeFileLane, transferID string, direction FileTransferDirection, forget bool) error {
	if err := validateNativeText([2]string{"transferID", transferID}); err != nil {
		return err
	}
	id := C.CString(transferID)
	defer C.free(unsafe.Pointer(id))
	var status C.int
	if forget {
		status = C.coakka_v2_go_file_lane_forget_transfer(b.ptr, (*C.coakka_v2_file_lane_t)(lane), id, C.uint32_t(direction))
	} else {
		status = C.coakka_v2_go_file_lane_cancel_transfer(b.ptr, (*C.coakka_v2_file_lane_t)(lane), id, C.uint32_t(direction))
	}
	return requireStatus(status, "file_lane_transfer_control")
}
func (b *nativeBindings) fileLaneStats(lane nativeFileLane) (FileLaneStats, error) {
	native := C.coakka_v2_file_lane_stats_t{struct_size: C.size_t(C.sizeof_coakka_v2_file_lane_stats_t)}
	if err := requireStatus(C.coakka_v2_go_file_lane_get_stats(b.ptr, (*C.coakka_v2_file_lane_t)(lane), &native), "file_lane_get_stats"); err != nil {
		return FileLaneStats{}, err
	}
	return FileLaneStats{uint64(native.queue_capacity), uint64(native.queued_sends), uint64(native.prepared_receives), uint64(native.active_sends), uint64(native.active_receives), uint64(native.retained_records), uint64(native.submitted_sends), uint64(native.prepared_receive_count), uint64(native.completed_sends), uint64(native.completed_receives), uint64(native.failed_sends), uint64(native.failed_receives), uint64(native.canceled_transfers), uint64(native.completed_send_bytes), uint64(native.completed_receive_bytes)}, nil
}
func (b *nativeBindings) fileSHA256(path string) (FileDigest, error) {
	if err := validateNativeText([2]string{"path", path}); err != nil {
		return FileDigest{}, err
	}
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	var digest [32]C.uint8_t
	var size C.uint64_t
	if err := requireStatus(C.coakka_v2_go_file_sha256_path(b.ptr, cPath, &digest[0], &size), "file_sha256_path"); err != nil {
		return FileDigest{}, err
	}
	var out [32]byte
	C.memcpy(unsafe.Pointer(&out[0]), unsafe.Pointer(&digest[0]), 32)
	return FileDigest{SHA256: out, Size: uint64(size)}, nil
}

func (b *nativeBindings) getAbiVersion() uint32 {
	return uint32(C.coakka_v2_go_get_abi_version(b.ptr))
}

func (b *nativeBindings) readRuntimeInfo() (RuntimeInfoSnapshot, error) {
	info := C.coakka_v2_runtime_info_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_runtime_info_t),
	}
	if err := requireStatus(C.coakka_v2_go_get_info(b.ptr, &info), "runtime_get_info"); err != nil {
		return RuntimeInfoSnapshot{}, err
	}
	return RuntimeInfoSnapshot{
		AbiVersion:        uint32(info.abi_version),
		FeatureFlags:      uint32(info.feature_flags),
		RuntimeVersion:    cString(info.runtime_version),
		GitCommit:         cString(info.git_commit),
		TransportProvider: cString(info.southbound_backend),
		AllocatorProvider: cString(info.allocator_backend),
		DocsHint:          cString(info.docs_hint),
	}, nil
}

func (b *nativeBindings) readRuntimeCapabilities() (RuntimeCapabilitiesSnapshot, error) {
	native := C.coakka_v2_runtime_capability_snapshot_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_runtime_capability_snapshot_t),
	}
	if err := requireStatus(
		C.coakka_v2_go_get_capabilities(b.ptr, &native),
		"runtime_get_capabilities",
	); err != nil {
		return RuntimeCapabilitiesSnapshot{}, err
	}
	return RuntimeCapabilitiesSnapshot{
		Edition:                       uint32(native.edition),
		LicenseStatus:                 uint32(native.license_status),
		CompiledCapabilities:          RuntimeCapability(native.compiled_capabilities),
		EntitledCapabilities:          RuntimeCapability(native.entitled_capabilities),
		EffectiveCapabilities:         RuntimeCapability(native.effective_capabilities),
		TCPConnectionDefaultsRevision: uint32(native.tcp_connection_defaults_revision),
	}, nil
}

func (b *nativeBindings) createRuntime(config ConnectorConfig) (nativeRuntime, error) {
	if err := validateNativeText([2]string{"config.SystemName", config.SystemName}, [2]string{"config.NodeID", config.NodeID}); err != nil {
		return nil, err
	}
	systemName := C.CString(config.SystemName)
	nodeID := C.CString(config.NodeID)
	defer C.free(unsafe.Pointer(systemName))
	defer C.free(unsafe.Pointer(nodeID))
	cfg := C.coakka_v2_runtime_config_t{
		system_name:    systemName,
		node_id:        nodeID,
		strict_no_drop: boolToCInt(config.StrictNoDrop),
		queue_capacity: C.int(config.QueueCapacity),
	}
	var connection *C.coakka_v2_tcp_connection_options_t
	if config.ConnectionStrategy != nil {
		value := C.coakka_v2_tcp_connection_options_t{
			struct_size: C.size_t(C.sizeof_coakka_v2_tcp_connection_options_t),
			fields:      C.uint64_t(transportConnectionFieldMode),
			mode:        C.uint32_t(config.ConnectionStrategy.Mode),
		}
		if config.ConnectionStrategy.MaxConnections != nil {
			value.fields |= C.uint64_t(transportConnectionFieldMaxConnection)
			value.max_connections = C.uint32_t(*config.ConnectionStrategy.MaxConnections)
		}
		if config.ConnectionStrategy.MaxRequestsPerConnection != nil {
			value.fields |= C.uint64_t(transportConnectionFieldMaxRequests)
			value.max_requests_per_connection = C.uint64_t(*config.ConnectionStrategy.MaxRequestsPerConnection)
		}
		if config.ConnectionStrategy.IdleTimeoutMillis != nil {
			value.fields |= C.uint64_t(transportConnectionFieldIdleTimeout)
			value.idle_timeout_ms = C.uint64_t(*config.ConnectionStrategy.IdleTimeoutMillis)
		}
		connection = &value
	}
	var runtime *C.coakka_v2_runtime_t
	var monitor *C.coakka_v2_startup_monitor_options_t
	if config.Monitor != nil {
		monitor = &C.coakka_v2_startup_monitor_options_t{
			struct_size: C.size_t(C.sizeof_coakka_v2_startup_monitor_options_t),
			enabled:     C.uint32_t(boolToUint32(config.Monitor.Enabled)),
		}
	}
	status := C.coakka_v2_go_create_runtime_native(
		b.ptr,
		&cfg,
		connection,
		monitor,
		&runtime,
	)
	if int(status) != COAKKAOK {
		return nil, nativeStatusError("runtime_create_native", int(status))
	}
	if runtime == nil {
		return nil, &CoreContractError{Operation: "runtime_create_native", Detail: "success without owner"}
	}
	return nativeRuntime(runtime), nil
}

func (b *nativeBindings) destroyRuntime(runtime nativeRuntime) {
	if runtime != nil {
		C.coakka_v2_go_destroy_runtime(b.ptr, (*C.coakka_v2_runtime_t)(runtime))
	}
}

func (b *nativeBindings) applyNetworkOptions(runtime nativeRuntime, network RuntimeNetworkConfig) error {
	// The absent group goes untouched to the native constructor's default.
	if network == (RuntimeNetworkConfig{}) {
		return nil
	}
	if err := validateNativeText([2]string{"network.BindHost", network.BindHost}, [2]string{"network.AdvertiseHost", network.AdvertiseHost}); err != nil {
		return err
	}
	var bindHost *C.char
	var advertiseHost *C.char
	if network.BindHost != "" {
		bindHost = C.CString(network.BindHost)
		defer C.free(unsafe.Pointer(bindHost))
	}
	if network.AdvertiseHost != "" {
		advertiseHost = C.CString(network.AdvertiseHost)
		defer C.free(unsafe.Pointer(advertiseHost))
	}
	fields := uint64(C.COAKKA_V2_NETWORK_FIELD_MODE)
	if network.Mode == RuntimeNetworkNode {
		fields = uint64(C.COAKKA_V2_NETWORK_ALL_FIELDS)
	}
	if network.BindHost != "" {
		fields |= uint64(C.COAKKA_V2_NETWORK_FIELD_BIND_HOST)
	}
	if network.BindPort != 0 {
		fields |= uint64(C.COAKKA_V2_NETWORK_FIELD_BIND_PORT)
	}
	if network.AdvertiseHost != "" {
		fields |= uint64(C.COAKKA_V2_NETWORK_FIELD_ADVERTISE_HOST)
	}
	if network.AdvertisePort != 0 {
		fields |= uint64(C.COAKKA_V2_NETWORK_FIELD_ADVERTISE_PORT)
	}
	options := C.coakka_v2_network_options_t{
		struct_size:    C.size_t(C.sizeof_coakka_v2_network_options_t),
		fields:         C.uint64_t(fields),
		mode:           C.uint32_t(network.Mode),
		bind_host:      bindHost,
		advertise_host: advertiseHost,
		bind_port:      C.uint16_t(network.BindPort),
		advertise_port: C.uint16_t(network.AdvertisePort),
	}
	return requireStatus(
		C.coakka_v2_go_apply_network_options(b.ptr, (*C.coakka_v2_runtime_t)(runtime), &options),
		"runtime_apply_network_options",
	)
}

func (b *nativeBindings) getHostHandles(runtime nativeRuntime, flags uint32) (HostHandlesSnapshot, error) {
	handles := C.coakka_v2_host_handles_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_host_handles_t),
		flags:       C.uint32_t(flags),
	}
	if err := requireStatus(C.coakka_v2_go_get_host_handles(b.ptr, (*C.coakka_v2_runtime_t)(runtime), &handles), "get_host_handles"); err != nil {
		return HostHandlesSnapshot{}, err
	}
	return HostHandlesSnapshot{
		Flags:                  uint32(handles.flags),
		RequestWriteFD:         int(handles.request_write_fd),
		ResponseReadFD:         int(handles.response_read_fd),
		DeadletterReadFD:       int(handles.deadletter_read_fd),
		ControlWriteFD:         int(handles.control_write_fd),
		MonitorReadFD:          int(handles.monitor_read_fd),
		DeliveredRequestReadFD: int(handles.delivered_request_read_fd),
	}, nil
}

func nativeHostHandles(handles HostHandlesSnapshot) C.coakka_v2_host_handles_t {
	return C.coakka_v2_host_handles_t{
		struct_size:               C.size_t(C.sizeof_coakka_v2_host_handles_t),
		flags:                     C.uint32_t(handles.Flags),
		request_write_fd:          C.int(handles.RequestWriteFD),
		response_read_fd:          C.int(handles.ResponseReadFD),
		deadletter_read_fd:        C.int(handles.DeadletterReadFD),
		control_write_fd:          C.int(handles.ControlWriteFD),
		monitor_read_fd:           C.int(handles.MonitorReadFD),
		delivered_request_read_fd: C.int(handles.DeliveredRequestReadFD),
	}
}

func (b *nativeBindings) closeHostHandles(handles HostHandlesSnapshot) {
	native := nativeHostHandles(handles)
	C.coakka_v2_go_host_handles_close(b.ptr, &native)
}

func (b *nativeBindings) createRequestReader(runtime nativeRuntime, handles HostHandlesSnapshot) (nativeRequestReader, error) {
	nativeHandles := nativeHostHandles(handles)
	var reader *C.coakka_v2_request_reader_t
	status := C.coakka_v2_go_request_reader_create(
		b.ptr,
		(*C.coakka_v2_runtime_t)(runtime),
		&nativeHandles,
		&reader,
	)
	if int(status) != COAKKAOK {
		return nil, nativeStatusError("request_reader_create", int(status))
	}
	if reader == nil {
		return nil, &CoreContractError{Operation: "request_reader_create", Detail: "success without owner"}
	}
	return nativeRequestReader(reader), nil
}

func (b *nativeBindings) destroyRequestReader(reader nativeRequestReader) {
	if reader != nil {
		C.coakka_v2_go_request_reader_destroy(b.ptr, (*C.coakka_v2_request_reader_t)(reader))
	}
}

func (b *nativeBindings) readDeliveredRequest(reader nativeRequestReader, timeoutMillis uint32, staging []byte) (nativeDeliveredRequest, int, error) {
	if len(staging) == 0 {
		return nativeDeliveredRequest{}, COAKKAErrInvalidArg, nativeStatusError("request_reader_staging", COAKKAErrInvalidArg)
	}
	result := C.coakka_v2_go_read_delivered_request_packed(
		b.ptr,
		(*C.coakka_v2_request_reader_t)(reader),
		C.uint32_t(timeoutMillis),
		bytePointer(staging),
		C.size_t(len(staging)),
	)
	status := int(result.status)
	native := result.value
	if status == COAKKAErrWouldBlock || status == COAKKAErrClosed {
		return nativeDeliveredRequest{}, status, nil
	}
	if status != COAKKAOK {
		return nativeDeliveredRequest{}, status, nativeStatusError("request_reader_read", status)
	}
	if native.large_packed != nil {
		defer C.free(unsafe.Pointer(native.large_packed))
	}
	frameLength := uint64(native.frame_len)
	targetLength := uint64(native.target_len)
	payloadLength := uint64(native.payload_len)
	packedLength := uint64(native.packed_len)
	if frameLength == 0 || frameLength > maxNativeProjectionBytes ||
		targetLength == 0 || targetLength > maxNativeProjectionBytes ||
		payloadLength > maxNativeProjectionBytes || packedLength > maxNativeProjectionBytes || native.one_way > 1 ||
		packedLength != frameLength+targetLength+1+payloadLength {
		return nativeDeliveredRequest{}, COAKKAErrIO, nativeStatusError("owned_frame", COAKKAErrIO)
	}
	var packed []byte
	if native.large_packed != nil {
		packed = C.GoBytes(unsafe.Pointer(native.large_packed), C.int(packedLength))
	} else {
		if packedLength > uint64(len(staging)) {
			return nativeDeliveredRequest{}, COAKKAErrIO, nativeStatusError("request_reader_staging", COAKKAErrIO)
		}
		packed = staging[:int(packedLength)]
	}
	frameEnd := int(frameLength)
	targetEnd := frameEnd + int(targetLength)
	payloadStart := targetEnd + 1
	if packed[targetEnd] != 0 {
		return nativeDeliveredRequest{}, COAKKAErrIO, nativeStatusError("request_target_termination", COAKKAErrIO)
	}
	request := nativeDeliveredRequest{
		OneWay: native.one_way == 1,
	}
	if native.large_packed != nil {
		request.Frame = packed[:frameEnd]
		request.Target = packed[frameEnd:targetEnd]
		request.TargetCString = packed[frameEnd:payloadStart]
		request.Payload = packed[payloadStart:]
	} else {
		owned := append([]byte(nil), packed...)
		request.Frame = owned[:frameEnd]
		request.Target = owned[frameEnd:targetEnd]
		request.TargetCString = owned[frameEnd:payloadStart]
		request.Payload = owned[payloadStart:]
	}
	return request, COAKKAOK, nil
}

func (b *nativeBindings) createAskClient(runtime nativeRuntime, handles HostHandlesSnapshot) (nativeAskClient, error) {
	nativeHandles := nativeHostHandles(handles)
	var client *C.coakka_v2_ask_client_t
	status := int(C.coakka_v2_go_ask_client_create_ex(
		b.ptr,
		(*C.coakka_v2_runtime_t)(runtime),
		&nativeHandles,
		&client,
	))
	return adoptAskClient(status, nativeAskClient(client))
}

// Cold language ownership projection only. Core owns every causal refusal and
// partial-start rollback; a failed result never transfers a client to Go.
func adoptAskClient(status int, client nativeAskClient) (nativeAskClient, error) {
	if status != COAKKAOK {
		return nil, nativeStatusError("ask_client_create_ex", status)
	}
	if client == nil {
		return nil, &CoreContractError{Operation: "ask_client_create_ex", Detail: "success without owner"}
	}
	return client, nil
}

func (b *nativeBindings) destroyAskClient(client nativeAskClient) {
	if client != nil {
		C.coakka_v2_go_ask_client_destroy(b.ptr, (*C.coakka_v2_ask_client_t)(client))
	}
}

func (b *nativeBindings) beginAsk(client nativeAskClient, frame []byte) (nativeAskTicket, error) {
	var ticket *C.coakka_v2_ask_ticket_t
	status := int(C.coakka_v2_go_ask_client_begin(
		b.ptr,
		(*C.coakka_v2_ask_client_t)(client),
		bytePointer(frame),
		C.size_t(len(frame)),
		&ticket,
	))
	if status != COAKKAOK {
		return nil, nativeStatusError("ask_client_begin", status)
	}
	if ticket == nil {
		return nil, &CoreContractError{Operation: "ask_client_begin", Detail: "success without owner"}
	}
	return nativeAskTicket(ticket), nil
}

func (b *nativeBindings) askPayload(
	client nativeAskClient,
	messageID nativeCString,
	source nativeCString,
	target nativeCString,
	replyTo nativeCString,
	payload []byte,
	requestTimeoutMillis uint32,
	waitTimeoutMillis uint32,
	deliveryHint DeliveryHint,
) (int, []byte, int, error) {
	if !validNativeCString(messageID) || !validNativeCString(source) ||
		!validNativeCString(target) || !validNativeCString(replyTo) {
		return 0, nil, COAKKAErrInvalidArg, fmt.Errorf("ask text must be NUL terminated")
	}
	result := C.coakka_v2_go_ask_payload(
		b.ptr,
		(*C.coakka_v2_ask_client_t)(client),
		(*C.char)(unsafe.Pointer(unsafe.SliceData(messageID))),
		(*C.char)(unsafe.Pointer(unsafe.SliceData(source))),
		(*C.char)(unsafe.Pointer(unsafe.SliceData(target))),
		(*C.char)(unsafe.Pointer(unsafe.SliceData(replyTo))),
		bytePointer(payload),
		C.size_t(len(payload)),
		C.uint32_t(requestTimeoutMillis),
		C.uint32_t(waitTimeoutMillis),
		C.uint32_t(deliveryHint),
	)
	status := int(result.status)
	kind := result.kind
	bytes := result.value.owner
	if bytes != nil {
		defer C.coakka_v2_go_client_bytes_release(b.ptr, bytes)
	}
	if status == COAKKAErrWouldBlock || status == COAKKAErrClosed {
		return int(kind), nil, status, nil
	}
	length := uint64(result.value.size)
	if status != COAKKAOK {
		return int(kind), nil, status, nativeStatusError("ask_payload", status)
	}
	if length > maxNativeProjectionBytes || result.value.inline_owned > 1 ||
		(result.value.inline_owned == 0 && length != 0 && bytes == nil) ||
		(result.value.inline_owned == 1 && (length > C.COAKKA_GO_INLINE_PAYLOAD_BYTES || bytes != nil)) ||
		(kind != nativeClientResultResponse && kind != nativeClientResultDeadletter) {
		return int(kind), nil, COAKKAErrIO, nativeStatusError("ask_payload_result", COAKKAErrIO)
	}
	if length == 0 {
		return int(kind), nil, COAKKAOK, nil
	}
	if result.value.inline_owned == 1 {
		// The receipt owns this fixed span in Go memory. Copy before returning;
		// neither its address nor a borrowed slice escapes the language boundary.
		owned := make([]byte, int(length))
		copy(owned, unsafe.Slice((*byte)(unsafe.Pointer(&result.value.inline_bytes[0])), int(length)))
		return int(kind), owned, COAKKAOK, nil
	}
	return int(kind), C.GoBytes(unsafe.Pointer(bytes), C.int(length)), COAKKAOK, nil
}

func validNativeCString(value nativeCString) bool {
	return len(value) >= 2 && value[len(value)-1] == 0
}

func (b *nativeBindings) awaitAsk(ticket nativeAskTicket, timeoutMillis uint32) (int, []byte, int, error) {
	var kind C.uint32_t
	var frame *C.uint8_t
	var frameLength C.size_t
	status := int(C.coakka_v2_go_ask_ticket_await(
		b.ptr,
		(*C.coakka_v2_ask_ticket_t)(ticket),
		C.uint32_t(timeoutMillis),
		&kind,
		&frame,
		&frameLength,
	))
	if frame != nil {
		defer C.coakka_v2_go_client_bytes_release(b.ptr, frame)
	}
	if status == COAKKAErrWouldBlock || status == COAKKAErrClosed {
		return int(kind), nil, status, nil
	}
	length := uint64(frameLength)
	if status != COAKKAOK {
		return int(kind), nil, status, nativeStatusError("ask_ticket_await", status)
	}
	if frame == nil || length == 0 || length > maxNativeProjectionBytes ||
		(kind != nativeClientResultResponse && kind != nativeClientResultDeadletter) {
		return int(kind), nil, COAKKAErrIO, nativeStatusError("ask_result", COAKKAErrIO)
	}
	return int(kind), C.GoBytes(unsafe.Pointer(frame), C.int(length)), COAKKAOK, nil
}

func (b *nativeBindings) cancelAsk(ticket nativeAskTicket) {
	if ticket != nil {
		C.coakka_v2_go_ask_ticket_cancel(b.ptr, (*C.coakka_v2_ask_ticket_t)(ticket))
	}
}

func (b *nativeBindings) destroyAskTicket(ticket nativeAskTicket) {
	if ticket != nil {
		C.coakka_v2_go_ask_ticket_destroy(b.ptr, (*C.coakka_v2_ask_ticket_t)(ticket))
	}
}

func (b *nativeBindings) submitRawReply(runtime nativeRuntime, requestFrame []byte, source string, payload []byte) error {
	if err := validateNativeText([2]string{"source", source}); err != nil {
		return err
	}
	sourceBytes := make([]byte, len(source)+1)
	copy(sourceBytes, source)
	return b.submitRawReplyCString(runtime, requestFrame, sourceBytes, payload)
}

func (b *nativeBindings) submitRawReplyCString(runtime nativeRuntime, requestFrame []byte, sourceCString []byte, payload []byte) error {
	if len(sourceCString) < 2 || sourceCString[len(sourceCString)-1] != 0 {
		return nativeStatusError("submit_raw_reply_source", COAKKAErrInvalidArg)
	}
	for _, value := range sourceCString[:len(sourceCString)-1] {
		if value == 0 {
			return nativeStatusError("submit_raw_reply_source", COAKKAErrInvalidArg)
		}
	}
	status := int(C.coakka_v2_go_submit_raw_reply(
		b.ptr,
		(*C.coakka_v2_runtime_t)(runtime),
		bytePointer(requestFrame),
		C.size_t(len(requestFrame)),
		(*C.char)(unsafe.Pointer(unsafe.SliceData(sourceCString))),
		bytePointer(payload),
		C.size_t(len(payload)),
	))
	if status != COAKKAOK {
		return nativeStatusError("submit_raw_reply", status)
	}
	return nil
}

func (b *nativeBindings) startRuntime(runtime nativeRuntime) error {
	return requireStatus(C.coakka_v2_go_start_runtime(b.ptr, (*C.coakka_v2_runtime_t)(runtime)), "runtime_start")
}

func (b *nativeBindings) stopRuntime(runtime nativeRuntime) error {
	return requireStatus(C.coakka_v2_go_stop_runtime(b.ptr, (*C.coakka_v2_runtime_t)(runtime)), "runtime_stop")
}

func (b *nativeBindings) getHealth(runtime nativeRuntime) (RuntimeHealthSnapshot, error) {
	health := C.coakka_v2_runtime_health_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_runtime_health_t),
	}
	if err := requireStatus(C.coakka_v2_go_get_health(b.ptr, (*C.coakka_v2_runtime_t)(runtime), &health), "runtime_get_health"); err != nil {
		return RuntimeHealthSnapshot{}, err
	}
	return RuntimeHealthSnapshot{
		RuntimeState:      int32(health.runtime_state),
		Flags:             uint32(health.flags),
		AppliedGeneration: uint64(health.applied_generation),
	}, nil
}

func (b *nativeBindings) getStats(runtime nativeRuntime) (RuntimeStatsSnapshot, error) {
	stats := C.coakka_v2_runtime_stats_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_runtime_stats_t),
	}
	if err := requireStatus(C.coakka_v2_go_get_stats(b.ptr, (*C.coakka_v2_runtime_t)(runtime), &stats), "runtime_get_stats"); err != nil {
		return RuntimeStatsSnapshot{}, err
	}
	return RuntimeStatsSnapshot{
		AppliedGeneration:                         uint64(stats.applied_generation),
		RouteCount:                                uint64(stats.route_count),
		RuntimeState:                              int32(stats.runtime_state),
		QueueRejectedCount:                        uint64(stats.queue_rejected_count),
		RouteMissCount:                            uint64(stats.route_miss_count),
		DeadletterCount:                           uint64(stats.deadletter_count),
		DeliveryFailedCount:                       uint64(stats.delivery_failed_count),
		ControlRejectedCount:                      uint64(stats.control_rejected_count),
		MonitorEventEmittedCount:                  uint64(stats.monitor_event_emitted_count),
		MonitorEventDroppedCount:                  uint64(stats.monitor_event_dropped_count),
		MonitorEventEmittedLifetimeCount:          uint64(stats.monitor_event_emitted_lifetime_count),
		MonitorEventDroppedLifetimeCount:          uint64(stats.monitor_event_dropped_lifetime_count),
		LocalWorkQueueCapacity:                    uint64(stats.local_work_queue_capacity),
		LocalWorkQueueDepth:                       uint64(stats.local_work_queue_depth),
		LocalWorkQueueHighWatermark:               uint64(stats.local_work_queue_high_watermark),
		DeliveredRequestOutboundDirectWriteCount:  uint64(stats.delivered_request_outbound_direct_write_count),
		ResponseOutboundDirectWriteCount:          uint64(stats.response_outbound_direct_write_count),
		DeadletterOutboundDirectWriteCount:        uint64(stats.deadletter_outbound_direct_write_count),
		RemoteOutboundQueueCapacity:               uint64(stats.remote_outbound_queue_capacity),
		RemoteOutboundQueueDepth:                  uint64(stats.remote_outbound_queue_depth),
		RemoteOutboundQueueHighWatermark:          uint64(stats.remote_outbound_queue_high_watermark),
		RemoteOutboundQueueRejectedCount:          uint64(stats.remote_outbound_queue_rejected_count),
		RemoteOutboundExpiredDropCount:            uint64(stats.remote_outbound_expired_drop_count),
		RemoteOutboundReplyReserveSlots:           uint64(stats.remote_outbound_reply_reserve_slots),
		RemoteOutboundReplyReservationRejectCount: uint64(stats.remote_outbound_reply_reservation_reject_count),
		RemoteOutboundOneWayDropCount:             uint64(stats.remote_outbound_one_way_drop_count),
	}, nil
}

func (b *nativeBindings) getConfig(runtime nativeRuntime) (RuntimeConfigSnapshot, error) {
	config := C.coakka_v2_runtime_config_view_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_runtime_config_view_t),
	}
	if err := requireStatus(C.coakka_v2_go_get_config(b.ptr, (*C.coakka_v2_runtime_t)(runtime), &config), "runtime_get_config"); err != nil {
		return RuntimeConfigSnapshot{}, err
	}
	return RuntimeConfigSnapshot{
		SystemName:                               cString(config.system_name),
		NodeID:                                   cString(config.node_id),
		RuntimeState:                             int32(config.runtime_state),
		SnapshotPresent:                          config.snapshot_present != 0,
		AppliedGeneration:                        uint64(config.applied_generation),
		RouteCount:                               uint64(config.route_count),
		TransportBindHost:                        cString(config.southbound_bind_host),
		TransportBindPort:                        uint16(config.southbound_bind_port),
		EffectiveIngressOverloadMode:             uint32(config.effective_ingress_overload_mode),
		EffectiveLocalDeliveryOverloadMode:       uint32(config.effective_local_delivery_overload_mode),
		EffectiveRemoteOutboundOverloadMode:      uint32(config.effective_remote_outbound_overload_mode),
		EffectiveRemoteOutboundReplyReserveSlots: uint64(config.effective_remote_outbound_reply_reserve_slots),
	}, nil
}

func (b *nativeBindings) getStartupConfig(runtime nativeRuntime) (RuntimeStartupConfigSnapshot, error) {
	native := C.coakka_v2_runtime_startup_config_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_runtime_startup_config_t),
	}
	if err := requireStatus(
		C.coakka_v2_go_get_startup_config(b.ptr, (*C.coakka_v2_runtime_t)(runtime), &native),
		"runtime_get_startup_config",
	); err != nil {
		return RuntimeStartupConfigSnapshot{}, err
	}
	return RuntimeStartupConfigSnapshot{
		DefaultsRevision:    uint32(native.defaults_revision),
		DefaultedGroups:     uint32(native.defaulted_groups),
		StrictNoDrop:        native.strict_no_drop != 0,
		QueueCapacity:       uint32(native.queue_capacity),
		RequestMaxFrameSize: uint64(native.request_max_frame_size),
		RuntimeState:        uint32(native.runtime_state),
		Execution: RuntimeStartupCPUInfo{
			RequestedPolicy:   uint32(native.execution.requested_policy),
			Supported:         native.execution.supported != 0,
			AvailableCPUCount: uint32(native.execution.available_cpu_count),
			SelectedCPUCount:  uint32(native.execution.selected_cpu_count),
			SelectedCPUIds:    [2]uint32{uint32(native.execution.selected_cpu_ids[0]), uint32(native.execution.selected_cpu_ids[1])},
			StartupApplied:    native.execution.startup_applied != 0,
		},
		Connection: connectionConfigSnapshot(native.connection),
		IO: RuntimeStartupIOInfo{
			Requested:  native.io.requested != 0,
			Compiled:   native.io.compiled != 0,
			Probed:     native.io.probed != 0,
			Supported:  native.io.supported != 0,
			Effective:  native.io.effective != 0,
			ProbeError: int32(native.io.probe_error),
		},
		Monitor: RuntimeStartupMonitorInfo{Enabled: native.monitor.enabled != 0},
	}, nil
}

func (b *nativeBindings) applyTCPConnectionStrategy(
	runtime nativeRuntime,
	spec RuntimeTCPConnectionStrategySpec,
) (RuntimeTCPConnectionApplyResult, error) {
	options := C.coakka_v2_tcp_connection_options_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_tcp_connection_options_t),
		fields:      C.uint64_t(transportConnectionFieldMode),
		mode:        C.uint32_t(spec.Mode),
	}
	if spec.MaxConnections != nil {
		options.fields |= C.uint64_t(transportConnectionFieldMaxConnection)
		options.max_connections = C.uint32_t(*spec.MaxConnections)
	}
	if spec.MaxRequestsPerConnection != nil {
		options.fields |= C.uint64_t(transportConnectionFieldMaxRequests)
		options.max_requests_per_connection = C.uint64_t(*spec.MaxRequestsPerConnection)
	}
	if spec.IdleTimeoutMillis != nil {
		options.fields |= C.uint64_t(transportConnectionFieldIdleTimeout)
		options.idle_timeout_ms = C.uint64_t(*spec.IdleTimeoutMillis)
	}
	native := C.coakka_v2_tcp_connection_apply_result_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_tcp_connection_apply_result_t),
	}
	status := C.coakka_v2_go_apply_tcp_connection_options_ex(
		b.ptr,
		(*C.coakka_v2_runtime_t)(runtime),
		&options,
		&native,
	)
	if int32(status) != int32(native.apply_status) {
		return RuntimeTCPConnectionApplyResult{}, fmt.Errorf(
			"tcp connection apply returned status=%d but result status=%d",
			int32(status),
			int32(native.apply_status),
		)
	}
	return RuntimeTCPConnectionApplyResult{
		Status:          CoakkaStatus(native.apply_status),
		Changed:         native.changed != 0,
		Reason:          RuntimeTransportApplyReason(native.reason),
		ReasonName:      cString(C.coakka_v2_go_transport_apply_reason_name(b.ptr, native.reason)),
		RuntimeState:    uint32(native.runtime_state),
		ValidationCode:  RuntimeTCPConnectionValidationCode(native.validation.code),
		ValidationField: uint64(native.validation.field),
		MinimumValue:    uint64(native.validation.minimum_value),
		MaximumValue:    uint64(native.validation.maximum_value),
		ActiveConfig:    connectionConfigSnapshot(native.effective_config),
	}, nil
}

func (b *nativeBindings) getTCPConnectionConfig(
	runtime nativeRuntime,
) (RuntimeTCPConnectionConfigSnapshot, error) {
	native := C.coakka_v2_tcp_connection_config_snapshot_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_tcp_connection_config_snapshot_t),
	}
	if err := requireStatus(
		C.coakka_v2_go_get_tcp_connection_config(
			b.ptr,
			(*C.coakka_v2_runtime_t)(runtime),
			&native,
		),
		"runtime_get_tcp_connection_config",
	); err != nil {
		return RuntimeTCPConnectionConfigSnapshot{}, err
	}
	return connectionConfigSnapshot(native), nil
}

func (b *nativeBindings) applyTCPSecurity(
	runtime nativeRuntime,
	spec RuntimeTCPSecuritySpec,
) (RuntimeTCPSecurityApplyResult, error) {
	for _, field := range []struct {
		name  string
		value string
	}{
		{name: "credentialId", value: spec.CredentialID},
		{name: "caCertificateFile", value: spec.CACertificateFile},
		{name: "identityCertificateFile", value: spec.IdentityCertificateFile},
		{name: "privateKeyFile", value: spec.PrivateKeyFile},
	} {
		if err := validateTransportText(field.value, field.name); err != nil {
			return RuntimeTCPSecurityApplyResult{}, err
		}
	}
	credentialID := C.CString(spec.CredentialID)
	caCertificateFile := C.CString(spec.CACertificateFile)
	identityCertificateFile := C.CString(spec.IdentityCertificateFile)
	privateKeyFile := C.CString(spec.PrivateKeyFile)
	defer C.free(unsafe.Pointer(credentialID))
	defer C.free(unsafe.Pointer(caCertificateFile))
	defer C.free(unsafe.Pointer(identityCertificateFile))
	defer C.free(unsafe.Pointer(privateKeyFile))

	options := C.coakka_v2_tcp_security_options_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_tcp_security_options_t),
		fields:      C.uint64_t(transportSecurityFieldMode),
		mode:        C.uint32_t(spec.Mode),
	}
	if spec.Mode != RuntimeTCPSecurityPlaintext {
		options.fields = C.uint64_t(transportSecurityAllFields)
		options.credential_source_kind = C.uint32_t(transportCredentialSourceFile)
		options.reload_mode = C.uint32_t(spec.ReloadMode)
		options.credential_generation = C.uint64_t(spec.CredentialGeneration)
		options.credential_id = credentialID
		options.ca_certificate_file = caCertificateFile
		options.identity_certificate_file = identityCertificateFile
		options.private_key_file = privateKeyFile
	} else {
		if spec.ReloadMode != RuntimeTLSReloadGraceful {
			options.fields |= C.uint64_t(transportSecurityFieldReloadMode)
			options.reload_mode = C.uint32_t(spec.ReloadMode)
		}
		if spec.CredentialGeneration != 0 {
			options.fields |= C.uint64_t(transportSecurityFieldGeneration)
			options.credential_generation = C.uint64_t(spec.CredentialGeneration)
		}
		if spec.CredentialID != "" {
			options.fields |= C.uint64_t(transportSecurityFieldCredentialID)
			options.credential_id = credentialID
		}
		if spec.CACertificateFile != "" || spec.IdentityCertificateFile != "" || spec.PrivateKeyFile != "" {
			options.fields |= C.uint64_t(transportSecurityFieldSource)
			options.credential_source_kind = C.uint32_t(transportCredentialSourceFile)
		}
		if spec.CACertificateFile != "" {
			options.fields |= C.uint64_t(transportSecurityFieldCAFile)
			options.ca_certificate_file = caCertificateFile
		}
		if spec.IdentityCertificateFile != "" {
			options.fields |= C.uint64_t(transportSecurityFieldIdentityFile)
			options.identity_certificate_file = identityCertificateFile
		}
		if spec.PrivateKeyFile != "" {
			options.fields |= C.uint64_t(transportSecurityFieldKeyFile)
			options.private_key_file = privateKeyFile
		}
	}
	native := C.coakka_v2_tcp_security_apply_result_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_tcp_security_apply_result_t),
	}
	status := C.coakka_v2_go_apply_tcp_security_options_ex(
		b.ptr,
		(*C.coakka_v2_runtime_t)(runtime),
		&options,
		&native,
	)
	if int32(status) != int32(native.apply_status) {
		return RuntimeTCPSecurityApplyResult{}, fmt.Errorf(
			"tcp security apply returned status=%d but result status=%d",
			int32(status),
			int32(native.apply_status),
		)
	}
	return RuntimeTCPSecurityApplyResult{
		Status:          CoakkaStatus(native.apply_status),
		Changed:         native.changed != 0,
		Reason:          RuntimeTransportApplyReason(native.reason),
		ReasonName:      cString(C.coakka_v2_go_transport_apply_reason_name(b.ptr, native.reason)),
		RuntimeState:    uint32(native.runtime_state),
		ValidationCode:  RuntimeTCPSecurityValidationCode(native.validation.code),
		ValidationField: uint64(native.validation.field),
		ActiveSecurity:  securityInfoSnapshot(native.active_security),
	}, nil
}

func (b *nativeBindings) getTCPSecurityInfo(
	runtime nativeRuntime,
) (RuntimeTCPSecurityInfoSnapshot, error) {
	native := C.coakka_v2_tcp_security_info_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_tcp_security_info_t),
	}
	if err := requireStatus(
		C.coakka_v2_go_get_tcp_security_info(
			b.ptr,
			(*C.coakka_v2_runtime_t)(runtime),
			&native,
		),
		"runtime_get_tcp_security_info",
	); err != nil {
		return RuntimeTCPSecurityInfoSnapshot{}, err
	}
	return securityInfoSnapshot(native), nil
}

func connectionConfigSnapshot(
	native C.coakka_v2_tcp_connection_config_snapshot_t,
) RuntimeTCPConnectionConfigSnapshot {
	return RuntimeTCPConnectionConfigSnapshot{
		DefaultsRevision:           uint32(native.defaults_revision),
		Mode:                       RuntimeTCPConnectionMode(native.mode),
		ApplicableFields:           uint64(native.applicable_fields),
		ExplicitlyConfiguredFields: uint64(native.explicitly_configured_fields),
		DefaultedFields:            uint64(native.defaulted_fields),
		ConfigurableFields:         uint64(native.configurable_fields),
		MaxConnections:             uint32(native.max_connections),
		MaxRequestsPerConnection:   uint64(native.max_requests_per_connection),
		IdleTimeoutMillis:          uint64(native.idle_timeout_ms),
	}
}

func securityInfoSnapshot(native C.coakka_v2_tcp_security_info_t) RuntimeTCPSecurityInfoSnapshot {
	return RuntimeTCPSecurityInfoSnapshot{
		Mode:                         RuntimeTCPSecurityMode(native.config.mode),
		CredentialSourceKind:         uint32(native.config.credential_source_kind),
		ReloadMode:                   RuntimeTLSReloadMode(native.config.reload_mode),
		ReloadStatus:                 uint32(native.config.reload_status),
		CredentialGeneration:         uint64(native.config.credential_generation),
		CredentialID:                 C.GoString(&native.identity.credential_id_value[0]),
		MinimumProtocolVersion:       uint32(native.identity.minimum_protocol_version),
		MaximumProtocolVersion:       uint32(native.identity.maximum_protocol_version),
		InboundVerificationFlags:     uint64(native.identity.inbound_verification_flags),
		OutboundVerificationFlags:    uint64(native.identity.outbound_verification_flags),
		IdentityNotBeforeUnixSeconds: int64(native.identity.identity_not_before_unix_seconds),
		IdentityNotAfterUnixSeconds:  int64(native.identity.identity_not_after_unix_seconds),
		IdentityFingerprintSHA256:    C.GoString(&native.identity.identity_fingerprint_sha256[0]),
	}
}

func transportABISizes() RuntimeTransportABISizes {
	return RuntimeTransportABISizes{
		Capabilities:         uintptr(C.sizeof_coakka_v2_runtime_capability_snapshot_t),
		ConnectionOptions:    uintptr(C.sizeof_coakka_v2_tcp_connection_options_t),
		ConnectionValidation: uintptr(C.sizeof_coakka_v2_tcp_connection_validation_t),
		ConnectionConfig:     uintptr(C.sizeof_coakka_v2_tcp_connection_config_snapshot_t),
		ConnectionApply:      uintptr(C.sizeof_coakka_v2_tcp_connection_apply_result_t),
		SecurityOptions:      uintptr(C.sizeof_coakka_v2_tcp_security_options_t),
		SecurityValidation:   uintptr(C.sizeof_coakka_v2_tcp_security_validation_t),
		SecurityConfig:       uintptr(C.sizeof_coakka_v2_tcp_security_config_snapshot_t),
		SecurityIdentity:     uintptr(C.sizeof_coakka_v2_tcp_security_identity_info_t),
		SecurityInfo:         uintptr(C.sizeof_coakka_v2_tcp_security_info_t),
		SecurityApply:        uintptr(C.sizeof_coakka_v2_tcp_security_apply_result_t),
	}
}

func (b *nativeBindings) submitEnvelope(runtime nativeRuntime, bytes []byte) error {
	return requireStatus(C.coakka_v2_go_submit_envelope(b.ptr, (*C.coakka_v2_runtime_t)(runtime), bytePointer(bytes), C.size_t(len(bytes))), "runtime_submit_envelope")
}

func (b *nativeBindings) applyControlEnvelope(runtime nativeRuntime, bytes []byte) error {
	return requireStatus(C.coakka_v2_go_apply_control_envelope(b.ptr, (*C.coakka_v2_runtime_t)(runtime), bytePointer(bytes), C.size_t(len(bytes))), "runtime_apply_control_envelope")
}

func (b *nativeBindings) monitorWaitConsume(fd int, timeoutMillis uint32) (uint64, int, error) {
	var signalCount C.uint64_t
	status := int(C.coakka_v2_go_monitor_wait_consume(
		b.ptr,
		C.int(fd),
		C.uint32_t(timeoutMillis),
		&signalCount,
	))
	if status == COAKKAErrWouldBlock || status == COAKKAErrClosed {
		return 0, status, nil
	}
	if status != COAKKAOK {
		return 0, status, nativeStatusError("monitor_wait_consume", status)
	}
	return uint64(signalCount), COAKKAOK, nil
}

func nativeStatusError(operation string, status int) error {
	return &CoreStatusError{Operation: operation, Status: status}
}

func requireStatus(status C.int, operation string) error {
	if int(status) == COAKKAOK {
		return nil
	}
	return nativeStatusError(operation, int(status))
}

func cString(ptr *C.char) string {
	if ptr == nil {
		return ""
	}
	return C.GoString(ptr)
}

func boolToCInt(value bool) C.int {
	if value {
		return 1
	}
	return 0
}

func boolToUint32(value bool) uint32 {
	if value {
		return 1
	}
	return 0
}

func bytePointer(bytes []byte) *C.uint8_t {
	if len(bytes) == 0 {
		return nil
	}
	return (*C.uint8_t)(unsafe.Pointer(unsafe.SliceData(bytes)))
}
