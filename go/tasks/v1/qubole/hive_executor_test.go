package qubole

import (
	"context"
	"errors"
	"testing"

	eventErrors "github.com/lyft/flyteidl/clients/go/events/errors"
	"github.com/lyft/flyteplugins/go/tasks/v1/qubole/client"
	clientMocks "github.com/lyft/flyteplugins/go/tasks/v1/qubole/client/mocks"
	"github.com/lyft/flyteplugins/go/tasks/v1/resourcemanager"
	resourceManagerMocks "github.com/lyft/flyteplugins/go/tasks/v1/resourcemanager/mocks"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"
	eventMocks "github.com/lyft/flyteplugins/go/tasks/v1/types/mocks"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// This line is here because we need to generate a mock for the cache as the flytestdlib does not have one.
// Remove this if we ever do generate one inside flytestdlib itself.
//go:generate mockery -dir ../../../../vendor/github.com/lyft/flytestdlib/utils -name AutoRefreshCache

func getDummyHiveExecutor() *HiveExecutor {
	return &HiveExecutor{
		OutputsResolver: types.OutputsResolver{},
		recorder:        &eventMocks.EventRecorder{},
		id:              "test-hive-executor",
		secretsManager:  MockSecretsManager{},
		executionsCache: &MockAutoRefreshCache{},
		metrics:         getHiveExecutorMetrics(promutils.NewTestScope()),
		quboleClient:    &clientMocks.QuboleClient{},
		redisClient:     nil,
		resourceManager: resourcemanager.NoopResourceManager{},
		executionBuffer: &resourceManagerMocks.ExecutionLooksideBuffer{},
	}
}

func TestUniqueCacheKey(t *testing.T) {
	mockTaskContext := CreateMockTaskContextWithRealTaskExecId()
	executor := getDummyHiveExecutor()

	out := executor.getUniqueCacheKey(mockTaskContext, 42)
	assert.Equal(t, "test-hive-job_42", out)
}

func TestTranslateCurrentState(t *testing.T) {
	t.Run("just one item", func(t *testing.T) {
		workItems := map[string]interface{}{
			"key_1": NewQuboleWorkItem(
				"key_1",
				"12345",
				QuboleWorkSucceeded,
				"default",
				[]string{},
				0,
			),
		}

		executor := getDummyHiveExecutor()
		taskStatus := executor.TranslateCurrentState(workItems)
		assert.Equal(t, types.TaskPhaseSucceeded, taskStatus.Phase)
	})

	t.Run("partial completion", func(t *testing.T) {
		workItems := map[string]interface{}{
			"key_1": NewQuboleWorkItem(
				"key_1",
				"12345",
				QuboleWorkSucceeded,
				"default",
				[]string{},
				0,
			),
			"key_2": NewQuboleWorkItem(
				"key_2",
				"45645",
				QuboleWorkRunning,
				"default",
				[]string{},
				0,
			),
		}

		executor := getDummyHiveExecutor()
		taskStatus := executor.TranslateCurrentState(workItems)
		assert.Equal(t, types.TaskPhaseRunning, taskStatus.Phase)
	})

	t.Run("any failure is a failure", func(t *testing.T) {
		workItems := map[string]interface{}{
			"key_1": NewQuboleWorkItem(
				"key_1",
				"12345",
				QuboleWorkSucceeded,
				"default",
				[]string{},
				0,
			),
			"key_2": NewQuboleWorkItem(
				"key_2",
				"45645",
				QuboleWorkFailed,
				"default",
				[]string{},
				0,
			),
		}

		executor := getDummyHiveExecutor()
		taskStatus := executor.TranslateCurrentState(workItems)
		assert.Equal(t, types.TaskPhaseRetryableFailure, taskStatus.Phase)
	})
}

func TestHiveExecutor_StartTask(t *testing.T) {
	ctx := context.Background()
	mockContext := CreateMockTaskContextWithRealTaskExecId()
	taskTemplate := createDummyHiveTaskTemplate("hive-task-id")

	executor := getDummyHiveExecutor()
	taskStatus, err := executor.StartTask(ctx, mockContext, taskTemplate, nil)
	assert.NoError(t, err)
	assert.Equal(t, types.TaskPhaseQueued, taskStatus.Phase)
	customState := taskStatus.State
	assert.Equal(t, 1, len(customState))

	workItem := customState["test-hive-job_0"].(QuboleWorkItem)

	assert.Equal(t, []string{"tag1", "tag2", "label-1:val1", "ns:test-namespace"}, workItem.Tags)
	assert.Equal(t, "cluster-label", workItem.ClusterLabel)
	assert.Equal(t, QuboleWorkNotStarted, workItem.Status)
}

func NewMockedHiveTaskExecutor() *HiveExecutor {
	mockCache := NewMockAutoRefreshCache()
	mockQubole := &clientMocks.QuboleClient{}
	mockResourceManager := &resourceManagerMocks.ResourceManager{}
	mockEventRecorder := &eventMocks.EventRecorder{}
	mockExecutionBuffer := &resourceManagerMocks.ExecutionLooksideBuffer{}

	return &HiveExecutor{
		OutputsResolver: types.OutputsResolver{},
		recorder:        mockEventRecorder,
		id:              "test-hive-executor",
		secretsManager:  MockSecretsManager{},
		executionsCache: mockCache,
		metrics:         getHiveExecutorMetrics(promutils.NewTestScope()),
		quboleClient:    mockQubole,
		redisClient:     nil,
		resourceManager: mockResourceManager,
		executionBuffer: mockExecutionBuffer,
	}
}

func TestHiveExecutor_CheckTaskStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("basic lifecycle", func(t *testing.T) {
		taskTemplate := createDummyHiveTaskTemplate("hive-task-id")

		// Get executor and add hooks for mocks
		executor := NewMockedHiveTaskExecutor()
		executor.resourceManager.(*resourceManagerMocks.ResourceManager).On("AllocateResource",
			mock.Anything, mock.Anything, mock.Anything).Return(resourcemanager.AllocationStatusGranted, nil)

		var eventRecorded = false
		executor.recorder.(*eventMocks.EventRecorder).On("RecordTaskEvent", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			eventRecorded = true
		})
		executor.executionBuffer.(*resourceManagerMocks.ExecutionLooksideBuffer).On("RetrieveExecution",
			mock.Anything, mock.Anything).Return("", resourcemanager.ExecutionNotFoundError)
		executor.executionBuffer.(*resourceManagerMocks.ExecutionLooksideBuffer).On("ConfirmExecution",
			mock.Anything, mock.Anything, mock.Anything).Return(nil)

		// StartTask
		taskStatus_0, err := executor.StartTask(ctx, CreateMockTaskContextWithRealTaskExecId(), taskTemplate, nil)
		assert.NoError(t, err)

		// Create a new mock task context to return the first custom state, the one constructed by the StartTask call
		mockContext := CreateMockTaskContextWithRealTaskExecId()
		mockContext.On("GetCustomState").Return(taskStatus_0.State)
		mockContext.On("GetPhase").Return(types.TaskPhaseQueued)
		mockContext.On("GetPhaseVersion").Return(uint32(0))

		// Call CheckTaskStatus twice
		// The first time the mock qubole client will create the query
		executor.quboleClient.(*clientMocks.QuboleClient).On("ExecuteHiveCommand", mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
			&client.QuboleCommandDetails{
				Status: client.QuboleStatusRunning,
				ID:     55482218961153,
			}, nil)
		taskStatus_1, err := executor.CheckTaskStatus(ctx, mockContext, taskTemplate)
		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhaseRunning, taskStatus_1.Phase)
		assert.True(t, eventRecorded)
		customState := taskStatus_1.State
		assert.Equal(t, 1, len(customState))
		workItem := customState["test-hive-job_0"].(QuboleWorkItem)
		assert.Equal(t, "55482218961153", workItem.CommandId)
		assert.Equal(t, QuboleWorkRunning, workItem.Status)

		// This bit mimics what the AutoRefreshCache would've been doing in the background.  Pull out the workitem,
		// update the status to succeeded, and put back into the cache.
		cachedWorkItem := executor.executionsCache.(MockAutoRefreshCache).values["test-hive-job_0"].(QuboleWorkItem)
		cachedWorkItem.Status = QuboleWorkSucceeded
		executor.executionsCache.(MockAutoRefreshCache).values["test-hive-job_0"] = cachedWorkItem

		// Second call to CheckTaskStatus, cache will say that it's finished
		// Reset eventRecorded
		eventRecorded = false
		mockContext = CreateMockTaskContextWithRealTaskExecId()
		mockContext.On("GetCustomState").Return(taskStatus_1.State)
		mockContext.On("GetPhase").Return(types.TaskPhaseRunning)
		taskStatus_2, err := executor.CheckTaskStatus(ctx, mockContext, taskTemplate)
		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhaseSucceeded, taskStatus_2.Phase)
		customState = taskStatus_2.State
		assert.True(t, eventRecorded)
		assert.Equal(t, 1, len(customState))
		workItem = customState["test-hive-job_0"].(QuboleWorkItem)
		assert.Equal(t, "55482218961153", workItem.CommandId)
		assert.Equal(t, QuboleWorkSucceeded, workItem.Status)
	})

	t.Run("resource manager gates properly", func(t *testing.T) {
		// Get executor and add hooks for mocks
		executor := NewMockedHiveTaskExecutor()
		executor.resourceManager.(*resourceManagerMocks.ResourceManager).On("AllocateResource", mock.Anything, mock.Anything, mock.Anything).
			Return(resourcemanager.AllocationStatusExhausted, nil)
		executor.recorder.(*eventMocks.EventRecorder).On("RecordTaskEvent", mock.Anything, mock.Anything).Return(nil)
		executor.executionBuffer.(*resourceManagerMocks.ExecutionLooksideBuffer).On("RetrieveExecution",
			mock.Anything, mock.Anything).Return("", resourcemanager.ExecutionNotFoundError)

		mockContext := CreateMockTaskContextWithRealTaskExecId()
		taskTemplate := createDummyHiveTaskTemplate("hive-task-id")

		taskStatus_0, err := executor.StartTask(ctx, mockContext, taskTemplate, nil)
		assert.NoError(t, err)
		mockContext.On("GetCustomState").Return(taskStatus_0.State)
		mockContext.On("GetPhase").Return(types.TaskPhaseQueued)

		// If the AllocateResource call doesn't return successfully, the query is not kicked off
		executor.quboleClient.(*clientMocks.QuboleClient).On("ExecuteHiveCommand",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
			&client.QuboleCommandDetails{}, nil)
		taskStatus_1, err := executor.CheckTaskStatus(ctx, mockContext, taskTemplate)

		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhaseRunning, taskStatus_1.Phase)
		customState := taskStatus_1.State
		assert.Equal(t, 1, len(customState))
		workItem := customState["test-hive-job_0"].(QuboleWorkItem)
		assert.Equal(t, QuboleWorkNotStarted, workItem.Status)
		assert.Equal(t, 0, len(executor.quboleClient.(*clientMocks.QuboleClient).Calls))
	})

	t.Run("executor doesn't launch when already in buffer", func(t *testing.T) {
		var allocationCalled = false

		// Get executor and add hooks for mocks
		executor := NewMockedHiveTaskExecutor()
		executor.resourceManager.(*resourceManagerMocks.ResourceManager).On("AllocateResource", mock.Anything,
			mock.Anything, mock.Anything).Return(resourcemanager.AllocationStatusGranted, nil).Run(func(args mock.Arguments) {
			allocationCalled = true
		})
		executor.recorder.(*eventMocks.EventRecorder).On("RecordTaskEvent", mock.Anything, mock.Anything).Return(nil)
		executor.executionBuffer.(*resourceManagerMocks.ExecutionLooksideBuffer).On("RetrieveExecution",
			mock.Anything, mock.Anything).Return("123456", nil)

		taskTemplate := createDummyHiveTaskTemplate("hive-task-id")
		taskStatus_0, err := executor.StartTask(ctx, CreateMockTaskContextWithRealTaskExecId(), taskTemplate, nil)
		assert.NoError(t, err)
		workItem_0 := taskStatus_0.State["test-hive-job_0"].(QuboleWorkItem)
		// object should be initialized in the not started state
		assert.Equal(t, QuboleWorkNotStarted, workItem_0.Status)

		// Create a new mock task context to return the first custom state, the one constructed by the StartTask call
		mockContext := CreateMockTaskContextWithRealTaskExecId()
		mockContext.On("GetCustomState").Return(taskStatus_0.State)
		mockContext.On("GetPhase").Return(types.TaskPhaseQueued)
		mockContext.On("GetPhaseVersion").Return(uint32(0))

		taskStatus_1, err := executor.CheckTaskStatus(ctx, mockContext, taskTemplate)
		assert.NoError(t, err)
		assert.False(t, allocationCalled)
		assert.Equal(t, types.TaskPhaseRunning, taskStatus_1.Phase)
		customState := taskStatus_1.State
		assert.Equal(t, 1, len(customState))
		workItem_1 := customState["test-hive-job_0"].(QuboleWorkItem)
		// If found in the lookaside buffer, CheckTaskStatus should set the status to running, and set the command ID
		assert.Equal(t, "123456", workItem_1.CommandId)
		assert.Equal(t, QuboleWorkRunning, workItem_1.Status)
	})
}

func TestHiveExecutor_SyncQuboleQuery(t *testing.T) {
	ctx := context.Background()
	executor := getDummyHiveExecutor()

	t.Run("command failed", func(t *testing.T) {
		mockQubole := &clientMocks.QuboleClient{}
		mockQubole.On("GetCommandStatus", mock.Anything, mock.Anything, mock.Anything).
			Return(client.QuboleStatusError, nil)
		executor.quboleClient = mockQubole

		workItem := QuboleWorkItem{
			CommandId: "123456789",
			Status:    QuboleWorkRunning,
		}
		x, action, err := executor.SyncQuboleQuery(ctx, workItem)
		newWorkItem := x.(QuboleWorkItem)
		assert.NoError(t, err)
		assert.Equal(t, QuboleWorkFailed, newWorkItem.Status)
		assert.Equal(t, utils.Update, action)
	})

	t.Run("command still running", func(t *testing.T) {
		mockQubole := &clientMocks.QuboleClient{}
		mockQubole.On("GetCommandStatus", mock.Anything, mock.Anything, mock.Anything).
			Return(client.QuboleStatusRunning, nil)
		executor.quboleClient = mockQubole

		workItem := QuboleWorkItem{
			CommandId: "123456789",
			Status:    QuboleWorkRunning,
		}
		x, action, err := executor.SyncQuboleQuery(ctx, workItem)
		newWorkItem := x.(QuboleWorkItem)
		assert.NoError(t, err)
		assert.Equal(t, QuboleWorkRunning, newWorkItem.Status)
		assert.Equal(t, utils.Unchanged, action)
	})

	t.Run("command succeeded", func(t *testing.T) {
		mockQubole := &clientMocks.QuboleClient{}
		mockQubole.On("GetCommandStatus", mock.Anything, mock.Anything, mock.Anything).
			Return(client.QuboleStatusDone, nil)
		executor.quboleClient = mockQubole

		mockResourceManager := &resourceManagerMocks.ResourceManager{}
		mockResourceManager.On("ReleaseResource", mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		executor.resourceManager = mockResourceManager

		workItem := QuboleWorkItem{
			CommandId: "123456789",
			Status:    QuboleWorkRunning,
		}
		x, action, err := executor.SyncQuboleQuery(ctx, workItem)
		newWorkItem := x.(QuboleWorkItem)
		assert.NoError(t, err)
		assert.Equal(t, QuboleWorkSucceeded, newWorkItem.Status)
		assert.Equal(t, 1, len(mockResourceManager.Calls))
		assert.Equal(t, utils.Update, action)
	})
}

func TestHiveExecutor_CheckTaskStatusStateMismatch(t *testing.T) {
	ctx := context.Background()
	taskTemplate := createDummyHiveTaskTemplate("hive-task-id")
	mockQubole := &clientMocks.QuboleClient{}

	mockResourceManager := &resourceManagerMocks.ResourceManager{}
	mockResourceManager.On("AllocateResource", mock.Anything, mock.Anything, mock.Anything).
		Return(resourcemanager.AllocationStatusGranted, nil)
	mockEventRecorder := &eventMocks.EventRecorder{}
	mockEventRecorder.On("RecordTaskEvent", mock.Anything, mock.Anything).Return(&eventErrors.EventError{Code: eventErrors.EventAlreadyInTerminalStateError,
		Cause: errors.New("already exists"),
	})

	executor := NewMockedHiveTaskExecutor()
	executor.recorder = mockEventRecorder
	executor.resourceManager = mockResourceManager
	executor.quboleClient = mockQubole
	executor.executionBuffer.(*resourceManagerMocks.ExecutionLooksideBuffer).On("RetrieveExecution",
		mock.Anything, mock.Anything).Return("", resourcemanager.ExecutionNotFoundError)
	executor.executionBuffer.(*resourceManagerMocks.ExecutionLooksideBuffer).On("ConfirmExecution",
		mock.Anything, mock.Anything, mock.Anything).Return(nil)
	taskstatus0, err := executor.StartTask(ctx, CreateMockTaskContextWithRealTaskExecId(), taskTemplate, nil)
	assert.NoError(t, err)

	// Create a new mock task context to return the first custom state, the one constructed by the StartTask call
	mockContext := CreateMockTaskContextWithRealTaskExecId()
	mockContext.On("GetCustomState").Return(taskstatus0.State)
	mockContext.On("GetPhase").Return(types.TaskPhaseSucceeded)
	mockContext.On("GetPhaseVersion").Return(uint32(0))

	// Unit test will call CheckTaskStatus twice
	// - the first time the mock qubole client will create the query
	mockQubole.On("ExecuteHiveCommand", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).Return(
		&client.QuboleCommandDetails{
			Status: client.QuboleStatusRunning,
			ID:     55482218961153,
		}, nil)
	taskstatus1, err := executor.CheckTaskStatus(ctx, mockContext, taskTemplate)
	assert.NoError(t, err)
	assert.Equal(t, types.TaskPhasePermanentFailure, taskstatus1.Phase)
	assert.Nil(t, taskstatus1.State)
}
