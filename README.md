# CoAkka Runtime Go 3.2.1

The Go connector embeds CoAkka Runtime in a Go application. It shares Core's
runtime, request/reply, deadletter, File and Stream contracts with the other
CoAkka connectors.

This connector uses module `github.com/phuong-tran/coakka-runtime-go/v3`,
version `v3.2.1`, over native Core 3.2/f9e5a632d/ABI1. It retains the public 3.0.0 API.

## Module and build requirements

Use the /v3 import path. The v3.2.1 tag must belong to this module path.
There are no external Go module dependencies. A working C11 cgo compiler is
required at build time; the resulting app contains the selected native image.

For the public File Lane API and delivery contract, see the
[file transfer guide](https://github.com/phuong-tran/coakka-samples/blob/main/docs/coakka-runtime/guides/runtime-file-transfer.md).

For registry availability, see [Current Packages](https://github.com/phuong-tran/coakka-samples/blob/main/docs/common/current-packages.md). Install this version with:

```sh
go get github.com/phuong-tran/coakka-runtime-go/v3@v3.2.1
```

For local qualification, use the exact selected module with an ordinary
`replace` directive. Do not fetch the older public module as a fallback.

The compatibility floor is Go1.22.12 for macOS ARM64, Linux ARM64/x64 and
Windows x64. Native Windows ARM64 uses its separately qualified Go1.25.5/cgo
profile. A build on another architecture does not prove execution there.
The selected native images target macOS13 ARM64, Linux ARM64 glibc2.35,
Linux x64 glibc2.36, Windows x64 and Windows11 ARM64; exact matching execution
and minimum-OS results are recorded in the package qualification record.

## Application shape

Import the connector with a local application alias:

```go
import coakka "github.com/phuong-tran/coakka-runtime-go/v3"
```

Start a runtime host with the application's identity and routes, register
handlers, make typed requests and propagate both operation and close failures.
The runnable [basic, deadletter and File fan-out samples](https://github.com/phuong-tran/coakka-samples/tree/main/coakka-runtime/go)
show complete ownership and cleanup through the selected package.

Core resolves omitted configuration to AUTO execution and MULTIPLEXED
connections. Core owns effective configuration, capability probes, admission,
deadlines, terminal causes and native shutdown. Inspect its accepted snapshots;
do not reconstruct native state from submitted preferences.

Go owns Go values, callback/goroutine scheduling and FFI memory safety.
Handler payloads and returned snapshots are copied before native leases retire.
Handlers must return promptly; cancellation and close are explicit lifecycle
operations. A failed close retains owners for the documented retry law.

The package's selected image is embedded at build time. Its digest-addressed
cache publishes atomically. Cold load, reuse and concurrent publishers must
converge on the same image; a corrupt existing entry refuses and remains
unchanged. No ambient library search or older runtime fallback is supplied.

## Migration from v1

Change imports and requirements to /v3 and recompile. Runtime metadata uses
TransportProvider/AllocatorProvider and TransportBindHost/TransportBindPort
rather than the old SouthboundBackend/AllocatorBackend and SouthboundBind*
fields. The module no longer depends on generated message bindings.
Review current constructors, typed values and error handling instead of
retaining accidental source or binary compatibility aliases.

Core-issued unknown numeric statuses and labels remain unknown values.
Feature-unavailable errors follow a successful capability inspection;
inspection failures retain their actual cause. Queue, frame and transport
defaults are resolved by Core outside request hot paths.

See [consumer deployment](CONSUMING.md), [transport configuration](TRANSPORT_CONFIGURATION.md),
[release controls](RELEASE.md) and [migration notes](RELEASE_NOTES.md).
The shared [documentation](https://github.com/phuong-tran/coakka-samples/blob/main/docs/README.md)
owns cross-language concepts and lifecycle guidance.

## License

Managed connector source uses Apache2.0. Native images use the file-scoped
CoAkka Native Artifact License1.2. Retain the supplied LICENSE, NOTICE,
NATIVE-LICENSE.md and PACKAGE-LICENSE.md. Publisher signing is currently absent.
