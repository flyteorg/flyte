/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package integration tests the runs → actions → executor CRD handoff on a
// real apiserver: the ActionsClient (actions/k8s) creates TaskAction CRs and
// the TaskActionReconciler consumes them, with a recording fake standing in
// for the InternalRunService. See docs/PROPOSAL_EXECUTOR_INTEGRATION_TESTS.md.
package integration

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel/metric/noop"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	actionsk8s "github.com/flyteorg/flyte/v2/actions/k8s"
	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/executor/pkg/controller"
	"github.com/flyteorg/flyte/v2/executor/pkg/plugin"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/v2/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"

	// Register the pod plugin so the registry can resolve container/python task types.
	_ "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/plugins/k8s/pod"
)

var (
	ctx           context.Context
	testEnv       *envtest.Environment
	k8sClient     client.WithWatch
	reconciler    *controller.TaskActionReconciler
	actionsClient *actionsk8s.ActionsClient
	runClient     *recordingRunClient
)

// TestMain stands up one envtest apiserver and wires BOTH halves of the
// handoff against it, mirroring production wiring in executor/setup.go and
// actions/setup.go.
func TestMain(m *testing.M) {
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())

	if err := flyteorgv1.AddToScheme(scheme.Scheme); err != nil {
		log.Fatalf("Failed to add TaskAction scheme: %v", err)
	}

	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}
	// Allow running from IDEs without KUBEBUILDER_ASSETS (same as the controller suite).
	if dir := firstEnvTestBinaryDir(); dir != "" {
		testEnv.BinaryAssetsDirectory = dir
	}

	cfg, err := testEnv.Start()
	if err != nil {
		log.Fatalf("Failed to start envtest: %v", err)
	}

	exitCode := 1
	defer func() {
		cancel()
		if err := testEnv.Stop(); err != nil {
			log.Printf("Warning: failed to stop envtest: %v", err)
		}
		os.Exit(exitCode)
	}()

	k8sClient, err = client.NewWithWatch(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		log.Printf("Failed to create k8s client: %v", err)
		return
	}

	// The manager provides the shared informer cache used by both the plugin
	// registry (executor side) and the ActionsClient watcher (actions side).
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:  scheme.Scheme,
		Metrics: metricsserver.Options{BindAddress: "0"},
	})
	if err != nil {
		log.Printf("Failed to create manager: %v", err)
		return
	}
	go func() {
		if err := mgr.Start(ctx); err != nil {
			log.Printf("Manager exited with error: %v", err)
		}
	}()
	if !mgr.GetCache().WaitForCacheSync(ctx) {
		log.Printf("Cache failed to sync")
		return
	}

	labeled.SetMetricKeys(
		contextutils.ProjectKey,
		contextutils.DomainKey,
		contextutils.WorkflowIDKey,
		contextutils.TaskIDKey,
	)

	// Executor half: plugin registry (real pod plugin) + reconciler.
	setupCtx := plugin.NewSetupContext(
		mgr, nil, nil, nil, nil,
		"TaskAction",
		promutils.NewScope("integration"),
	)
	registry := plugin.NewRegistry(setupCtx, pluginmachinery.PluginRegistry())
	if err := registry.Initialize(ctx); err != nil {
		log.Printf("Failed to initialize plugin registry: %v", err)
		return
	}

	dataStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewScope("integration_storage"))
	if err != nil {
		log.Printf("Failed to create data store: %v", err)
		return
	}

	reconciler = controller.NewTaskActionReconciler(
		k8sClient, scheme.Scheme, registry, dataStore,
		&fakeEventsClient{}, "", noop.NewMeterProvider(), mgr.GetCache(),
	)
	reconciler.Recorder = events.NewFakeRecorder(100)

	// Actions half: the client that Enqueues CRs and watches them back.
	runClient = &recordingRunClient{}
	var acErr error
	actionsClient, acErr = actionsk8s.NewActionsClient(k8sClient, mgr.GetCache(), "flyte", 32, 1, runClient, 0, promutils.NewScope("integration_actions"))
	if acErr != nil {
		log.Printf("Failed to create actions client: %v", acErr)
		return
	}
	if err := actionsClient.StartWatching(ctx); err != nil {
		log.Printf("Failed to start actions watcher: %v", err)
		return
	}

	exitCode = m.Run()
}

// firstEnvTestBinaryDir locates envtest binaries under executor/bin/k8s for
// IDE runs; CI sets KUBEBUILDER_ASSETS instead.
func firstEnvTestBinaryDir() string {
	basePath := filepath.Join("..", "..", "bin", "k8s")
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return filepath.Join(basePath, entry.Name())
		}
	}
	return ""
}

// fakeEventsClient is a no-op EventsProxyServiceClient for the reconciler.
type fakeEventsClient struct{}

func (f *fakeEventsClient) Record(_ context.Context, _ *connect.Request[workflow.RecordRequest]) (*connect.Response[workflow.RecordResponse], error) {
	return connect.NewResponse(&workflow.RecordResponse{}), nil
}

// recordingRunClient captures what the ActionsClient forwards to the internal
// run service — the DB-write half of the handoff under test. Streaming
// methods are unused by the ActionsClient and left unimplemented.
type recordingRunClient struct {
	mu            sync.Mutex
	recorded      []*workflow.RecordActionRequest
	statusUpdates []*workflow.UpdateActionStatusRequest
}

func (r *recordingRunClient) RecordAction(_ context.Context, req *connect.Request[workflow.RecordActionRequest]) (*connect.Response[workflow.RecordActionResponse], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recorded = append(r.recorded, req.Msg)
	return connect.NewResponse(&workflow.RecordActionResponse{}), nil
}

func (r *recordingRunClient) UpdateActionStatus(_ context.Context, req *connect.Request[workflow.UpdateActionStatusRequest]) (*connect.Response[workflow.UpdateActionStatusResponse], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.statusUpdates = append(r.statusUpdates, req.Msg)
	return connect.NewResponse(&workflow.UpdateActionStatusResponse{}), nil
}

func (r *recordingRunClient) RecordActionEvents(_ context.Context, _ *connect.Request[workflow.RecordActionEventsRequest]) (*connect.Response[workflow.RecordActionEventsResponse], error) {
	return connect.NewResponse(&workflow.RecordActionEventsResponse{}), nil
}

func (r *recordingRunClient) RecordActionStream(_ context.Context) *connect.BidiStreamForClient[workflow.RecordActionStreamRequest, workflow.RecordActionStreamResponse] {
	panic("not used in tests")
}

func (r *recordingRunClient) UpdateActionStatusStream(_ context.Context) *connect.BidiStreamForClient[workflow.UpdateActionStatusStreamRequest, workflow.UpdateActionStatusStreamResponse] {
	panic("not used in tests")
}

func (r *recordingRunClient) RecordActionEventStream(_ context.Context) *connect.BidiStreamForClient[workflow.RecordActionEventStreamRequest, workflow.RecordActionEventStreamResponse] {
	panic("not used in tests")
}

func (r *recordingRunClient) recordedActions() []*workflow.RecordActionRequest {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*workflow.RecordActionRequest, len(r.recorded))
	copy(out, r.recorded)
	return out
}

func (r *recordingRunClient) recordedStatusUpdates() []*workflow.UpdateActionStatusRequest {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*workflow.UpdateActionStatusRequest, len(r.statusUpdates))
	copy(out, r.statusUpdates)
	return out
}
