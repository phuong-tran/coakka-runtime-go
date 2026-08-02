package coakka_v2_connector

import "fmt"

func ReadRuntimeCapabilities(runtimeLibPath string) (RuntimeCapabilitiesSnapshot, error) {
	resolvedRuntimeLib, err := (RuntimeLibraryResolver{}).Resolve(runtimeLibPath, nil)
	if err != nil {
		return RuntimeCapabilitiesSnapshot{}, err
	}
	bindings, err := openNativeBindings(resolvedRuntimeLib)
	if err != nil {
		return RuntimeCapabilitiesSnapshot{}, err
	}
	defer bindings.close()
	if bindings.getAbiVersion() != COAKKAABIVersion {
		return RuntimeCapabilitiesSnapshot{}, fmt.Errorf(
			"unexpected ABI version: %d",
			bindings.getAbiVersion(),
		)
	}
	return bindings.readRuntimeCapabilities()
}

func (c *GoRuntimeClient) RuntimeCapabilities() (RuntimeCapabilitiesSnapshot, error) {
	if err := c.requireTransportOpen(); err != nil {
		return RuntimeCapabilitiesSnapshot{}, err
	}
	c.transportMu.Lock()
	defer c.transportMu.Unlock()
	if err := c.requireTransportOpen(); err != nil {
		return RuntimeCapabilitiesSnapshot{}, err
	}
	return c.bindings.readRuntimeCapabilities()
}

func (c *GoRuntimeClient) TCPConnectionConfig() (RuntimeTCPConnectionConfigSnapshot, error) {
	if err := c.requireTransportOpen(); err != nil {
		return RuntimeTCPConnectionConfigSnapshot{}, err
	}
	c.transportMu.Lock()
	defer c.transportMu.Unlock()
	if err := c.requireTransportOpen(); err != nil {
		return RuntimeTCPConnectionConfigSnapshot{}, err
	}
	return c.bindings.getTCPConnectionConfig(c.runtime)
}

func (c *GoRuntimeClient) TCPSecurityInfo() (RuntimeTCPSecurityInfoSnapshot, error) {
	if err := c.requireTransportOpen(); err != nil {
		return RuntimeTCPSecurityInfoSnapshot{}, err
	}
	c.transportMu.Lock()
	defer c.transportMu.Unlock()
	if err := c.requireTransportOpen(); err != nil {
		return RuntimeTCPSecurityInfoSnapshot{}, err
	}
	return c.bindings.getTCPSecurityInfo(c.runtime)
}

func (c *GoRuntimeClient) StartupTCPConnectionResult() (RuntimeTCPConnectionApplyResult, bool) {
	if c.startupConnectionResult == nil {
		return RuntimeTCPConnectionApplyResult{}, false
	}
	return *c.startupConnectionResult, true
}

func (c *GoRuntimeClient) StartupTCPSecurityResult() (RuntimeTCPSecurityApplyResult, bool) {
	if c.startupSecurityResult == nil {
		return RuntimeTCPSecurityApplyResult{}, false
	}
	return *c.startupSecurityResult, true
}

func (c *GoRuntimeClient) ApplyTCPConnectionStrategy(
	spec RuntimeTCPConnectionStrategySpec,
) (RuntimeTCPConnectionApplyResult, error) {
	if err := c.requireTransportOpen(); err != nil {
		return RuntimeTCPConnectionApplyResult{}, err
	}
	c.transportMu.Lock()
	defer c.transportMu.Unlock()
	if err := c.requireTransportOpen(); err != nil {
		return RuntimeTCPConnectionApplyResult{}, err
	}
	return c.bindings.applyTCPConnectionStrategy(c.runtime, spec)
}

func (c *GoRuntimeClient) ApplyTCPSecurity(
	spec RuntimeTCPSecuritySpec,
) (RuntimeTCPSecurityApplyResult, error) {
	if err := c.requireTransportOpen(); err != nil {
		return RuntimeTCPSecurityApplyResult{}, err
	}
	c.transportMu.Lock()
	defer c.transportMu.Unlock()
	if err := c.requireTransportOpen(); err != nil {
		return RuntimeTCPSecurityApplyResult{}, err
	}
	return c.bindings.applyTCPSecurity(c.runtime, spec)
}

func (c *GoRuntimeClient) requireTransportOpen() error {
	select {
	case <-c.closedCh:
		return fmt.Errorf("runtime client closed")
	default:
		return nil
	}
}

func (o *ConnectorOrchestrator) RuntimeCapabilities() (RuntimeCapabilitiesSnapshot, error) {
	return o.Client.RuntimeCapabilities()
}

func (o *ConnectorOrchestrator) TCPConnectionConfig() (RuntimeTCPConnectionConfigSnapshot, error) {
	return o.Client.TCPConnectionConfig()
}

func (o *ConnectorOrchestrator) TCPSecurityInfo() (RuntimeTCPSecurityInfoSnapshot, error) {
	return o.Client.TCPSecurityInfo()
}

func (o *ConnectorOrchestrator) StartupTCPConnectionResult() (RuntimeTCPConnectionApplyResult, bool) {
	return o.Client.StartupTCPConnectionResult()
}

func (o *ConnectorOrchestrator) StartupTCPSecurityResult() (RuntimeTCPSecurityApplyResult, bool) {
	return o.Client.StartupTCPSecurityResult()
}

func (o *ConnectorOrchestrator) ApplyTCPConnectionStrategy(
	spec RuntimeTCPConnectionStrategySpec,
) (RuntimeTCPConnectionApplyResult, error) {
	return o.Client.ApplyTCPConnectionStrategy(spec)
}

func (o *ConnectorOrchestrator) ApplyTCPSecurity(
	spec RuntimeTCPSecuritySpec,
) (RuntimeTCPSecurityApplyResult, error) {
	return o.Client.ApplyTCPSecurity(spec)
}
