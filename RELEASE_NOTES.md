# Go Connector Release Notes

## v1.7.1

Lowers the module compatibility floor from Go `1.23.0` to Go `1.22`, the
minimum required by `google.golang.org/protobuf v1.36.6`. The package and
JVM-Go demo pass module-tidy and source tests with both Go `1.22.12` and the
current stable toolchain. The demo no longer retains an unused
`golang.org/x/sys` dependency that raised its floor to Go 1.23.

This patch keeps connector version `2.4.0`, native generation
`2.4.0+c2f53117`, the public Go API, runtime ABI, lifecycle, bounded queues,
transport behavior, and all five native payloads unchanged. Go `1.22` is a
compatibility floor, not a recommendation to run an unsupported toolchain in
production; use a currently supported Go release for production builds.

## v1.7.0

Publishes connector version `2.4.0` with exact native generation
`2.4.0+c2f53117` for all five supported OS/CPU payloads. Runtime startup now
uses an explicit network participation mode: `EMBEDDED` and `OUTBOUND_ONLY`
do not open a listener, while `NETWORK_NODE` requires explicit bind and
advertise endpoints and fails closed when bind fails. File Lane and Stream
Lane remain available through the same bounded, lifecycle-owned connector
surface. The module remains on semantic major `v1` because its public path has
no `/v2` suffix.

## v1.6.0

Adds concurrent-safe Stream Lane bindings with `runtime/cgo.Handle` callback
ownership, panic containment, credit and pressure snapshots, draining close,
and exact native generation `2.3.0+a83ab412`. The module remains on semantic
major `v1` because its public path has no `/v2` suffix.

## v1.5.0

This module release adds `OpenFileLane` for connector version `2.1.0` over
native generation `2.1.0+60ddf70d`. It remains on semantic major `v1`
because the public module path has no `/v2` suffix.
The concurrent-safe API provides bounded lane configuration, receive
preparation, sender submission, SHA-256, sequence-based progress waits,
cancellation, retained terminal records, stats, and stop-before-drain close.
Direct TCP, TLS, and mutual TLS are startup profiles.

The module packages exact Linux ARM64/x86-64, macOS ARM64, and Windows
ARM64/x86-64 payloads. Unit tests, payload verification, the clean module
consumer, and a `9 MiB + 731 byte` file transfer with SHA-256 equality pass on
macOS ARM64. Matching-host execution claims remain platform-specific.

## v1.4.1

This release adds runtime-v2 transport configuration while preserving the
public C ABI semantics:

- exact cgo layouts for capability, connection, security, and structured apply
  result blocks
- capability discovery before optional feature selection
- startup-configured connection strategy and TLS/mTLS policy
- atomic results with active-state preservation after rejection
- same-mode newer-generation TLS/mTLS credential reload
- copied non-secret TLS identity snapshots
- a process-lifetime runtime-library identity that avoids native static
  registry teardown/re-registration across sequential host lifecycles
- OS-specific C shims for dynamic loading and runtime fd wait/read/close

macOS ARM64 source tests pass against the baseline capability profile and the
exact full-capability runtime generation `1.4.1+9e02a51d`, including failed key
mismatch preservation and successful generation reload. The same cgo source
cross-compiles to Windows x86-64 PE and Linux ARM64 ELF. Those cross-builds are
compile evidence only; this package receipt makes no connector or consumer
execution claim for Linux or Windows.

The module payload verifier passes with the exact release-metadata digests for
macOS ARM64, Linux ARM64/x86-64, and Windows ARM64/x86-64, and its macOS
consumer smoke passes. This does not claim Linux/Windows execution. Publisher
signing is absent.

Canonical guides: [connection strategies](https://github.com/phuong-tran/coakka-publish/blob/main/docs/connection-strategies.md),
[TLS/mTLS](https://github.com/phuong-tran/coakka-publish/blob/main/docs/tls-and-mtls.md),
[troubleshooting](https://github.com/phuong-tran/coakka-publish/blob/main/docs/troubleshooting.md),
and [contact/support](https://github.com/phuong-tran/coakka-publish/blob/main/docs/contact-and-support.md).
