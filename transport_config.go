package coakka_v2_connector

import (
	"fmt"
	"strings"
)

const (
	transportConnectionFieldMode          uint64 = 1 << 0
	transportConnectionFieldMaxConnection uint64 = 1 << 1
	transportConnectionFieldMaxRequests   uint64 = 1 << 2
	transportConnectionFieldIdleTimeout   uint64 = 1 << 3
	transportSecurityFieldMode            uint64 = 1 << 0
	transportSecurityFieldSource          uint64 = 1 << 1
	transportSecurityFieldReloadMode      uint64 = 1 << 2
	transportSecurityFieldGeneration      uint64 = 1 << 3
	transportSecurityFieldCredentialID    uint64 = 1 << 4
	transportSecurityFieldCAFile          uint64 = 1 << 5
	transportSecurityFieldIdentityFile    uint64 = 1 << 6
	transportSecurityFieldKeyFile         uint64 = 1 << 7
	transportSecurityAllFields                   = (1 << 8) - 1
	transportCredentialSourceFile         uint32 = 1
)

type CoakkaStatus int32

const (
	CoakkaStatusOK                 CoakkaStatus = 0
	CoakkaStatusInvalidArgument    CoakkaStatus = -1
	CoakkaStatusNoMemory           CoakkaStatus = -2
	CoakkaStatusBadState           CoakkaStatus = -3
	CoakkaStatusStaleGeneration    CoakkaStatus = -4
	CoakkaStatusIO                 CoakkaStatus = -5
	CoakkaStatusWouldBlock         CoakkaStatus = -6
	CoakkaStatusClosed             CoakkaStatus = -7
	CoakkaStatusFeatureUnavailable CoakkaStatus = -8
	CoakkaStatusFeatureNotEntitled CoakkaStatus = -9
)

type RuntimeCapability uint64

const (
	RuntimeCapabilityTCPBoundedPool            RuntimeCapability = 1 << 0
	RuntimeCapabilityTCPPoolTuning             RuntimeCapability = 1 << 1
	RuntimeCapabilityTCPTLS                    RuntimeCapability = 1 << 2
	RuntimeCapabilityTCPMutualTLS              RuntimeCapability = 1 << 3
	RuntimeCapabilityTLSCredentialReload       RuntimeCapability = 1 << 4
	RuntimeCapabilityTLSExternalProvider       RuntimeCapability = 1 << 5
	RuntimeCapabilityTCPPersistentSingleFlight RuntimeCapability = 1 << 6
	RuntimeCapabilityTCPMultiplexing           RuntimeCapability = 1 << 7
)

type RuntimeTCPConnectionMode uint32

const (
	RuntimeTCPConnectionPerExchange            RuntimeTCPConnectionMode = 0
	RuntimeTCPConnectionBoundedPool            RuntimeTCPConnectionMode = 1
	RuntimeTCPConnectionPersistentSingleFlight RuntimeTCPConnectionMode = 2
	RuntimeTCPConnectionMultiplexing           RuntimeTCPConnectionMode = 3
)

type RuntimeTCPSecurityMode uint32

const (
	RuntimeTCPSecurityPlaintext RuntimeTCPSecurityMode = 0
	RuntimeTCPSecurityTLS       RuntimeTCPSecurityMode = 1
	RuntimeTCPSecurityMutualTLS RuntimeTCPSecurityMode = 2
)

type RuntimeTLSReloadMode uint32

const (
	RuntimeTLSReloadGraceful                 RuntimeTLSReloadMode = 0
	RuntimeTLSReloadDrainExistingConnections RuntimeTLSReloadMode = 1
)

type RuntimeTransportApplyReason uint32

const (
	RuntimeTransportApplyReasonNone                      RuntimeTransportApplyReason = 0
	RuntimeTransportApplyReasonInvalidArgument           RuntimeTransportApplyReason = 1
	RuntimeTransportApplyReasonFeatureUnavailable        RuntimeTransportApplyReason = 2
	RuntimeTransportApplyReasonFeatureNotEntitled        RuntimeTransportApplyReason = 3
	RuntimeTransportApplyReasonRuntimeNotConfigurable    RuntimeTransportApplyReason = 4
	RuntimeTransportApplyReasonSecurityModeChange        RuntimeTransportApplyReason = 5
	RuntimeTransportApplyReasonStaleCredentialGeneration RuntimeTransportApplyReason = 6
	RuntimeTransportApplyReasonCredentialRejected        RuntimeTransportApplyReason = 7
	RuntimeTransportApplyReasonResourceFailure           RuntimeTransportApplyReason = 8
	RuntimeTransportApplyReasonAdapterRejected           RuntimeTransportApplyReason = 9
)

type RuntimeTCPConnectionValidationCode uint32

const (
	RuntimeTCPConnectionValid              RuntimeTCPConnectionValidationCode = 0
	RuntimeTCPConnectionInvalidStructSize  RuntimeTCPConnectionValidationCode = 1
	RuntimeTCPConnectionUnknownField       RuntimeTCPConnectionValidationCode = 2
	RuntimeTCPConnectionModeRequired       RuntimeTCPConnectionValidationCode = 3
	RuntimeTCPConnectionUnknownMode        RuntimeTCPConnectionValidationCode = 4
	RuntimeTCPConnectionFieldNotApplicable RuntimeTCPConnectionValidationCode = 5
	RuntimeTCPConnectionValueOutOfRange    RuntimeTCPConnectionValidationCode = 6
	RuntimeTCPConnectionFeatureUnavailable RuntimeTCPConnectionValidationCode = 7
	RuntimeTCPConnectionFeatureNotEntitled RuntimeTCPConnectionValidationCode = 8
	RuntimeTCPConnectionReservedNonzero    RuntimeTCPConnectionValidationCode = 9
	RuntimeTCPConnectionFieldOutsideStruct RuntimeTCPConnectionValidationCode = 10
	RuntimeTCPConnectionValueWithoutField  RuntimeTCPConnectionValidationCode = 11
)

type RuntimeTCPSecurityValidationCode uint32

const (
	RuntimeTCPSecurityValid                RuntimeTCPSecurityValidationCode = 0
	RuntimeTCPSecurityInvalidStructSize    RuntimeTCPSecurityValidationCode = 1
	RuntimeTCPSecurityUnknownField         RuntimeTCPSecurityValidationCode = 2
	RuntimeTCPSecurityModeRequired         RuntimeTCPSecurityValidationCode = 3
	RuntimeTCPSecurityUnknownMode          RuntimeTCPSecurityValidationCode = 4
	RuntimeTCPSecurityReservedNonzero      RuntimeTCPSecurityValidationCode = 5
	RuntimeTCPSecurityFieldOutsideStruct   RuntimeTCPSecurityValidationCode = 6
	RuntimeTCPSecurityFieldNotApplicable   RuntimeTCPSecurityValidationCode = 7
	RuntimeTCPSecurityRequiredFieldMissing RuntimeTCPSecurityValidationCode = 8
	RuntimeTCPSecuritySourceUnavailable    RuntimeTCPSecurityValidationCode = 9
	RuntimeTCPSecurityFeatureUnavailable   RuntimeTCPSecurityValidationCode = 10
	RuntimeTCPSecurityInvalidGeneration    RuntimeTCPSecurityValidationCode = 11
	RuntimeTCPSecurityCredentialIDTooLong  RuntimeTCPSecurityValidationCode = 12
	RuntimeTCPSecurityValueWithoutField    RuntimeTCPSecurityValidationCode = 13
)

type RuntimeTCPConnectionStrategySpec struct {
	Mode                     RuntimeTCPConnectionMode
	MaxConnections           *uint32
	MaxRequestsPerConnection *uint64
	IdleTimeoutMillis        *uint64
}

type RuntimeTCPSecuritySpec struct {
	Mode                    RuntimeTCPSecurityMode
	ReloadMode              RuntimeTLSReloadMode
	CredentialGeneration    uint64
	CredentialID            string
	CACertificateFile       string
	IdentityCertificateFile string
	PrivateKeyFile          string
}

type RuntimeCapabilitiesSnapshot struct {
	Edition                       uint32
	LicenseStatus                 uint32
	CompiledCapabilities          RuntimeCapability
	EntitledCapabilities          RuntimeCapability
	EffectiveCapabilities         RuntimeCapability
	TCPConnectionDefaultsRevision uint32
}

func (s RuntimeCapabilitiesSnapshot) Supports(capabilities RuntimeCapability) bool {
	return s.EffectiveCapabilities&capabilities == capabilities
}

type RuntimeTCPConnectionConfigSnapshot struct {
	DefaultsRevision           uint32
	Mode                       RuntimeTCPConnectionMode
	ApplicableFields           uint64
	ExplicitlyConfiguredFields uint64
	DefaultedFields            uint64
	ConfigurableFields         uint64
	MaxConnections             uint32
	MaxRequestsPerConnection   uint64
	IdleTimeoutMillis          uint64
}

type RuntimeTCPConnectionApplyResult struct {
	Status          CoakkaStatus
	Changed         bool
	Reason          RuntimeTransportApplyReason
	ReasonName      string
	RuntimeState    uint32
	ValidationCode  RuntimeTCPConnectionValidationCode
	ValidationField uint64
	MinimumValue    uint64
	MaximumValue    uint64
	ActiveConfig    RuntimeTCPConnectionConfigSnapshot
}

func (r RuntimeTCPConnectionApplyResult) Applied() bool {
	return r.Status == CoakkaStatusOK
}

type RuntimeTCPSecurityInfoSnapshot struct {
	Mode                         RuntimeTCPSecurityMode
	CredentialSourceKind         uint32
	ReloadMode                   RuntimeTLSReloadMode
	ReloadStatus                 uint32
	CredentialGeneration         uint64
	CredentialID                 string
	MinimumProtocolVersion       uint32
	MaximumProtocolVersion       uint32
	InboundVerificationFlags     uint64
	OutboundVerificationFlags    uint64
	IdentityNotBeforeUnixSeconds int64
	IdentityNotAfterUnixSeconds  int64
	IdentityFingerprintSHA256    string
}

type RuntimeTCPSecurityApplyResult struct {
	Status          CoakkaStatus
	Changed         bool
	Reason          RuntimeTransportApplyReason
	ReasonName      string
	RuntimeState    uint32
	ValidationCode  RuntimeTCPSecurityValidationCode
	ValidationField uint64
	ActiveSecurity  RuntimeTCPSecurityInfoSnapshot
}

func (r RuntimeTCPSecurityApplyResult) Applied() bool {
	return r.Status == CoakkaStatusOK
}

type RuntimeTCPConnectionApplyError struct {
	Result RuntimeTCPConnectionApplyResult
}

func (e *RuntimeTCPConnectionApplyError) Error() string {
	return fmt.Sprintf(
		"tcp connection strategy apply failed status=%d reason=%s validation=%d category=CONFIGURATION",
		e.Result.Status,
		e.Result.ReasonName,
		e.Result.ValidationCode,
	)
}

type RuntimeTCPSecurityApplyError struct {
	Result RuntimeTCPSecurityApplyResult
}

func (e *RuntimeTCPSecurityApplyError) Error() string {
	return fmt.Sprintf(
		"tcp security apply failed status=%d reason=%s validation=%d category=CONFIGURATION",
		e.Result.Status,
		e.Result.ReasonName,
		e.Result.ValidationCode,
	)
}

type RuntimeTransportABISizes struct {
	Capabilities         uintptr
	ConnectionOptions    uintptr
	ConnectionValidation uintptr
	ConnectionConfig     uintptr
	ConnectionApply      uintptr
	SecurityOptions      uintptr
	SecurityValidation   uintptr
	SecurityConfig       uintptr
	SecurityIdentity     uintptr
	SecurityInfo         uintptr
	SecurityApply        uintptr
}

func TransportABISizes() RuntimeTransportABISizes {
	return transportABISizes()
}

func validateTransportText(value string, field string) error {
	if strings.IndexByte(value, 0) >= 0 {
		return fmt.Errorf("%s must not contain NUL", field)
	}
	return nil
}
