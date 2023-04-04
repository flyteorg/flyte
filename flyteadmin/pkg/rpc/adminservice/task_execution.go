package adminservice

import (
	"context"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (m *AdminService) CreateTaskEvent(
	ctx context.Context, request *admin.TaskExecutionEventRequest) (*admin.TaskExecutionEventResponse, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}

	var response *admin.TaskExecutionEventResponse
	var err error
	m.Metrics.taskExecutionEndpointMetrics.createEvent.Time(func() {
		response, err = m.TaskExecutionManager.CreateTaskExecutionEvent(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.taskExecutionEndpointMetrics.createEvent)
	}
	m.Metrics.taskExecutionEndpointMetrics.createEvent.Success()
	return response, nil
}

func (m *AdminService) GetTaskExecution(
	ctx context.Context, request *admin.TaskExecutionGetRequest) (*admin.TaskExecution, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	// NOTE: When the Get HTTP endpoint is called the resource type is implicit (from the URL) so we must add it
	// to the request.
	if request.Id != nil && request.Id.TaskId != nil && request.Id.TaskId.ResourceType == core.ResourceType_UNSPECIFIED {
		logger.Infof(ctx, "Adding resource type for unspecified value in request: [%+v]", request)
		request.Id.TaskId.ResourceType = core.ResourceType_TASK
	}
	if err := validation.ValidateTaskExecutionIdentifier(request.Id); err != nil {
		return nil, err
	}

	var response *admin.TaskExecution
	var err error
	m.Metrics.taskExecutionEndpointMetrics.get.Time(func() {
		response, err = m.TaskExecutionManager.GetTaskExecution(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.taskExecutionEndpointMetrics.get)
	}
	m.Metrics.taskExecutionEndpointMetrics.get.Success()
	return response, nil
}

func (m *AdminService) ListTaskExecutions(
	ctx context.Context, request *admin.TaskExecutionListRequest) (*admin.TaskExecutionList, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Nil request")
	}
	if err := validation.ValidateTaskExecutionListRequest(*request); err != nil {
		return nil, err
	}

	var response *admin.TaskExecutionList
	var err error
	m.Metrics.taskExecutionEndpointMetrics.list.Time(func() {
		response, err = m.TaskExecutionManager.ListTaskExecutions(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.taskExecutionEndpointMetrics.list)
	}
	m.Metrics.taskExecutionEndpointMetrics.list.Success()
	return response, nil
}

func (m *AdminService) GetTaskExecutionData(
	ctx context.Context, request *admin.TaskExecutionGetDataRequest) (*admin.TaskExecutionGetDataResponse, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	// NOTE: When the Get HTTP endpoint is called the resource type is implicit (from the URL) so we must add it
	// to the request.
	if request.Id != nil && request.Id.TaskId != nil && request.Id.TaskId.ResourceType == core.ResourceType_UNSPECIFIED {
		logger.Infof(ctx, "Adding resource type for unspecified value in request: [%+v]", request)
		request.Id.TaskId.ResourceType = core.ResourceType_TASK
	}
	var response *admin.TaskExecutionGetDataResponse
	var err error
	m.Metrics.taskExecutionEndpointMetrics.getData.Time(func() {
		response, err = m.TaskExecutionManager.GetTaskExecutionData(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.taskExecutionEndpointMetrics.getData)
	}
	m.Metrics.taskExecutionEndpointMetrics.getData.Success()
	return response, nil
}
