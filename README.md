# CoAkka Runtime Go Connector

[![CI](https://github.com/phuong-tran/coakka-runtime-go/actions/workflows/go-ci.yml/badge.svg)](https://github.com/phuong-tran/coakka-runtime-go/actions/workflows/go-ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/phuong-tran/coakka-runtime-go.svg)](https://pkg.go.dev/github.com/phuong-tran/coakka-runtime-go)
[![Version](https://img.shields.io/badge/version-v1.3.10-blue)](https://github.com/phuong-tran/coakka-runtime-go/tree/v1.3.10)
[![Release](https://img.shields.io/badge/release-v1.3.10-informational)](https://github.com/phuong-tran/coakka-runtime-go/releases/tag/v1.3.10)
[![License](https://img.shields.io/badge/license-Apache--2.0-green)](LICENSE)
[![Funding](https://img.shields.io/badge/funding-Ko--fi-ff5f5f)](https://ko-fi.com/phuongnamtran)

Go module:

```sh
go get github.com/phuong-tran/coakka-runtime-go@v1.3.10
```

This package is the Go connector for CoAkka runtime v2. It embeds native
runtime generation `1.3.2+caff6d6d` for macOS, Linux, and Windows.

## License

The Go connector source is Apache-2.0 licensed. The bundled native runtime
libraries under `native/` use the CoAkka Public Artifact Preview terms in
`NATIVE-LICENSE.md`.

## New To CoAkka

CoAkka is a native-backed runtime and logger toolkit for application-owned
work. It helps an app route work by target name, handle request/reply,
deadletters, bounded queues, diagnostics, and native-backed logging without
turning every internal boundary into another hand-written HTTP endpoint.

Use these public repositories to orient first:

| Repository | Use it for | Link |
| --- | --- | --- |
| `coakka-samples` | Runnable examples and code you can inspect first. | https://github.com/phuong-tran/coakka-samples |
| `coakka-publish` | Released packages, native archives, manifests, checksums, compatibility matrix, and release notes. | https://github.com/phuong-tran/coakka-publish |
| `coakka-runtime-go` | Public Go module source for this package. | https://github.com/phuong-tran/coakka-runtime-go |
| `coakka-logger-go` | Public Go logger module source. | https://github.com/phuong-tran/coakka-logger-go |

Run the matching sample:

```sh
git clone https://github.com/phuong-tran/coakka-samples.git
cd coakka-samples
bash run.sh runtime go basic
```

Read the deeper package docs:

- [Why CoAkka Runtime matters](docs/coakka-runtime.md)
- [CoAkka ecosystem map](docs/coakka-ecosystem.md)

Try the module without cloning any CoAkka repo. The example uses the same
customer command that often becomes fake backend HTTP in a growing app:

```sh
mkdir coakka-runtime-go-first-run
cd coakka-runtime-go-first-run
go mod init coakka-runtime-go-first-run
go get github.com/phuong-tran/coakka-runtime-go@v1.3.10
```

## Quick Start

```go
package main

import (
	"encoding/json"
	"fmt"
	"time"

	coakka "github.com/phuong-tran/coakka-runtime-go"
)

func main() {
	const target = "samples.customer.store.create"
	store := map[string]map[string]any{}

	runtimeHost, err := coakka.StartRuntimeHost(coakka.ConnectorStartSpec{
		SystemName: "customer-app",
		NodeID:     "customer-app-node-1",
		Routes:     []coakka.RouteSpec{coakka.LocalRouteDefault(target)},
	}, "")
	if err != nil {
		panic(err)
	}
	defer runtimeHost.Close()

	err = runtimeHost.RegisterHandler(target, func(request *coakka.Envelope) *coakka.Envelope {
		var draft map[string]any
		_ = json.Unmarshal(request.GetPayload(), &draft)
		customer := map[string]any{
			"id":        draft["id"],
			"name":      draft["name"],
			"createdBy": request.GetSource(),
		}
		store[customer["id"].(string)] = customer

		reply, _ := coakka.MakeJSONReplyFromRequestIdentity(request, target, map[string]any{
			"status":      "created",
			"customer":    customer,
			"storedCount": len(store),
		})
		return reply
	}, true)
	if err != nil {
		panic(err)
	}

	response, err := runtimeHost.AskJSON(
		"customer-api",
		target,
		map[string]any{"id": "cust-001", "name": "Ada Lovelace"},
		coakka.NewPayloadIdentity("samples.customer.create.request.v1", 1, coakka.PayloadFormatJSON),
		2*time.Second,
		"create_customer",
		coakka.DeliveryHintRouterDefault,
		nil,
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(response)
}
```

One process owns one active runtime host. Start the host, register handlers for
targets this process owns, send typed requests to target names, then close the
host during application shutdown.

## Development

Quick verification:

```bash
cd go
go test ./...
```

Optional live runtime integration smoke:

```bash
export COAKKA_GO_INTEGRATION=1
cd go
go test ./...
```

This integration lane runs a helper subprocess to avoid `dlopen` collisions in
the `go test` binary on macOS.

Package smoke with the embedded native runtime:

```bash
cd go
bash scripts/smoke-packaged-package.sh
```

Package release tarball:

```bash
cd go
bash scripts/package-release.sh
```

The archive is written to:

```text
go/coakka-v2-connector-go-1.3.10.tar.gz
```

Public Go module export:

```bash
cd go
bash scripts/export-module-repo.sh /tmp/coakka-runtime-go-module
```

The exported directory is the root of public module
`github.com/phuong-tran/coakka-runtime-go`.

Main public surface:

- `StartRuntimeHost(startSpec, runtimeLibPath)` as the preferred single-process
  lifecycle entrypoint
- `StartConnectorOrchestrator(startSpec, runtimeLibPath)`
- `GoRuntimeClient`
- `PayloadIdentity`
- `NewTextPayloadIdentity(...)` and `PayloadFormatText` for text-first samples
- `LocalRouteDefault(...)` / `LocalRoute(...)` for same-process targets
- `SubmitRequestTyped(...)`, `SubmitRequestJSON(...)`, `SubmitRequestRaw(...)`
- `TerminalEvents(ctx, buffer)`
- `MakeJSONReplyFromRequestIdentity(...)`
- `RuntimeControlClient`
- `RuntimeMonitor`
- delivered-request lane enabled by default for request/reply hosts; set
  `DisableSeparateDeliveredRequestLane: true` only for advanced, measured,
  mostly one-way hosts

## Before / After

Before, the browser/API edge can be real HTTP, but teams often add a second
private backend HTTP endpoint only so work owned by the same app or team has
an address:

```go
func createCustomerBackendHTTP(w http.ResponseWriter, r *http.Request) {
	var draft map[string]any
	_ = json.NewDecoder(r.Body).Decode(&draft)
	customer := storeCreate(draft)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   "created",
		"customer": customer,
	})
}

func createCustomerPublicAPI(w http.ResponseWriter, r *http.Request) {
	reply, _ := http.Post(
		"http://customer-store/backend/customers",
		"application/json",
		r.Body,
	)
	defer reply.Body.Close()
	_, _ = io.Copy(w, reply.Body)
}
```

After, the public API can stay HTTP, but the fake backend URL becomes a CoAkka
target:

```go
func createCustomerPublicAPI(w http.ResponseWriter, r *http.Request) {
	var draft map[string]any
	_ = json.NewDecoder(r.Body).Decode(&draft)

	response, err := runtimeHost.AskJSON(
	"customer-api",
	"samples.customer.store.create",
	draft,
	coakka.NewPayloadIdentity("samples.customer.create.request.v1", 1, coakka.PayloadFormatJSON),
	5*time.Second,
	"create_customer",
	coakka.DeliveryHintRouterDefault,
	nil,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusGatewayTimeout)
		return
	}
	_ = json.NewEncoder(w).Encode(response)
}
```

The change is not "replace HTTP." HTTP still belongs at real browser/API or
legacy edges. CoAkka removes backend HTTP that exists only to call capabilities
owned by the same app or team by URL.

`StartConnectorOrchestrator` remains available for existing code. New examples
prefer `StartRuntimeHost` so the first screen reads as one embedded runtime
owner, not a remote connector setup.

Native runtime resolution order:

- explicit `runtimeLibPath`
- `$COAKKA_RUNTIME_LIB`
- packaged native under `native/<platform>/`
- local fallback under `lib/`

Request/reply lane in Go has two host API shapes over the same runtime contract:

- `Ask...`: submit and wait inline
- `SubmitRequest...` + `TerminalEvents(...)`: submit now, consume terminal
  outcome (`response` or `deadletter`) later through a channel

`TerminalEvents(...)` is a connector-owned API shape, not a separate transport
mode.

Hot-path reading note:

- False sharing is not the first-order hot-path concern in this Go layer in the
  same way it can be for native runtime internals.
- The more likely cost centers are the `cgo` boundary, native read/write calls,
  protobuf marshal/unmarshal work, `TerminalEvents(...)` subscriber churn, and
  goroutine handoff topology.
- Cacheline or padding hardening should be revisited only if this layer later
  owns packed shared state, off-heap rings, or other cacheline-sensitive layout.

Cross-language demo web lives under `examples/` when that workspace is present.
