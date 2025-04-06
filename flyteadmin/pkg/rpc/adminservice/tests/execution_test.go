package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"

	flyteAdminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/mocks"
	repoErrors "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

const requestID = "request id"

var workflowExecutionIdentifier = core.WorkflowExecutionIdentifier{
	Name:    "Name",
	Domain:  "Domain",
	Project: "Project",
}

func TestCreateExecutionHappyCase(t *testing.T) {
	ctx := context.Background()

	mockExecutionManager := mocks.ExecutionInterface{}
	mockExecutionManager.EXPECT().CreateExecution(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context,
			request *admin.ExecutionCreateRequest, requestedAt time.Time) (*admin.ExecutionCreateResponse, error) {
			return &admin.ExecutionCreateResponse{
				Id: &core.WorkflowExecutionIdentifier{
					Project: request.GetProject(),
					Domain:  request.GetDomain(),
					Name:    request.GetName(),
				},
			}, nil
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		executionManager: &mockExecutionManager,
	})

	resp, err := mockServer.CreateExecution(ctx, &admin.ExecutionCreateRequest{
		Name:    "Name",
		Domain:  "Domain",
		Project: "Project",
	})
	assert.True(t, proto.Equal(&workflowExecutionIdentifier, resp.GetId()))
	assert.NoError(t, err)
}

func TestCreateExecutionError(t *testing.T) {
	ctx := context.Background()

	mockExecutionManager := mocks.ExecutionInterface{}
	mockExecutionManager.EXPECT().CreateExecution(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context,
			request *admin.ExecutionCreateRequest, requestedAt time.Time) (*admin.ExecutionCreateResponse, error) {
			return nil, repoErrors.GetMissingEntityError("execution", &core.Identifier{
				Project: request.GetProject(),
				Domain:  request.GetDomain(),
				Name:    request.GetName(),
			})
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		executionManager: &mockExecutionManager,
	})

	resp, err := mockServer.CreateExecution(ctx, &admin.ExecutionCreateRequest{
		Project: "Project",
		Domain:  "Domain",
		Name:    "Name",
	})
	assert.Nil(t, resp)
	utils.AssertEqualWithSanitizedRegex(t,
		`missing entity of type execution with identifier project:"Project" domain:"Domain" name:"Name"`, err.Error())
}

func TestRelaunchExecutionHappyCase(t *testing.T) {
	ctx := context.Background()

	mockExecutionManager := mocks.ExecutionInterface{}
	mockExecutionManager.EXPECT().RelaunchExecution(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context,
			request *admin.ExecutionRelaunchRequest, requestedAt time.Time) (*admin.ExecutionCreateResponse, error) {
			return &admin.ExecutionCreateResponse{
				Id: &core.WorkflowExecutionIdentifier{
					Project: request.GetId().GetProject(),
					Domain:  request.GetId().GetDomain(),
					Name:    request.GetName(),
				},
			}, nil
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		executionManager: &mockExecutionManager,
	})

	resp, err := mockServer.RelaunchExecution(ctx, &admin.ExecutionRelaunchRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
		},
		Name: "name",
	})
	assert.Equal(t, "project", resp.GetId().GetProject())
	assert.Equal(t, "domain", resp.GetId().GetDomain())
	assert.Equal(t, "name", resp.GetId().GetName())
	assert.NoError(t, err)
}

func TestRelaunchExecutionError(t *testing.T) {
	ctx := context.Background()

	mockExecutionManager := mocks.ExecutionInterface{}
	mockExecutionManager.EXPECT().RelaunchExecution(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context,
			request *admin.ExecutionRelaunchRequest, requestedAt time.Time) (*admin.ExecutionCreateResponse, error) {
			return nil, repoErrors.GetMissingEntityError("execution", request.GetId())
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		executionManager: &mockExecutionManager,
	})

	resp, err := mockServer.RelaunchExecution(ctx, &admin.ExecutionRelaunchRequest{
		Name: "Name",
	})
	assert.Nil(t, resp)
	assert.EqualError(t, err,
		"missing entity of type execution with identifier <nil>")
}

func TestRecoverExecutionHappyCase(t *testing.T) {
	ctx := context.Background()

	mockExecutionManager := mocks.ExecutionInterface{}
	mockExecutionManager.EXPECT().RecoverExecution(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context,
			request *admin.ExecutionRecoverRequest, requestedAt time.Time) (*admin.ExecutionCreateResponse, error) {
			return &admin.ExecutionCreateResponse{
				Id: &core.WorkflowExecutionIdentifier{
					Project: request.GetId().GetProject(),
					Domain:  request.GetId().GetDomain(),
					Name:    request.GetName(),
				},
			}, nil
		})

	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		executionManager: &mockExecutionManager,
	})

	resp, err := mockServer.RecoverExecution(ctx, &admin.ExecutionRecoverRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
		},
		Name: "name",
	})
	assert.Equal(t, "project", resp.GetId().GetProject())
	assert.Equal(t, "domain", resp.GetId().GetDomain())
	assert.Equal(t, "name", resp.GetId().GetName())
	assert.NoError(t, err)
}

func TestRecoverExecutionError(t *testing.T) {
	ctx := context.Background()

	mockExecutionManager := mocks.ExecutionInterface{}
	mockExecutionManager.EXPECT().RecoverExecution(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context,
			request *admin.ExecutionRecoverRequest, requestedAt time.Time) (*admin.ExecutionCreateResponse, error) {
			return nil, repoErrors.GetMissingEntityError("execution", request.GetId())
		})
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		executionManager: &mockExecutionManager,
	})

	resp, err := mockServer.RecoverExecution(ctx, &admin.ExecutionRecoverRequest{
		Name: "Name",
	})
	assert.Nil(t, resp)
	assert.EqualError(t, err,
		"missing entity of type execution with identifier <nil>")
}

func TestCreateWorkflowEvent(t *testing.T) {
	phase := core.WorkflowExecution_RUNNING
	mockExecutionManager := mocks.ExecutionInterface{}
	mockExecutionManager.EXPECT().CreateWorkflowEvent(mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context, request *admin.WorkflowExecutionEventRequest) (
			*admin.WorkflowExecutionEventResponse, error) {
			assert.Equal(t, requestID, request.GetRequestId())
			assert.NotNil(t, request.GetEvent())
			assert.True(t, proto.Equal(&workflowExecutionIdentifier, request.GetEvent().GetExecutionId()))
			assert.Equal(t, phase, request.GetEvent().GetPhase())
			return &admin.WorkflowExecutionEventResponse{}, nil
		})
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		executionManager: &mockExecutionManager,
	})
	resp, err := mockServer.CreateWorkflowEvent(context.Background(), &admin.WorkflowExecutionEventRequest{
		RequestId: requestID,
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &workflowExecutionIdentifier,
			Phase:       phase,
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}
func TestCreateWorkflowEventErr(t *testing.T) {
	mockExecutionManager := mocks.ExecutionInterface{}
	mockExecutionManager.EXPECT().CreateWorkflowEvent(mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context, request *admin.WorkflowExecutionEventRequest) (
			*admin.WorkflowExecutionEventResponse, error) {
			return nil, errors.New("expected error")
		})
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		executionManager: &mockExecutionManager,
	})
	resp, err := mockServer.CreateWorkflowEvent(context.Background(), &admin.WorkflowExecutionEventRequest{
		RequestId: requestID,
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &workflowExecutionIdentifier,
			Phase:       core.WorkflowExecution_RUNNING,
		},
	})
	assert.EqualError(t, err, "expected error")
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
	assert.Nil(t, resp)
}

func TestGetExecution(t *testing.T) {
	response := &admin.Execution{
		Id: &workflowExecutionIdentifier,
	}
	mockExecutionManager := mocks.ExecutionInterface{}
	mockExecutionManager.EXPECT().GetExecution(mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context,
			request *admin.WorkflowExecutionGetRequest) (*admin.Execution, error) {
			assert.True(t, proto.Equal(&workflowExecutionIdentifier, request.GetId()))
			return response, nil
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		executionManager: &mockExecutionManager,
	})

	actualResponse, err := mockServer.GetExecution(context.Background(), &admin.WorkflowExecutionGetRequest{
		Id: &workflowExecutionIdentifier,
	})
	assert.NoError(t, err)
	assert.True(t, proto.Equal(response, actualResponse))
}

func TestGetExecutionError(t *testing.T) {
	mockExecutionManager := mocks.ExecutionInterface{}
	mockExecutionManager.EXPECT().GetExecution(mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context,
			request *admin.WorkflowExecutionGetRequest) (*admin.Execution, error) {
			return nil, errors.New("expected error")
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		executionManager: &mockExecutionManager,
	})

	actualResponse, err := mockServer.GetExecution(context.Background(), &admin.WorkflowExecutionGetRequest{
		Id: &workflowExecutionIdentifier,
	})
	assert.EqualError(t, err, "expected error")
	assert.Nil(t, actualResponse)
}

func TestUpdateExecution(t *testing.T) {
	response := &admin.ExecutionUpdateResponse{}
	mockExecutionManager := mocks.ExecutionInterface{}
	mockExecutionManager.EXPECT().UpdateExecution(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context,
			request *admin.ExecutionUpdateRequest, requestedAt time.Time) (*admin.ExecutionUpdateResponse, error) {
			assert.True(t, proto.Equal(&workflowExecutionIdentifier, request.GetId()))
			return response, nil
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		executionManager: &mockExecutionManager,
	})

	actualResponse, err := mockServer.UpdateExecution(context.Background(), &admin.ExecutionUpdateRequest{
		Id: &workflowExecutionIdentifier,
	})
	assert.NoError(t, err)
	assert.True(t, proto.Equal(response, actualResponse))
}

func TestUpdateExecutionError(t *testing.T) {
	mockExecutionManager := mocks.ExecutionInterface{}
	mockExecutionManager.EXPECT().UpdateExecution(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context,
			request *admin.ExecutionUpdateRequest, requestedAt time.Time) (*admin.ExecutionUpdateResponse, error) {
			return nil, errors.New("expected error")
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		executionManager: &mockExecutionManager,
	})

	actualResponse, err := mockServer.UpdateExecution(context.Background(), &admin.ExecutionUpdateRequest{
		Id: &workflowExecutionIdentifier,
	})
	assert.EqualError(t, err, "expected error")
	assert.Nil(t, actualResponse)
}

func TestListExecutions(t *testing.T) {
	mockExecutionManager := mocks.ExecutionInterface{}
	mockExecutionManager.EXPECT().ListExecutions(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, request *admin.ResourceListRequest) (
		*admin.ExecutionList, error) {
		assert.Equal(t, "project", request.GetId().GetProject())
		assert.Equal(t, "domain", request.GetId().GetDomain())
		assert.Equal(t, uint32(1), request.GetLimit())
		return &admin.ExecutionList{
			Executions: []*admin.Execution{
				{
					Id: &workflowExecutionIdentifier,
				},
			},
		}, nil
	})

	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		executionManager: &mockExecutionManager,
	})

	response, err := mockServer.ListExecutions(context.Background(), &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
		},
		Limit: 1,
	})
	assert.NoError(t, err)
	assert.Len(t, response.GetExecutions(), 1)
}

func TestListExecutionsError(t *testing.T) {
	mockExecutionManager := mocks.ExecutionInterface{}
	mockExecutionManager.EXPECT().ListExecutions(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, request *admin.ResourceListRequest) (
		*admin.ExecutionList, error) {
		return nil, errors.New("expected error")
	})

	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		executionManager: &mockExecutionManager,
	})

	response, err := mockServer.ListExecutions(context.Background(), &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
		},
		Limit: 1,
	})
	assert.EqualError(t, err, "expected error")
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
	assert.Nil(t, response)
}

func TestTerminateExecution(t *testing.T) {
	mockExecutionManager := mocks.ExecutionInterface{}
	identifier := core.WorkflowExecutionIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}
	abortCause := "abort cause"
	mockExecutionManager.EXPECT().TerminateExecution(mock.Anything, mock.Anything).RunAndReturn(func(
		ctx context.Context, request *admin.ExecutionTerminateRequest) (*admin.ExecutionTerminateResponse, error) {
		assert.True(t, proto.Equal(&identifier, request.GetId()))
		assert.Equal(t, abortCause, request.GetCause())
		return &admin.ExecutionTerminateResponse{}, nil
	})
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		executionManager: &mockExecutionManager,
	})
	_, err := mockServer.TerminateExecution(context.Background(), &admin.ExecutionTerminateRequest{
		Id:    &identifier,
		Cause: abortCause,
	})
	assert.Nil(t, err)
}

func TestTerminateExecution_Error(t *testing.T) {
	mockExecutionManager := mocks.ExecutionInterface{}
	identifier := core.WorkflowExecutionIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}
	abortCause := "abort cause"
	mockExecutionManager.EXPECT().TerminateExecution(mock.Anything, mock.Anything).RunAndReturn(func(
		ctx context.Context, request *admin.ExecutionTerminateRequest) (*admin.ExecutionTerminateResponse, error) {
		return nil, errors.New("expected error")
	})
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		executionManager: &mockExecutionManager,
	})
	response, err := mockServer.TerminateExecution(context.Background(), &admin.ExecutionTerminateRequest{
		Id:    &identifier,
		Cause: abortCause,
	})
	assert.EqualError(t, err, "expected error")
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
	assert.Nil(t, response)
}
