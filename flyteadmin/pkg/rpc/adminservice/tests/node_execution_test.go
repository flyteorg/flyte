package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"

	flyteAdminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
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
	mockNodeExecutionManager := mocks.NodeExecutionInterface{}
	mockNodeExecutionManager.EXPECT().CreateNodeEvent(mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context, request *admin.NodeExecutionEventRequest) (
			*admin.NodeExecutionEventResponse, error) {
			assert.Equal(t, requestID, request.GetRequestId())
			assert.NotNil(t, request.GetEvent())
			assert.True(t, proto.Equal(&nodeExecutionID, request.GetEvent().GetId()))
			assert.Equal(t, phase, request.GetEvent().GetPhase())
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
	mockNodeExecutionManager := mocks.NodeExecutionInterface{}
	mockNodeExecutionManager.EXPECT().CreateNodeEvent(mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context, request *admin.NodeExecutionEventRequest) (
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
	mockNodeExecutionManager := mocks.NodeExecutionInterface{}
	mockNodeExecutionManager.EXPECT().GetNodeExecution(mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context,
			request *admin.NodeExecutionGetRequest) (*admin.NodeExecution, error) {
			assert.True(t, proto.Equal(&nodeExecutionID, request.GetId()))
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
	mockNodeExecutionManager := mocks.NodeExecutionInterface{}
	mockNodeExecutionManager.EXPECT().GetNodeExecution(mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context,
			request *admin.NodeExecutionGetRequest) (*admin.NodeExecution, error) {
			assert.True(t, proto.Equal(&nodeExecutionID, request.GetId()))
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
	mockNodeExecutionManager := mocks.NodeExecutionInterface{}
	filters := "encoded filters probably"
	mockNodeExecutionManager.EXPECT().ListNodeExecutions(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, request *admin.NodeExecutionListRequest) (
		*admin.NodeExecutionList, error) {
		assert.Equal(t, filters, request.GetFilters())
		assert.Equal(t, uint32(1), request.GetLimit())
		assert.Equal(t, "20", request.GetToken())
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
	assert.Len(t, response.GetNodeExecutions(), 1)
}

func TestListNodeExecutionsError(t *testing.T) {
	mockNodeExecutionManager := mocks.NodeExecutionInterface{}
	mockNodeExecutionManager.EXPECT().ListNodeExecutions(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, request *admin.NodeExecutionListRequest) (
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
	mockNodeExecutionManager := mocks.NodeExecutionInterface{}
	filters := "encoded filters probably"
	mockNodeExecutionManager.EXPECT().ListNodeExecutionsForTask(mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context, request *admin.NodeExecutionForTaskListRequest) (
			*admin.NodeExecutionList, error) {
			assert.Equal(t, filters, request.GetFilters())
			assert.Equal(t, uint32(1), request.GetLimit())
			assert.Equal(t, "20", request.GetToken())
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
	assert.Len(t, response.GetNodeExecutions(), 1)
}

func TestListNodeExecutionsForTaskError(t *testing.T) {
	mockNodeExecutionManager := mocks.NodeExecutionInterface{}
	mockNodeExecutionManager.EXPECT().ListNodeExecutionsForTask(mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context, request *admin.NodeExecutionForTaskListRequest) (
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
	mockNodeExecutionManager := mocks.NodeExecutionInterface{}
	mockNodeExecutionManager.EXPECT().GetNodeExecutionData(mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context,
			request *admin.NodeExecutionGetDataRequest) (*admin.NodeExecutionGetDataResponse, error) {
			assert.True(t, proto.Equal(&nodeExecutionID, request.GetId()))
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
	}, resp.GetInputs()))
	assert.True(t, proto.Equal(&admin.UrlBlob{
		Url:   "outputs",
		Bytes: 200,
	}, resp.GetOutputs()))
}
