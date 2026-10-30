package coakka_v2_connector

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type RuntimeLibraryResolver struct{}

// Resolve selects an explicit caller image or the package's sealed asset.
// No source-directory or unrelated-library fallback participates in loading.
func (RuntimeLibraryResolver) Resolve(explicitPath string) (string, error) {
	if explicitPath != "" {
		return requireExistingLibrary(explicitPath, "explicit runtimeLibPath")
	}
	if path := os.Getenv("COAKKA_RUNTIME_LIB"); path != "" {
		return requireExistingLibrary(path, "COAKKA_RUNTIME_LIB")
	}
	return resolvePackagedRuntimeNative()
}

func requireExistingLibrary(path string, source string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("%s: %w", source, err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%s is not a regular file: %s", source, path)
	}
	return filepath.Abs(path)
}

type RuntimeControlClient struct {
	orchestrator *ConnectorOrchestrator
}

func (c *RuntimeControlClient) ApplySnapshot(generation uint64, routes []RouteSpec, sourceConnector string, seq uint64) error {
	return c.orchestrator.ApplySnapshot(generation, routes, sourceConnector, seq)
}

func (c *RuntimeControlClient) ApplySnapshotWithPolicy(
	generation uint64,
	routes []RouteSpec,
	sourceConnector string,
	seq uint64,
	overloadPolicy *RuntimeOverloadPolicy,
) error {
	return c.orchestrator.ApplySnapshotWithPolicy(generation, routes, sourceConnector, seq, overloadPolicy)
}

type RuntimeMonitor struct {
	orchestrator *ConnectorOrchestrator
}

func (m *RuntimeMonitor) IsEnabled() (bool, error) {
	return m.orchestrator.MonitorIsEnabled()
}

func (m *RuntimeMonitor) Snapshot(signalCount uint64) (MonitorSnapshot, error) {
	return m.orchestrator.MonitorSnapshot(signalCount)
}

func (m *RuntimeMonitor) AwaitNext(timeout time.Duration) (*MonitorSnapshot, error) {
	return m.orchestrator.AwaitNextMonitor(timeout)
}

func (m *RuntimeMonitor) AwaitAppliedGenerationAtLeast(generation uint64, timeout time.Duration) (*MonitorSnapshot, error) {
	return m.orchestrator.AwaitAppliedGenerationAtLeast(generation, timeout)
}

type ConnectorOrchestrator struct {
	RuntimeLibPath string
	StartSpec      ConnectorStartSpec
	Client         *GoRuntimeClient
	Control        *RuntimeControlClient
	Monitor        *RuntimeMonitor
}

type RuntimeHost = ConnectorOrchestrator

func StartRuntimeHost(startSpec ConnectorStartSpec, runtimeLibPath string) (*RuntimeHost, error) {
	return StartConnectorOrchestrator(startSpec, runtimeLibPath)
}

func StartConnectorOrchestrator(startSpec ConnectorStartSpec, runtimeLibPath string) (*ConnectorOrchestrator, error) {
	resolvedRuntimeLib, err := (RuntimeLibraryResolver{}).Resolve(runtimeLibPath)
	if err != nil {
		return nil, err
	}
	normalizedSpec := startSpec.Normalized()
	if err := normalizedSpec.RequireValid(); err != nil {
		return nil, err
	}
	client, err := NewGoRuntimeClient(resolvedRuntimeLib, normalizedSpec.ToConnectorConfig())
	if err != nil {
		return nil, err
	}
	orchestrator := &ConnectorOrchestrator{
		RuntimeLibPath: resolvedRuntimeLib,
		StartSpec:      normalizedSpec,
		Client:         client,
	}
	orchestrator.Control = &RuntimeControlClient{orchestrator: orchestrator}
	orchestrator.Monitor = &RuntimeMonitor{orchestrator: orchestrator}
	return orchestrator, nil
}

func (o *ConnectorOrchestrator) AskTyped(source string, target string, payload []byte, payloadIdentity PayloadIdentity, timeout time.Duration, operation string, deliveryHint DeliveryHint, headers map[string]string) (*Envelope, error) {
	return o.Client.AskTyped(source, target, payload, payloadIdentity, timeout, operation, deliveryHint, headers)
}

func (o *ConnectorOrchestrator) AskJSON(source string, target string, payload any, payloadIdentity PayloadIdentity, timeout time.Duration, operation string, deliveryHint DeliveryHint, headers map[string]string) (map[string]any, error) {
	return o.Client.AskJSON(source, target, payload, payloadIdentity, timeout, operation, deliveryHint, headers)
}

func (o *ConnectorOrchestrator) AskPayload(source string, target string, payload []byte, timeout time.Duration, deliveryHint DeliveryHint) ([]byte, error) {
	return o.Client.AskPayload(source, target, payload, timeout, deliveryHint)
}

func (o *ConnectorOrchestrator) AskRaw(request *Envelope, timeout time.Duration) (*Envelope, error) {
	return o.Client.AskRaw(request, timeout)
}

func (o *ConnectorOrchestrator) SubmitRequestTyped(source string, target string, payload []byte, payloadIdentity PayloadIdentity, timeout time.Duration, operation string, deliveryHint DeliveryHint, headers map[string]string) (*SubmittedRequest, error) {
	return o.Client.SubmitRequestTyped(source, target, payload, payloadIdentity, timeout, operation, deliveryHint, headers)
}

func (o *ConnectorOrchestrator) SubmitRequestJSON(source string, target string, payload any, payloadIdentity PayloadIdentity, timeout time.Duration, operation string, deliveryHint DeliveryHint, headers map[string]string) (*SubmittedRequest, error) {
	return o.Client.SubmitRequestJSON(source, target, payload, payloadIdentity, timeout, operation, deliveryHint, headers)
}

func (o *ConnectorOrchestrator) SubmitRequestRaw(request *Envelope) (*SubmittedRequest, error) {
	return o.Client.SubmitRequestRaw(request)
}

func (o *ConnectorOrchestrator) TerminalEvents(ctx context.Context, buffer int) (<-chan RequestTerminalEvent, error) {
	return o.Client.TerminalEvents(ctx, buffer)
}

func (o *ConnectorOrchestrator) Deadletters(ctx context.Context, buffer int) (<-chan ObservedDeadletter, error) {
	return o.Client.Deadletters(ctx, buffer)
}

func (o *ConnectorOrchestrator) SendOneWayTyped(source string, target string, payload []byte, payloadIdentity PayloadIdentity, deliveryHint DeliveryHint, headers map[string]string) error {
	return o.Client.SendOneWayTyped(source, target, payload, payloadIdentity, deliveryHint, headers)
}

func (o *ConnectorOrchestrator) SendOneWayJSON(source string, target string, payload any, payloadIdentity PayloadIdentity, deliveryHint DeliveryHint, headers map[string]string) error {
	return o.Client.SendOneWayJSON(source, target, payload, payloadIdentity, deliveryHint, headers)
}

func (o *ConnectorOrchestrator) SubmitEnvelope(envelope *Envelope) error {
	return o.Client.SubmitEnvelope(envelope)
}

func (o *ConnectorOrchestrator) SubmitTypedEnvelope(envelope *Envelope) error {
	return o.Client.SubmitTypedEnvelope(envelope)
}

func (o *ConnectorOrchestrator) SubmitRawEnvelope(envelope *Envelope) error {
	return o.Client.SubmitRawEnvelope(envelope)
}

func (o *ConnectorOrchestrator) RegisterHandler(target string, handler HandlerFn, typedReplies bool) error {
	return o.Client.RegisterHandler(target, handler, typedReplies)
}

func (o *ConnectorOrchestrator) RegisterRawHandler(target string, handler HandlerFn) error {
	return o.Client.RegisterRawHandler(target, handler)
}

func (o *ConnectorOrchestrator) RegisterPayloadHandler(target string, handler PayloadHandlerFn) error {
	return o.Client.RegisterPayloadHandler(target, handler)
}

func (o *ConnectorOrchestrator) ClientStats() RuntimeClientStats {
	return o.Client.SnapshotStats()
}

func (o *ConnectorOrchestrator) RuntimeInfo() (RuntimeInfoSnapshot, error) {
	return o.Client.RuntimeInfoSnapshot()
}

func (o *ConnectorOrchestrator) RuntimeConfig() (RuntimeConfigSnapshot, error) {
	return o.Client.RuntimeConfigSnapshot()
}

func (o *ConnectorOrchestrator) StartupConfig() (RuntimeStartupConfigSnapshot, error) {
	return o.Client.RuntimeStartupConfigSnapshot()
}

func (o *ConnectorOrchestrator) Health() (RuntimeHealthSnapshot, error) {
	return o.Client.HealthSnapshot()
}

func (o *ConnectorOrchestrator) Stats() (RuntimeStatsSnapshot, error) {
	return o.Client.StatsSnapshot()
}

func (o *ConnectorOrchestrator) RuntimeSnapshot() (RuntimeSnapshot, error) {
	return o.Client.RuntimeSnapshot()
}

func (o *ConnectorOrchestrator) ApplySnapshot(generation uint64, routes []RouteSpec, sourceConnector string, seq uint64) error {
	return o.Client.ApplySnapshot(generation, routes, sourceConnector, seq)
}

func (o *ConnectorOrchestrator) ApplySnapshotWithPolicy(
	generation uint64,
	routes []RouteSpec,
	sourceConnector string,
	seq uint64,
	overloadPolicy *RuntimeOverloadPolicy,
) error {
	return o.Client.ApplySnapshotWithPolicy(generation, routes, sourceConnector, seq, overloadPolicy)
}

func (o *ConnectorOrchestrator) MonitorIsEnabled() (bool, error) {
	return o.Client.MonitorIsEnabled()
}

func (o *ConnectorOrchestrator) MonitorSnapshot(signalCount uint64) (MonitorSnapshot, error) {
	return o.Client.MonitorSnapshot(signalCount)
}

func (o *ConnectorOrchestrator) AwaitNextMonitor(timeout time.Duration) (*MonitorSnapshot, error) {
	return o.Client.AwaitNextMonitor(timeout)
}

func (o *ConnectorOrchestrator) AwaitAppliedGenerationAtLeast(generation uint64, timeout time.Duration) (*MonitorSnapshot, error) {
	return o.Client.AwaitAppliedGenerationAtLeast(generation, timeout)
}

func (o *ConnectorOrchestrator) Close() error { return o.Client.Close() }
