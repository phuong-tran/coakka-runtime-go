# Go Connector Release Notes

## v1.5.1

This patch keeps connector version `2.1.0` and native generation
`2.1.0+60ddf70d` unchanged. It replaces the private connector-repository link
with the canonical public file-lane contract and adds a regression test that
rejects private documentation links in the packaged README.

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
