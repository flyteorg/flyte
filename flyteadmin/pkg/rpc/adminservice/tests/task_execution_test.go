package tests

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"

	flyteAdminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/golang/protobuf/ptypes"
)

func TestTaskExecution(t *testing.T) {
	executionID := core.WorkflowExecutionIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}
	nodeExecutionID := core.NodeExecutionIdentifier{
		NodeId:      "node id",
		ExecutionId: &executionID,
	}

	taskID := &core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "p",
		Domain:       "d",
		Version:      "v",
		Name:         "n",
	}

	phase := core.TaskExecution_RUNNING
	occurredAt := ptypes.TimestampNow()
	retryAttempt := uint32(1)

	const requestID = "request id"

	t.Run("TestCreateTaskEvent", func(t *testing.T) {
		mockTaskExecutionManager := mocks.MockTaskExecutionManager{}
		mockTaskExecutionManager.SetCreateTaskEventCallback(
			func(ctx context.Context, request admin.TaskExecutionEventRequest) (
				*admin.TaskExecutionEventResponse, error) {
				assert.Equal(t, requestID, request.RequestId)
				assert.NotNil(t, request.Event)
				assert.True(t, proto.Equal(taskID, request.Event.TaskId))
				assert.Equal(t, phase, request.Event.Phase)
				assert.Equal(t, retryAttempt, request.Event.RetryAttempt)
				return &admin.TaskExecutionEventResponse{}, nil
			})
		mockServer := NewMockAdminServer(NewMockAdminServerInput{
			taskExecutionManager: &mockTaskExecutionManager,
		})
		resp, err := mockServer.CreateTaskEvent(context.Background(), &admin.TaskExecutionEventRequest{
			RequestId: requestID,
			Event: &event.TaskExecutionEvent{
				TaskId:                taskID,
				ParentNodeExecutionId: &nodeExecutionID,
				RetryAttempt:          retryAttempt,
				Phase:                 phase,
				OccurredAt:            occurredAt,
			},
		})
		assert.Nil(t, err)
		assert.NotNil(t, resp)
	})

	//	TEMP: uncomment when we turn on task execution events end to end
	// t.Run("TestCreateTaskEventErr", func(t *testing.T) {
	// 	mockTaskExecutionManager := mocks.MockTaskExecutionManager{}
	// 	mockTaskExecutionManager.SetCreateTaskEventCallback(
	// 		func(ctx context.Context, request admin.TaskExecutionEventRequest) (
	// 			*admin.TaskExecutionEventResponse, error) {
	// 			return nil, errors.New("expected error")
	// 		})
	// 	mockServer := NewMockAdminServer(NewMockAdminServerInput{
	// 		taskExecutionManager: &mockTaskExecutionManager,
	// 	})
	// 	resp, err := mockServer.CreateTaskEvent(context.Background(), &admin.TaskExecutionEventRequest{
	// 		RequestId: requestID,
	// 		Event: &event.TaskExecutionEvent{
	// 			TaskId:                taskID,
	// 			Phase:                 phase,
	// 			OccurredAt:            occurredAt,
	// 			RetryAttempt:          retryAttempt,
	// 			ParentNodeExecutionId: &nodeExecutionID,
	// 		},
	// 	})
	// 	assert.EqualError(t, err, "rpc error: code = Internal desc = expected error")
	// 	assert.Nil(t, resp)
	// })
	//
	// t.Run("TestCreateTaskEventMissingTimestamp", func(t *testing.T) {
	// 	mockTaskExecutionManager := mocks.MockTaskExecutionManager{}
	// 	mockTaskExecutionManager.SetCreateTaskEventCallback(
	// 		func(ctx context.Context, request admin.TaskExecutionEventRequest) (
	// 			*admin.TaskExecutionEventResponse, error) {
	// 			t.Fatal("Parameters should be checked before this call")
	// 			return nil, nil
	// 		})
	// 	mockServer := NewMockAdminServer(NewMockAdminServerInput{
	// 		taskExecutionManager: &mockTaskExecutionManager,
	// 	})
	// 	resp, err := mockServer.CreateTaskEvent(context.Background(), &admin.TaskExecutionEventRequest{
	// 		RequestId: requestID,
	// 		Event:     &event.TaskExecutionEvent{},
	// 	})
	// 	assert.EqualError(t, err, "missing occurred_at")
	// 	assert.Nil(t, resp)
	// })
	//
	// t.Run("TestCreateTaskEventMissingNodeExecutionId", func(t *testing.T) {
	// 	mockTaskExecutionManager := mocks.MockTaskExecutionManager{}
	// 	mockTaskExecutionManager.SetCreateTaskEventCallback(
	// 		func(ctx context.Context, request admin.TaskExecutionEventRequest) (
	// 			*admin.TaskExecutionEventResponse, error) {
	// 			t.Fatal("Parameters should be checked before this call")
	// 			return nil, nil
	// 		})
	// 	mockServer := NewMockAdminServer(NewMockAdminServerInput{
	// 		taskExecutionManager: &mockTaskExecutionManager,
	// 	})
	// 	resp, err := mockServer.CreateTaskEvent(context.Background(), &admin.TaskExecutionEventRequest{
	// 		RequestId: requestID,
	// 		Event: &event.TaskExecutionEvent{
	// 			TaskId:     taskID,
	// 			OccurredAt: occurredAt,
	// 			Phase:      phase,
	// 		},
	// 	})
	// 	assert.EqualError(t, err, "missing node_execution_id")
	// 	assert.Nil(t, resp)
	// })

	t.Run("TestGetTaskExecution", func(t *testing.T) {
		mockTaskExecutionManager := mocks.MockTaskExecutionManager{}
		mockTaskExecutionManager.SetGetTaskExecutionCallback(
			func(ctx context.Context, request admin.TaskExecutionGetRequest) (
				*admin.TaskExecution, error) {
				assert.Equal(t, taskID, request.Id.TaskId)
				assert.Equal(t, nodeExecutionID, *request.Id.NodeExecutionId)
				assert.Equal(t, retryAttempt, request.Id.RetryAttempt)
				return &admin.TaskExecution{}, nil
			})
		mockServer := NewMockAdminServer(NewMockAdminServerInput{
			taskExecutionManager: &mockTaskExecutionManager,
		})
		resp, err := mockServer.GetTaskExecution(context.Background(), &admin.TaskExecutionGetRequest{
			Id: &core.TaskExecutionIdentifier{
				TaskId:          taskID,
				NodeExecutionId: &nodeExecutionID,
				RetryAttempt:    retryAttempt,
			},
		})
		assert.Nil(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("TestGetTaskExecutionErr", func(t *testing.T) {
		mockTaskExecutionManager := mocks.MockTaskExecutionManager{}
		mockTaskExecutionManager.SetGetTaskExecutionCallback(
			func(ctx context.Context, request admin.TaskExecutionGetRequest) (
				*admin.TaskExecution, error) {
				return nil, errors.New("expected error")
			})
		mockServer := NewMockAdminServer(NewMockAdminServerInput{
			taskExecutionManager: &mockTaskExecutionManager,
		})
		resp, err := mockServer.GetTaskExecution(context.Background(), &admin.TaskExecutionGetRequest{
			Id: &core.TaskExecutionIdentifier{
				TaskId:          taskID,
				NodeExecutionId: &nodeExecutionID,
				RetryAttempt:    retryAttempt,
			},
		})
		assert.EqualError(t, err, "expected error")
		assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
		assert.Nil(t, resp)
	})

	t.Run("TestGetTaskExecutionMissingId", func(t *testing.T) {
		mockTaskExecutionManager := mocks.MockTaskExecutionManager{}
		mockTaskExecutionManager.SetGetTaskExecutionCallback(
			func(ctx context.Context, request admin.TaskExecutionGetRequest) (
				*admin.TaskExecution, error) {
				t.Fatal("Parameters should be checked before this call")
				return nil, nil
			})
		mockServer := NewMockAdminServer(NewMockAdminServerInput{
			taskExecutionManager: &mockTaskExecutionManager,
		})
		resp, err := mockServer.GetTaskExecution(context.Background(), &admin.TaskExecutionGetRequest{
			Id: &core.TaskExecutionIdentifier{
				NodeExecutionId: &nodeExecutionID,
				RetryAttempt:    1,
			},
		})
		assert.EqualError(t, err, "missing task_id")
		assert.Nil(t, resp)
	})

	t.Run("TestGetTaskExecutionMissingNodeExecutionId", func(t *testing.T) {
		mockTaskExecutionManager := mocks.MockTaskExecutionManager{}
		mockTaskExecutionManager.SetGetTaskExecutionCallback(
			func(ctx context.Context, request admin.TaskExecutionGetRequest) (
				*admin.TaskExecution, error) {
				t.Fatal("Parameters should be checked before this call")
				return nil, nil
			})
		mockServer := NewMockAdminServer(NewMockAdminServerInput{
			taskExecutionManager: &mockTaskExecutionManager,
		})
		resp, err := mockServer.GetTaskExecution(context.Background(), &admin.TaskExecutionGetRequest{
			Id: &core.TaskExecutionIdentifier{
				TaskId:       taskID,
				RetryAttempt: 1,
			},
		})
		assert.EqualError(t, err, "missing node_execution_id")
		assert.Nil(t, resp)
	})

	// List endpoint tests
	t.Run("TestListTaskExecutions", func(t *testing.T) {
		mockTaskExecutionManager := mocks.MockTaskExecutionManager{}
		mockTaskExecutionManager.SetListTaskExecutionsCallback(
			func(ctx context.Context, request admin.TaskExecutionListRequest) (
				*admin.TaskExecutionList, error) {
				assert.Equal(t, "1", request.Token)
				assert.Equal(t, uint32(99), request.Limit)
				assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
					NodeId: "nodey",
					ExecutionId: &core.WorkflowExecutionIdentifier{
						Project: "project",
						Domain:  "domain",
						Name:    "name",
					},
				}, request.NodeExecutionId))
				return &admin.TaskExecutionList{}, nil
			})
		mockServer := NewMockAdminServer(NewMockAdminServerInput{
			taskExecutionManager: &mockTaskExecutionManager,
		})
		resp, err := mockServer.ListTaskExecutions(context.Background(), &admin.TaskExecutionListRequest{
			NodeExecutionId: &core.NodeExecutionIdentifier{
				NodeId: "nodey",
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
			},
			Token: "1",
			Limit: 99,
		})
		assert.Nil(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("TestListTaskExecutions_NoLimit", func(t *testing.T) {
		mockTaskExecutionManager := mocks.MockTaskExecutionManager{}
		mockTaskExecutionManager.SetListTaskExecutionsCallback(
			func(ctx context.Context, request admin.TaskExecutionListRequest) (
				*admin.TaskExecutionList, error) {
				return &admin.TaskExecutionList{}, nil
			})
		mockServer := NewMockAdminServer(NewMockAdminServerInput{
			taskExecutionManager: &mockTaskExecutionManager,
		})
		resp, err := mockServer.ListTaskExecutions(context.Background(), &admin.TaskExecutionListRequest{
			NodeExecutionId: &core.NodeExecutionIdentifier{
				NodeId: "nodey",
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
			},
			Token: "1",
		})
		assert.EqualError(t, err, "invalid value for limit")
		assert.Nil(t, resp)
	})

	t.Run("TestListTaskExecutions_NoFilters", func(t *testing.T) {
		mockTaskExecutionManager := mocks.MockTaskExecutionManager{}
		mockTaskExecutionManager.SetListTaskExecutionsCallback(
			func(ctx context.Context, request admin.TaskExecutionListRequest) (
				*admin.TaskExecutionList, error) {
				return &admin.TaskExecutionList{}, nil
			})
		mockServer := NewMockAdminServer(NewMockAdminServerInput{
			taskExecutionManager: &mockTaskExecutionManager,
		})
		resp, err := mockServer.ListTaskExecutions(context.Background(), &admin.TaskExecutionListRequest{
			Token: "1",
			Limit: 99,
		})
		assert.EqualError(t, err, "missing id")
		assert.Nil(t, resp)
	})
}

func TestGetTaskExecutionData(t *testing.T) {
	mockTaskExecutionManager := mocks.MockTaskExecutionManager{}
	mockTaskExecutionManager.SetGetTaskExecutionDataCallback(
		func(ctx context.Context, request admin.TaskExecutionGetDataRequest) (
			*admin.TaskExecutionGetDataResponse, error) {
			return &admin.TaskExecutionGetDataResponse{
				Inputs: &admin.UrlBlob{
					Url:   "inputs",
					Bytes: 100,
				},
				Outputs: &admin.UrlBlob{
					Url:   "outputs",
					Bytes: 200,
				},
			}, nil
		})
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		taskExecutionManager: &mockTaskExecutionManager,
	})
	taskID := &core.Identifier{
		Project: "p",
		Domain:  "d",
		Version: "v",
		Name:    "n",
	}

	resp, err := mockServer.GetTaskExecutionData(context.Background(), &admin.TaskExecutionGetDataRequest{
		Id: &core.TaskExecutionIdentifier{
			TaskId:          taskID,
			NodeExecutionId: &nodeExecutionID,
			RetryAttempt:    1,
		},
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.UrlBlob{
		Url:   "inputs",
		Bytes: 100,
	}, resp.Inputs))
	assert.True(t, proto.Equal(&admin.UrlBlob{
		Url:   "outputs",
		Bytes: 200,
	}, resp.Outputs))
}
