package k8splugins

import (
	"context"
	"testing"

	"errors"
	structpb "github.com/golang/protobuf/ptypes/struct"
	eventErrors "github.com/lyft/flyteidl/clients/go/events/errors"
	utils2 "github.com/lyft/flyteplugins/go/tasks/v1/utils"

	adminMocks "github.com/lyft/flyteidl/clients/go/admin/mocks"
	"github.com/lyft/flyteidl/clients/go/coreutils"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/lyft/flyteplugins/go/tasks/v1/flytek8s"
	flytek8sMocks "github.com/lyft/flyteplugins/go/tasks/v1/flytek8s/mocks"
	"github.com/lyft/flyteplugins/go/tasks/v1/k8splugins/mocks"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"
	tasksMocks "github.com/lyft/flyteplugins/go/tasks/v1/types/mocks"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"
	"github.com/lyft/flytestdlib/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
)

//go:generate mockery -dir ../../../../vendor/github.com/lyft/flytestdlib/utils -name AutoRefreshCache

type mockAdminService struct {
	*adminMocks.AdminServiceClient
	executionPhaseCache map[string]core.WorkflowExecution_Phase
}

type mockAutorefreshCache struct {
	*mocks.AutoRefreshCache
	waitableExec *waitableTaskExecutor
}

func getMockTaskContext() *tasksMocks.TaskContext {
	taskCtx := &tasksMocks.TaskContext{}
	taskCtx.On("GetNamespace").Return("ns")
	taskCtx.On("GetAnnotations").Return(map[string]string{"aKey": "aVal"})
	taskCtx.On("GetLabels").Return(map[string]string{"lKey": "lVal"})
	taskCtx.On("GetOwnerReference").Return(metav1.OwnerReference{Name: "x"})
	taskCtx.On("GetInputsFile").Return(storage.DataReference("/fake/inputs.pb"))
	taskCtx.On("GetDataDir").Return(storage.DataReference("/fake/"))
	taskCtx.On("GetErrorFile").Return(storage.DataReference("/fake/error.pb"))
	taskCtx.On("GetOutputsFile").Return(storage.DataReference("/fake/inputs.pb"))
	taskCtx.On("GetPhaseVersion").Return(uint32(1))

	id := &tasksMocks.TaskExecutionID{}
	id.On("GetGeneratedName").Return("test")
	id.On("GetID").Return(core.TaskExecutionIdentifier{})
	taskCtx.On("GetTaskExecutionID").Return(id)

	to := &tasksMocks.TaskOverrides{}
	to.On("GetResources").Return(resourceRequirements)
	taskCtx.On("GetOverrides").Return(to)
	return taskCtx
}

func newDummpyTaskEventRecorder() types.EventRecorder {
	s := &tasksMocks.EventRecorder{}
	s.On("RecordTaskEvent", mock.Anything, mock.Anything).Return(nil)
	return s
}

func (m mockAdminService) GetExecution(ctx context.Context, in *admin.WorkflowExecutionGetRequest, opts ...grpc.CallOption) (
	*admin.Execution, error) {
	wfExec := &admin.Execution{}
	wfExec.Id = in.Id
	wfExec.Closure = &admin.ExecutionClosure{}
	if existingPhase, found := m.executionPhaseCache[in.Id.String()]; found {
		if existingPhase < core.WorkflowExecution_SUCCEEDED {
			wfExec.Closure.Phase = existingPhase + 1
		} else {
			wfExec.Closure.Phase = existingPhase
		}
	} else {
		wfExec.Closure.Phase = core.WorkflowExecution_QUEUED
	}

	m.executionPhaseCache[in.Id.String()] = wfExec.Closure.Phase

	return wfExec, nil
}

func (m mockAutorefreshCache) Start(ctx context.Context) {
}

func (m mockAutorefreshCache) GetOrCreate(item utils.CacheItem) (utils.CacheItem, error) {
	w := item.(*waitableWrapper)
	item, _, err := m.waitableExec.syncItem(context.TODO(), w)
	return item, err
}

func setupMockExecutor(t testing.TB) *waitableTaskExecutor {
	ctx := context.Background()
	waitableExec := &waitableTaskExecutor{
		containerTaskExecutor: containerTaskExecutor{},
	}

	waitableExec.K8sTaskExecutor = flytek8s.NewK8sTaskExecutorForResource(waitableTaskType, &v1.Pod{}, waitableExec,
		flytek8s.DefaultInformerResyncDuration)

	mockAdmin := &mockAdminService{
		AdminServiceClient:  &adminMocks.AdminServiceClient{},
		executionPhaseCache: map[string]core.WorkflowExecution_Phase{},
	}
	mockAdmin.On("TerminateExecution", mock.Anything, mock.Anything).Return(nil, nil).Times(2)
	waitableExec.adminClient = mockAdmin

	mockCache := &mockAutorefreshCache{
		AutoRefreshCache: &mocks.AutoRefreshCache{},
		waitableExec:     waitableExec,
	}

	mem, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	ref, err := mem.ConstructReference(ctx, mem.GetBaseContainerFQN(ctx), "fake", "inputs.pb")
	assert.NoError(t, err)
	assert.NoError(t, mem.WriteProtobuf(ctx, ref,
		storage.Options{}, coreutils.MustMakeLiteral(map[string]interface{}{
			"w": 1,
		})))

	assert.NoError(t, waitableExec.Initialize(context.TODO(), types.ExecutorInitializationParameters{
		EventRecorder: newDummpyTaskEventRecorder(),
		DataStore:     mem,
		MetricsScope:  promutils.NewTestScope(),
		OwnerKind:     "Pod",
		EnqueueOwner: func(name k8sTypes.NamespacedName) error {
			return nil
		},
	}))

	// Wait until after Initialize is called since it sets the executionsCache there too, but we want
	// to use the mock one
	waitableExec.executionsCache = mockCache
	mockCache.Start(ctx)

	return waitableExec
}

func simulateMarshalUnMarshal(t testing.TB, in map[string]interface{}) (out map[string]interface{}) {
	raw, err := json.Marshal(in)
	assert.NoError(t, err)

	out = map[string]interface{}{}
	assert.NoError(t, json.Unmarshal(raw, &out))
	return out
}

func createWaitableLiteral(t testing.TB, execName string, phase core.WorkflowExecution_Phase) (*plugins.Waitable, *core.Literal) {
	stObj := &structpb.Struct{}
	expected := &plugins.Waitable{
		WfExecId: &core.WorkflowExecutionIdentifier{
			Name:    execName,
			Project: "exec_proj",
			Domain:  "exec_domain",
		},
		Phase: phase,
	}
	assert.NoError(t, utils2.MarshalStruct(expected, stObj))

	waitableLiteral := &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Generic{
					Generic: stObj,
				},
			},
		},
	}

	return expected, waitableLiteral
}

func TestWaitableTaskExecutor(t *testing.T) {
	ctx := context.Background()
	mockRuntimeClient := flytek8sMocks.NewMockRuntimeClient()
	flytek8s.InitializeFake()
	assert.NoError(t, flytek8s.InjectCache(&flytek8sMocks.Cache{}))
	assert.NoError(t, flytek8s.InjectClient(mockRuntimeClient))


	t.Run("Initialize", func(t *testing.T) {
		_, err := newWaitableTaskExecutor(context.TODO())
		assert.NoError(t, err)
	})

	t.Run("StartTask", func(t *testing.T) {
		taskCtx := getMockTaskContext()
		expected, waitableLiteral := createWaitableLiteral(t, "exec_name", 0)
		exec := setupMockExecutor(t)

		status, err := exec.StartTask(ctx, taskCtx, &core.TaskTemplate{}, coreutils.MustMakeLiteral(map[string]interface{}{
			"w": waitableLiteral,
		}).GetMap())

		assert.NoError(t, err)
		assert.Contains(t, status.State, waitablesKey)
		wr := status.State[waitablesKey].([]*waitableWrapper)
		assert.Equal(t, (&waitableWrapper{Waitable: &plugins.Waitable{
			WfExecId: expected.WfExecId,
			Phase:    core.WorkflowExecution_QUEUED,
		}}).String(), wr[0].String())
		assert.Equal(t, types.TaskPhaseQueued, status.Phase)
	})

	t.Run("StartTaskStateMismatch", func(t *testing.T) {
		taskCtx := getMockTaskContext()
		_, waitableLiteral := createWaitableLiteral(t, "exec_name", 0)
		mockRecorder := &tasksMocks.EventRecorder{}
		mockRecorder.On("RecordTaskEvent", mock.Anything, mock.Anything).Return(&eventErrors.EventError{Code: eventErrors.EventAlreadyInTerminalStateError,
			Cause: errors.New("already exists"),
		})
		exec := setupMockExecutor(t)
		exec.recorder = mockRecorder
		status, err := exec.StartTask(ctx, taskCtx, &core.TaskTemplate{}, coreutils.MustMakeLiteral(map[string]interface{}{
			"w": waitableLiteral,
		}).GetMap())

		assert.NoError(t, err)
		assert.Nil(t, status.State)
		assert.Equal(t, types.TaskPhasePermanentFailure, status.Phase)
	})

	t.Run("CheckTaskUntilSuccess", func(t *testing.T) {
		taskCtx := getMockTaskContext()
		_, waitableLiteral := createWaitableLiteral(t, "exec_should_succeed", 0)
		stObj := &structpb.Struct{}
		wrongTypeProto := &plugins.ArrayJob{
			Size: 2,
		}
		assert.NoError(t, utils2.MarshalStruct(wrongTypeProto, stObj))
		taskTemplate := &core.TaskTemplate{}
		exec := setupMockExecutor(t)
		status, err := exec.StartTask(ctx, taskCtx, taskTemplate, coreutils.MustMakeLiteral(map[string]interface{}{
			"w":            waitableLiteral,
			"wArr":         coreutils.MustMakeLiteral([]interface{}{waitableLiteral, waitableLiteral}),
			"otherGeneric": &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{Value: &core.Scalar_Generic{Generic: stObj}}}},
		}).GetMap())

		assert.NoError(t, err)

		taskCtx = getMockTaskContext()
		taskCtx.On("GetCustomState").Return(simulateMarshalUnMarshal(t, status.State))
		taskCtx.On("GetPhase").Return(status.Phase)
		taskCtx.On("GetK8sServiceAccount").Return("")

		for status.Phase != types.TaskPhaseSucceeded {
			status, err = exec.CheckTaskStatus(ctx, taskCtx, taskTemplate)
			assert.NoError(t, err)
			if err != nil {
				assert.FailNow(t, "failed to check status. Err: %v", err)
			}

			taskCtx = getMockTaskContext()
			taskCtx.On("GetCustomState").Return(simulateMarshalUnMarshal(t, status.State))
			taskCtx.On("GetPhase").Return(status.Phase)

			assert.NoError(t, advancePodPhases(ctx, mockRuntimeClient))
		}

		assert.Equal(t, types.TaskPhaseSucceeded, status.Phase)
	})

	t.Run("KillTask", func(t *testing.T) {
		taskCtx := getMockTaskContext()
		expected, waitableLiteral := createWaitableLiteral(t, "exec_should_terminate", 0)
		exec := setupMockExecutor(t)

		status, err := exec.StartTask(ctx, taskCtx, &core.TaskTemplate{}, coreutils.MustMakeLiteral(map[string]interface{}{
			"w": waitableLiteral,
		}).GetMap())

		assert.NoError(t, err)
		taskCtx.On("GetCustomState").Return(simulateMarshalUnMarshal(t, status.State))
		taskCtx.On("GetPhase").Return(status.Phase)

		assert.Contains(t, status.State, waitablesKey)
		wr := status.State[waitablesKey].([]*waitableWrapper)
		assert.Equal(t, (&waitableWrapper{Waitable: &plugins.Waitable{
			WfExecId: expected.WfExecId,
			Phase:    core.WorkflowExecution_QUEUED,
		}}).String(), wr[0].String())
		assert.Equal(t, types.TaskPhaseQueued, status.Phase)

		assert.NoError(t, exec.KillTask(ctx, taskCtx, "cause I like to"))
	})
}

func TestUpdateWaitableLiterals(t *testing.T) {
	_, l1 := createWaitableLiteral(t, "a.a.exec", 0)
	_, l2 := createWaitableLiteral(t, "b.exec", 0)
	originalLiterals := coreutils.MustMakeLiteral(map[string]interface{}{
		"a": map[string]interface{}{
			"a.a": l1,
		},
		"b": l2,
	})

	subLiterals, waitables := discoverWaitableInputs(originalLiterals)
	assert.Len(t, subLiterals, 2)
	assert.Len(t, waitables, 2)

	for _, w := range waitables {
		if w.WfExecId.Name == "a.a.exec" {
			w.Phase = core.WorkflowExecution_FAILED
		} else {
			w.Phase = core.WorkflowExecution_SUCCEEDED
		}
	}

	assert.NoError(t, updateWaitableLiterals(subLiterals, waitables))

	_, l1 = createWaitableLiteral(t, "a.a.exec", core.WorkflowExecution_FAILED)
	_, l2 = createWaitableLiteral(t, "b.exec", core.WorkflowExecution_SUCCEEDED)
	expectedLiterals := coreutils.MustMakeLiteral(map[string]interface{}{
		"a": map[string]interface{}{
			"a.a": l1,
		},
		"b": l2,
	})

	assert.Equal(t, expectedLiterals, originalLiterals)
}

func TestToWaitableWrapperSlice(t *testing.T) {
	input := []*waitableWrapper{
		{
			Waitable: &plugins.Waitable{
				Phase: core.WorkflowExecution_SUCCEEDED,
			},
		},
	}

	ctx := context.Background()
	t.Run("WaitableWrapper Slice", func(t *testing.T) {
		res, err := toWaitableWrapperSlice(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, input, res)
	})

	t.Run("Wrong type", func(t *testing.T) {
		_, err := toWaitableWrapperSlice(ctx, "wrong type")
		assert.Error(t, err)
	})

	t.Run("Interface", func(t *testing.T) {
		a := make([]interface{}, 0, len(input))
		for _, w := range input {
			a = append(a, w)
		}

		res, err := toWaitableWrapperSlice(context.TODO(), a)
		assert.NoError(t, err)
		assert.Equal(t, input, res)
	})

	t.Run("Json", func(t *testing.T) {
		raw, err := json.Marshal(input)
		assert.NoError(t, err)

		var a []interface{}
		assert.NoError(t, json.Unmarshal(raw, &a))

		res, err := toWaitableWrapperSlice(context.TODO(), a)
		assert.NoError(t, err)
		assert.Equal(t, input, res)
	})

	t.Run("Wrong type in slice", func(t *testing.T) {
		a := make([]interface{}, 0, len(input))
		for _, w := range input {
			a = append(a, w)
		}

		a = append(a, "wrong type")

		_, err := toWaitableWrapperSlice(context.TODO(), a)
		assert.Error(t, err)
	})
}
