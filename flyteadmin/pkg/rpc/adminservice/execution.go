package adminservice

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

func (m *AdminService) CreateExecution(
	ctx context.Context, request *admin.ExecutionCreateRequest) (*admin.ExecutionCreateResponse, error) {
	requestedAt := time.Now()
	var response *admin.ExecutionCreateResponse
	var err error
	m.Metrics.executionEndpointMetrics.create.Time(func() {
		response, err = m.ExecutionManager.CreateExecution(ctx, request, requestedAt)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.executionEndpointMetrics.create)
	}
	m.Metrics.executionEndpointMetrics.create.Success()
	return response, nil
}

func (m *AdminService) RelaunchExecution(
	ctx context.Context, request *admin.ExecutionRelaunchRequest) (*admin.ExecutionCreateResponse, error) {
	requestedAt := time.Now()
	var response *admin.ExecutionCreateResponse
	var err error
	m.Metrics.executionEndpointMetrics.relaunch.Time(func() {
		response, err = m.ExecutionManager.RelaunchExecution(ctx, request, requestedAt)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.executionEndpointMetrics.relaunch)
	}
	m.Metrics.executionEndpointMetrics.relaunch.Success()
	return response, nil
}

func (m *AdminService) RecoverExecution(
	ctx context.Context, request *admin.ExecutionRecoverRequest) (*admin.ExecutionCreateResponse, error) {
	requestedAt := time.Now()
	var response *admin.ExecutionCreateResponse
	var err error
	m.Metrics.executionEndpointMetrics.recover.Time(func() {
		response, err = m.ExecutionManager.RecoverExecution(ctx, request, requestedAt)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.executionEndpointMetrics.relaunch)
	}
	m.Metrics.executionEndpointMetrics.relaunch.Success()
	return response, nil
}

func (m *AdminService) CreateWorkflowEvent(
	ctx context.Context, request *admin.WorkflowExecutionEventRequest) (*admin.WorkflowExecutionEventResponse, error) {
	var response *admin.WorkflowExecutionEventResponse
	var err error
	m.Metrics.executionEndpointMetrics.createEvent.Time(func() {
		response, err = m.ExecutionManager.CreateWorkflowEvent(ctx, request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.executionEndpointMetrics.createEvent)
	}
	m.Metrics.executionEndpointMetrics.createEvent.Success()

	return response, nil
}

func (m *AdminService) GetExecution(
	ctx context.Context, request *admin.WorkflowExecutionGetRequest) (*admin.Execution, error) {
	var response *admin.Execution
	var err error
	m.Metrics.executionEndpointMetrics.get.Time(func() {
		response, err = m.ExecutionManager.GetExecution(ctx, request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.executionEndpointMetrics.get)
	}
	m.Metrics.executionEndpointMetrics.get.Success()
	return response, nil
}

func (m *AdminService) UpdateExecution(
	ctx context.Context, request *admin.ExecutionUpdateRequest) (*admin.ExecutionUpdateResponse, error) {
	requestedAt := time.Now()
	var response *admin.ExecutionUpdateResponse
	var err error
	m.Metrics.executionEndpointMetrics.update.Time(func() {
		response, err = m.ExecutionManager.UpdateExecution(ctx, request, requestedAt)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.executionEndpointMetrics.update)
	}
	m.Metrics.executionEndpointMetrics.update.Success()
	return response, nil
}

func (m *AdminService) GetExecutionData(
	ctx context.Context, request *admin.WorkflowExecutionGetDataRequest) (*admin.WorkflowExecutionGetDataResponse, error) {
	var response *admin.WorkflowExecutionGetDataResponse
	var err error
	m.Metrics.executionEndpointMetrics.getData.Time(func() {
		response, err = m.ExecutionManager.GetExecutionData(ctx, request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.executionEndpointMetrics.getData)
	}
	m.Metrics.executionEndpointMetrics.getData.Success()
	// DEBUG: CASE-643
	if response.GetFullOutputs().GetLiterals() == nil || len(response.GetFullOutputs().GetLiterals()) == 0 {
		logger.Debugf(ctx, "GetExecutionData().FullOutputs empty for execution %v. FullOutputs: %v. Outputs: %v", request.GetId(), response.GetFullOutputs(), response.GetOutputs())
	}
	return response, nil
}

func (m *AdminService) GetExecutionMetrics(
	ctx context.Context, request *admin.WorkflowExecutionGetMetricsRequest) (*admin.WorkflowExecutionGetMetricsResponse, error) {
	var response *admin.WorkflowExecutionGetMetricsResponse
	var err error
	m.Metrics.executionEndpointMetrics.getMetrics.Time(func() {
		response, err = m.MetricsManager.GetExecutionMetrics(ctx, request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.executionEndpointMetrics.getMetrics)
	}
	m.Metrics.executionEndpointMetrics.getMetrics.Success()
	return response, nil
}

func (m *AdminService) ListExecutions(
	ctx context.Context, request *admin.ResourceListRequest) (*admin.ExecutionList, error) {
	var response *admin.ExecutionList
	var err error
	m.Metrics.executionEndpointMetrics.list.Time(func() {
		response, err = m.ExecutionManager.ListExecutions(ctx, request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.executionEndpointMetrics.list)
	}
	m.Metrics.executionEndpointMetrics.list.Success()
	return response, nil
}

func (m *AdminService) TerminateExecution(
	ctx context.Context, request *admin.ExecutionTerminateRequest) (*admin.ExecutionTerminateResponse, error) {
	var response *admin.ExecutionTerminateResponse
	var err error
	m.Metrics.executionEndpointMetrics.terminate.Time(func() {
		response, err = m.ExecutionManager.TerminateExecution(ctx, request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.executionEndpointMetrics.terminate)
	}
	m.Metrics.executionEndpointMetrics.terminate.Success()
	return response, nil
}

func (m *AdminService) GetExecutionCounts(
	ctx context.Context, request *admin.ExecutionCountsGetRequest) (*admin.ExecutionCountsGetResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ExecutionCountsGetResponse
	var err error
	m.Metrics.executionEndpointMetrics.count.Time(func() {
		response, err = m.ExecutionManager.GetExecutionCounts(ctx, request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.executionEndpointMetrics.count)
	}
	m.Metrics.executionEndpointMetrics.count.Success()
	return response, nil
}

func (m *AdminService) GetRunningExecutionsCount(
	ctx context.Context, request *admin.RunningExecutionsCountGetRequest) (*admin.RunningExecutionsCountGetResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.RunningExecutionsCountGetResponse
	var err error
	m.Metrics.executionEndpointMetrics.runningCount.Time(func() {
		response, err = m.ExecutionManager.GetRunningExecutionsCount(ctx, request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.executionEndpointMetrics.runningCount)
	}
	m.Metrics.executionEndpointMetrics.runningCount.Success()
	return response, nil
}
