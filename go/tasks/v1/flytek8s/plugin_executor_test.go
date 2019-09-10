package flytek8s_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	eventErrors "github.com/lyft/flyteidl/clients/go/events/errors"

	k8serrs "k8s.io/apimachinery/pkg/api/errors"

	taskerrs "github.com/lyft/flyteplugins/go/tasks/v1/errors"

	"github.com/lyft/flytestdlib/promutils"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/event"

	"github.com/lyft/flyteplugins/go/tasks/v1/events"
	"github.com/lyft/flyteplugins/go/tasks/v1/flytek8s"
	"github.com/lyft/flyteplugins/go/tasks/v1/flytek8s/mocks"
	"github.com/lyft/flytestdlib/storage"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"
	mocks2 "github.com/lyft/flyteplugins/go/tasks/v1/types/mocks"
	"github.com/stretchr/testify/assert"
	k8sBatch "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type k8sSampleHandler struct {
}

func (k8sSampleHandler) BuildResource(ctx context.Context, taskCtx types.TaskContext, task *core.TaskTemplate, inputs *core.LiteralMap) (flytek8s.K8sResource, error) {
	panic("implement me")
}

func (k8sSampleHandler) BuildIdentityResource(ctx context.Context, taskCtx types.TaskContext) (flytek8s.K8sResource, error) {
	panic("implement me")
}

func (k8sSampleHandler) GetTaskStatus(ctx context.Context, taskCtx types.TaskContext, resource flytek8s.K8sResource) (types.TaskStatus, *events.TaskEventInfo, error) {
	panic("implement me")
}

func (k8sSampleHandler) GetProperties() types.ExecutorProperties {
	panic("implement me")
}

func ExampleNewK8sTaskExecutorForResource() {
	exec := flytek8s.NewK8sTaskExecutorForResource("SampleHandler", &k8sBatch.Job{}, k8sSampleHandler{}, time.Second*1)
	fmt.Printf("Created executor: %v\n", exec.GetID())

	// Output:
	// Created executor: SampleHandler
}

func getMockTaskContext() *mocks2.TaskContext {
	taskCtx := &mocks2.TaskContext{}
	taskCtx.On("GetNamespace").Return("ns")
	taskCtx.On("GetAnnotations").Return(map[string]string{"aKey": "aVal"})
	taskCtx.On("GetLabels").Return(map[string]string{"lKey": "lVal"})
	taskCtx.On("GetOwnerReference").Return(v12.OwnerReference{Name: "x"})
	taskCtx.On("GetOutputsFile").Return(storage.DataReference("outputs"))
	taskCtx.On("GetInputsFile").Return(storage.DataReference("inputs"))
	taskCtx.On("GetErrorFile").Return(storage.DataReference("error"))

	id := &mocks2.TaskExecutionID{}
	id.On("GetGeneratedName").Return("test")
	id.On("GetID").Return(core.TaskExecutionIdentifier{})
	taskCtx.On("GetTaskExecutionID").Return(id)
	return taskCtx
}

func init() {
	_ = flytek8s.InitializeFake()
	cacheMock := mocks.Cache{}
	if err := flytek8s.InjectCache(&cacheMock); err != nil {
		panic(err)
	}
}

func createExecutorInitializationParams(t testing.TB, evtRecorder *mocks2.EventRecorder) types.ExecutorInitializationParameters {
	store, err := storage.NewDataStore(&storage.Config{
		Type: storage.TypeMemory,
	}, promutils.NewTestScope())
	assert.NoError(t, err)

	return types.ExecutorInitializationParameters{
		EventRecorder: evtRecorder,
		DataStore:     store,
		EnqueueOwner:  func(ownerId k8stypes.NamespacedName) error { return nil },
		OwnerKind:     "x",
		MetricsScope:  promutils.NewTestScope(),
	}
}

func TestK8sTaskExecutor_StartTask(t *testing.T) {

	ctx := context.TODO()
	tctx := getMockTaskContext()
	var tmpl *core.TaskTemplate
	var inputs *core.LiteralMap
	c := flytek8s.InitializeFake()

	t.Run("jobQueued", func(t *testing.T) {
		// common setup code
		mockResourceHandler := &mocks.K8sResourceHandler{}
		evRecorder := &mocks2.EventRecorder{}
		k := flytek8s.NewK8sTaskExecutorForResource("x", &v1.Pod{}, mockResourceHandler, time.Second)
		mockResourceHandler.On("BuildResource", mock.Anything, tctx, tmpl, inputs).Return(&v1.Pod{}, nil)
		err := k.Initialize(ctx, createExecutorInitializationParams(t, evRecorder))
		assert.NoError(t, err)

		evRecorder.On("RecordTaskEvent", mock.MatchedBy(func(c context.Context) bool { return true }),
			mock.MatchedBy(func(e *event.TaskExecutionEvent) bool { return e.Phase == core.TaskExecution_QUEUED })).Return(nil)

		status, err := k.StartTask(ctx, tctx, nil, nil)
		assert.NoError(t, err)
		assert.Nil(t, status.State)
		assert.Equal(t, types.TaskPhaseQueued, status.Phase)
		createdPod := &v1.Pod{}
		flytek8s.AddObjectMetadata(tctx, createdPod)
		assert.NoError(t, c.Get(ctx, k8stypes.NamespacedName{Namespace: tctx.GetNamespace(), Name: tctx.GetTaskExecutionID().GetGeneratedName()}, createdPod))
		assert.Equal(t, tctx.GetTaskExecutionID().GetGeneratedName(), createdPod.Name)
		assert.NoError(t, c.Delete(ctx, createdPod))
	})

	t.Run("jobAlreadyExists", func(t *testing.T) {
		// common setup code
		mockResourceHandler := &mocks.K8sResourceHandler{}
		evRecorder := &mocks2.EventRecorder{}
		k := flytek8s.NewK8sTaskExecutorForResource("x", &v1.Pod{}, mockResourceHandler, time.Second)
		mockResourceHandler.On("BuildResource", mock.Anything, tctx, tmpl, inputs).Return(&v1.Pod{}, nil)
		err := k.Initialize(ctx, createExecutorInitializationParams(t, evRecorder))
		assert.NoError(t, err)

		expectedNewStatus := types.TaskStatusQueued
		mockResourceHandler.On("GetTaskStatus", mock.Anything, mock.Anything, mock.MatchedBy(func(o *v1.Pod) bool { return true })).Return(expectedNewStatus, nil, nil)

		evRecorder.On("RecordTaskEvent", mock.MatchedBy(func(c context.Context) bool { return true }),
			mock.MatchedBy(func(e *event.TaskExecutionEvent) bool { return e.Phase == core.TaskExecution_QUEUED })).Return(nil)

		status, err := k.StartTask(ctx, tctx, nil, nil)
		assert.NoError(t, err)
		assert.Nil(t, status.State)
		assert.Equal(t, types.TaskPhaseQueued, status.Phase)
		createdPod := &v1.Pod{}
		flytek8s.AddObjectMetadata(tctx, createdPod)
		assert.NoError(t, c.Get(ctx, k8stypes.NamespacedName{Namespace: tctx.GetNamespace(), Name: tctx.GetTaskExecutionID().GetGeneratedName()}, createdPod))
		assert.Equal(t, tctx.GetTaskExecutionID().GetGeneratedName(), createdPod.Name)
	})

	t.Run("jobDifferentTerminalState", func(t *testing.T) {
		// common setup code
		mockResourceHandler := &mocks.K8sResourceHandler{}
		evRecorder := &mocks2.EventRecorder{}
		k := flytek8s.NewK8sTaskExecutorForResource("x", &v1.Pod{}, mockResourceHandler, time.Second)
		mockResourceHandler.On("BuildResource", mock.Anything, tctx, tmpl, inputs).Return(&v1.Pod{}, nil)
		err := k.Initialize(ctx, createExecutorInitializationParams(t, evRecorder))
		assert.NoError(t, err)

		expectedNewStatus := types.TaskStatusQueued
		mockResourceHandler.On("GetTaskStatus", mock.Anything, mock.Anything, mock.MatchedBy(func(o *v1.Pod) bool { return true })).Return(expectedNewStatus, nil, nil)

		evRecorder.On("RecordTaskEvent", mock.MatchedBy(func(c context.Context) bool { return true }),
			mock.MatchedBy(func(e *event.TaskExecutionEvent) bool {
				return e.Phase == core.TaskExecution_QUEUED
			})).Return(&eventErrors.EventError{Code: eventErrors.EventAlreadyInTerminalStateError,
			Cause: errors.New("already exists"),
		})

		status, err := k.StartTask(ctx, tctx, nil, nil)
		assert.NoError(t, err)
		assert.Nil(t, status.State)
		assert.Equal(t, types.TaskPhasePermanentFailure, status.Phase)
	})

	t.Run("jobQuotaExceeded", func(t *testing.T) {
		// common setup code
		mockResourceHandler := &mocks.K8sResourceHandler{}
		evRecorder := &mocks2.EventRecorder{}
		k := flytek8s.NewK8sTaskExecutorForResource("x", &v1.Pod{}, mockResourceHandler, time.Second)
		mockResourceHandler.On("BuildResource", mock.Anything, tctx, tmpl, inputs).Return(&v1.Pod{}, nil)
		err := k.Initialize(ctx, createExecutorInitializationParams(t, evRecorder))
		assert.NoError(t, err)

		evRecorder.On("RecordTaskEvent", mock.MatchedBy(func(c context.Context) bool { return true }),
			mock.MatchedBy(func(e *event.TaskExecutionEvent) bool { return e.Phase == core.TaskExecution_QUEUED })).Return(nil)

		// override create to return quota exceeded
		mockRuntimeClient := mocks.NewMockRuntimeClient()
		mockRuntimeClient.CreateCb = func(ctx context.Context, obj runtime.Object) (err error) {
			return k8serrs.NewForbidden(schema.GroupResource{}, "", errors.New("exceeded quota"))
		}
		if err := flytek8s.InjectClient(mockRuntimeClient); err != nil {
			assert.NoError(t, err)
		}

		status, err := k.StartTask(ctx, tctx, nil, nil)
		assert.NoError(t, err)
		assert.Nil(t, status.State)
		assert.Equal(t, types.TaskPhaseNotReady, status.Phase)

		// reset the client back to fake client
		if err := flytek8s.InjectClient(fake.NewFakeClient()); err != nil {
			assert.NoError(t, err)
		}
	})

	t.Run("jobForbidden", func(t *testing.T) {
		// common setup code
		mockResourceHandler := &mocks.K8sResourceHandler{}
		evRecorder := &mocks2.EventRecorder{}
		k := flytek8s.NewK8sTaskExecutorForResource("x", &v1.Pod{}, mockResourceHandler, time.Second)
		mockResourceHandler.On("BuildResource", mock.Anything, tctx, tmpl, inputs).Return(&v1.Pod{}, nil)
		err := k.Initialize(ctx, createExecutorInitializationParams(t, evRecorder))
		assert.NoError(t, err)

		evRecorder.On("RecordTaskEvent", mock.MatchedBy(func(c context.Context) bool { return true }),
			mock.MatchedBy(func(e *event.TaskExecutionEvent) bool { return e.Phase == core.TaskExecution_FAILED })).Return(nil)

		// override create to return forbidden
		mockRuntimeClient := mocks.NewMockRuntimeClient()
		mockRuntimeClient.CreateCb = func(ctx context.Context, obj runtime.Object) (err error) {
			return k8serrs.NewForbidden(schema.GroupResource{}, "", nil)
		}
		if err := flytek8s.InjectClient(mockRuntimeClient); err != nil {
			assert.NoError(t, err)
		}

		status, err := k.StartTask(ctx, tctx, nil, nil)
		assert.NoError(t, err)
		assert.Nil(t, status.State)
		assert.Equal(t, types.TaskPhasePermanentFailure, status.Phase)

		// reset the client back to fake client
		if err := flytek8s.InjectClient(fake.NewFakeClient()); err != nil {
			assert.NoError(t, err)
		}
	})
}

func TestK8sTaskExecutor_CheckTaskStatus(t *testing.T) {
	ctx := context.TODO()
	c := flytek8s.InitializeFake()

	t.Run("phaseChange", func(t *testing.T) {

		evRecorder := &mocks2.EventRecorder{}
		tctx := getMockTaskContext()
		mockResourceHandler := &mocks.K8sResourceHandler{}
		k := flytek8s.NewK8sTaskExecutorForResource("x", &v1.Pod{}, mockResourceHandler, time.Second)
		mockResourceHandler.On("BuildIdentityResource", mock.Anything, tctx).Return(&v1.Pod{}, nil)

		assert.NoError(t, k.Initialize(ctx, createExecutorInitializationParams(t, evRecorder)))
		testPod := &v1.Pod{}
		testPod.SetName(tctx.GetTaskExecutionID().GetGeneratedName())
		testPod.SetNamespace(tctx.GetNamespace())
		testPod.SetOwnerReferences([]v12.OwnerReference{tctx.GetOwnerReference()})
		err := c.Delete(ctx, testPod)
		if err != nil {
			assert.True(t, k8serrs.IsNotFound(err))
		}

		assert.NoError(t, c.Create(ctx, testPod))
		defer func() {
			assert.NoError(t, c.Delete(ctx, testPod))
		}()

		info := &events.TaskEventInfo{}
		info.Logs = []*core.TaskLog{{Uri: "log1"}}
		expectedOldPhase := types.TaskPhaseQueued
		expectedNewStatus := types.TaskStatusRunning
		expectedNewStatus.PhaseVersion = uint32(1)
		tctx.On("GetPhase").Return(expectedOldPhase)
		tctx.On("GetPhaseVersion").Return(uint32(1))
		mockResourceHandler.On("GetTaskStatus", mock.Anything, mock.Anything, mock.MatchedBy(func(o *v1.Pod) bool { return true })).Return(expectedNewStatus, nil, nil)

		evRecorder.On("RecordTaskEvent", mock.MatchedBy(func(c context.Context) bool { return true }),
			mock.MatchedBy(func(e *event.TaskExecutionEvent) bool { return true })).Return(nil)

		s, err := k.CheckTaskStatus(ctx, tctx, nil)
		assert.Nil(t, s.State)
		assert.NoError(t, err)
		assert.Equal(t, expectedNewStatus, s)
	})

	t.Run("PhaseMismatch", func(t *testing.T) {

		evRecorder := &mocks2.EventRecorder{}
		tctx := getMockTaskContext()
		mockResourceHandler := &mocks.K8sResourceHandler{}
		k := flytek8s.NewK8sTaskExecutorForResource("x", &v1.Pod{}, mockResourceHandler, time.Second)
		mockResourceHandler.On("BuildIdentityResource", mock.Anything, tctx).Return(&v1.Pod{}, nil)

		assert.NoError(t, k.Initialize(ctx, createExecutorInitializationParams(t, evRecorder)))
		testPod := &v1.Pod{}
		testPod.SetName(tctx.GetTaskExecutionID().GetGeneratedName())
		testPod.SetNamespace(tctx.GetNamespace())
		testPod.SetOwnerReferences([]v12.OwnerReference{tctx.GetOwnerReference()})
		err := c.Delete(ctx, testPod)
		if err != nil {
			assert.True(t, k8serrs.IsNotFound(err))
		}

		assert.NoError(t, c.Create(ctx, testPod))
		defer func() {
			assert.NoError(t, c.Delete(ctx, testPod))
		}()

		info := &events.TaskEventInfo{}
		info.Logs = []*core.TaskLog{{Uri: "log1"}}
		expectedOldPhase := types.TaskPhaseRunning
		expectedNewStatus := types.TaskStatusSucceeded
		expectedNewStatus.PhaseVersion = uint32(1)
		tctx.On("GetPhase").Return(expectedOldPhase)
		tctx.On("GetPhaseVersion").Return(uint32(1))
		mockResourceHandler.On("GetTaskStatus", mock.Anything, mock.Anything, mock.MatchedBy(func(o *v1.Pod) bool { return true })).Return(expectedNewStatus, nil, nil)

		evRecorder.On("RecordTaskEvent", mock.MatchedBy(func(c context.Context) bool { return true }),
			mock.MatchedBy(func(e *event.TaskExecutionEvent) bool {
				return e.Phase == core.TaskExecution_SUCCEEDED
			})).Return(&eventErrors.EventError{Code: eventErrors.EventAlreadyInTerminalStateError,
			Cause: errors.New("already exists"),
		})

		s, err := k.CheckTaskStatus(ctx, tctx, nil)
		assert.NoError(t, err)
		assert.Nil(t, s.State)
		assert.Equal(t, types.TaskPhasePermanentFailure, s.Phase)
	})

	t.Run("noChange", func(t *testing.T) {

		evRecorder := &mocks2.EventRecorder{}
		tctx := getMockTaskContext()
		mockResourceHandler := &mocks.K8sResourceHandler{}
		k := flytek8s.NewK8sTaskExecutorForResource("x", &v1.Pod{}, mockResourceHandler, time.Second)
		mockResourceHandler.On("BuildIdentityResource", mock.Anything, tctx).Return(&v1.Pod{}, nil)

		assert.NoError(t, k.Initialize(ctx, createExecutorInitializationParams(t, evRecorder)))
		testPod := &v1.Pod{}
		testPod.SetName(tctx.GetTaskExecutionID().GetGeneratedName())
		testPod.SetNamespace(tctx.GetNamespace())
		testPod.SetOwnerReferences([]v12.OwnerReference{tctx.GetOwnerReference()})
		err := c.Delete(ctx, testPod)
		if err != nil {
			assert.True(t, k8serrs.IsNotFound(err))
		}

		assert.NoError(t, c.Create(ctx, testPod))
		defer func() {
			assert.NoError(t, c.Delete(ctx, testPod))
		}()

		info := &events.TaskEventInfo{}
		info.Logs = []*core.TaskLog{{Uri: "log1"}}
		expectedOldPhase := types.TaskPhaseRunning
		expectedNewStatus := types.TaskStatusRunning
		expectedNewStatus.PhaseVersion = uint32(1)
		tctx.On("GetPhase").Return(expectedOldPhase)
		tctx.On("GetPhaseVersion").Return(uint32(1))
		mockResourceHandler.On("GetTaskStatus", mock.Anything, mock.Anything, mock.MatchedBy(func(o *v1.Pod) bool { return true })).Return(expectedNewStatus, nil, nil)

		s, err := k.CheckTaskStatus(ctx, tctx, nil)
		assert.Nil(t, s.State)
		assert.NoError(t, err)
		assert.Equal(t, expectedNewStatus, s)
	})

	t.Run("resourceNotFound", func(t *testing.T) {

		evRecorder := &mocks2.EventRecorder{}
		tctx := getMockTaskContext()
		mockResourceHandler := &mocks.K8sResourceHandler{}
		k := flytek8s.NewK8sTaskExecutorForResource("x", &v1.Pod{}, mockResourceHandler, time.Second)
		mockResourceHandler.On("BuildIdentityResource", mock.Anything, tctx).Return(&v1.Pod{}, nil)

		assert.NoError(t, k.Initialize(ctx, createExecutorInitializationParams(t, evRecorder)))
		_ = flytek8s.InitializeFake()

		info := &events.TaskEventInfo{}
		info.Logs = []*core.TaskLog{{Uri: "log1"}}
		expectedOldPhase := types.TaskPhaseRunning
		tctx.On("GetPhase").Return(expectedOldPhase)
		tctx.On("GetPhaseVersion").Return(uint32(0))

		evRecorder.On("RecordTaskEvent", mock.MatchedBy(func(c context.Context) bool { return true }),
			mock.MatchedBy(func(e *event.TaskExecutionEvent) bool { return true })).Return(nil)

		s, err := k.CheckTaskStatus(ctx, tctx, nil)
		assert.Nil(t, s.State)
		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhaseRetryableFailure, s.Phase, "Expected failure got %s", s.Phase.String())
	})

	t.Run("errorFileExit", func(t *testing.T) {

		evRecorder := &mocks2.EventRecorder{}
		tctx := getMockTaskContext()
		mockResourceHandler := &mocks.K8sResourceHandler{}
		k := flytek8s.NewK8sTaskExecutorForResource("x", &v1.Pod{}, mockResourceHandler, time.Second)
		mockResourceHandler.On("BuildIdentityResource", mock.Anything, tctx).Return(&v1.Pod{}, nil)

		params := createExecutorInitializationParams(t, evRecorder)
		store := params.DataStore
		assert.NoError(t, k.Initialize(ctx, params))
		testPod := &v1.Pod{}
		testPod.SetName(tctx.GetTaskExecutionID().GetGeneratedName())
		testPod.SetNamespace(tctx.GetNamespace())
		testPod.SetOwnerReferences([]v12.OwnerReference{tctx.GetOwnerReference()})

		err := c.Delete(ctx, testPod)
		if err != nil {
			assert.True(t, k8serrs.IsNotFound(err))
		}

		assert.NoError(t, c.Create(ctx, testPod))
		defer func() {
			assert.NoError(t, c.Delete(ctx, testPod))
		}()

		assert.NoError(t, store.WriteProtobuf(ctx, tctx.GetErrorFile(), storage.Options{}, &core.ErrorDocument{
			Error: &core.ContainerError{
				Kind:    core.ContainerError_NON_RECOVERABLE,
				Code:    "code",
				Message: "pleh",
			},
		}))

		info := &events.TaskEventInfo{}
		info.Logs = []*core.TaskLog{{Uri: "log1"}}
		expectedOldPhase := types.TaskPhaseQueued
		tctx.On("GetPhase").Return(expectedOldPhase)
		tctx.On("GetPhaseVersion").Return(uint32(0))
		mockResourceHandler.On("GetTaskStatus", mock.Anything, mock.Anything, mock.MatchedBy(func(o *v1.Pod) bool { return true })).Return(types.TaskStatusSucceeded, nil, nil)

		evRecorder.On("RecordTaskEvent", mock.MatchedBy(func(c context.Context) bool { return true }),
			mock.MatchedBy(func(e *event.TaskExecutionEvent) bool { return true })).Return(nil)

		s, err := k.CheckTaskStatus(ctx, tctx, nil)
		assert.Nil(t, s.State)
		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhasePermanentFailure, s.Phase)
	})

	t.Run("errorFileExitRecoverable", func(t *testing.T) {
		evRecorder := &mocks2.EventRecorder{}
		tctx := getMockTaskContext()
		mockResourceHandler := &mocks.K8sResourceHandler{}
		k := flytek8s.NewK8sTaskExecutorForResource("x", &v1.Pod{}, mockResourceHandler, time.Second)
		mockResourceHandler.On("BuildIdentityResource", mock.Anything, tctx).Return(&v1.Pod{}, nil)

		params := createExecutorInitializationParams(t, evRecorder)
		store := params.DataStore

		assert.NoError(t, k.Initialize(ctx, params))
		testPod := &v1.Pod{}
		testPod.SetName(tctx.GetTaskExecutionID().GetGeneratedName())
		testPod.SetNamespace(tctx.GetNamespace())
		testPod.SetOwnerReferences([]v12.OwnerReference{tctx.GetOwnerReference()})

		err := c.Delete(ctx, testPod)
		if err != nil {
			assert.True(t, k8serrs.IsNotFound(err))
		}

		assert.NoError(t, c.Create(ctx, testPod))
		defer func() {
			assert.NoError(t, c.Delete(ctx, testPod))
		}()

		assert.NoError(t, store.WriteProtobuf(ctx, tctx.GetErrorFile(), storage.Options{}, &core.ErrorDocument{
			Error: &core.ContainerError{
				Kind:    core.ContainerError_RECOVERABLE,
				Code:    "code",
				Message: "pleh",
			},
		}))

		info := &events.TaskEventInfo{}
		info.Logs = []*core.TaskLog{{Uri: "log1"}}
		expectedOldPhase := types.TaskPhaseQueued
		tctx.On("GetPhase").Return(expectedOldPhase)
		tctx.On("GetPhaseVersion").Return(uint32(0))
		mockResourceHandler.On("GetTaskStatus", mock.Anything, mock.Anything, mock.MatchedBy(func(o *v1.Pod) bool { return true })).Return(types.TaskStatusSucceeded, nil, nil)

		evRecorder.On("RecordTaskEvent", mock.MatchedBy(func(c context.Context) bool { return true }),
			mock.MatchedBy(func(e *event.TaskExecutionEvent) bool { return e.Phase == core.TaskExecution_FAILED })).Return(nil)

		s, err := k.CheckTaskStatus(ctx, tctx, nil)
		assert.Nil(t, s.State)
		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhaseRetryableFailure, s.Phase)
	})

	t.Run("nodeGetsDeleted", func(t *testing.T) {
		evRecorder := &mocks2.EventRecorder{}
		tctx := getMockTaskContext()
		mockResourceHandler := &mocks.K8sResourceHandler{}
		k := flytek8s.NewK8sTaskExecutorForResource("x", &v1.Pod{}, mockResourceHandler, time.Second)
		mockResourceHandler.On("BuildIdentityResource", mock.Anything, tctx).Return(&v1.Pod{}, nil)
		params := createExecutorInitializationParams(t, evRecorder)
		assert.NoError(t, k.Initialize(ctx, params))
		testPod := &v1.Pod{}
		testPod.SetName(tctx.GetTaskExecutionID().GetGeneratedName())
		testPod.SetNamespace(tctx.GetNamespace())
		testPod.SetOwnerReferences([]v12.OwnerReference{tctx.GetOwnerReference()})

		testPod.SetFinalizers([]string{"test_finalizer"})

		// Ensure that the pod is not there
		err := c.Delete(ctx, testPod)
		assert.True(t, k8serrs.IsNotFound(err))

		// Add a deletion timestamp to the pod definition and then create it
		tt := time.Now()
		k8sTime := v12.Time{
			Time: tt,
		}
		testPod.SetDeletionTimestamp(&k8sTime)
		assert.NoError(t, c.Create(ctx, testPod))

		// Make sure that the phase doesn't change so no events are recorded
		expectedOldPhase := types.TaskPhaseQueued
		tctx.On("GetPhase").Return(expectedOldPhase)
		tctx.On("GetPhaseVersion").Return(uint32(0))
		mockResourceHandler.On("GetTaskStatus", mock.Anything, mock.Anything, mock.MatchedBy(func(o *v1.Pod) bool { return true })).Return(types.TaskStatusQueued, nil, nil)

		s, err := k.CheckTaskStatus(ctx, tctx, nil)
		assert.Nil(t, s.State)
		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhaseRetryableFailure, s.Phase)
	})
}

func TestK8sTaskExecutor_HandleTaskSuccess(t *testing.T) {
	ctx := context.TODO()

	tctx := getMockTaskContext()
	mockResourceHandler := &mocks.K8sResourceHandler{}
	k := flytek8s.NewK8sTaskExecutorForResource("x", &v1.Pod{}, mockResourceHandler, time.Second)

	t.Run("no-errorfile", func(t *testing.T) {
		assert.NoError(t, k.Initialize(ctx, createExecutorInitializationParams(t, nil)))
		s, err := k.HandleTaskSuccess(ctx, tctx)
		assert.NoError(t, err)
		assert.Equal(t, s.Phase, types.TaskPhaseSucceeded)
	})

	t.Run("retryable-error", func(t *testing.T) {
		params := createExecutorInitializationParams(t, nil)
		store := params.DataStore
		msg := &core.ErrorDocument{
			Error: &core.ContainerError{
				Kind:    core.ContainerError_RECOVERABLE,
				Code:    "x",
				Message: "y",
			},
		}
		assert.NoError(t, store.WriteProtobuf(ctx, tctx.GetErrorFile(), storage.Options{}, msg))
		assert.NoError(t, k.Initialize(ctx, params))
		s, err := k.HandleTaskSuccess(ctx, tctx)
		assert.NoError(t, err)
		assert.Equal(t, s.Phase, types.TaskPhaseRetryableFailure)
		c, ok := taskerrs.GetErrorCode(s.Err)
		assert.True(t, ok)
		assert.Equal(t, c, "x")
	})

	t.Run("nonretryable-error", func(t *testing.T) {
		params := createExecutorInitializationParams(t, nil)
		store := params.DataStore
		msg := &core.ErrorDocument{
			Error: &core.ContainerError{
				Kind:    core.ContainerError_NON_RECOVERABLE,
				Code:    "m",
				Message: "n",
			},
		}
		assert.NoError(t, store.WriteProtobuf(ctx, tctx.GetErrorFile(), storage.Options{}, msg))
		assert.NoError(t, k.Initialize(ctx, params))
		s, err := k.HandleTaskSuccess(ctx, tctx)
		assert.NoError(t, err)
		assert.Equal(t, s.Phase, types.TaskPhasePermanentFailure)
		c, ok := taskerrs.GetErrorCode(s.Err)
		assert.True(t, ok)
		assert.Equal(t, c, "m")
	})

	t.Run("corrupted", func(t *testing.T) {
		params := createExecutorInitializationParams(t, nil)
		store := params.DataStore
		r := bytes.NewReader([]byte{'x'})
		assert.NoError(t, store.WriteRaw(ctx, tctx.GetErrorFile(), r.Size(), storage.Options{}, r))
		assert.NoError(t, k.Initialize(ctx, params))
		s, err := k.HandleTaskSuccess(ctx, tctx)
		assert.Error(t, err)
		assert.Equal(t, s.Phase, types.TaskPhaseUndefined)
	})
}
