// Package coakka_v2_connector is the Go connector for CoAkka runtime v2.
//
// Install the public module with:
//
//	go get github.com/phuong-tran/coakka-runtime-go@v1.3.4
//
// CoAkka is a native-backed runtime and logger toolkit for application-owned
// work. It lets an app route work by target name, handle request/reply,
// observe deadletters, use bounded queues, and read diagnostics without turning
// every internal boundary into another hand-written HTTP endpoint.
//
// Start one runtime host per process, register handlers only for target names
// this process owns, send typed requests to target names, and close the host
// during application shutdown.
//
//	runtimeHost, err := coakka.StartRuntimeHost(coakka.ConnectorStartSpec{
//		SystemName: "sample",
//		NodeID:     "sample-node",
//		Routes:     []coakka.RouteSpec{coakka.LocalRouteDefault("samples.echo")},
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
