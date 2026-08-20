# Go Runtime Transport Configuration

The public CoAkka C ABI is the semantic authority. The Go connector copies all
results and snapshots into Go-owned values; it does not invent lifecycle,
capability, validation, or edition behavior.

## Startup Contract

Set `Network`, `ConnectionStrategy`, and `Security` on `ConnectorStartSpec` or
`ConnectorConfig`. The connector creates the native runtime, applies the
network options and selected transport policy in `CREATED`, exports host
handles, applies the initial route snapshot, then starts the runtime. A
rejected startup apply destroys the handle. Transport-policy errors retain the
structured native result.

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
	Network:    coakka.EmbeddedNetwork(),
	ConnectionStrategy: &coakka.RuntimeTCPConnectionStrategySpec{Mode: mode},
	Security: &coakka.RuntimeTCPSecuritySpec{
		Mode: coakka.RuntimeTCPSecurityPlaintext,
	},
}, runtimeLibrary)
```

## Network Participation

`Network` is explicit about whether this process owns a runtime listener:

| Constructor | Listener behavior | Required addresses |
| --- | --- | --- |
| `EmbeddedNetwork()` | Opens no runtime network listener. This is also the zero-value default. | None; local routes use port `0`. |
| `OutboundOnlyNetwork()` | Opens no listener but allows outbound runtime connections to remote network nodes. | None on the local process. |
| `NetworkNode(bindHost, bindPort, advertiseHost, advertisePort)` | Binds a runtime listener and advertises a reachable endpoint. | Bind and advertise host/port are required; a zero advertise port inherits the bind port. |

Use a network node only when another process or machine must connect to this
runtime. Bind may use an interface wildcard, but advertise must be a concrete
host reachable by peers. The connector rejects missing addresses, wildcard
advertise hosts, listener fields on non-listening modes, and unknown modes
before creating a live host. Native bind failure fails startup closed.

```go
host, err := coakka.StartRuntimeHost(coakka.ConnectorStartSpec{
	SystemName: "orders",
	NodeID:     "orders-edge-1",
	Routes:     []coakka.RouteSpec{coakka.LocalRouteDefault("orders.create")},
	Network:    coakka.NetworkNode("0.0.0.0", 19102, "10.20.0.15", 19102),
}, runtimeLibrary)
```

Do not add a listener merely because application code and the runtime are in
the same process. See the canonical
[runtime network modes](https://github.com/phuong-tran/coakka-samples/blob/main/docs/runtime-network-modes.md)
guide for embedded, outbound-only, and cross-process deployment shapes.

## Public Functions

| Function | Purpose and parameters | Default/result | Ownership, thread safety, blocking, lifecycle, atomicity, errors, edition |
| --- | --- | --- | --- |
| `ReadRuntimeCapabilities(runtimeLibPath)` | Resolve/load one runtime library and read compiled, entitled, and effective capability bits. Empty path follows normal resolver order. | Copied `RuntimeCapabilitiesSnapshot`. | Synchronous and process-safe. The first loaded runtime-library identity remains loaded for process lifetime; a different later path is rejected. Available in all editions. |
| `capabilities.Supports(bits)` | Require every requested bit to be effective. | Boolean; the empty bit set is supported. | Pure copied-value operation, thread-safe and non-blocking. It never infers capability from edition/package name. |
| `EmbeddedNetwork()` / `OutboundOnlyNetwork()` / `NetworkNode(...)` | Select listener ownership before runtime start. | Zero value is embedded; only a network node accepts bind/advertise fields. | Validation is synchronous. Network options are immutable after start; apply or bind failure releases the runtime handle. |
| `StartRuntimeHost(startSpec, runtimeLibPath)` / `StartConnectorOrchestrator(...)` | Own the single active process runtime. Network and optional transport specs live on `startSpec`. | Omitted network mode is embedded; omitted transport policy preserves native defaults. | Applies before start. Credential file loading can block. Rejection releases the runtime handle; transport-policy rejection returns a typed error with active state. One active host per process. |
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

`RuntimeNetworkConfig` defaults to `EMBEDDED`, which owns no listener and needs
no port. `OUTBOUND_ONLY` also owns no listener. `NETWORK_NODE` requires both a
bind endpoint and a non-wildcard advertise endpoint. `LocalRouteDefault(...)`
uses port `0` because a local route is process-owned rather than a loopback TCP
endpoint.

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

See the canonical [connection strategy guide](https://github.com/phuong-tran/coakka-samples/blob/main/docs/connection-strategies.md)
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

See the canonical [TLS/mTLS guide](https://github.com/phuong-tran/coakka-samples/blob/main/docs/tls-and-mtls.md)
for direct runtime TLS/mTLS without a service-mesh data plane, Kubernetes,
controlled networks, LAN/edge, RPi, BeagleBone, bare metal, industrial Android,
and rotation semantics.

## Errors, Platforms, And Troubleshooting

Structured rejections preserve status, stable reason/name, validation
code/field/range, runtime state, `Changed`, and the active snapshot. The dynamic
runtime module remains loaded for process lifetime so native static registries
are not torn down and registered twice across sequential host lifecycles.

The module keeps and digest-verifies Windows x86-64, macOS ARM64, and Linux
ARM64 native artifacts. Cross-compilation and payload presence are not execution
evidence; exact connector and consumer evidence is reported per platform. Use
[common troubleshooting](https://github.com/phuong-tran/coakka-samples/blob/main/docs/troubleshooting.md)
for loader, OS/CPU, dependency, certificate, Gatekeeper, Authenticode, digest,
and signing status. Publisher signing is currently absent.
Contact `gabrielgun1983@gmail.com` or use the public issue tracker.
