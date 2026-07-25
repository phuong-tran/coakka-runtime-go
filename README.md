# Go Connector

Package Go runtime connector nằm trong thư mục `go/`.

Verify nhanh:

```bash
cd go
go test ./...
```

Live runtime integration smoke có thể bật thêm:

```bash
export COAKKA_GO_INTEGRATION=1
cd go
go test ./...
```

Integration lane này chạy helper subprocess riêng để tránh va chạm `dlopen` trong `go test` binary trên macOS.

Package smoke với embedded native runtime:

```bash
cd go
bash scripts/smoke-packaged-package.sh
```

Package release tarball:

```bash
cd go
bash scripts/package-release.sh
```

Archive được ghi ra:

```text
go/coakka-v2-connector-go-1.3.2.tar.gz
```

Public Go module export:

```bash
cd go
bash scripts/export-module-repo.sh /tmp/coakka-runtime-go-module
```

The exported directory is the root of module
`github.com/phuong-tran/coakka-runtime-go`. After that public repository exists
and is tagged as `v1.3.2`, consumers can use `go get` without a local
`replace`.

Public surface chính:

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

Truoc do, local consumer phai nhin thay ten orchestration noi bo truoc:

```go
startSpec := coakka.ConnectorStartSpec{
	SystemName: "customer-local",
	NodeID:     "customer-local-node",
	Routes:     nil,
}
requestIdentity := coakka.NewPayloadIdentity(
	"customer.create.request.v1",
	1,
	coakka.PayloadFormatJSON,
)

connector, err := coakka.StartConnectorOrchestrator(startSpec, "")
if err != nil {
	return err
}
defer connector.Close()

response, err := connector.AskJSON(
	"customer-api",
	"customer.create",
	map[string]any{"name": "Ada"},
	requestIdentity,
	2*time.Second,
	"create",
	coakka.DeliveryHintRouterDefault,
	nil,
)
```

Sau do, van la runtime single-process ay, nhung entrypoint doc dung theo vai tro
application-owned host:

```go
startSpec := coakka.ConnectorStartSpec{
	SystemName: "customer-local",
	NodeID:     "customer-local-node",
	Routes:     nil,
}
requestIdentity := coakka.NewPayloadIdentity(
	"customer.create.request.v1",
	1,
	coakka.PayloadFormatJSON,
)

runtime, err := coakka.StartRuntimeHost(startSpec, "")
if err != nil {
	return err
}
defer runtime.Close()

response, err := runtime.AskJSON(
	"customer-api",
	"customer.create",
	map[string]any{"name": "Ada"},
	requestIdentity,
	2*time.Second,
	"create",
	coakka.DeliveryHintRouterDefault,
	nil,
)
```

`StartConnectorOrchestrator` van giu cho code cu. Code local-first moi nen dung
`StartRuntimeHost` de nguoi doc thay ngay day la mot runtime host embedded trong
process hien tai, chua phai remote/Kubernetes setup.

Native runtime resolution order:

- explicit `runtimeLibPath`
- `$COAKKA_RUNTIME_LIB`
- packaged native under `native/<platform>/`
- local fallback under `lib/`

Request/reply lane trong Go hiện có hai host API shape trên cùng runtime contract:

- `Ask...`: submit rồi chờ inline
- `SubmitRequest...` + `TerminalEvents(...)`: submit trước, bắt terminal outcome (`response` hoặc `deadletter`) sau qua channel

`TerminalEvents(...)` là connector-owned API shape, không phải transport mode riêng.

Hot-path reading note:

- false-sharing hiện không phải mối lo hot-path cấp 1 ở layer Go này theo cùng
  nghĩa như native C++ connector
- cost center dễ đáng ngờ hơn hiện tại là:
  - `cgo` boundary và native read/write calls
  - protobuf marshal/unmarshal
  - channel/subscriber churn quanh `TerminalEvents(...)`
  - goroutine handoff topology
- chỉ nên quay lại cacheline/padding style hardening nếu layer này sau đó
  chuyển sang packed shared state, off-heap rings, hoặc layout nhạy cacheline hơn

Cross-language demo web lives under `examples/` when that workspace is present.
