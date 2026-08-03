# Go Connector Release Notes

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
