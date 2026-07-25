package coakka_v2_connector

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var (
	orchestratorFactoryMu sync.Mutex
	activeOrchestrator    *ConnectorOrchestrator
)

var defaultLibraryCandidates = []string{
	"libcoakka_runtime_v2.dll",
	"libcoakka_runtime_v2.dylib",
	"libcoakka_runtime_v2.so",
}

type RuntimeLibraryResolver struct{}

func (RuntimeLibraryResolver) Resolve(explicitPath string, candidateNames []string) (string, error) {
	if len(candidateNames) == 0 {
		candidateNames = defaultLibraryCandidates
	}
	if strings := explicitPath; strings != "" {
		return requireExistingLibrary(filepath.Clean(strings), "explicit runtimeLibPath")
	}
	if configuredPath := filepath.Clean(os.Getenv("COAKKA_RUNTIME_LIB")); configuredPath != "." && configuredPath != "" {
		return requireExistingLibrary(configuredPath, "$COAKKA_RUNTIME_LIB")
	}
	if embedded, ok := resolvePackagedRuntimeNative(); ok {
		return embedded, nil
	}
	for _, candidate := range searchLibraryCandidates(candidateNames) {
		if _, err := os.Stat(candidate); err == nil {
			return filepath.Abs(candidate)
		}
	}
	return "", fmt.Errorf(
		"native runtime library was not found. Set COAKKA_RUNTIME_LIB, pass runtimeLibPath explicitly, package one of %v under native/%s, or place one under a repo-local lib/ directory",
		runtimeResourceFileNames(runtime.GOOS),
		RuntimePlatformID(runtime.GOOS, runtime.GOARCH),
	)
}

func searchLibraryCandidates(candidateNames []string) []string {
	cwd, _ := os.Getwd()
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile()), ".."))
	roots := []string{cwd, filepath.Join(cwd, "lib"), repoRoot, filepath.Join(repoRoot, "lib")}
	var out []string
	seen := map[string]struct{}{}
	for _, root := range roots {
		for _, name := range candidateNames {
			candidate := filepath.Join(root, name)
			if filepath.Base(root) != "lib" {
				candidate = filepath.Join(root, "lib", name)
			}
			if _, ok := seen[candidate]; !ok {
				out = append(out, candidate)
				seen[candidate] = struct{}{}
			}
		}
	}
	return out
}

func requireExistingLibrary(path string, source string) (string, error) {
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("%s does not exist: %s", source, path)
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

func (m *RuntimeMonitor) IsEnabled() bool {
	return m.orchestrator.MonitorIsEnabled()
}

func (m *RuntimeMonitor) Snapshot(signalCount uint64) MonitorSnapshot {
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
	closed         bool
}

type RuntimeHost = ConnectorOrchestrator

func StartRuntimeHost(startSpec ConnectorStartSpec, runtimeLibPath string) (*RuntimeHost, error) {
	return StartConnectorOrchestrator(startSpec, runtimeLibPath)
}

func StartConnectorOrchestrator(startSpec ConnectorStartSpec, runtimeLibPath string) (*ConnectorOrchestrator, error) {
	resolvedRuntimeLib, err := (RuntimeLibraryResolver{}).Resolve(runtimeLibPath, nil)
	if err != nil {
		return nil, err
	}
	normalizedSpec := startSpec.Normalized()
	if err := normalizedSpec.RequireValid(); err != nil {
		return nil, err
	}
	orchestratorFactoryMu.Lock()
	defer orchestratorFactoryMu.Unlock()
	if activeOrchestrator != nil {
		return nil, errors.New("ConnectorOrchestrator already started for this Go process")
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
	activeOrchestrator = orchestrator
	return orchestrator, nil
}

func (o *ConnectorOrchestrator) AskTyped(source string, target string, payload []byte, payloadIdentity PayloadIdentity, timeout time.Duration, operation string, deliveryHint DeliveryHint, headers map[string]string) (*Envelope, error) {
	return o.Client.AskTyped(source, target, payload, payloadIdentity, timeout, operation, deliveryHint, headers)
}

func (o *ConnectorOrchestrator) AskJSON(source string, target string, payload any, payloadIdentity PayloadIdentity, timeout time.Duration, operation string, deliveryHint DeliveryHint, headers map[string]string) (map[string]any, error) {
	return o.Client.AskJSON(source, target, payload, payloadIdentity, timeout, operation, deliveryHint, headers)
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

func (o *ConnectorOrchestrator) TerminalEvents(ctx context.Context, buffer int) <-chan RequestTerminalEvent {
	return o.Client.TerminalEvents(ctx, buffer)
}

func (o *ConnectorOrchestrator) Deadletters(ctx context.Context, buffer int) <-chan ObservedDeadletter {
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

func (o *ConnectorOrchestrator) ClientStats() RuntimeClientStats {
	return o.Client.SnapshotStats()
}

func (o *ConnectorOrchestrator) RuntimeInfo() RuntimeInfoSnapshot {
	return o.Client.RuntimeInfoSnapshot()
}

func (o *ConnectorOrchestrator) RuntimeConfig() RuntimeConfigSnapshot {
	return o.Client.RuntimeConfigSnapshot()
}

func (o *ConnectorOrchestrator) Health() RuntimeHealthSnapshot {
	return o.Client.HealthSnapshot()
}

func (o *ConnectorOrchestrator) Stats() RuntimeStatsSnapshot {
	return o.Client.StatsSnapshot()
}

func (o *ConnectorOrchestrator) RuntimeSnapshot() RuntimeSnapshot {
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

func (o *ConnectorOrchestrator) MonitorIsEnabled() bool {
	return o.Client.MonitorIsEnabled()
}

func (o *ConnectorOrchestrator) MonitorSnapshot(signalCount uint64) MonitorSnapshot {
	return o.Client.MonitorSnapshot(signalCount)
}

func (o *ConnectorOrchestrator) AwaitNextMonitor(timeout time.Duration) (*MonitorSnapshot, error) {
	return o.Client.AwaitNextMonitor(timeout)
}

func (o *ConnectorOrchestrator) AwaitAppliedGenerationAtLeast(generation uint64, timeout time.Duration) (*MonitorSnapshot, error) {
	return o.Client.AwaitAppliedGenerationAtLeast(generation, timeout)
}

func (o *ConnectorOrchestrator) Close() error {
	orchestratorFactoryMu.Lock()
	defer orchestratorFactoryMu.Unlock()
	if o.closed {
		return nil
	}
	o.closed = true
	if activeOrchestrator == o {
		activeOrchestrator = nil
	}
	return o.Client.Close()
}

func currentFile() string {
	_, file, _, _ := runtime.Caller(0)
	return file
}
