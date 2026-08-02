# Go Runtime Transport Configuration

The public CoAkka C ABI is the semantic authority. The Go connector copies all
results and snapshots into Go-owned values; it does not invent lifecycle,
capability, validation, or edition behavior.

## Startup Contract

Set `ConnectionStrategy` and `Security` on `ConnectorStartSpec` or
`ConnectorConfig`. The connector creates the native runtime, applies selected
policy in `CREATED`, exports host handles, applies the initial route snapshot,
then starts the runtime. A rejected startup apply destroys the handle and
returns `*RuntimeTCPConnectionApplyError` or `*RuntimeTCPSecurityApplyError`;
the error retains the structured native result.

Omit both pointers to preserve runtime defaults. Query capabilities before
selecting an optional mode.

```go
capabilities, err := coakka.ReadRuntimeCapabilities(runtimeLibrary)
if err != nil {
	return err
}
mode := coakka.RuntimeTCPConnectionPerExchange
if capabilities.Supports(coakka.RuntimeCapabilityTCPBoundedPool) {
	mode = coakka.RuntimeTCPConnectionBoundedPool
}

host, err := coakka.StartRuntimeHost(coakka.ConnectorStartSpec{
	SystemName: "orders",
	NodeID:     "orders-1",
	Routes:     []coakka.RouteSpec{coakka.LocalRouteDefault("orders.create")},
	ConnectionStrategy: &coakka.RuntimeTCPConnectionStrategySpec{Mode: mode},
	Security: &coakka.RuntimeTCPSecuritySpec{
		Mode: coakka.RuntimeTCPSecurityPlaintext,
	},
}, runtimeLibrary)
```

## Public Functions

| Function | Purpose and parameters | Default/result | Ownership, thread safety, blocking, lifecycle, atomicity, errors, edition |
| --- | --- | --- | --- |
| `ReadRuntimeCapabilities(runtimeLibPath)` | Resolve/load one runtime library and read compiled, entitled, and effective capability bits. Empty path follows normal resolver order. | Copied `RuntimeCapabilitiesSnapshot`. | Synchronous and process-safe. The first loaded runtime-library identity remains loaded for process lifetime; a different later path is rejected. Available in all editions. |
| `capabilities.Supports(bits)` | Require every requested bit to be effective. | Boolean; the empty bit set is supported. | Pure copied-value operation, thread-safe and non-blocking. It never infers capability from edition/package name. |
| `StartRuntimeHost(startSpec, runtimeLibPath)` / `StartConnectorOrchestrator(...)` | Own the single active process runtime. Optional startup transport specs live on `startSpec`. | Omitted policy preserves native defaults; explicit startup results are retained. | Applies before start. Credential file loading can block. Rejection returns a typed error with active state and releases the runtime handle. One active host per process. |
| `host.RuntimeCapabilities()` | Read capability truth through the host's loaded binding. | Copied snapshot plus bridge error. | Serialized with transport apply/close; valid only while open. |
| `host.TCPConnectionConfig()` | Read effective mode, tuning values, and explicit/default provenance. | Copied `RuntimeTCPConnectionConfigSnapshot`. | Serialized with apply/close, synchronous, open-host only. Availability of modes/tuning follows capabilities. |
| `host.TCPSecurityInfo()` | Read active non-secret mode, generation, credential ID, protocol/verification metadata, validity, and fingerprint. | Copied `RuntimeTCPSecurityInfoSnapshot`. | Excludes paths, PEM, keys, tokens, and raw diagnostics. Serialized with apply/close. TLS fields depend on effective capabilities. |
| `host.StartupTCPConnectionResult()` | Inspect explicit startup connection apply. | `(result, false)` when no spec was selected. | Copied value, non-blocking, valid after close. |
| `host.StartupTCPSecurityResult()` | Inspect explicit startup security apply. | `(result, false)` when no spec was selected. | Copied non-secret value, non-blocking, valid after close. |
| `host.ApplyTCPConnectionStrategy(spec)` | Attempt an apply and return the state remaining after it. Optional tuning uses pointer fields; nil means absent. | Structured result plus bridge/conversion error. Native policy rejection is a result, not a Go error. | Serialized and atomic. A started host normally returns `BadState` and preserves active config because connection mode is startup-only. |
| `host.ApplyTCPSecurity(spec)` | Apply startup-shaped security or reload a strictly newer generation in the same live TLS/mTLS mode. | Structured result plus bridge/conversion error. | Credential strings are borrowed only for the synchronous call and freed afterward. Failed/stale reload preserves active generation. File I/O may block. TLS/mTLS/reload are capability-gated. |
| `connectionResult.Applied()` / `securityResult.Applied()` | Test native status `OK`. | Boolean; `Changed` remains separate. | Pure copied-value operation. A successful no-op can be applied without changing state. |
| `TransportABISizes()` | Expose compiled cgo sizes for conformance diagnostics/tests. | Eleven ABI block sizes. | Pure process-local metadata. It does not prove execution on another OS/architecture. |

## Specs And Defaults

`RuntimeTCPConnectionStrategySpec.Mode` defaults to `PER_EXCHANGE` because the
Go zero value maps to the core zero value. `MaxConnections`,
`MaxRequestsPerConnection`, and `IdleTimeoutMillis` are pointers: nil omits the
field; a non-nil zero is an explicit value that core validation may reject.

`RuntimeTCPSecuritySpec` zero value is plaintext with graceful reload,
generation zero, and no credential fields. TLS/mTLS requires file-backed CA,
certificate, and private-key paths, a positive generation, and a non-secret
credential ID. Credential values in plaintext mode are forwarded for stable
core rejection rather than discarded. Strings containing NUL are rejected at
the Go/C boundary.

See the canonical [connection strategy guide](https://github.com/phuong-tran/coakka-publish/blob/main/docs/connection-strategies.md)
for `PER_EXCHANGE`, `BOUNDED_POOL`, `PERSISTENT_SINGLE_FLIGHT`, and
`MULTIPLEXING` behavior, defaults, edition matrix, and tuning availability.

## TLS Reload

```go
result, err := host.ApplyTCPSecurity(coakka.RuntimeTCPSecuritySpec{
	Mode:                    coakka.RuntimeTCPSecurityTLS,
	CredentialGeneration:    2,
	CredentialID:            "orders-edge-2026-08-b",
	CACertificateFile:       "/run/secrets/coakka/ca.pem",
	IdentityCertificateFile: "/run/secrets/coakka/tls.crt",
	PrivateKeyFile:          "/run/secrets/coakka/tls.key",
})
if err != nil {
	return err
}
if !result.Applied() {
	return fmt.Errorf(
		"reload rejected: %s; active generation=%d",
		result.ReasonName,
		result.ActiveSecurity.CredentialGeneration,
	)
}
```

The caller owns path strings and host secret lifecycle. The runtime loads and
validates material during the synchronous apply, then atomically publishes an
immutable TLS context. It never returns secret paths or bytes through public
snapshots.

See the canonical [TLS/mTLS guide](https://github.com/phuong-tran/coakka-publish/blob/main/docs/tls-and-mtls.md)
for Kubernetes ingress/service-mesh guidance, controlled networks, LAN/edge,
RPi, BeagleBone, bare metal, industrial Android, and rotation semantics.

## Errors, Platforms, And Troubleshooting

Structured rejections preserve status, stable reason/name, validation
code/field/range, runtime state, `Changed`, and the active snapshot. The dynamic
runtime module remains loaded for process lifetime so native static registries
are not torn down and registered twice across sequential host lifecycles.

The module keeps and digest-verifies Windows x86-64, macOS ARM64, and Linux
ARM64 native artifacts. Cross-compilation and payload presence are not execution
evidence; exact connector and consumer evidence is reported per platform. Use
[common troubleshooting](https://github.com/phuong-tran/coakka-publish/blob/main/docs/troubleshooting.md)
for loader, OS/CPU, dependency, certificate, Gatekeeper, Authenticode, digest,
and signing status. Publisher signing is currently absent.
Contact `gabrielgun1983@gmail.com` or use the public issue tracker.
