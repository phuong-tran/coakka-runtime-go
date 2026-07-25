# Consuming The Go Runtime Package

The current Go runtime artifact is a source package plus embedded native
runtime libraries. The module path is already fixed as
`github.com/phuong-tran/coakka-runtime-go`, but the public Go module repository
must exist and be tagged before users can install it with `go get`.

Until that repository is opened, extract the archive and use a local `replace`.

Example consumer `go.mod`:

```go
module my-runtime-consumer

go 1.23.0

require github.com/phuong-tran/coakka-runtime-go v0.0.0

replace github.com/phuong-tran/coakka-runtime-go => ./coakka-v2-connector-go-1.3.2
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

After the public module repository is opened and tagged, the local `replace`
line goes away and consumers should use:

```sh
go get github.com/phuong-tran/coakka-runtime-go@v1.3.2
```
