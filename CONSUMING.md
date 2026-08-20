# Consuming The Go Runtime Package

This is the Go connector for the polyglot, multi-language, multi-platform
CoAkka Runtime ecosystem. The public Go module includes the matching native
runtime libraries:

```sh
go get github.com/phuong-tran/coakka-runtime-go@v1.8.2
```

Every module release records its connector version and bundled native runtime
generation separately.

Example consumer `go.mod`:

```go
module my-runtime-consumer

go 1.22

require github.com/phuong-tran/coakka-runtime-go v1.8.2
```

Go `1.22` is the module compatibility floor. For production builds, use a
currently supported Go release; CI verifies both Go `1.22.12` compatibility
and the current stable toolchain.

Example:

```go
package main

import (
	"time"

	connector "github.com/phuong-tran/coakka-runtime-go"
)

func main() {
	runtimeHost, err := connector.StartRuntimeHost(connector.ConnectorStartSpec{
		SystemName: "sample",
		NodeID:     "sample-node",
		Routes:     []connector.RouteSpec{connector.LocalRouteDefault("svc.echo")},
	}, "")
	if err != nil {
		panic(err)
	}
	defer runtimeHost.Close()

	_, _ = runtimeHost.AwaitNextMonitor(10 * time.Millisecond)
}
```

Library resolution order:

- explicit path passed to `StartRuntimeHost(startSpec, runtimeLibPath)`
- `$COAKKA_RUNTIME_LIB`
- packaged native library under `native/<platform>/`
- local fallback candidates under `lib/`

Package targets:

- `macos-aarch64`
- `linux-aarch64`
- `linux-x86_64`
- `windows-aarch64`
- `windows-x86_64`

All five native digests and binary formats are verified during packaging.
Exact module `v1.8.2` request/reply passes on macOS ARM64. Module source and
consumer-shaped package tests pass on Linux x86-64 with Go `1.22.12` and the
current stable toolchain. Exact native generation `2.5.0+4b65d0b2256037bf7fc180bfa6df8c41efc1dd6a` retains the
previous matching-host request/reply evidence on Linux
ARM64/x86-64. Both Windows payloads pass package, export, dependency, and
digest gates; matching Go-on-Windows execution is not recorded. Read
[Transport Configuration](TRANSPORT_CONFIGURATION.md) before selecting a
non-default connection mode or TLS/mTLS, and use the canonical
[troubleshooting guide](https://github.com/phuong-tran/coakka-samples/blob/main/docs/troubleshooting.md)
for loader, architecture, certificate, and publisher-trust failures.

One Go process may start one active runtime host. `StartConnectorOrchestrator`
remains as the compatibility name for the same lifecycle owner.

For plain text request identities, prefer:

```go
identity := connector.NewTextPayloadIdentity("demo.echo.request.v1")
```

`PayloadFormatPlainText` remains as a compatibility alias; new user-facing code
should use `PayloadFormatText`.

Repository:

- `https://github.com/phuong-tran/coakka-runtime-go`
