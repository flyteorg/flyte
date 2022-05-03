package adminservice

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (m *AdminService) CreateWorkflow(
	ctx context.Context,
	request *admin.WorkflowCreateRequest) (*admin.WorkflowCreateResponse, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.WorkflowCreateResponse
	var err error
	m.Metrics.workflowEndpointMetrics.create.Time(func() {
		response, err = m.WorkflowManager.CreateWorkflow(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowEndpointMetrics.create)
	}
	m.Metrics.workflowEndpointMetrics.create.Success()
	return response, nil
}

func (m *AdminService) GetWorkflow(ctx context.Context, request *admin.ObjectGetRequest) (*admin.Workflow, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	// NOTE: When the Get HTTP endpoint is called the resource type is implicit (from the URL) so we must add it
	// to the request.
	if request.Id != nil && request.Id.ResourceType == core.ResourceType_UNSPECIFIED {
		logger.Infof(ctx, "Adding resource type for unspecified value in request: [%+v]", request)
		request.Id.ResourceType = core.ResourceType_WORKFLOW
	}
	var response *admin.Workflow
	var err error
	m.Metrics.workflowEndpointMetrics.get.Time(func() {
		response, err = m.WorkflowManager.GetWorkflow(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowEndpointMetrics.get)
	}
	m.Metrics.workflowEndpointMetrics.get.Success()
	return response, nil
}

func (m *AdminService) ListWorkflowIds(ctx context.Context, request *admin.NamedEntityIdentifierListRequest) (
	*admin.NamedEntityIdentifierList, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}

	var response *admin.NamedEntityIdentifierList
	var err error
	m.Metrics.workflowEndpointMetrics.listIds.Time(func() {
		response, err = m.WorkflowManager.ListWorkflowIdentifiers(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowEndpointMetrics.listIds)
	}

	m.Metrics.workflowEndpointMetrics.listIds.Success()
	return response, nil
}

func (m *AdminService) ListWorkflows(ctx context.Context, request *admin.ResourceListRequest) (*admin.WorkflowList, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.WorkflowList
	var err error
	m.Metrics.workflowEndpointMetrics.list.Time(func() {
		response, err = m.WorkflowManager.ListWorkflows(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowEndpointMetrics.list)
	}
	m.Metrics.workflowEndpointMetrics.list.Success()
	return response, nil
}
