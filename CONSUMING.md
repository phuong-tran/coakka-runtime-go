# Consuming The Go Runtime Package

The current Go runtime package is a public Go module plus embedded native
runtime libraries:

```sh
go get github.com/phuong-tran/coakka-runtime-go@v1.3.10
```

Example consumer `go.mod`:

```go
module my-runtime-consumer

go 1.23.0

require github.com/phuong-tran/coakka-runtime-go v1.3.10
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

Current packaged platforms:

- `macos-aarch64`
- `linux-aarch64`
- `linux-x86_64`
- `windows-aarch64`
- `windows-x86_64`

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
