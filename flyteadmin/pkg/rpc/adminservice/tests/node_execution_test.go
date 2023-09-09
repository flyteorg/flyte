package tests

import (
	"context"
	"errors"
	"testing"

	flyteAdminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"google.golang.org/grpc/codes"

	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteadmin/pkg/manager/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/stretchr/testify/assert"
)

var executionID = core.WorkflowExecutionIdentifier{
	Project: "project",
	Domain:  "domain",
	Name:    "name",
}
var nodeExecutionID = core.NodeExecutionIdentifier{
	NodeId:      "node id",
	ExecutionId: &executionID,
}

func TestCreateNodeEvent(t *testing.T) {
	phase := core.NodeExecution_RUNNING
	mockNodeExecutionManager := mocks.MockNodeExecutionManager{}
	mockNodeExecutionManager.SetCreateNodeEventCallback(
		func(ctx context.Context, request admin.NodeExecutionEventRequest) (
			*admin.NodeExecutionEventResponse, error) {
			assert.Equal(t, requestID, request.RequestId)
			assert.NotNil(t, request.Event)
			assert.True(t, proto.Equal(&nodeExecutionID, request.Event.Id))
			assert.Equal(t, phase, request.Event.Phase)
			return &admin.NodeExecutionEventResponse{}, nil
		})
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		nodeExecutionManager: &mockNodeExecutionManager,
	})
	resp, err := mockServer.CreateNodeEvent(context.Background(), &admin.NodeExecutionEventRequest{
		RequestId: requestID,
		Event: &event.NodeExecutionEvent{
			Id:    &nodeExecutionID,
			Phase: phase,
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestCreateNodeEventErr(t *testing.T) {
	mockNodeExecutionManager := mocks.MockNodeExecutionManager{}
	mockNodeExecutionManager.SetCreateNodeEventCallback(
		func(ctx context.Context, request admin.NodeExecutionEventRequest) (
			*admin.NodeExecutionEventResponse, error) {
			return nil, errors.New("expected error")
		})
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		nodeExecutionManager: &mockNodeExecutionManager,
	})
	resp, err := mockServer.CreateNodeEvent(context.Background(), &admin.NodeExecutionEventRequest{
		RequestId: requestID,
		Event: &event.NodeExecutionEvent{
			Id:    &nodeExecutionID,
			Phase: core.NodeExecution_SKIPPED,
		},
	})
	assert.EqualError(t, err, "expected error")
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
	assert.Nil(t, resp)
}

func TestGetNodeExecution(t *testing.T) {
	response := &admin.NodeExecution{
		Id: &nodeExecutionID,
	}
	mockNodeExecutionManager := mocks.MockNodeExecutionManager{}
	mockNodeExecutionManager.SetGetNodeExecutionFunc(
		func(ctx context.Context,
			request admin.NodeExecutionGetRequest) (*admin.NodeExecution, error) {
			assert.True(t, proto.Equal(&nodeExecutionID, request.Id))
			return response, nil
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		nodeExecutionManager: &mockNodeExecutionManager,
	})

	actualResponse, err := mockServer.GetNodeExecution(context.Background(), &admin.NodeExecutionGetRequest{
		Id: &nodeExecutionID,
	})
	assert.NoError(t, err)
	assert.True(t, proto.Equal(response, actualResponse))
}

func TestGetNodeExecutionError(t *testing.T) {
	mockNodeExecutionManager := mocks.MockNodeExecutionManager{}
	mockNodeExecutionManager.SetGetNodeExecutionFunc(
		func(ctx context.Context,
			request admin.NodeExecutionGetRequest) (*admin.NodeExecution, error) {
			assert.True(t, proto.Equal(&nodeExecutionID, request.Id))
			return nil, errors.New("expected error")
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		nodeExecutionManager: &mockNodeExecutionManager,
	})

	actualResponse, err := mockServer.GetNodeExecution(context.Background(), &admin.NodeExecutionGetRequest{
		Id: &nodeExecutionID,
	})
	assert.EqualError(t, err, "expected error")
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
	assert.Nil(t, actualResponse)
}

func TestListNodeExecutions(t *testing.T) {
	mockNodeExecutionManager := mocks.MockNodeExecutionManager{}
	filters := "encoded filters probably"
	mockNodeExecutionManager.SetListNodeExecutionsFunc(func(ctx context.Context, request admin.NodeExecutionListRequest) (
		*admin.NodeExecutionList, error) {
		assert.Equal(t, filters, request.Filters)
		assert.Equal(t, uint32(1), request.Limit)
		assert.Equal(t, "20", request.Token)
		return &admin.NodeExecutionList{
			NodeExecutions: []*admin.NodeExecution{
				{
					Id: &nodeExecutionID,
				},
			},
		}, nil
	})

	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		nodeExecutionManager: &mockNodeExecutionManager,
	})

	response, err := mockServer.ListNodeExecutions(context.Background(), &admin.NodeExecutionListRequest{
		Filters: filters,
		Limit:   1,
		Token:   "20",
	})
	assert.NoError(t, err)
	assert.Len(t, response.NodeExecutions, 1)
}

func TestListNodeExecutionsError(t *testing.T) {
	mockNodeExecutionManager := mocks.MockNodeExecutionManager{}
	mockNodeExecutionManager.SetListNodeExecutionsFunc(func(ctx context.Context, request admin.NodeExecutionListRequest) (
		*admin.NodeExecutionList, error) {
		return nil, errors.New("expected error")
	})

	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		nodeExecutionManager: &mockNodeExecutionManager,
	})

	response, err := mockServer.ListNodeExecutions(context.Background(), &admin.NodeExecutionListRequest{
		Limit: 1,
		Token: "20",
	})
	assert.EqualError(t, err, "expected error")
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
	assert.Nil(t, response)
}

func TestListNodeExecutionsForTask(t *testing.T) {
	mockNodeExecutionManager := mocks.MockNodeExecutionManager{}
	filters := "encoded filters probably"
	mockNodeExecutionManager.SetListNodeExecutionsForTaskFunc(
		func(ctx context.Context, request admin.NodeExecutionForTaskListRequest) (
			*admin.NodeExecutionList, error) {
			assert.Equal(t, filters, request.Filters)
			assert.Equal(t, uint32(1), request.Limit)
			assert.Equal(t, "20", request.Token)
			return &admin.NodeExecutionList{
				NodeExecutions: []*admin.NodeExecution{
					{
						Id: &nodeExecutionID,
					},
				},
			}, nil
		})

	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		nodeExecutionManager: &mockNodeExecutionManager,
	})

	response, err := mockServer.ListNodeExecutionsForTask(context.Background(), &admin.NodeExecutionForTaskListRequest{
		Filters: filters,
		Limit:   1,
		Token:   "20",
	})
	assert.NoError(t, err)
	assert.Len(t, response.NodeExecutions, 1)
}

func TestListNodeExecutionsForTaskError(t *testing.T) {
	mockNodeExecutionManager := mocks.MockNodeExecutionManager{}
	mockNodeExecutionManager.SetListNodeExecutionsForTaskFunc(
		func(ctx context.Context, request admin.NodeExecutionForTaskListRequest) (
			*admin.NodeExecutionList, error) {
			return nil, errors.New("expected error")
		})

	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		nodeExecutionManager: &mockNodeExecutionManager,
	})

	response, err := mockServer.ListNodeExecutionsForTask(context.Background(), &admin.NodeExecutionForTaskListRequest{
		Limit: 1,
		Token: "20",
	})
	assert.EqualError(t, err, "expected error")
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
	assert.Nil(t, response)
}

func TestGetNodeExecutionData(t *testing.T) {
	mockNodeExecutionManager := mocks.MockNodeExecutionManager{}
	mockNodeExecutionManager.SetGetNodeExecutionDataFunc(
		func(ctx context.Context,
			request admin.NodeExecutionGetDataRequest) (*admin.NodeExecutionGetDataResponse, error) {
			assert.True(t, proto.Equal(&nodeExecutionID, request.Id))
			return &admin.NodeExecutionGetDataResponse{
				Inputs: &admin.UrlBlob{
					Url:   "inputs",
					Bytes: 100,
				},
				Outputs: &admin.UrlBlob{
					Url:   "outputs",
					Bytes: 200,
				},
			}, nil
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		nodeExecutionManager: &mockNodeExecutionManager,
	})

	resp, err := mockServer.GetNodeExecutionData(context.Background(), &admin.NodeExecutionGetDataRequest{
		Id: &nodeExecutionID,
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
