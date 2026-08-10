package coakka_v2_connector

import (
	"errors"
	"fmt"
	"sync"
)

const (
	// FileLaneSender enables sender work on a lane.
	FileLaneSender uint32 = 1
	// FileLaneReceiver enables receiver work on a lane.
	FileLaneReceiver uint32 = 1 << 1
)

// FileLaneSecurityMode selects transport protection for file bytes.
type FileLaneSecurityMode uint32

const (
	FileLaneDirect FileLaneSecurityMode = iota
	FileLaneTLS
	FileLaneMutualTLS
)

// FileTransferDirection identifies one side of a retained transfer record.
type FileTransferDirection uint32

const (
	FileTransferSend    FileTransferDirection = 1
	FileTransferReceive FileTransferDirection = 2
)

// FileTransferState is an observable file-transfer lifecycle state.
type FileTransferState uint32

const (
	FileTransferPrepared     FileTransferState = 1
	FileTransferQueued       FileTransferState = 2
	FileTransferConnecting   FileTransferState = 3
	FileTransferTransferring FileTransferState = 4
	FileTransferVerifying    FileTransferState = 5
	FileTransferCompleted    FileTransferState = 6
	FileTransferPaused       FileTransferState = 7
	FileTransferRejected     FileTransferState = 8
	FileTransferFailed       FileTransferState = 9
	FileTransferCanceled     FileTransferState = 10
)

// FileTransferResult is a stable terminal outcome reported by one peer.
type FileTransferResult uint32

const (
	FileResultNone FileTransferResult = iota
	FileResultOK
	FileResultNotPrepared
	FileResultTokenMismatch
	FileResultMetadataMismatch
	FileResultSizeLimit
	FileResultStorageIO
	FileResultIntegrityMismatch
	FileResultNetworkIO
	FileResultTimeout
	FileResultQueueFull
	FileResultProtocolError
	FileResultSourceChanged
	FileResultInternalError
	FileResultCanceledByHost
	FileResultTLSConfigInvalid
	FileResultTLSHandshakeFailed
	FileResultPeerCertUntrusted
	FileResultPeerCertExpired
	FileResultPeerIdentityMismatch
	FileResultClientCertRequired
)

// FileLaneSecurityConfig names TLS material read when a lane starts.
// Private-key paths and per-transfer tokens must never be logged.
type FileLaneSecurityConfig struct {
	Mode                    FileLaneSecurityMode
	CredentialGeneration    uint64
	CredentialID            string
	CACertificateFile       string
	IdentityCertificateFile string
	PrivateKeyFile          string
}

// FileLaneConfig controls a bounded lane. Zero tuning fields select native defaults.
// Size fields are bytes and time fields are milliseconds.
type FileLaneConfig struct {
	Flags                  uint32
	BindHost               string
	BindPort               uint16
	QueueCapacity          uint64
	MaxFileSize            uint64
	IOTimeoutMillis        uint32
	CheckpointBytes        uint64
	ProgressBytes          uint64
	ProgressIntervalMillis uint32
	SenderWorkerCount      uint32
	ReceiverWorkerCount    uint32
	Security               *FileLaneSecurityConfig
}

// DefaultFileLaneConfig returns a loopback lane with sender and receiver enabled.
func DefaultFileLaneConfig() FileLaneConfig {
	return FileLaneConfig{Flags: FileLaneSender | FileLaneReceiver, BindHost: "127.0.0.1"}
}
func (c FileLaneConfig) normalized() FileLaneConfig {
	if c.Flags == 0 {
		c.Flags = FileLaneSender | FileLaneReceiver
	}
	if c.BindHost == "" {
		c.BindHost = "127.0.0.1"
	}
	return c
}
func (c FileLaneConfig) validate() error {
	if c.Flags & ^uint32(FileLaneSender|FileLaneReceiver) != 0 || c.Flags == 0 {
		return errors.New("file lane requires valid sender or receiver flags")
	}
	if c.SenderWorkerCount > 4 || c.ReceiverWorkerCount > 4 {
		return errors.New("file lane worker counts must be in [0, 4]")
	}
	return nil
}

// FileReceiveSpec authorizes one destination and exact content identity.
type FileReceiveSpec struct {
	TransferID         string
	AuthorizationToken string
	DestinationPath    string
	ExpectedSize       uint64
	ExpectedSHA256     [32]byte
}

// FileSendSpec names the source and endpoint for a previously prepared receive.
type FileSendSpec struct {
	TransferID         string
	AuthorizationToken string
	RemoteHost         string
	RemotePort         uint16
	SourcePath         string
	ExpectedSize       uint64
	ExpectedSHA256     [32]byte
	TimeoutMillis      uint32
}

// FileDigest contains a file's SHA-256 and exact byte count.
type FileDigest struct {
	SHA256 [32]byte
	Size   uint64
}

// FileTransferSnapshot is an immutable progress view.
// Its monotonic timestamps are process-local and are not wall-clock values.
// ProgressMilli ranges from 0 to 100000 (100.000%); UpdateSequence advances on retained changes.
type FileTransferSnapshot struct {
	Direction                                                                     FileTransferDirection
	State                                                                         FileTransferState
	Result                                                                        FileTransferResult
	ExpectedSize, TransferredBytes, CommittedOffset                               uint64
	ProgressMilli                                                                 uint32
	CancelRequested                                                               bool
	UpdateSequence, SubmittedMonoNS, StartedMonoNS, UpdatedMonoNS, TerminalMonoNS uint64
	Detail                                                                        string
}

// Terminal reports whether this side has reached a terminal state.
func (s FileTransferSnapshot) Terminal() bool { return s.State >= FileTransferCompleted }

// Succeeded reports COMPLETED plus OK for this side of the transfer.
func (s FileTransferSnapshot) Succeeded() bool {
	return s.State == FileTransferCompleted && s.Result == FileResultOK
}

// FileLaneStats contains bounded queue, active-work, terminal, and byte counters.
type FileLaneStats struct{ QueueCapacity, QueuedSends, PreparedReceives, ActiveSends, ActiveReceives, RetainedRecords, SubmittedSends, PreparedReceiveCount, CompletedSends, CompletedReceives, FailedSends, FailedReceives, CanceledTransfers, CompletedSendBytes, CompletedReceiveBytes uint64 }

// FileLane owns an independent native bulk-transfer lane.
//
// Prepare the receiver before submitting the sender. Continue WaitTransfer from
// each UpdateSequence until both peers report success, then Forget retained
// records. Close stops the lane, wakes blocked waits, and drains active calls.
// File bytes do not belong in runtime Envelope payloads.
type FileLane struct {
	bindings *nativeBindings
	lane     nativeFileLane
	mu       sync.Mutex
	drained  *sync.Cond
	closing  bool
	active   int
}

// OpenFileLane opens and starts a lane using the selected native runtime.
func OpenFileLane(config FileLaneConfig, runtimeLibPath string) (*FileLane, error) {
	config = config.normalized()
	if err := config.validate(); err != nil {
		return nil, err
	}
	path, err := (RuntimeLibraryResolver{}).Resolve(runtimeLibPath, nil)
	if err != nil {
		return nil, err
	}
	bindings, err := openNativeBindings(path)
	if err != nil {
		return nil, err
	}
	if !bindings.fileLaneAvailable() {
		bindings.close()
		return nil, errors.New("native runtime does not export file-lane ABI; install the next runtime native release")
	}
	lane, err := bindings.createFileLane(config)
	if err != nil {
		bindings.close()
		return nil, err
	}
	if err := bindings.startFileLane(lane); err != nil {
		bindings.destroyFileLane(lane)
		bindings.close()
		return nil, err
	}
	result := &FileLane{bindings: bindings, lane: lane}
	result.drained = sync.NewCond(&result.mu)
	return result, nil
}

// FileSHA256 computes the exact source identity using the native implementation.
func FileSHA256(path, runtimeLibPath string) (FileDigest, error) {
	resolved, err := (RuntimeLibraryResolver{}).Resolve(runtimeLibPath, nil)
	if err != nil {
		return FileDigest{}, err
	}
	bindings, err := openNativeBindings(resolved)
	if err != nil {
		return FileDigest{}, err
	}
	defer bindings.close()
	if !bindings.fileLaneAvailable() {
		return FileDigest{}, errors.New("native runtime does not export file-lane ABI")
	}
	return bindings.fileSHA256(path)
}
func (l *FileLane) acquire() (nativeFileLane, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closing || l.lane == nil {
		return nil, errors.New("file lane is closed")
	}
	l.active++
	return l.lane, nil
}
func (l *FileLane) release() {
	l.mu.Lock()
	l.active--
	if l.active == 0 {
		l.drained.Broadcast()
	}
	l.mu.Unlock()
}

// BoundPort returns the receiver port selected when the lane started.
func (l *FileLane) BoundPort() (uint16, error) {
	lane, err := l.acquire()
	if err != nil {
		return 0, err
	}
	defer l.release()
	return l.bindings.fileLaneBoundPort(lane)
}

// PrepareReceive authorizes one destination and expected content identity.
func (l *FileLane) PrepareReceive(spec FileReceiveSpec) error {
	lane, err := l.acquire()
	if err != nil {
		return err
	}
	defer l.release()
	return l.bindings.prepareFileReceive(lane, spec)
}

// SubmitSend queues a send after the remote application prepares the receive.
func (l *FileLane) SubmitSend(spec FileSendSpec) error {
	if spec.RemotePort == 0 {
		return errors.New("remote port must be in [1, 65535]")
	}
	lane, err := l.acquire()
	if err != nil {
		return err
	}
	defer l.release()
	return l.bindings.submitFileSend(lane, spec)
}

// Transfer returns the current copied snapshot without waiting.
func (l *FileLane) Transfer(transferID string, direction FileTransferDirection) (FileTransferSnapshot, error) {
	lane, err := l.acquire()
	if err != nil {
		return FileTransferSnapshot{}, err
	}
	defer l.release()
	return l.bindings.fileTransfer(lane, transferID, direction, 0, 0, false)
}

// WaitTransfer blocks until the sequence advances, timeoutMillis expires, or the lane stops.
// afterSequence is the last update already handled; zero requests the current state.
func (l *FileLane) WaitTransfer(transferID string, direction FileTransferDirection, afterSequence uint64, timeoutMillis uint32) (FileTransferSnapshot, error) {
	lane, err := l.acquire()
	if err != nil {
		return FileTransferSnapshot{}, err
	}
	defer l.release()
	return l.bindings.fileTransfer(lane, transferID, direction, afterSequence, timeoutMillis, true)
}

// Cancel requests cooperative cancellation; observe terminal state before Forget.
func (l *FileLane) Cancel(transferID string, direction FileTransferDirection) error {
	lane, err := l.acquire()
	if err != nil {
		return err
	}
	defer l.release()
	return l.bindings.cancelFileTransfer(lane, transferID, direction, false)
}

// Forget releases one retained terminal record after its outcome is recorded.
func (l *FileLane) Forget(transferID string, direction FileTransferDirection) error {
	lane, err := l.acquire()
	if err != nil {
		return err
	}
	defer l.release()
	return l.bindings.cancelFileTransfer(lane, transferID, direction, true)
}

// Stats returns a copied lane-level observability snapshot.
func (l *FileLane) Stats() (FileLaneStats, error) {
	lane, err := l.acquire()
	if err != nil {
		return FileLaneStats{}, err
	}
	defer l.release()
	return l.bindings.fileLaneStats(lane)
}

// Close stops the lane, wakes blocked waits, drains calls, and releases native state.
func (l *FileLane) Close() error {
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
	status := l.bindings.stopFileLane(lane)
	l.mu.Lock()
	for l.active != 0 {
		l.drained.Wait()
	}
	l.bindings.destroyFileLane(lane)
	l.bindings.close()
	l.lane = nil
	l.drained.Broadcast()
	l.mu.Unlock()
	if status != 0 && status != -7 {
		return fmt.Errorf("file_lane_stop failed: native status %d", status)
	}
	return nil
}
