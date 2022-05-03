package adminservice

import (
	"context"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (m *AdminService) CreateNodeEvent(
	ctx context.Context, request *admin.NodeExecutionEventRequest) (*admin.NodeExecutionEventResponse, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.NodeExecutionEventResponse
	var err error
	m.Metrics.nodeExecutionEndpointMetrics.createEvent.Time(func() {
		response, err = m.NodeExecutionManager.CreateNodeEvent(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.nodeExecutionEndpointMetrics.createEvent)
	}
	m.Metrics.nodeExecutionEndpointMetrics.createEvent.Success()
	return response, nil
}

func (m *AdminService) GetNodeExecution(
	ctx context.Context, request *admin.NodeExecutionGetRequest) (*admin.NodeExecution, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.NodeExecution
	var err error
	m.Metrics.nodeExecutionEndpointMetrics.get.Time(func() {
		response, err = m.NodeExecutionManager.GetNodeExecution(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.nodeExecutionEndpointMetrics.get)
	}
	m.Metrics.nodeExecutionEndpointMetrics.get.Success()
	return response, nil
}

func (m *AdminService) ListNodeExecutions(
	ctx context.Context, request *admin.NodeExecutionListRequest) (*admin.NodeExecutionList, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.NodeExecutionList
	var err error
	m.Metrics.nodeExecutionEndpointMetrics.list.Time(func() {
		response, err = m.NodeExecutionManager.ListNodeExecutions(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.nodeExecutionEndpointMetrics.list)
	}
	m.Metrics.nodeExecutionEndpointMetrics.list.Success()
	return response, nil
}

func (m *AdminService) ListNodeExecutionsForTask(
	ctx context.Context, request *admin.NodeExecutionForTaskListRequest) (*admin.NodeExecutionList, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	// NOTE: When the Get HTTP endpoint is called the resource type is implicit (from the URL) so we must add it
	// to the request.
	if request.TaskExecutionId != nil && request.TaskExecutionId.TaskId != nil &&
		request.TaskExecutionId.TaskId.ResourceType == core.ResourceType_UNSPECIFIED {
		logger.Infof(ctx, "Adding resource type for unspecified value in request: [%+v]", request)
		request.TaskExecutionId.TaskId.ResourceType = core.ResourceType_TASK
	}
	var response *admin.NodeExecutionList
	var err error
	m.Metrics.nodeExecutionEndpointMetrics.listChildren.Time(func() {
		response, err = m.NodeExecutionManager.ListNodeExecutionsForTask(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.nodeExecutionEndpointMetrics.listChildren)
	}
	m.Metrics.nodeExecutionEndpointMetrics.listChildren.Success()
	return response, nil
}

func (m *AdminService) GetNodeExecutionData(
	ctx context.Context, request *admin.NodeExecutionGetDataRequest) (*admin.NodeExecutionGetDataResponse, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.NodeExecutionGetDataResponse
	var err error
	m.Metrics.nodeExecutionEndpointMetrics.getData.Time(func() {
		response, err = m.NodeExecutionManager.GetNodeExecutionData(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.nodeExecutionEndpointMetrics.getData)
	}
	m.Metrics.nodeExecutionEndpointMetrics.getData.Success()
	return response, nil
}
