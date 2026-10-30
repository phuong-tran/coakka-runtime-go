package coakka_v2_connector

/*
#include "runtime_bridge.h"
*/
import "C"

// RuntimeNetworkSnapshot contains accepted Core facts, not input preferences.
type RuntimeNetworkSnapshot struct {
	Mode           RuntimeNetworkMode `json:"mode"`
	ExplicitPolicy bool               `json:"explicitPolicy"`
	BindHost       string             `json:"bindHost"`
	BindPort       uint16             `json:"bindPort"`
	AdvertiseHost  string             `json:"advertiseHost"`
	AdvertisePort  uint16             `json:"advertisePort"`
}

func (b *nativeBindings) getNetworkConfig(runtime nativeRuntime) (RuntimeNetworkSnapshot, error) {
	value := C.coakka_v2_network_config_snapshot_t{struct_size: C.sizeof_coakka_v2_network_config_snapshot_t}
	status := C.coakka_v2_go_get_network_config(b.ptr, (*C.coakka_v2_runtime_t)(runtime), &value)
	if err := requireStatus(status, "runtime_get_network_config"); err != nil {
		return RuntimeNetworkSnapshot{}, err
	}
	if value.explicit_policy > 1 {
		return RuntimeNetworkSnapshot{}, &CoreContractError{Operation: "runtime_get_network_config", Detail: "invalid explicit boolean"}
	}
	// Core strings borrow until apply/destroy. The caller serializes inspection
	// with close; network policy is immutable after startup.
	return RuntimeNetworkSnapshot{Mode: RuntimeNetworkMode(value.mode), ExplicitPolicy: value.explicit_policy != 0,
		BindHost: C.GoString(value.bind_host), BindPort: uint16(value.bind_port),
		AdvertiseHost: C.GoString(value.advertise_host), AdvertisePort: uint16(value.advertise_port)}, nil
}

func (c *GoRuntimeClient) NetworkConfig() (RuntimeNetworkSnapshot, error) {
	if !c.beginCall() {
		return RuntimeNetworkSnapshot{}, ErrClosed
	}
	defer c.endCall()
	c.transportMu.Lock()
	defer c.transportMu.Unlock()
	return c.bindings.getNetworkConfig(c.runtime)
}

func (o *ConnectorOrchestrator) NetworkConfig() (RuntimeNetworkSnapshot, error) {
	return o.Client.NetworkConfig()
}
