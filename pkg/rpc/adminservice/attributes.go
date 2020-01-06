package adminservice

import (
	"context"
	"time"

	"github.com/lyft/flyteadmin/pkg/audit"

	"github.com/lyft/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (m *AdminService) UpdateWorkflowAttributes(ctx context.Context, request *admin.WorkflowAttributesUpdateRequest) (
	*admin.WorkflowAttributesUpdateResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.WorkflowAttributesUpdateResponse
	var err error
	m.Metrics.workflowAttributesEndpointMetrics.update.Time(func() {
		response, err = m.WorkflowAttributesManager.UpdateWorkflowAttributes(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"UpdateWorkflowAttributes",
		map[string]string{
			audit.Project: request.Attributes.Project,
			audit.Domain:  request.Attributes.Domain,
			audit.Name:    request.Attributes.Workflow,
		},
		audit.ReadWrite,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowAttributesEndpointMetrics.update)
	}

	return response, nil
}

func (m *AdminService) GetWorkflowAttributes(ctx context.Context, request *admin.WorkflowAttributesGetRequest) (
	*admin.WorkflowAttributesGetResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.WorkflowAttributesGetResponse
	var err error
	m.Metrics.workflowAttributesEndpointMetrics.get.Time(func() {
		response, err = m.WorkflowAttributesManager.GetWorkflowAttributes(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"GetWorkflowAttributes",
		map[string]string{
			audit.Project: request.Project,
			audit.Domain:  request.Domain,
			audit.Name:    request.Workflow,
		},
		audit.ReadOnly,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowAttributesEndpointMetrics.get)
	}

	return response, nil
}

func (m *AdminService) DeleteWorkflowAttributes(ctx context.Context, request *admin.WorkflowAttributesDeleteRequest) (
	*admin.WorkflowAttributesDeleteResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.WorkflowAttributesDeleteResponse
	var err error
	m.Metrics.workflowAttributesEndpointMetrics.delete.Time(func() {
		response, err = m.WorkflowAttributesManager.DeleteWorkflowAttributes(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"DeleteWorkflowAttributes",
		map[string]string{
			audit.Project: request.Project,
			audit.Domain:  request.Domain,
			audit.Name:    request.Workflow,
		},
		audit.ReadWrite,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowAttributesEndpointMetrics.delete)
	}

	return response, nil
}

func (m *AdminService) UpdateProjectDomainAttributes(ctx context.Context, request *admin.ProjectDomainAttributesUpdateRequest) (
	*admin.ProjectDomainAttributesUpdateResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ProjectDomainAttributesUpdateResponse
	var err error
	m.Metrics.projectDomainAttributesEndpointMetrics.update.Time(func() {
		response, err = m.ProjectDomainAttributesManager.UpdateProjectDomainAttributes(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"UpdateProjectDomainAttributes",
		map[string]string{
			audit.Project: request.Attributes.Project,
			audit.Domain:  request.Attributes.Domain,
		},
		audit.ReadWrite,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.projectDomainAttributesEndpointMetrics.update)
	}

	return response, nil
}

func (m *AdminService) GetProjectDomainAttributes(ctx context.Context, request *admin.ProjectDomainAttributesGetRequest) (
	*admin.ProjectDomainAttributesGetResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ProjectDomainAttributesGetResponse
	var err error
	m.Metrics.workflowAttributesEndpointMetrics.get.Time(func() {
		response, err = m.ProjectDomainAttributesManager.GetProjectDomainAttributes(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"GetProjectDomainAttributes",
		map[string]string{
			audit.Project: request.Project,
			audit.Domain:  request.Domain,
		},
		audit.ReadOnly,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowAttributesEndpointMetrics.get)
	}

	return response, nil
}

func (m *AdminService) DeleteProjectDomainAttributes(ctx context.Context, request *admin.ProjectDomainAttributesDeleteRequest) (
	*admin.ProjectDomainAttributesDeleteResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ProjectDomainAttributesDeleteResponse
	var err error
	m.Metrics.workflowAttributesEndpointMetrics.delete.Time(func() {
		response, err = m.ProjectDomainAttributesManager.DeleteProjectDomainAttributes(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"DeleteProjectDomainAttributes",
		map[string]string{
			audit.Project: request.Project,
			audit.Domain:  request.Domain,
		},
		audit.ReadWrite,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowAttributesEndpointMetrics.delete)
	}

	return response, nil
}

func (m *AdminService) UpdateProjectAttributes(ctx context.Context, request *admin.ProjectAttributesUpdateRequest) (
	*admin.ProjectAttributesUpdateResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
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
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"UpdateProjectAttributes",
		map[string]string{
			audit.Project: request.Attributes.Project,
		},
		audit.ReadWrite,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)

	return response, nil
}

func (m *AdminService) GetProjectAttributes(ctx context.Context, request *admin.ProjectAttributesGetRequest) (
	*admin.ProjectAttributesGetResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ProjectAttributesGetResponse
	var err error
	m.Metrics.workflowAttributesEndpointMetrics.get.Time(func() {
		response, err = m.ProjectAttributesManager.GetProjectAttributes(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"GetProjectAttributes",
		map[string]string{
			audit.Project: request.Project,
		},
		audit.ReadOnly,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowAttributesEndpointMetrics.get)
	}

	return response, nil
}

func (m *AdminService) DeleteProjectAttributes(ctx context.Context, request *admin.ProjectAttributesDeleteRequest) (
	*admin.ProjectAttributesDeleteResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ProjectAttributesDeleteResponse
	var err error
	m.Metrics.workflowAttributesEndpointMetrics.delete.Time(func() {
		response, err = m.ProjectAttributesManager.DeleteProjectAttributes(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"DeleteProjectAttributes",
		map[string]string{
			audit.Project: request.Project,
		},
		audit.ReadWrite,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowAttributesEndpointMetrics.delete)
	}

	return response, nil
}
