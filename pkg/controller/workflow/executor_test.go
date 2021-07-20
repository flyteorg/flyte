package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"
	pluginCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flytestdlib/logger"

	eventsErr "github.com/flyteorg/flyteidl/clients/go/events/errors"

	mocks2 "github.com/flyteorg/flytepropeller/pkg/controller/executors/mocks"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/catalog"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/fakeplugins"

	wfErrors "github.com/flyteorg/flytepropeller/pkg/controller/workflow/errors"

	"github.com/flyteorg/flyteidl/clients/go/events"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/flyteorg/flytestdlib/yamlutils"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/record"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes"
	recoveryMocks "github.com/flyteorg/flytepropeller/pkg/controller/nodes/recovery/mocks"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
)

var (
	testScope      = promutils.NewScope("test_wfexec")
	fakeKubeClient = mocks2.NewFakeKubeClient()
)

const (
	maxOutputSize = 10 * 1024
)

type fakeRemoteWritePlugin struct {
	pluginCore.Plugin
	enableAsserts bool
	t             assert.TestingT
}

func (f fakeRemoteWritePlugin) Handle(ctx context.Context, tCtx pluginCore.TaskExecutionContext) (pluginCore.Transition, error) {
	logger.Infof(ctx, "----------------------------------------------------------------------------------------------")
	logger.Infof(ctx, "Handle called for %s", tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName())

	defer func() {
		logger.Infof(ctx, "Handle completed for %s", tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName())
		logger.Infof(ctx, "----------------------------------------------------------------------------------------------")
	}()
	trns, err := f.Plugin.Handle(ctx, tCtx)
	if err != nil {
		return trns, err
	}
	if trns.Info().Phase() == pluginCore.PhaseSuccess {
		tk, err := tCtx.TaskReader().Read(ctx)
		assert.NoError(f.t, err)
		outputVars := tk.GetInterface().Outputs.Variables
		o := &core.LiteralMap{
			Literals: make(map[string]*core.Literal, len(outputVars)),
		}
		for k, v := range outputVars {
			l, err := coreutils.MakeDefaultLiteralForType(v.Type)
			if f.enableAsserts && !assert.NoError(f.t, err) {
				assert.FailNow(f.t, "Failed to create default output for node [%v] Type [%v]", tCtx.TaskExecutionMetadata().GetTaskExecutionID(), v.Type)
			}
			o.Literals[k] = l
		}
		assert.NoError(f.t, tCtx.DataStore().WriteProtobuf(ctx, tCtx.OutputWriter().GetOutputPath(), storage.Options{}, o))
		assert.NoError(f.t, tCtx.OutputWriter().Put(ctx, ioutils.NewRemoteFileOutputReader(ctx, tCtx.DataStore(), tCtx.OutputWriter(), tCtx.MaxDatasetSizeBytes())))
	}
	return trns, err
}

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey,
		contextutils.TaskIDKey)
}

func createInmemoryDataStore(t testing.TB, scope promutils.Scope) *storage.DataStore {
	cfg := storage.Config{
		Type: storage.TypeMemory,
	}
	d, err := storage.NewDataStore(&cfg, scope)
	assert.NoError(t, err)
	return d
}

func StdOutEventRecorder() record.EventRecorder {
	eventChan := make(chan string)
	recorder := &record.FakeRecorder{
		Events: eventChan,
	}

	go func() {
		defer close(eventChan)
		for {
			s := <-eventChan
			if s == "" {
				return
			}
			fmt.Printf("Event: [%v]\n", s)
		}
	}()
	return recorder
}

func createHappyPathTaskExecutor(t assert.TestingT, enableAsserts bool) pluginCore.PluginEntry {
	f := func(ctx context.Context, iCtx pluginCore.SetupContext) (plugin pluginCore.Plugin, e error) {
		return fakeRemoteWritePlugin{
			Plugin: fakeplugins.NewReplayer(
				"test",
				pluginCore.PluginProperties{},
				[]fakeplugins.HandleResponse{
					fakeplugins.NewHandleTransition(pluginCore.DoTransition(pluginCore.PhaseInfoRunning(1, nil))),
					fakeplugins.NewHandleTransition(pluginCore.DoTransition(pluginCore.PhaseInfoSuccess(nil))),
				},
				[]error{
					nil,
				},
				[]error{
					nil,
				},
			),
			enableAsserts: enableAsserts,
			t:             t,
		}, nil
	}
	return pluginCore.PluginEntry{
		ID:                  "test",
		RegisteredTaskTypes: []string{"7"},
		LoadPlugin:          f,
		IsDefault:           true,
	}
}

func createFailingTaskExecutor(t assert.TestingT) pluginCore.PluginEntry {
	f := func(ctx context.Context, iCtx pluginCore.SetupContext) (plugin pluginCore.Plugin, e error) {
		return fakeRemoteWritePlugin{
			Plugin: fakeplugins.NewReplayer(
				"test",
				pluginCore.PluginProperties{},
				[]fakeplugins.HandleResponse{
					fakeplugins.NewHandleTransition(pluginCore.DoTransition(pluginCore.PhaseInfoRunning(1, nil))),
					fakeplugins.NewHandleTransition(pluginCore.DoTransition(pluginCore.PhaseInfoFailure("code", "message", nil))),
				},
				[]error{
					nil,
				},
				[]error{
					nil,
				},
			),
			enableAsserts: false,
			t:             t,
		}, nil
	}
	return pluginCore.PluginEntry{
		ID:                  "test",
		RegisteredTaskTypes: []string{"7"},
		LoadPlugin:          f,
		IsDefault:           true,
	}
}

func createTaskExecutorErrorInCheck(t assert.TestingT) pluginCore.PluginEntry {
	f := func(ctx context.Context, iCtx pluginCore.SetupContext) (plugin pluginCore.Plugin, e error) {
		return fakeRemoteWritePlugin{
			Plugin: fakeplugins.NewReplayer(
				"test",
				pluginCore.PluginProperties{},
				[]fakeplugins.HandleResponse{
					fakeplugins.NewHandleTransition(pluginCore.DoTransition(pluginCore.PhaseInfoRunning(1, nil))),
					fakeplugins.NewHandleError(fmt.Errorf("error")),
				},
				[]error{
					nil,
				},
				[]error{
					nil,
				},
			),
			enableAsserts: false,
			t:             t,
		}, nil
	}
	return pluginCore.PluginEntry{
		ID:                  "test",
		RegisteredTaskTypes: []string{"7"},
		LoadPlugin:          f,
		IsDefault:           true,
	}
}

func TestWorkflowExecutor_HandleFlyteWorkflow_Error(t *testing.T) {
	ctx := context.Background()
	store := createInmemoryDataStore(t, testScope.NewSubScope("12"))
	recorder := StdOutEventRecorder()
	_, err := events.ConstructEventSink(ctx, events.GetConfig(ctx))
	assert.NoError(t, err)

	te := createTaskExecutorErrorInCheck(t)
	pluginmachinery.PluginRegistry().RegisterCorePlugin(te)
	enqueueWorkflow := func(workflowId v1alpha1.WorkflowID) {}

	eventSink := events.NewMockEventSink()
	catalogClient, err := catalog.NewCatalogClient(ctx)
	assert.NoError(t, err)
	recoveryClient := &recoveryMocks.RecoveryClient{}

	adminClient := launchplan.NewFailFastLaunchPlanExecutor()
	nodeExec, err := nodes.NewExecutor(ctx, config.GetConfig().NodeConfig, store, enqueueWorkflow, eventSink, adminClient,
		adminClient, maxOutputSize, "s3://bucket", fakeKubeClient, catalogClient, recoveryClient, promutils.NewTestScope())
	assert.NoError(t, err)
	executor, err := NewExecutor(ctx, store, enqueueWorkflow, eventSink, recorder, "", nodeExec, promutils.NewTestScope())
	assert.NoError(t, err)

	assert.NoError(t, executor.Initialize(ctx))

	wJSON, err := yamlutils.ReadYamlFileAsJSON("testdata/benchmark_wf.yaml")
	if assert.NoError(t, err) {
		w := &v1alpha1.FlyteWorkflow{
			RawOutputDataConfig: v1alpha1.RawOutputDataConfig{RawOutputDataConfig: &admin.RawOutputDataConfig{}},
		}
		if assert.NoError(t, json.Unmarshal(wJSON, w)) {
			// For benchmark workflow, we know how many rounds it needs
			// Number of rounds = 7 + 1
			for i := 0; i < 11; i++ {
				err := executor.HandleFlyteWorkflow(ctx, w)
				for k, v := range w.Status.NodeStatus {
					fmt.Printf("Node[%v=%v],", k, v.Phase.String())
					// Reset dirty manually for tests.
					v.ResetDirty()
				}
				fmt.Printf("\n")

				if i < 4 {
					assert.NoError(t, err, "Round %d", i)
				} else {
					assert.Error(t, err, "Round %d", i)
				}
			}
			assert.Equal(t, v1alpha1.WorkflowPhaseRunning.String(), w.Status.Phase.String(), "Message: [%v]", w.Status.Message)
		}
	}
}

func walkAndPrint(conns v1alpha1.Connections, ns map[v1alpha1.NodeID]*v1alpha1.NodeStatus) {
	ds := []v1alpha1.NodeID{v1alpha1.StartNodeID}
	visited := map[v1alpha1.NodeID]bool{}
	for k := range ns {
		visited[k] = false
	}
	for len(ds) > 0 {
		sub := sets.NewString()
		for _, x := range ds {
			sub.Insert(conns.Downstream[x]...)
			if !visited[x] {
				s, ok := ns[x]
				if ok {
					fmt.Printf("| %s: %s, %s |", x, s.Phase, s.Message)
					visited[x] = true
				} else {
					fmt.Printf("| %s: Not considered |", x)
				}
			}
		}
		fmt.Println()
		ds = sub.List()
	}
}

func TestWorkflowExecutor_HandleFlyteWorkflow(t *testing.T) {
	ctx := context.Background()
	store := createInmemoryDataStore(t, testScope.NewSubScope("13"))
	recorder := StdOutEventRecorder()
	_, err := events.ConstructEventSink(ctx, events.GetConfig(ctx))
	assert.NoError(t, err)

	te := createHappyPathTaskExecutor(t, true)
	pluginmachinery.PluginRegistry().RegisterCorePlugin(te)

	enqueueWorkflow := func(workflowId v1alpha1.WorkflowID) {}

	eventSink := events.NewMockEventSink()
	catalogClient, err := catalog.NewCatalogClient(ctx)
	assert.NoError(t, err)
	recoveryClient := &recoveryMocks.RecoveryClient{}

	adminClient := launchplan.NewFailFastLaunchPlanExecutor()
	nodeExec, err := nodes.NewExecutor(ctx, config.GetConfig().NodeConfig, store, enqueueWorkflow, eventSink, adminClient,
		adminClient, maxOutputSize, "s3://bucket", fakeKubeClient, catalogClient, recoveryClient, promutils.NewTestScope())
	assert.NoError(t, err)

	executor, err := NewExecutor(ctx, store, enqueueWorkflow, eventSink, recorder, "", nodeExec, promutils.NewTestScope())
	assert.NoError(t, err)

	assert.NoError(t, executor.Initialize(ctx))

	wJSON, err := yamlutils.ReadYamlFileAsJSON("testdata/benchmark_wf.yaml")
	if assert.NoError(t, err) {
		w := &v1alpha1.FlyteWorkflow{
			RawOutputDataConfig: v1alpha1.RawOutputDataConfig{RawOutputDataConfig: &admin.RawOutputDataConfig{}},
		}
		if assert.NoError(t, json.Unmarshal(wJSON, w)) {
			// For benchmark workflow, we know how many rounds it needs
			// Number of rounds = 28
			// + WF (x1)
			// | start-node: Succeeded, successfully completed | (x1)
			// | add-one-and-print-0: Succeeded, completed successfully || add-one-and-print-3: Succeeded, completed successfully || print-every-time-0: Succeeded, completed successfully | (x3)
			// | sum-non-none-0: Succeeded, completed successfully | (x3)
			// | add-one-and-print-1: Succeeded, completed successfully || sum-and-print-0: Succeeded, completed successfully | (x3)
			// | add-one-and-print-2: Succeeded, completed successfully | (x3)
			// + WF (x2)
			// Also there is some overlap
			for i := 0; i < 28; i++ {
				err := executor.HandleFlyteWorkflow(ctx, w)
				if err != nil {
					t.Log(err)
				}

				assert.NoError(t, err)
				fmt.Printf("Round[%d] Workflow[%v]\n", i, w.Status.Phase.String())
				walkAndPrint(w.Connections, w.Status.NodeStatus)
				for _, v := range w.Status.NodeStatus {
					// Reset dirty manually for tests.
					v.ResetDirty()
				}
				fmt.Printf("\n")
			}

			assert.Equal(t, v1alpha1.WorkflowPhaseSuccess.String(), w.Status.Phase.String(), "Message: [%v]", w.Status.Message)
		}
	}
}

func BenchmarkWorkflowExecutor(b *testing.B) {
	scope := promutils.NewScope("test3")
	ctx := context.Background()
	store := createInmemoryDataStore(b, scope.NewSubScope(strconv.Itoa(b.N)))
	recorder := StdOutEventRecorder()
	_, err := events.ConstructEventSink(ctx, events.GetConfig(ctx))
	assert.NoError(b, err)

	te := createHappyPathTaskExecutor(b, false)
	pluginmachinery.PluginRegistry().RegisterCorePlugin(te)
	enqueueWorkflow := func(workflowId v1alpha1.WorkflowID) {}

	eventSink := events.NewMockEventSink()
	catalogClient, err := catalog.NewCatalogClient(ctx)
	assert.NoError(b, err)
	recoveryClient := &recoveryMocks.RecoveryClient{}
	adminClient := launchplan.NewFailFastLaunchPlanExecutor()
	nodeExec, err := nodes.NewExecutor(ctx, config.GetConfig().NodeConfig, store, enqueueWorkflow, eventSink, adminClient,
		adminClient, maxOutputSize, "s3://bucket", fakeKubeClient, catalogClient, recoveryClient, scope)
	assert.NoError(b, err)

	executor, err := NewExecutor(ctx, store, enqueueWorkflow, eventSink, recorder, "", nodeExec, promutils.NewTestScope())
	assert.NoError(b, err)

	assert.NoError(b, executor.Initialize(ctx))
	b.ReportAllocs()

	wJSON, err := yamlutils.ReadYamlFileAsJSON("testdata/benchmark_wf.yaml")
	if err != nil {
		assert.FailNow(b, "Got error reading the testdata")
	}
	w := &v1alpha1.FlyteWorkflow{}
	err = json.Unmarshal(wJSON, w)
	if err != nil {
		assert.FailNow(b, "Got error unmarshalling the testdata")
	}

	// Current benchmark 2ms/op
	for i := 0; i < b.N; i++ {
		deepW := w.DeepCopy()
		// For benchmark workflow, we know how many rounds it needs
		// Number of rounds = 7 + 1
		for i := 0; i < 8; i++ {
			err := executor.HandleFlyteWorkflow(ctx, deepW)
			if err != nil {
				assert.FailNow(b, "Run the unit test first. Benchmark should not fail")
			}
		}
		if deepW.Status.Phase != v1alpha1.WorkflowPhaseSuccess {
			assert.FailNow(b, "Workflow did not end in the expected state")
		}
	}
}

func TestWorkflowExecutor_HandleFlyteWorkflow_Failing(t *testing.T) {
	ctx := context.Background()
	store := createInmemoryDataStore(t, promutils.NewTestScope())
	recorder := StdOutEventRecorder()
	_, err := events.ConstructEventSink(ctx, events.GetConfig(ctx))
	assert.NoError(t, err)

	te := createFailingTaskExecutor(t)
	pluginmachinery.PluginRegistry().RegisterCorePlugin(te)

	enqueueWorkflow := func(workflowId v1alpha1.WorkflowID) {}

	recordedRunning := false
	recordedFailed := false
	recordedFailing := true
	eventSink := events.NewMockEventSink()
	mockSink := eventSink.(*events.MockEventSink)
	mockSink.SinkCb = func(ctx context.Context, message proto.Message) error {
		e, ok := message.(*event.WorkflowExecutionEvent)

		if ok {
			assert.True(t, ok)
			switch e.Phase {
			case core.WorkflowExecution_RUNNING:
				occuredAt, err := ptypes.Timestamp(e.OccurredAt)
				assert.NoError(t, err)

				assert.WithinDuration(t, occuredAt, time.Now(), time.Millisecond*5)
				recordedRunning = true
			case core.WorkflowExecution_FAILING:
				occuredAt, err := ptypes.Timestamp(e.OccurredAt)
				assert.NoError(t, err)

				assert.WithinDuration(t, occuredAt, time.Now(), time.Millisecond*5)
				recordedFailing = true
			case core.WorkflowExecution_FAILED:
				occuredAt, err := ptypes.Timestamp(e.OccurredAt)
				assert.NoError(t, err)

				assert.WithinDuration(t, occuredAt, time.Now(), time.Millisecond*5)
				recordedFailed = true
			default:
				return fmt.Errorf("MockWorkflowRecorder should not have entered into any other states [%v]", e.Phase)
			}
		}
		return nil
	}
	catalogClient, err := catalog.NewCatalogClient(ctx)
	assert.NoError(t, err)
	recoveryClient := &recoveryMocks.RecoveryClient{}
	adminClient := launchplan.NewFailFastLaunchPlanExecutor()
	nodeExec, err := nodes.NewExecutor(ctx, config.GetConfig().NodeConfig, store, enqueueWorkflow, eventSink, adminClient,
		adminClient, maxOutputSize, "s3://bucket", fakeKubeClient, catalogClient, recoveryClient, promutils.NewTestScope())
	assert.NoError(t, err)
	executor, err := NewExecutor(ctx, store, enqueueWorkflow, eventSink, recorder, "", nodeExec, promutils.NewTestScope())
	assert.NoError(t, err)

	assert.NoError(t, executor.Initialize(ctx))

	wJSON, err := yamlutils.ReadYamlFileAsJSON("testdata/benchmark_wf.yaml")
	if assert.NoError(t, err) {
		w := &v1alpha1.FlyteWorkflow{
			RawOutputDataConfig: v1alpha1.RawOutputDataConfig{RawOutputDataConfig: &admin.RawOutputDataConfig{}},
		}
		if assert.NoError(t, json.Unmarshal(wJSON, w)) {
			// For benchmark workflow, we will run into the first failure on round 6

			roundsToFail := 7
			for i := 0; i < roundsToFail; i++ {
				err := executor.HandleFlyteWorkflow(ctx, w)
				assert.Nil(t, err, "Round [%v]", i)
				fmt.Printf("Round[%d] Workflow[%v]\n", i, w.Status.Phase.String())
				walkAndPrint(w.Connections, w.Status.NodeStatus)
				for _, v := range w.Status.NodeStatus {
					// Reset dirty manually for tests.
					v.ResetDirty()
				}
				fmt.Printf("\n")

				if i == roundsToFail-1 {
					assert.Equal(t, v1alpha1.WorkflowPhaseFailed, w.Status.Phase)
				} else {
					assert.NotEqual(t, v1alpha1.WorkflowPhaseFailed, w.Status.Phase, "For Round [%v] got phase [%v]", i, w.Status.Phase.String())
				}

			}

			assert.Equal(t, v1alpha1.WorkflowPhaseFailed.String(), w.Status.Phase.String(), "Message: [%v]", w.Status.Message)
		}
	}
	assert.True(t, recordedRunning)
	assert.True(t, recordedFailing)
	assert.True(t, recordedFailed)
}

func TestWorkflowExecutor_HandleFlyteWorkflow_Events(t *testing.T) {
	ctx := context.Background()
	store := createInmemoryDataStore(t, promutils.NewTestScope())
	recorder := StdOutEventRecorder()
	_, err := events.ConstructEventSink(ctx, events.GetConfig(ctx))
	assert.NoError(t, err)

	te := createHappyPathTaskExecutor(t, true)
	pluginmachinery.PluginRegistry().RegisterCorePlugin(te)

	enqueueWorkflow := func(workflowId v1alpha1.WorkflowID) {}

	recordedRunning := false
	recordedSuccess := false
	recordedFailing := true
	eventSink := events.NewMockEventSink()
	mockSink := eventSink.(*events.MockEventSink)
	mockSink.SinkCb = func(ctx context.Context, message proto.Message) error {
		e, ok := message.(*event.WorkflowExecutionEvent)
		if ok {
			switch e.Phase {
			case core.WorkflowExecution_RUNNING:
				occuredAt, err := ptypes.Timestamp(e.OccurredAt)
				assert.NoError(t, err)

				assert.WithinDuration(t, occuredAt, time.Now(), time.Millisecond*5)
				recordedRunning = true
			case core.WorkflowExecution_SUCCEEDING:
				occuredAt, err := ptypes.Timestamp(e.OccurredAt)
				assert.NoError(t, err)

				assert.WithinDuration(t, occuredAt, time.Now(), time.Millisecond*5)
				recordedFailing = true
			case core.WorkflowExecution_SUCCEEDED:
				occuredAt, err := ptypes.Timestamp(e.OccurredAt)
				assert.NoError(t, err)

				assert.WithinDuration(t, occuredAt, time.Now(), time.Millisecond*5)
				recordedSuccess = true
			default:
				return fmt.Errorf("MockWorkflowRecorder should not have entered into any other states, received [%v]", e.Phase.String())
			}
		}
		return nil
	}
	catalogClient, err := catalog.NewCatalogClient(ctx)
	assert.NoError(t, err)
	adminClient := launchplan.NewFailFastLaunchPlanExecutor()
	recoveryClient := &recoveryMocks.RecoveryClient{}
	nodeExec, err := nodes.NewExecutor(ctx, config.GetConfig().NodeConfig, store, enqueueWorkflow, eventSink, adminClient,
		adminClient, maxOutputSize, "s3://bucket", fakeKubeClient, catalogClient, recoveryClient, promutils.NewTestScope())
	assert.NoError(t, err)
	executor, err := NewExecutor(ctx, store, enqueueWorkflow, eventSink, recorder, "metadata", nodeExec, promutils.NewTestScope())
	assert.NoError(t, err)

	assert.NoError(t, executor.Initialize(ctx))

	wJSON, err := yamlutils.ReadYamlFileAsJSON("testdata/benchmark_wf.yaml")
	if assert.NoError(t, err) {
		w := &v1alpha1.FlyteWorkflow{
			RawOutputDataConfig: v1alpha1.RawOutputDataConfig{RawOutputDataConfig: &admin.RawOutputDataConfig{}},
		}
		if assert.NoError(t, json.Unmarshal(wJSON, w)) {
			// For benchmark workflow, we know how many rounds it needs
			// Number of rounds = 28 ?
			for i := 0; i < 28; i++ {
				err := executor.HandleFlyteWorkflow(ctx, w)
				assert.NoError(t, err)
				fmt.Printf("Round[%d] Workflow[%v]\n", i, w.Status.Phase.String())
				walkAndPrint(w.Connections, w.Status.NodeStatus)
				for _, v := range w.Status.NodeStatus {
					// Reset dirty manually for tests.
					v.ResetDirty()
				}
				fmt.Printf("\n")
			}

			assert.Equal(t, v1alpha1.WorkflowPhaseSuccess.String(), w.Status.Phase.String(), "Message: [%v]", w.Status.Message)
		}
	}
	assert.True(t, recordedRunning)
	assert.True(t, recordedFailing)
	assert.True(t, recordedSuccess)
}

func TestWorkflowExecutor_HandleFlyteWorkflow_EventFailure(t *testing.T) {
	ctx := context.Background()
	store := createInmemoryDataStore(t, promutils.NewTestScope())
	recorder := StdOutEventRecorder()
	_, err := events.ConstructEventSink(ctx, events.GetConfig(ctx))
	assert.NoError(t, err)

	te := createHappyPathTaskExecutor(t, true)
	pluginmachinery.PluginRegistry().RegisterCorePlugin(te)

	enqueueWorkflow := func(workflowId v1alpha1.WorkflowID) {}

	wJSON, err := yamlutils.ReadYamlFileAsJSON("testdata/benchmark_wf.yaml")
	assert.NoError(t, err)

	nodeEventSink := events.NewMockEventSink()
	catalogClient, err := catalog.NewCatalogClient(ctx)
	assert.NoError(t, err)
	recoveryClient := &recoveryMocks.RecoveryClient{}

	adminClient := launchplan.NewFailFastLaunchPlanExecutor()
	nodeExec, err := nodes.NewExecutor(ctx, config.GetConfig().NodeConfig, store, enqueueWorkflow, nodeEventSink, adminClient,
		adminClient, maxOutputSize, "s3://bucket", fakeKubeClient, catalogClient, recoveryClient, promutils.NewTestScope())
	assert.NoError(t, err)

	t.Run("EventAlreadyInTerminalStateError", func(t *testing.T) {

		eventSink := events.NewMockEventSink()
		mockSink := eventSink.(*events.MockEventSink)
		mockSink.SinkCb = func(ctx context.Context, message proto.Message) error {
			return &eventsErr.EventError{Code: eventsErr.EventAlreadyInTerminalStateError,
				Cause: errors.New("already exists"),
			}
		}
		executor, err := NewExecutor(ctx, store, enqueueWorkflow, mockSink, recorder, "metadata", nodeExec, promutils.NewTestScope())
		assert.NoError(t, err)
		w := &v1alpha1.FlyteWorkflow{}
		assert.NoError(t, json.Unmarshal(wJSON, w))

		assert.NoError(t, executor.Initialize(ctx))
		err = executor.HandleFlyteWorkflow(ctx, w)
		assert.Equal(t, v1alpha1.WorkflowPhaseFailed.String(), w.Status.Phase.String())

		assert.NoError(t, err)
	})

	t.Run("EventSinkAlreadyExistsError", func(t *testing.T) {
		eventSink := events.NewMockEventSink()
		mockSink := eventSink.(*events.MockEventSink)
		mockSink.SinkCb = func(ctx context.Context, message proto.Message) error {
			return &eventsErr.EventError{Code: eventsErr.AlreadyExists,
				Cause: errors.New("already exists"),
			}
		}
		executor, err := NewExecutor(ctx, store, enqueueWorkflow, eventSink, recorder, "metadata", nodeExec, promutils.NewTestScope())
		assert.NoError(t, err)
		w := &v1alpha1.FlyteWorkflow{}
		assert.NoError(t, json.Unmarshal(wJSON, w))

		err = executor.HandleFlyteWorkflow(ctx, w)
		assert.NoError(t, err)
	})

	t.Run("EventSinkGenericError", func(t *testing.T) {
		eventSink := events.NewMockEventSink()
		mockSink := eventSink.(*events.MockEventSink)
		mockSink.SinkCb = func(ctx context.Context, message proto.Message) error {
			return &eventsErr.EventError{Code: eventsErr.EventSinkError,
				Cause: errors.New("generic exists"),
			}
		}
		executor, err := NewExecutor(ctx, store, enqueueWorkflow, eventSink, recorder, "metadata", nodeExec, promutils.NewTestScope())
		assert.NoError(t, err)
		w := &v1alpha1.FlyteWorkflow{}
		assert.NoError(t, json.Unmarshal(wJSON, w))

		err = executor.HandleFlyteWorkflow(ctx, w)
		assert.Error(t, err)
		assert.True(t, wfErrors.Matches(err, wfErrors.EventRecordingError))
	})
}

func TestWorkflowExecutor_HandleAbortedWorkflow(t *testing.T) {
	ctx := context.TODO()

	t.Run("user-initiated-fail", func(t *testing.T) {

		nodeExec := &mocks2.Node{}
		wExec := &workflowExecutor{
			nodeExecutor: nodeExec,
			metrics:      newMetrics(promutils.NewTestScope()),
		}

		nodeExec.OnAbortHandlerMatch(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

		w := &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				DeletionTimestamp: &v1.Time{},
			},
			Status: v1alpha1.WorkflowStatus{
				FailedAttempts: 1,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{
					v1alpha1.StartNodeID: {},
				},
			},
		}

		assert.Error(t, wExec.HandleAbortedWorkflow(ctx, w, 5))

		assert.Equal(t, uint32(1), w.Status.FailedAttempts)
	})

	t.Run("user-initiated-success", func(t *testing.T) {

		var evs []*event.WorkflowExecutionEvent
		nodeExec := &mocks2.Node{}
		wExec := &workflowExecutor{
			nodeExecutor: nodeExec,
			wfRecorder: &events.MockRecorder{
				RecordWorkflowEventCb: func(ctx context.Context, event *event.WorkflowExecutionEvent) error {
					evs = append(evs, event)
					return nil
				},
			},
			metrics: newMetrics(promutils.NewTestScope()),
		}

		nodeExec.OnAbortHandlerMatch(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		w := &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				DeletionTimestamp: &v1.Time{},
			},
			Status: v1alpha1.WorkflowStatus{
				FailedAttempts: 1,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{
					v1alpha1.StartNodeID: {},
				},
			},
		}

		assert.NoError(t, wExec.HandleAbortedWorkflow(ctx, w, 5))

		assert.Equal(t, uint32(1), w.Status.FailedAttempts)
		assert.Len(t, evs, 1)
	})

	t.Run("user-initiated-attempts-exhausted", func(t *testing.T) {

		var evs []*event.WorkflowExecutionEvent
		nodeExec := &mocks2.Node{}
		wExec := &workflowExecutor{
			nodeExecutor: nodeExec,
			wfRecorder: &events.MockRecorder{
				RecordWorkflowEventCb: func(ctx context.Context, event *event.WorkflowExecutionEvent) error {
					evs = append(evs, event)
					return nil
				},
			},
			metrics: newMetrics(promutils.NewTestScope()),
		}

		nodeExec.OnAbortHandlerMatch(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		w := &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				DeletionTimestamp: &v1.Time{},
			},
			Status: v1alpha1.WorkflowStatus{
				FailedAttempts: 6,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{
					v1alpha1.StartNodeID: {},
				},
			},
		}

		assert.NoError(t, wExec.HandleAbortedWorkflow(ctx, w, 5))

		assert.Equal(t, uint32(6), w.Status.FailedAttempts)
		assert.Len(t, evs, 1)
	})

	t.Run("failure-abort-success", func(t *testing.T) {
		var evs []*event.WorkflowExecutionEvent
		nodeExec := &mocks2.Node{}
		wExec := &workflowExecutor{
			nodeExecutor: nodeExec,
			wfRecorder: &events.MockRecorder{
				RecordWorkflowEventCb: func(ctx context.Context, event *event.WorkflowExecutionEvent) error {
					evs = append(evs, event)
					return nil
				},
			},
			metrics: newMetrics(promutils.NewTestScope()),
		}

		nodeExec.OnAbortHandlerMatch(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		w := &v1alpha1.FlyteWorkflow{
			Status: v1alpha1.WorkflowStatus{
				FailedAttempts: 5,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{
					v1alpha1.StartNodeID: {},
				},
			},
		}

		assert.NoError(t, wExec.HandleAbortedWorkflow(ctx, w, 5))

		assert.Equal(t, uint32(5), w.Status.FailedAttempts)
		assert.Len(t, evs, 1)
	})

	t.Run("failure-abort-failed", func(t *testing.T) {

		nodeExec := &mocks2.Node{}
		wExec := &workflowExecutor{
			nodeExecutor: nodeExec,
			metrics:      newMetrics(promutils.NewTestScope()),
		}

		nodeExec.OnAbortHandlerMatch(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("err"))

		w := &v1alpha1.FlyteWorkflow{
			Status: v1alpha1.WorkflowStatus{
				FailedAttempts: 1,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{
					v1alpha1.StartNodeID: {},
				},
			},
		}

		assert.Error(t, wExec.HandleAbortedWorkflow(ctx, w, 5))

		assert.Equal(t, uint32(1), w.Status.FailedAttempts)
	})
}
