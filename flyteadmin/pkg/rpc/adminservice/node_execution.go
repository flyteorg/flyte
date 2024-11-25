package adminservice

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

func (m *AdminService) CreateNodeEvent(
	ctx context.Context, request *admin.NodeExecutionEventRequest) (*admin.NodeExecutionEventResponse, error) {
	var response *admin.NodeExecutionEventResponse
	var err error
	m.Metrics.nodeExecutionEndpointMetrics.createEvent.Time(func() {
		response, err = m.NodeExecutionManager.CreateNodeEvent(ctx, request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.nodeExecutionEndpointMetrics.createEvent)
	}
	m.Metrics.nodeExecutionEndpointMetrics.createEvent.Success()
	return response, nil
}

func (m *AdminService) GetNodeExecution(
	ctx context.Context, request *admin.NodeExecutionGetRequest) (*admin.NodeExecution, error) {
	var response *admin.NodeExecution
	var err error
	m.Metrics.nodeExecutionEndpointMetrics.get.Time(func() {
		response, err = m.NodeExecutionManager.GetNodeExecution(ctx, request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.nodeExecutionEndpointMetrics.get)
	}
	m.Metrics.nodeExecutionEndpointMetrics.get.Success()
	return response, nil
}

func (m *AdminService) GetDynamicNodeWorkflow(ctx context.Context, request *admin.GetDynamicNodeWorkflowRequest) (*admin.DynamicNodeWorkflowResponse, error) {
	var response *admin.DynamicNodeWorkflowResponse
	var err error
	m.Metrics.nodeExecutionEndpointMetrics.getDynamicNodeWorkflow.Time(func() {
		response, err = m.NodeExecutionManager.GetDynamicNodeWorkflow(ctx, request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowEndpointMetrics.get)
	}
	m.Metrics.nodeExecutionEndpointMetrics.getDynamicNodeWorkflow.Success()
	return response, nil
}

func (m *AdminService) ListNodeExecutions(
	ctx context.Context, request *admin.NodeExecutionListRequest) (*admin.NodeExecutionList, error) {
	var response *admin.NodeExecutionList
	var err error
	m.Metrics.nodeExecutionEndpointMetrics.list.Time(func() {
		response, err = m.NodeExecutionManager.ListNodeExecutions(ctx, request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.nodeExecutionEndpointMetrics.list)
	}
	m.Metrics.nodeExecutionEndpointMetrics.list.Success()
	return response, nil
}

func (m *AdminService) ListNodeExecutionsForTask(
	ctx context.Context, request *admin.NodeExecutionForTaskListRequest) (*admin.NodeExecutionList, error) {
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
		response, err = m.NodeExecutionManager.ListNodeExecutionsForTask(ctx, request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.nodeExecutionEndpointMetrics.listChildren)
	}
	m.Metrics.nodeExecutionEndpointMetrics.listChildren.Success()
	return response, nil
}

func (m *AdminService) GetNodeExecutionData(
	ctx context.Context, request *admin.NodeExecutionGetDataRequest) (*admin.NodeExecutionGetDataResponse, error) {
	var response *admin.NodeExecutionGetDataResponse
	var err error
	m.Metrics.nodeExecutionEndpointMetrics.getData.Time(func() {
		response, err = m.NodeExecutionManager.GetNodeExecutionData(ctx, request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.nodeExecutionEndpointMetrics.getData)
	}
	m.Metrics.nodeExecutionEndpointMetrics.getData.Success()
	return response, nil
}
