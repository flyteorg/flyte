package adminservice

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (m *AdminService) UpdateWorkflowAttributes(ctx context.Context, request *admin.WorkflowAttributesUpdateRequest) (
	*admin.WorkflowAttributesUpdateResponse, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.WorkflowAttributesUpdateResponse
	var err error
	m.Metrics.workflowAttributesEndpointMetrics.update.Time(func() {
		response, err = m.WorkflowAttributesManager.UpdateWorkflowAttributes(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowAttributesEndpointMetrics.update)
	}

	return response, nil
}

func (m *AdminService) UpdateProjectDomainAttributes(ctx context.Context, request *admin.ProjectDomainAttributesUpdateRequest) (
	*admin.ProjectDomainAttributesUpdateResponse, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ProjectDomainAttributesUpdateResponse
	var err error
	m.Metrics.projectDomainAttributesEndpointMetrics.update.Time(func() {
		response, err = m.ProjectDomainAttributesManager.UpdateProjectDomainAttributes(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.projectDomainAttributesEndpointMetrics.update)
	}

	return response, nil
}

func (m *AdminService) UpdateProjectAttributes(ctx context.Context, request *admin.ProjectAttributesUpdateRequest) (
	*admin.ProjectAttributesUpdateResponse, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ProjectAttributesUpdateResponse
	var err error
	m.Metrics.projectAttributesEndpointMetrics.update.Time(func() {
		response, err = m.ProjectAttributesManager.UpdateProjectAttributes(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.projectAttributesEndpointMetrics.update)
	}

	return response, nil
}
