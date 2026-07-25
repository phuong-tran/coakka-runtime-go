// Package coakka_v2_connector is the Go connector for CoAkka runtime v2.
//
// Install the public module with:
//
//	go get github.com/phuong-tran/coakka-runtime-go@v1.3.8
//
// CoAkka is a native-backed runtime and logger toolkit for application-owned
// work. It lets an app route work by target name, handle request/reply,
// observe deadletters, use bounded queues, and read diagnostics without turning
// every internal boundary into another hand-written HTTP endpoint.
//
// Use it when the browser/API edge can stay HTTP, but backend-to-backend HTTP
// exists only because a capability owned by the same app or team needed a URL.
// In the after shape, callers ask a CoAkka target such as
// "samples.customer.store.create" and receive an explicit reply or deadletter.
//
// Start one runtime host per process, register handlers only for target names
// this process owns, send typed requests to target names, and close the host
// during application shutdown.
//
//	runtimeHost, err := coakka.StartRuntimeHost(coakka.ConnectorStartSpec{
//		SystemName: "customer-app",
//		NodeID:     "customer-app-node-1",
//		Routes:     []coakka.RouteSpec{coakka.LocalRouteDefault("samples.customer.store.create")},
//	}, "")
//	if err != nil {
//		panic(err)
//	}
//	defer runtimeHost.Close()
//
// Public repositories:
//
//   - https://github.com/phuong-tran/coakka-runtime-go
//   - https://github.com/phuong-tran/coakka-logger-go
//   - https://github.com/phuong-tran/coakka-publish
//   - https://github.com/phuong-tran/coakka-samples
package coakka_v2_connector
