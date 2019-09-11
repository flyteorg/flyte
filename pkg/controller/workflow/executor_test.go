package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	mocks2 "github.com/lyft/flytepropeller/pkg/controller/executors/mocks"

	eventsErr "github.com/lyft/flyteidl/clients/go/events/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	wfErrors "github.com/lyft/flytepropeller/pkg/controller/workflow/errors"

	"time"

	"github.com/lyft/flyteplugins/go/tasks/v1/flytek8s"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/lyft/flyteidl/clients/go/events"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/event"
	pluginV1 "github.com/lyft/flyteplugins/go/tasks/v1/types"
	"github.com/lyft/flyteplugins/go/tasks/v1/types/mocks"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/catalog"
	"github.com/lyft/flytepropeller/pkg/controller/nodes"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/task"
	"github.com/lyft/flytepropeller/pkg/utils"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"
	"github.com/lyft/flytestdlib/yamlutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/client-go/tools/record"
)

var (
	testScope      = promutils.NewScope("test_wfexec")
	fakeKubeClient = mocks2.NewFakeKubeClient()
)

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

func createHappyPathTaskExecutor(t assert.TestingT, store *storage.DataStore, enableAsserts bool) pluginV1.Executor {
	exec := &mocks.Executor{}
	exec.On("GetID").Return("task")
	exec.On("GetProperties").Return(pluginV1.ExecutorProperties{})
	exec.On("Initialize",
		mock.AnythingOfType(reflect.TypeOf(context.TODO()).String()),
		mock.AnythingOfType(reflect.TypeOf(pluginV1.ExecutorInitializationParameters{}).String()),
	).Return(nil)

	exec.On("ResolveOutputs",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).
		Return(func(ctx context.Context, taskCtx pluginV1.TaskContext, varNames ...string) (values map[string]*core.Literal) {
			d := &handler.Data{}
			outputsFileRef := v1alpha1.GetOutputsFile(taskCtx.GetDataDir())
			assert.NoError(t, store.ReadProtobuf(ctx, outputsFileRef, d))
			assert.NotNil(t, d.Literals)

			values = make(map[string]*core.Literal, len(varNames))
			for _, varName := range varNames {
				l, ok := d.Literals[varName]
				assert.True(t, ok, "Expect var %v in task outputs.", varName)

				values[varName] = l
			}

			return values
		}, func(ctx context.Context, taskCtx pluginV1.TaskContext, varNames ...string) error {
			return nil
		})

	startFn := func(ctx context.Context, taskCtx pluginV1.TaskContext, task *core.TaskTemplate, _ *core.LiteralMap) pluginV1.TaskStatus {
		outputVars := task.GetInterface().Outputs.Variables
		o := &core.LiteralMap{
			Literals: make(map[string]*core.Literal, len(outputVars)),
		}
		for k, v := range outputVars {
			l, err := utils.MakeDefaultLiteralForType(v.Type)
			if enableAsserts && !assert.NoError(t, err) {
				assert.FailNow(t, "Failed to create default output for node [%v] Type [%v]", taskCtx.GetTaskExecutionID(), v.Type)
			}
			o.Literals[k] = l
		}
		assert.NoError(t, store.WriteProtobuf(ctx, v1alpha1.GetOutputsFile(taskCtx.GetDataDir()), storage.Options{}, o))

		return pluginV1.TaskStatusRunning
	}

	exec.On("StartTask",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(o pluginV1.TaskContext) bool { return true }),
		mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
		mock.MatchedBy(func(o *core.LiteralMap) bool { return true }),
	).Return(startFn, nil)

	checkStatusFn := func(_ context.Context, taskCtx pluginV1.TaskContext, _ *core.TaskTemplate) pluginV1.TaskStatus {
		if enableAsserts {
			assert.NotEmpty(t, taskCtx.GetDataDir())
		}
		return pluginV1.TaskStatusSucceeded
	}
	exec.On("CheckTaskStatus",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(o pluginV1.TaskContext) bool { return true }),
		mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
	).Return(checkStatusFn, nil)

	return exec
}

func createFailingTaskExecutor() pluginV1.Executor {
	exec := &mocks.Executor{}
	exec.On("GetID").Return("task")
	exec.On("GetProperties").Return(pluginV1.ExecutorProperties{})
	exec.On("Initialize",
		mock.AnythingOfType(reflect.TypeOf(context.TODO()).String()),
		mock.AnythingOfType(reflect.TypeOf(pluginV1.ExecutorInitializationParameters{}).String()),
	).Return(nil)

	exec.On("StartTask",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(o pluginV1.TaskContext) bool { return true }),
		mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
		mock.MatchedBy(func(o *core.LiteralMap) bool { return true }),
	).Return(pluginV1.TaskStatusRunning, nil)

	exec.On("CheckTaskStatus",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(o pluginV1.TaskContext) bool { return true }),
		mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
	).Return(pluginV1.TaskStatusPermanentFailure(errors.New("failed")), nil)

	exec.On("KillTask",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(o pluginV1.TaskContext) bool { return true }),
		mock.Anything,
	).Return(nil)

	return exec
}

func createTaskExecutorErrorInCheck() pluginV1.Executor {
	exec := &mocks.Executor{}
	exec.On("GetID").Return("task")
	exec.On("GetProperties").Return(pluginV1.ExecutorProperties{})
	exec.On("Initialize",
		mock.AnythingOfType(reflect.TypeOf(context.TODO()).String()),
		mock.AnythingOfType(reflect.TypeOf(pluginV1.ExecutorInitializationParameters{}).String()),
	).Return(nil)

	exec.On("StartTask",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(o pluginV1.TaskContext) bool { return true }),
		mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
		mock.MatchedBy(func(o *core.LiteralMap) bool { return true }),
	).Return(pluginV1.TaskStatusRunning, nil)

	exec.On("CheckTaskStatus",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(o pluginV1.TaskContext) bool { return true }),
		mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
	).Return(pluginV1.TaskStatusUndefined, errors.New("check failed"))

	return exec
}

func createSingletonTaskExecutorFactory(te pluginV1.Executor) task.Factory {
	return &task.FactoryFuncs{
		GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginV1.Executor, error) {
			return te, nil
		},
		ListAllTaskExecutorsCb: func() []pluginV1.Executor {
			return []pluginV1.Executor{te}
		},
	}
}

func init() {
	flytek8s.InitializeFake()
}

func TestWorkflowExecutor_HandleFlyteWorkflow_Error(t *testing.T) {
	ctx := context.Background()
	store := createInmemoryDataStore(t, testScope.NewSubScope("12"))
	recorder := StdOutEventRecorder()
	_, err := events.ConstructEventSink(ctx, events.GetConfig(ctx))
	assert.NoError(t, err)

	te := createTaskExecutorErrorInCheck()
	tf := createSingletonTaskExecutorFactory(te)
	task.SetTestFactory(tf)
	assert.True(t, task.IsTestModeEnabled())

	enqueueWorkflow := func(workflowId v1alpha1.WorkflowID) {}

	eventSink := events.NewMockEventSink()
	catalogClient, _ := catalog.NewCatalogClient(ctx, store)
	nodeExec, err := nodes.NewExecutor(ctx, store, enqueueWorkflow, time.Second, eventSink, launchplan.NewFailFastLaunchPlanExecutor(), catalogClient, fakeKubeClient, promutils.NewTestScope())
	assert.NoError(t, err)
	executor, err := NewExecutor(ctx, store, enqueueWorkflow, eventSink, recorder, "", nodeExec, promutils.NewTestScope())
	assert.NoError(t, err)

	assert.NoError(t, executor.Initialize(ctx))

	wJSON, err := yamlutils.ReadYamlFileAsJSON("testdata/benchmark_wf.yaml")
	if assert.NoError(t, err) {
		w := &v1alpha1.FlyteWorkflow{}
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

func TestWorkflowExecutor_HandleFlyteWorkflow(t *testing.T) {
	ctx := context.Background()
	store := createInmemoryDataStore(t, testScope.NewSubScope("13"))
	recorder := StdOutEventRecorder()
	_, err := events.ConstructEventSink(ctx, events.GetConfig(ctx))
	assert.NoError(t, err)

	te := createHappyPathTaskExecutor(t, store, true)
	tf := createSingletonTaskExecutorFactory(te)
	task.SetTestFactory(tf)
	assert.True(t, task.IsTestModeEnabled())

	enqueueWorkflow := func(workflowId v1alpha1.WorkflowID) {}

	eventSink := events.NewMockEventSink()
	catalogClient, _ := catalog.NewCatalogClient(ctx, store)
	nodeExec, err := nodes.NewExecutor(ctx, store, enqueueWorkflow, time.Second, eventSink, launchplan.NewFailFastLaunchPlanExecutor(), catalogClient, fakeKubeClient, promutils.NewTestScope())
	assert.NoError(t, err)

	executor, err := NewExecutor(ctx, store, enqueueWorkflow, eventSink, recorder, "", nodeExec, promutils.NewTestScope())
	assert.NoError(t, err)

	assert.NoError(t, executor.Initialize(ctx))

	wJSON, err := yamlutils.ReadYamlFileAsJSON("testdata/benchmark_wf.yaml")
	if assert.NoError(t, err) {
		w := &v1alpha1.FlyteWorkflow{}
		if assert.NoError(t, json.Unmarshal(wJSON, w)) {
			// For benchmark workflow, we know how many rounds it needs
			// Number of rounds = 12 ?
			for i := 0; i < 12; i++ {
				err := executor.HandleFlyteWorkflow(ctx, w)
				if err != nil {
					t.Log(err)
				}

				assert.NoError(t, err)
				fmt.Printf("Round[%d] Workflow[%v]\n", i, w.Status.Phase.String())
				for k, v := range w.Status.NodeStatus {
					fmt.Printf("Node[%v=%v],", k, v.Phase.String())
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

	te := createHappyPathTaskExecutor(b, store, false)
	tf := createSingletonTaskExecutorFactory(te)
	task.SetTestFactory(tf)
	assert.True(b, task.IsTestModeEnabled())
	enqueueWorkflow := func(workflowId v1alpha1.WorkflowID) {}

	eventSink := events.NewMockEventSink()
	catalogClient, _ := catalog.NewCatalogClient(ctx, store)
	nodeExec, err := nodes.NewExecutor(ctx, store, enqueueWorkflow, time.Second, eventSink, launchplan.NewFailFastLaunchPlanExecutor(), catalogClient, fakeKubeClient, scope)
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

	te := createFailingTaskExecutor()
	tf := createSingletonTaskExecutorFactory(te)
	task.SetTestFactory(tf)
	assert.True(t, task.IsTestModeEnabled())

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
	catalogClient, _ := catalog.NewCatalogClient(ctx, store)
	nodeExec, err := nodes.NewExecutor(ctx, store, enqueueWorkflow, time.Second, eventSink, launchplan.NewFailFastLaunchPlanExecutor(), catalogClient, fakeKubeClient, promutils.NewTestScope())
	assert.NoError(t, err)
	executor, err := NewExecutor(ctx, store, enqueueWorkflow, eventSink, recorder, "", nodeExec, promutils.NewTestScope())
	assert.NoError(t, err)

	assert.NoError(t, executor.Initialize(ctx))

	wJSON, err := yamlutils.ReadYamlFileAsJSON("testdata/benchmark_wf.yaml")
	if assert.NoError(t, err) {
		w := &v1alpha1.FlyteWorkflow{}
		if assert.NoError(t, json.Unmarshal(wJSON, w)) {
			// For benchmark workflow, we will run into the first failure on round 6

			for i := 0; i < 6; i++ {
				err := executor.HandleFlyteWorkflow(ctx, w)
				assert.Nil(t, err, "Round [%v]", i)
				fmt.Printf("Round[%d] Workflow[%v]\n", i, w.Status.Phase.String())
				for k, v := range w.Status.NodeStatus {
					fmt.Printf("Node[%v=%v],", k, v.Phase.String())
					// Reset dirty manually for tests.
					v.ResetDirty()
				}
				fmt.Printf("\n")

				if i == 5 {
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

	te := createHappyPathTaskExecutor(t, store, true)
	tf := createSingletonTaskExecutorFactory(te)
	task.SetTestFactory(tf)
	assert.True(t, task.IsTestModeEnabled())

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
	catalogClient, _ := catalog.NewCatalogClient(ctx, store)
	nodeExec, err := nodes.NewExecutor(ctx, store, enqueueWorkflow, time.Second, eventSink, launchplan.NewFailFastLaunchPlanExecutor(), catalogClient, fakeKubeClient, promutils.NewTestScope())
	assert.NoError(t, err)
	executor, err := NewExecutor(ctx, store, enqueueWorkflow, eventSink, recorder, "metadata", nodeExec, promutils.NewTestScope())
	assert.NoError(t, err)

	assert.NoError(t, executor.Initialize(ctx))

	wJSON, err := yamlutils.ReadYamlFileAsJSON("testdata/benchmark_wf.yaml")
	if assert.NoError(t, err) {
		w := &v1alpha1.FlyteWorkflow{}
		if assert.NoError(t, json.Unmarshal(wJSON, w)) {
			// For benchmark workflow, we know how many rounds it needs
			// Number of rounds = 12 ?
			for i := 0; i < 12; i++ {
				err := executor.HandleFlyteWorkflow(ctx, w)
				assert.NoError(t, err)
				fmt.Printf("Round[%d] Workflow[%v]\n", i, w.Status.Phase.String())
				for k, v := range w.Status.NodeStatus {
					fmt.Printf("Node[%v=%v],", k, v.Phase.String())
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

	te := createHappyPathTaskExecutor(t, store, true)
	tf := createSingletonTaskExecutorFactory(te)
	task.SetTestFactory(tf)
	assert.True(t, task.IsTestModeEnabled())

	enqueueWorkflow := func(workflowId v1alpha1.WorkflowID) {}

	wJSON, err := yamlutils.ReadYamlFileAsJSON("testdata/benchmark_wf.yaml")
	assert.NoError(t, err)

	nodeEventSink := events.NewMockEventSink()
	catalogClient, _ := catalog.NewCatalogClient(ctx, store)
	nodeExec, err := nodes.NewExecutor(ctx, store, enqueueWorkflow, time.Second, nodeEventSink, launchplan.NewFailFastLaunchPlanExecutor(), catalogClient, fakeKubeClient, promutils.NewTestScope())
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
