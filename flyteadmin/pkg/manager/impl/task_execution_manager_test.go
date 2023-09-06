package impl

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flyteadmin/pkg/common"

	commonMocks "github.com/flyteorg/flyteadmin/pkg/common/mocks"
	dataMocks "github.com/flyteorg/flyteadmin/pkg/data/mocks"
	flyteAdminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

var taskStartedAt = time.Now().UTC()
var sampleTaskEventOccurredAt, _ = ptypes.TimestampProto(taskStartedAt)

var sampleTaskID = &core.Identifier{
	ResourceType: core.ResourceType_TASK,
	Project:      "project",
	Domain:       "domain",
	Name:         "task-id",
	Version:      "task-v",
}

var sampleNodeExecID = &core.NodeExecutionIdentifier{
	NodeId: "node-id",
	ExecutionId: &core.WorkflowExecutionIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	},
}

var taskEventRequest = admin.TaskExecutionEventRequest{
	RequestId: "request id",
	Event: &event.TaskExecutionEvent{
		ProducerId:            "propeller",
		TaskId:                sampleTaskID,
		ParentNodeExecutionId: sampleNodeExecID,
		OccurredAt:            sampleTaskEventOccurredAt,
		Phase:                 core.TaskExecution_RUNNING,
		RetryAttempt:          uint32(1),
		InputValue: &event.TaskExecutionEvent_InputUri{
			InputUri: "input uri",
		},
	},
}

var mockTaskExecutionRemoteURL = dataMocks.NewMockRemoteURL()

var retryAttemptValue = uint32(1)

func addGetWorkflowExecutionCallback(repository interfaces.Repository) {
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
			return models.Execution{
				ExecutionKey: models.ExecutionKey{
					Project: sampleNodeExecID.ExecutionId.Project,
					Domain:  sampleNodeExecID.ExecutionId.Domain,
					Name:    sampleNodeExecID.ExecutionId.Name,
				},
				Cluster: "propeller",
			}, nil
		},
	)

}

func addGetNodeExecutionCallback(repository interfaces.Repository) {
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
			return models.NodeExecution{
				NodeExecutionKey: models.NodeExecutionKey{
					NodeID: sampleNodeExecID.NodeId,
					ExecutionKey: models.ExecutionKey{
						Project: sampleNodeExecID.ExecutionId.Project,
						Domain:  sampleNodeExecID.ExecutionId.Domain,
						Name:    sampleNodeExecID.ExecutionId.Name,
					},
				},
			}, nil
		},
	)
}

func addGetTaskCallback(repository interfaces.Repository) {
	repository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetGetCallback(
		func(input interfaces.Identifier) (models.Task, error) {
			return models.Task{
				TaskKey: models.TaskKey{
					Project: sampleTaskID.Project,
					Domain:  sampleTaskID.Domain,
					Name:    sampleTaskID.Name,
					Version: sampleTaskID.Version,
				},
			}, nil
		},
	)
}

func TestCreateTaskEvent(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	addGetWorkflowExecutionCallback(repository)
	addGetNodeExecutionCallback(repository)
	addGetTaskCallback(repository)

	getTaskCalled := false
	// See if we check for existing task executions, and return none found
	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.GetTaskExecutionInput) (models.TaskExecution, error) {
			getTaskCalled = true
			assert.Equal(t, core.ResourceType_TASK, input.TaskExecutionID.TaskId.ResourceType)
			assert.Equal(t, "task-id", input.TaskExecutionID.TaskId.Name)
			assert.Equal(t, "project", input.TaskExecutionID.TaskId.Project)
			assert.Equal(t, "domain", input.TaskExecutionID.TaskId.Domain)
			assert.Equal(t, "task-v", input.TaskExecutionID.TaskId.Version)
			assert.Equal(t, "node-id", input.TaskExecutionID.NodeExecutionId.NodeId)
			assert.Equal(t, "project", input.TaskExecutionID.NodeExecutionId.ExecutionId.Project)
			assert.Equal(t, "domain", input.TaskExecutionID.NodeExecutionId.ExecutionId.Domain)
			assert.Equal(t, "name", input.TaskExecutionID.NodeExecutionId.ExecutionId.Name)
			return models.TaskExecution{}, flyteAdminErrors.NewFlyteAdminError(codes.NotFound, "foo")
		})

	expectedClosure := admin.TaskExecutionClosure{
		Phase:     core.TaskExecution_RUNNING,
		StartedAt: sampleTaskEventOccurredAt,
		UpdatedAt: sampleTaskEventOccurredAt,
		CreatedAt: sampleTaskEventOccurredAt,
	}
	expectedClosureBytes, _ := proto.Marshal(&expectedClosure)

	createTaskCalled := false
	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetCreateCallback(
		func(ctx context.Context, input models.TaskExecution) error {
			createTaskCalled = true
			assert.Equal(t, models.TaskExecution{
				TaskExecutionKey: models.TaskExecutionKey{
					TaskKey: models.TaskKey{
						Project: sampleTaskID.Project,
						Domain:  sampleTaskID.Domain,
						Name:    sampleTaskID.Name,
						Version: sampleTaskID.Version,
					},
					NodeExecutionKey: models.NodeExecutionKey{
						NodeID: sampleNodeExecID.NodeId,
						ExecutionKey: models.ExecutionKey{
							Project: sampleNodeExecID.ExecutionId.Project,
							Domain:  sampleNodeExecID.ExecutionId.Domain,
							Name:    sampleNodeExecID.ExecutionId.Name,
						},
					},
					RetryAttempt: &retryAttemptValue,
				},
				Phase:                  core.TaskExecution_RUNNING.String(),
				InputURI:               "input uri",
				StartedAt:              &taskStartedAt,
				TaskExecutionCreatedAt: &taskStartedAt,
				TaskExecutionUpdatedAt: &taskStartedAt,
				Closure:                expectedClosureBytes,
			}, input)
			return nil
		})
	taskExecManager := NewTaskExecutionManager(repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockTaskExecutionRemoteURL, nil, nil)
	resp, err := taskExecManager.CreateTaskExecutionEvent(context.Background(), taskEventRequest)
	assert.True(t, getTaskCalled)
	assert.True(t, createTaskCalled)
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetCreateCallback(
		func(ctx context.Context, input models.TaskExecution) error {
			return errors.New("failed to insert record into task table")
		})
	taskExecManager = NewTaskExecutionManager(repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockTaskExecutionRemoteURL, nil, nil)
	resp, err = taskExecManager.CreateTaskExecutionEvent(context.Background(), taskEventRequest)
	assert.NotNil(t, err)
	assert.Nil(t, resp)
}

func TestCreateTaskEvent_Update(t *testing.T) {
	taskCompletedAt := taskStartedAt.Add(time.Minute)
	taskEventCompletedAtProto, _ := ptypes.TimestampProto(taskCompletedAt)

	repository := repositoryMocks.NewMockRepository()
	addGetWorkflowExecutionCallback(repository)
	addGetNodeExecutionCallback(repository)
	addGetTaskCallback(repository)

	getTaskCalled := false
	// See if we check for existing task executions, and return one
	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.GetTaskExecutionInput) (models.TaskExecution, error) {
			getTaskCalled = true
			runningTaskClosure := &admin.TaskExecutionClosure{
				StartedAt: sampleTaskEventOccurredAt,
				CreatedAt: sampleTaskEventOccurredAt,
				UpdatedAt: sampleTaskEventOccurredAt,
				Phase:     core.TaskExecution_RUNNING,
			}
			runningTaskClosureBytes, _ := proto.Marshal(runningTaskClosure)

			return models.TaskExecution{
				TaskExecutionKey: models.TaskExecutionKey{
					TaskKey: models.TaskKey{
						Project: sampleTaskID.Project,
						Domain:  sampleTaskID.Domain,
						Name:    sampleTaskID.Name,
						Version: sampleTaskID.Version,
					},
					NodeExecutionKey: models.NodeExecutionKey{
						NodeID: sampleNodeExecID.NodeId,
						ExecutionKey: models.ExecutionKey{
							Project: sampleNodeExecID.ExecutionId.Project,
							Domain:  sampleNodeExecID.ExecutionId.Domain,
							Name:    sampleNodeExecID.ExecutionId.Name,
						},
					},
				},
				InputURI:               "input uri",
				Closure:                runningTaskClosureBytes,
				StartedAt:              &taskStartedAt,
				TaskExecutionCreatedAt: &taskStartedAt,
				TaskExecutionUpdatedAt: &taskStartedAt,
				Phase:                  core.TaskExecution_RUNNING.String(),
			}, nil
		},
	)

	expectedLogs := []*core.TaskLog{{Uri: "test-log1.txt"}}
	expectedOutputResult := &admin.TaskExecutionClosure_OutputUri{
		OutputUri: "test-output.pb",
	}
	expectedClosure := &admin.TaskExecutionClosure{
		StartedAt:    sampleTaskEventOccurredAt,
		CreatedAt:    sampleTaskEventOccurredAt,
		UpdatedAt:    taskEventCompletedAtProto,
		Phase:        core.TaskExecution_SUCCEEDED,
		Duration:     ptypes.DurationProto(time.Minute),
		OutputResult: expectedOutputResult,
		Logs:         expectedLogs,
	}

	expectedClosureBytes, _ := proto.Marshal(expectedClosure)

	updateTaskCalled := false
	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetUpdateCallback(
		func(ctx context.Context, input models.TaskExecution) error {
			updateTaskCalled = true
			assert.EqualValues(t, models.TaskExecution{
				TaskExecutionKey: models.TaskExecutionKey{
					TaskKey: models.TaskKey{
						Project: sampleTaskID.Project,
						Domain:  sampleTaskID.Domain,
						Name:    sampleTaskID.Name,
						Version: sampleTaskID.Version,
					},
					NodeExecutionKey: models.NodeExecutionKey{
						NodeID: sampleNodeExecID.NodeId,
						ExecutionKey: models.ExecutionKey{
							Project: sampleNodeExecID.ExecutionId.Project,
							Domain:  sampleNodeExecID.ExecutionId.Domain,
							Name:    sampleNodeExecID.ExecutionId.Name,
						},
					},
				},
				Phase:                  core.TaskExecution_SUCCEEDED.String(),
				InputURI:               "input uri",
				StartedAt:              &taskStartedAt,
				TaskExecutionUpdatedAt: &taskCompletedAt,
				TaskExecutionCreatedAt: &taskStartedAt,
				Closure:                expectedClosureBytes,
				Duration:               time.Minute,
			}, input)
			return nil
		})

	// fill in event for a completed task execution
	taskEventRequest.Event.Phase = core.TaskExecution_SUCCEEDED
	taskEventRequest.Event.OccurredAt = taskEventCompletedAtProto
	taskEventRequest.Event.Logs = expectedLogs
	taskEventRequest.Event.OutputResult = &event.TaskExecutionEvent_OutputUri{
		OutputUri: expectedOutputResult.OutputUri,
	}

	taskExecManager := NewTaskExecutionManager(repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockTaskExecutionRemoteURL, &mockPublisher, &mockPublisher)
	resp, err := taskExecManager.CreateTaskExecutionEvent(context.Background(), taskEventRequest)
	assert.True(t, getTaskCalled)
	assert.True(t, updateTaskCalled)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestCreateTaskEvent_MissingExecution(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	expectedErr := flyteAdminErrors.NewFlyteAdminErrorf(codes.Internal, "expected error")
	addGetWorkflowExecutionCallback(repository)
	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.GetTaskExecutionInput) (models.TaskExecution, error) {
			return models.TaskExecution{}, flyteAdminErrors.NewFlyteAdminError(codes.NotFound, "foo")
		})
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetExistsCallback(
		func(
			ctx context.Context, input interfaces.NodeExecutionResource) (bool, error) {
			return false, expectedErr
		})
	taskExecManager := NewTaskExecutionManager(repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockTaskExecutionRemoteURL, nil, nil)
	resp, err := taskExecManager.CreateTaskExecutionEvent(context.Background(), taskEventRequest)
	assert.EqualError(t, err, "Failed to get existing node execution id: [node_id:\"node-id\""+
		" execution_id:<project:\"project\" domain:\"domain\" name:\"name\" > ] "+
		"with err: expected error")
	assert.Nil(t, resp)

	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetExistsCallback(
		func(
			ctx context.Context, input interfaces.NodeExecutionResource) (bool, error) {
			return false, nil
		})
	taskExecManager = NewTaskExecutionManager(repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockTaskExecutionRemoteURL, nil, nil)
	resp, err = taskExecManager.CreateTaskExecutionEvent(context.Background(), taskEventRequest)
	assert.EqualError(t, err, "failed to get existing node execution id: [node_id:\"node-id\""+
		" execution_id:<project:\"project\" domain:\"domain\" name:\"name\" > ]")
	assert.Nil(t, resp)
}

func TestCreateTaskEvent_CreateDatabaseError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	addGetWorkflowExecutionCallback(repository)
	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.GetTaskExecutionInput) (models.TaskExecution, error) {
			return models.TaskExecution{}, flyteAdminErrors.NewFlyteAdminError(codes.NotFound, "foo")
		})

	expectedErr := errors.New("expected error")
	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetCreateCallback(
		func(ctx context.Context, input models.TaskExecution) error {
			return expectedErr
		})
	taskExecManager := NewTaskExecutionManager(repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockTaskExecutionRemoteURL, nil, nil)
	resp, err := taskExecManager.CreateTaskExecutionEvent(context.Background(), taskEventRequest)
	assert.EqualError(t, err, expectedErr.Error())
	assert.Nil(t, resp)
}

func TestCreateTaskEvent_UpdateDatabaseError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	addGetWorkflowExecutionCallback(repository)
	addGetNodeExecutionCallback(repository)

	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.GetTaskExecutionInput) (models.TaskExecution, error) {
			return models.TaskExecution{
				TaskExecutionKey: models.TaskExecutionKey{
					TaskKey: models.TaskKey{
						Project: sampleTaskID.Project,
						Domain:  sampleTaskID.Domain,
						Name:    sampleTaskID.Name,
						Version: sampleTaskID.Version,
					},
					NodeExecutionKey: models.NodeExecutionKey{
						NodeID: sampleNodeExecID.NodeId,
						ExecutionKey: models.ExecutionKey{
							Project: sampleNodeExecID.ExecutionId.Project,
							Domain:  sampleNodeExecID.ExecutionId.Domain,
							Name:    sampleNodeExecID.ExecutionId.Name,
						},
					},
					RetryAttempt: &retryAttemptValue,
				},
				StartedAt: &taskStartedAt,
				Phase:     core.TaskExecution_RUNNING.String(),
			}, nil
		})

	expectedErr := errors.New("expected error")
	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetUpdateCallback(
		func(ctx context.Context, execution models.TaskExecution) error {
			return expectedErr
		})
	nodeExecManager := NewTaskExecutionManager(repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockTaskExecutionRemoteURL, nil, nil)
	resp, err := nodeExecManager.CreateTaskExecutionEvent(context.Background(), taskEventRequest)
	assert.EqualError(t, err, expectedErr.Error())
	assert.Nil(t, resp)
}

func TestCreateTaskEvent_UpdateTerminalEventError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	addGetWorkflowExecutionCallback(repository)
	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.GetTaskExecutionInput) (models.TaskExecution, error) {
			return models.TaskExecution{
				TaskExecutionKey: models.TaskExecutionKey{
					TaskKey: models.TaskKey{
						Project: sampleTaskID.Project,
						Domain:  sampleTaskID.Domain,
						Name:    sampleTaskID.Name,
						Version: sampleTaskID.Version,
					},
					NodeExecutionKey: models.NodeExecutionKey{
						NodeID: sampleNodeExecID.NodeId,
						ExecutionKey: models.ExecutionKey{
							Project: sampleNodeExecID.ExecutionId.Project,
							Domain:  sampleNodeExecID.ExecutionId.Domain,
							Name:    sampleNodeExecID.ExecutionId.Name,
						},
					},
					RetryAttempt: &retryAttemptValue,
				},
				StartedAt: &taskStartedAt,
				Phase:     core.TaskExecution_SUCCEEDED.String(),
			}, nil
		})
	taskEventRequest.Event.Phase = core.TaskExecution_RUNNING
	taskExecManager := NewTaskExecutionManager(repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockTaskExecutionRemoteURL, nil, nil)
	resp, err := taskExecManager.CreateTaskExecutionEvent(context.Background(), taskEventRequest)

	assert.Nil(t, resp)

	adminError := err.(flyteAdminErrors.FlyteAdminError)
	assert.Equal(t, adminError.Code(), codes.FailedPrecondition)
	details, ok := adminError.GRPCStatus().Details()[0].(*admin.EventFailureReason)
	assert.True(t, ok)
	_, ok = details.GetReason().(*admin.EventFailureReason_AlreadyInTerminalState)
	assert.True(t, ok)
}

func TestCreateTaskEvent_PhaseVersionChange(t *testing.T) {
	taskUpdatedAt := taskStartedAt.Add(time.Minute)
	taskEventUpdatedAtProto, _ := ptypes.TimestampProto(taskUpdatedAt)

	repository := repositoryMocks.NewMockRepository()
	addGetWorkflowExecutionCallback(repository)
	addGetNodeExecutionCallback(repository)
	addGetTaskCallback(repository)

	getTaskCalled := false
	// See if we check for existing task executions, and return one
	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.GetTaskExecutionInput) (models.TaskExecution, error) {
			getTaskCalled = true

			return models.TaskExecution{
				TaskExecutionKey: models.TaskExecutionKey{
					TaskKey: models.TaskKey{
						Project: sampleTaskID.Project,
						Domain:  sampleTaskID.Domain,
						Name:    sampleTaskID.Name,
						Version: sampleTaskID.Version,
					},
					NodeExecutionKey: models.NodeExecutionKey{
						NodeID: sampleNodeExecID.NodeId,
						ExecutionKey: models.ExecutionKey{
							Project: sampleNodeExecID.ExecutionId.Project,
							Domain:  sampleNodeExecID.ExecutionId.Domain,
							Name:    sampleNodeExecID.ExecutionId.Name,
						},
					},
				},
				InputURI:               "input uri",
				StartedAt:              &taskStartedAt,
				TaskExecutionCreatedAt: &taskStartedAt,
				TaskExecutionUpdatedAt: &taskStartedAt,
				Phase:                  core.TaskExecution_RUNNING.String(),
			}, nil
		},
	)

	updateTaskCalled := false
	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetUpdateCallback(
		func(ctx context.Context, input models.TaskExecution) error {
			updateTaskCalled = true
			assert.Equal(t, uint32(1), input.PhaseVersion)
			return nil
		})

	// fill in event for a completed task execution
	taskEventRequest.Event.Phase = core.TaskExecution_RUNNING
	taskEventRequest.Event.PhaseVersion = uint32(1)
	taskEventRequest.Event.OccurredAt = taskEventUpdatedAtProto

	taskExecManager := NewTaskExecutionManager(repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockTaskExecutionRemoteURL, &mockPublisher, &mockPublisher)
	resp, err := taskExecManager.CreateTaskExecutionEvent(context.Background(), taskEventRequest)
	assert.True(t, getTaskCalled)
	assert.True(t, updateTaskCalled)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestGetTaskExecution(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	addGetWorkflowExecutionCallback(repository)
	addGetNodeExecutionCallback(repository)
	addGetTaskCallback(repository)

	expectedLogs := []*core.TaskLog{{Uri: "test-log1.txt"}}
	expectedOutputResult := &admin.TaskExecutionClosure_OutputUri{
		OutputUri: "test-output.pb",
	}
	expectedClosure := &admin.TaskExecutionClosure{
		StartedAt:    sampleTaskEventOccurredAt,
		Phase:        core.TaskExecution_SUCCEEDED,
		Duration:     ptypes.DurationProto(time.Minute),
		OutputResult: expectedOutputResult,
		Logs:         expectedLogs,
	}

	closureBytes, _ := proto.Marshal(expectedClosure)

	getTaskCalled := false
	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.GetTaskExecutionInput) (models.TaskExecution, error) {
			getTaskCalled = true
			assert.Equal(t, sampleTaskID, input.TaskExecutionID.TaskId)
			assert.Equal(t, sampleNodeExecID, input.TaskExecutionID.NodeExecutionId)
			assert.Equal(t, uint32(1), input.TaskExecutionID.RetryAttempt)
			return models.TaskExecution{
				TaskExecutionKey: models.TaskExecutionKey{
					TaskKey: models.TaskKey{
						Project: sampleTaskID.Project,
						Domain:  sampleTaskID.Domain,
						Name:    sampleTaskID.Name,
						Version: sampleTaskID.Version,
					},
					NodeExecutionKey: models.NodeExecutionKey{
						NodeID: sampleNodeExecID.NodeId,
						ExecutionKey: models.ExecutionKey{
							Project: sampleNodeExecID.ExecutionId.Project,
							Domain:  sampleNodeExecID.ExecutionId.Domain,
							Name:    sampleNodeExecID.ExecutionId.Name,
						},
					},
					RetryAttempt: &retryAttemptValue,
				},
				Phase:     core.TaskExecution_SUCCEEDED.String(),
				InputURI:  "input-uri.pb",
				StartedAt: &taskStartedAt,
				Closure:   closureBytes,
			}, nil
		})
	taskExecManager := NewTaskExecutionManager(repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockTaskExecutionRemoteURL, nil, nil)
	taskExecution, err := taskExecManager.GetTaskExecution(context.Background(), admin.TaskExecutionGetRequest{
		Id: &core.TaskExecutionIdentifier{
			TaskId:          sampleTaskID,
			NodeExecutionId: sampleNodeExecID,
			RetryAttempt:    1,
		},
	})
	assert.Nil(t, err)
	assert.True(t, getTaskCalled)
	assert.True(t, proto.Equal(&admin.TaskExecution{
		Id: &core.TaskExecutionIdentifier{
			TaskId:          sampleTaskID,
			NodeExecutionId: sampleNodeExecID,
			RetryAttempt:    uint32(1),
		},
		InputUri: "input-uri.pb",
		Closure:  expectedClosure,
	}, taskExecution))
}

func TestGetTaskExecution_TransformerError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.GetTaskExecutionInput) (models.TaskExecution, error) {
			return models.TaskExecution{
				TaskExecutionKey: models.TaskExecutionKey{
					TaskKey: models.TaskKey{
						Project: sampleTaskID.Project,
						Domain:  sampleTaskID.Domain,
						Name:    sampleTaskID.Name,
						Version: sampleTaskID.Version,
					},
					NodeExecutionKey: models.NodeExecutionKey{
						NodeID: sampleNodeExecID.NodeId,
						ExecutionKey: models.ExecutionKey{
							Project: sampleNodeExecID.ExecutionId.Project,
							Domain:  sampleNodeExecID.ExecutionId.Domain,
							Name:    sampleNodeExecID.ExecutionId.Name,
						},
					},
					RetryAttempt: &retryAttemptValue,
				},
				Phase:     core.TaskExecution_SUCCEEDED.String(),
				InputURI:  "input-uri.pb",
				StartedAt: &taskStartedAt,
				Closure:   []byte("i'm an invalid task closure"),
			}, nil
		})
	taskExecManager := NewTaskExecutionManager(repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockTaskExecutionRemoteURL, nil, nil)
	taskExecution, err := taskExecManager.GetTaskExecution(context.Background(), admin.TaskExecutionGetRequest{
		Id: &core.TaskExecutionIdentifier{
			TaskId:          sampleTaskID,
			NodeExecutionId: sampleNodeExecID,
			RetryAttempt:    1,
		},
	})
	assert.Nil(t, taskExecution)
	assert.Equal(t, err.(flyteAdminErrors.FlyteAdminError).Code(), codes.Internal)
}

func TestListTaskExecutions(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()

	expectedLogs := []*core.TaskLog{{Uri: "test-log1.txt"}}
	extraLongErrMsg := string(make([]byte, 2*100))
	expectedOutputResult := &admin.TaskExecutionClosure_Error{
		Error: &core.ExecutionError{
			Message: extraLongErrMsg,
		},
	}
	expectedClosure := &admin.TaskExecutionClosure{
		StartedAt:    sampleTaskEventOccurredAt,
		Phase:        core.TaskExecution_SUCCEEDED,
		Duration:     ptypes.DurationProto(time.Minute),
		OutputResult: expectedOutputResult,
		Logs:         expectedLogs,
	}

	closureBytes, _ := proto.Marshal(expectedClosure)

	firstRetryAttempt := uint32(1)
	secondRetryAttempt := uint32(2)
	listTaskExecutionsCalled := false
	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetListCallback(
		func(ctx context.Context, input interfaces.ListResourceInput) (interfaces.TaskExecutionCollectionOutput, error) {
			listTaskExecutionsCalled = true
			assert.Equal(t, 99, input.Limit)
			assert.Equal(t, 1, input.Offset)

			assert.Len(t, input.InlineFilters, 4)
			assert.Equal(t, common.Execution, input.InlineFilters[0].GetEntity())
			queryExpr, _ := input.InlineFilters[0].GetGormQueryExpr()
			assert.Equal(t, "exec project b", queryExpr.Args)
			assert.Equal(t, "execution_project = ?", queryExpr.Query)

			assert.Equal(t, common.Execution, input.InlineFilters[1].GetEntity())
			queryExpr, _ = input.InlineFilters[1].GetGormQueryExpr()
			assert.Equal(t, "exec domain b", queryExpr.Args)
			assert.Equal(t, "execution_domain = ?", queryExpr.Query)

			assert.Equal(t, common.Execution, input.InlineFilters[2].GetEntity())
			queryExpr, _ = input.InlineFilters[2].GetGormQueryExpr()
			assert.Equal(t, "exec name b", queryExpr.Args)
			assert.Equal(t, "execution_name = ?", queryExpr.Query)

			assert.Equal(t, common.NodeExecution, input.InlineFilters[3].GetEntity())
			queryExpr, _ = input.InlineFilters[3].GetGormQueryExpr()
			assert.Equal(t, "nodey b", queryExpr.Args)
			assert.Equal(t, "node_id = ?", queryExpr.Query)

			return interfaces.TaskExecutionCollectionOutput{
				TaskExecutions: []models.TaskExecution{
					{
						TaskExecutionKey: models.TaskExecutionKey{
							TaskKey: models.TaskKey{
								Project: "task project a",
								Domain:  "task domain a",
								Name:    "task name a",
								Version: "task version a",
							},
							NodeExecutionKey: models.NodeExecutionKey{
								NodeID: "nodey a",
								ExecutionKey: models.ExecutionKey{
									Project: "exec project a",
									Domain:  "exec domain a",
									Name:    "exec name a",
								},
							},
							RetryAttempt: &firstRetryAttempt,
						},
						Phase:     core.TaskExecution_SUCCEEDED.String(),
						InputURI:  "input-uri.pb",
						StartedAt: &taskStartedAt,
						Closure:   closureBytes,
					},
					{
						TaskExecutionKey: models.TaskExecutionKey{
							TaskKey: models.TaskKey{
								Project: "task project b",
								Domain:  "task domain b",
								Name:    "task name b",
								Version: "task version b",
							},
							NodeExecutionKey: models.NodeExecutionKey{
								NodeID: "nodey b",
								ExecutionKey: models.ExecutionKey{
									Project: "exec project b",
									Domain:  "exec domain b",
									Name:    "exec name b",
								},
							},
							RetryAttempt: &secondRetryAttempt,
						},
						Phase:     core.TaskExecution_SUCCEEDED.String(),
						InputURI:  "input-uri2.pb",
						StartedAt: &taskStartedAt,
						Closure:   closureBytes,
					},
				},
			}, nil
		})
	taskExecManager := NewTaskExecutionManager(repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockTaskExecutionRemoteURL, nil, nil)
	taskExecutions, err := taskExecManager.ListTaskExecutions(context.Background(), admin.TaskExecutionListRequest{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: "nodey b",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "exec project b",
				Domain:  "exec domain b",
				Name:    "exec name b",
			},
		},
		Token: "1",
		Limit: 99,
	})
	assert.Nil(t, err)
	assert.True(t, listTaskExecutionsCalled)

	assert.True(t, proto.Equal(&admin.TaskExecution{
		Id: &core.TaskExecutionIdentifier{
			RetryAttempt: firstRetryAttempt,
			NodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "exec project a",
					Domain:  "exec domain a",
					Name:    "exec name a",
				},
				NodeId: "nodey a",
			},
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "task project a",
				Domain:       "task domain a",
				Name:         "task name a",
				Version:      "task version a",
			},
		},
		InputUri: "input-uri.pb",
		Closure:  expectedClosure,
	}, taskExecutions.TaskExecutions[0]))
	assert.True(t, proto.Equal(&admin.TaskExecution{
		Id: &core.TaskExecutionIdentifier{
			RetryAttempt: secondRetryAttempt,
			NodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "exec project b",
					Domain:  "exec domain b",
					Name:    "exec name b",
				},
				NodeId: "nodey b",
			},
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "task project b",
				Domain:       "task domain b",
				Name:         "task name b",
				Version:      "task version b",
			},
		},
		InputUri: "input-uri2.pb",
		Closure:  expectedClosure,
	}, taskExecutions.TaskExecutions[1]))
}

func TestListTaskExecutions_NoFilters(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()

	listTaskCalled := false
	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetListCallback(
		func(ctx context.Context, input interfaces.ListResourceInput) (interfaces.TaskExecutionCollectionOutput, error) {
			listTaskCalled = true
			return interfaces.TaskExecutionCollectionOutput{}, nil
		})
	taskExecManager := NewTaskExecutionManager(repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockTaskExecutionRemoteURL, nil, nil)
	_, err := taskExecManager.ListTaskExecutions(context.Background(), admin.TaskExecutionListRequest{
		Token: "1",
		Limit: 99,
	})
	assert.NotNil(t, err)
	assert.False(t, listTaskCalled)
}

func TestListTaskExecutions_NoLimit(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()

	getTaskCalled := false
	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetListCallback(
		func(ctx context.Context, input interfaces.ListResourceInput) (interfaces.TaskExecutionCollectionOutput, error) {
			getTaskCalled = true
			return interfaces.TaskExecutionCollectionOutput{}, nil
		})
	taskExecManager := NewTaskExecutionManager(repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockTaskExecutionRemoteURL, nil, nil)
	_, err := taskExecManager.ListTaskExecutions(context.Background(), admin.TaskExecutionListRequest{
		Limit: 0,
	})
	assert.NotNil(t, err)
	assert.False(t, getTaskCalled)
}

func TestListTaskExecutions_NothingToReturn(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()

	var listExecutionsCalled bool
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetListCallback(
		func(ctx context.Context, input interfaces.ListResourceInput) (
			interfaces.ExecutionCollectionOutput, error) {
			listExecutionsCalled = true
			return interfaces.ExecutionCollectionOutput{}, nil
		})
	var listNodeExecutionsCalled bool
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetListCallback(
		func(ctx context.Context, input interfaces.ListResourceInput) (
			interfaces.NodeExecutionCollectionOutput, error) {
			listNodeExecutionsCalled = true
			return interfaces.NodeExecutionCollectionOutput{}, nil
		})
	var listTasksCalled bool
	repository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetListCallback(
		func(input interfaces.ListResourceInput) (interfaces.TaskCollectionOutput, error) {
			listTasksCalled = true
			return interfaces.TaskCollectionOutput{}, nil
		})
	taskExecManager := NewTaskExecutionManager(repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockTaskExecutionRemoteURL, nil, nil)
	_, err := taskExecManager.ListTaskExecutions(context.Background(), admin.TaskExecutionListRequest{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "exec project b",
				Domain:  "exec domain b",
				Name:    "exec name b",
			},
			NodeId: "nodey b",
		},
		Token: "1",
		Limit: 99,
	})
	assert.Nil(t, err)
	assert.False(t, listExecutionsCalled)
	assert.False(t, listNodeExecutionsCalled)
	assert.False(t, listTasksCalled)
}

func TestGetTaskExecutionData(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	addGetWorkflowExecutionCallback(repository)
	addGetNodeExecutionCallback(repository)
	addGetTaskCallback(repository)

	expectedOutputResult := &admin.TaskExecutionClosure_OutputUri{
		OutputUri: "test-output.pb",
	}
	expectedClosure := &admin.TaskExecutionClosure{
		OutputResult: expectedOutputResult,
	}

	closureBytes, _ := proto.Marshal(expectedClosure)

	getTaskCalled := false
	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.GetTaskExecutionInput) (models.TaskExecution, error) {
			getTaskCalled = true
			return models.TaskExecution{
				TaskExecutionKey: models.TaskExecutionKey{
					TaskKey: models.TaskKey{
						Project: sampleTaskID.Project,
						Domain:  sampleTaskID.Domain,
						Name:    sampleTaskID.Name,
						Version: sampleTaskID.Version,
					},
					NodeExecutionKey: models.NodeExecutionKey{
						NodeID: sampleNodeExecID.NodeId,
						ExecutionKey: models.ExecutionKey{
							Project: sampleNodeExecID.ExecutionId.Project,
							Domain:  sampleNodeExecID.ExecutionId.Domain,
							Name:    sampleNodeExecID.ExecutionId.Name,
						},
					},
					RetryAttempt: &retryAttemptValue,
				},
				Phase:     core.TaskExecution_SUCCEEDED.String(),
				InputURI:  "input-uri.pb",
				StartedAt: &taskStartedAt,
				Closure:   closureBytes,
			}, nil
		})
	mockTaskExecutionRemoteURL = dataMocks.NewMockRemoteURL()
	mockTaskExecutionRemoteURL.(*dataMocks.MockRemoteURL).GetCallback = func(ctx context.Context, uri string) (admin.UrlBlob, error) {
		if uri == "input-uri.pb" {
			return admin.UrlBlob{
				Url:   "inputs",
				Bytes: 100,
			}, nil
		} else if uri == "test-output.pb" {
			return admin.UrlBlob{
				Url:   "outputs",
				Bytes: 200,
			}, nil
		}

		return admin.UrlBlob{}, errors.New("unexpected input")
	}
	mockStorage := commonMocks.GetMockStorageClient()
	fullInputs := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo": testutils.MakeStringLiteral("foo-value-1"),
		},
	}
	fullOutputs := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"bar": testutils.MakeStringLiteral("bar-value-1"),
		},
	}
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).ReadProtobufCb = func(
		ctx context.Context, reference storage.DataReference, msg proto.Message) error {
		if reference.String() == "input-uri.pb" {
			marshalled, _ := proto.Marshal(fullInputs)
			_ = proto.Unmarshal(marshalled, msg)
			return nil
		} else if reference.String() == "test-output.pb" {
			marshalled, _ := proto.Marshal(fullOutputs)
			_ = proto.Unmarshal(marshalled, msg)
			return nil
		}
		return fmt.Errorf("unexpected call to find value in storage [%v]", reference.String())
	}
	taskExecManager := NewTaskExecutionManager(repository, getMockExecutionsConfigProvider(), mockStorage, mockScope.NewTestScope(), mockTaskExecutionRemoteURL, nil, nil)
	dataResponse, err := taskExecManager.GetTaskExecutionData(context.Background(), admin.TaskExecutionGetDataRequest{
		Id: &core.TaskExecutionIdentifier{
			TaskId:          sampleTaskID,
			NodeExecutionId: sampleNodeExecID,
			RetryAttempt:    1,
		},
	})
	assert.Nil(t, err)
	assert.True(t, getTaskCalled)
	assert.True(t, proto.Equal(&admin.TaskExecutionGetDataResponse{
		Inputs: &admin.UrlBlob{
			Url:   "inputs",
			Bytes: 100,
		},
		Outputs: &admin.UrlBlob{
			Url:   "outputs",
			Bytes: 200,
		},
		FullInputs:  fullInputs,
		FullOutputs: fullOutputs,
		FlyteUrls: &admin.FlyteURLs{
			Inputs:  "flyte://v1/project/domain/name/node-id/1/i",
			Outputs: "flyte://v1/project/domain/name/node-id/1/o",
			Deck:    "",
		},
	}, dataResponse))
}
