# Consuming The Go Runtime Package

This is the Go connector for the polyglot, multi-language, multi-platform
CoAkka Runtime ecosystem. The public Go module includes the matching native
runtime libraries:

```sh
go get github.com/phuong-tran/coakka-runtime-go@v1.4.1
```

Every module release records its connector version and bundled native runtime
generation separately.

Example consumer `go.mod`:

```go
module my-runtime-consumer

go 1.23.0

require github.com/phuong-tran/coakka-runtime-go v1.4.1
```

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
- `windows-x86_64`

All five native digests are verified during packaging. Current connector
execution evidence covers macOS ARM64; this package receipt makes no Go
execution claim for Linux or Windows. Read
[Transport Configuration](TRANSPORT_CONFIGURATION.md) before selecting a
non-default connection mode or TLS/mTLS, and use the canonical
[troubleshooting guide](https://github.com/phuong-tran/coakka-publish/blob/main/docs/troubleshooting.md)
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
