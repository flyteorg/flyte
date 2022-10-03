package adminservice

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
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
		response, err = m.ResourceManager.UpdateWorkflowAttributes(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowAttributesEndpointMetrics.update)
	}

	return response, nil
}

func (m *AdminService) GetWorkflowAttributes(ctx context.Context, request *admin.WorkflowAttributesGetRequest) (
	*admin.WorkflowAttributesGetResponse, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.WorkflowAttributesGetResponse
	var err error
	m.Metrics.workflowAttributesEndpointMetrics.get.Time(func() {
		response, err = m.ResourceManager.GetWorkflowAttributes(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowAttributesEndpointMetrics.get)
	}

	return response, nil
}

func (m *AdminService) DeleteWorkflowAttributes(ctx context.Context, request *admin.WorkflowAttributesDeleteRequest) (
	*admin.WorkflowAttributesDeleteResponse, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.WorkflowAttributesDeleteResponse
	var err error
	m.Metrics.workflowAttributesEndpointMetrics.delete.Time(func() {
		response, err = m.ResourceManager.DeleteWorkflowAttributes(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowAttributesEndpointMetrics.delete)
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
		response, err = m.ResourceManager.UpdateProjectDomainAttributes(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.projectDomainAttributesEndpointMetrics.update)
	}

	return response, nil
}

func (m *AdminService) GetProjectDomainAttributes(ctx context.Context, request *admin.ProjectDomainAttributesGetRequest) (
	*admin.ProjectDomainAttributesGetResponse, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ProjectDomainAttributesGetResponse
	var err error
	m.Metrics.workflowAttributesEndpointMetrics.get.Time(func() {
		response, err = m.ResourceManager.GetProjectDomainAttributes(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowAttributesEndpointMetrics.get)
	}

	return response, nil
}

func (m *AdminService) DeleteProjectDomainAttributes(ctx context.Context, request *admin.ProjectDomainAttributesDeleteRequest) (
	*admin.ProjectDomainAttributesDeleteResponse, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ProjectDomainAttributesDeleteResponse
	var err error
	m.Metrics.workflowAttributesEndpointMetrics.delete.Time(func() {
		response, err = m.ResourceManager.DeleteProjectDomainAttributes(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowAttributesEndpointMetrics.delete)
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
	m.Metrics.projectAttributesEndpointMetrics.get.Time(func() {
		response, err = m.ResourceManager.UpdateProjectAttributes(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.projectAttributesEndpointMetrics.get)
	}

	return response, nil
}

func (m *AdminService) GetProjectAttributes(ctx context.Context, request *admin.ProjectAttributesGetRequest) (
	*admin.ProjectAttributesGetResponse, error) {

	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ProjectAttributesGetResponse
	var err error
	m.Metrics.projectAttributesEndpointMetrics.get.Time(func() {
		response, err = m.ResourceManager.GetProjectAttributes(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.projectAttributesEndpointMetrics.get)
	}

	return response, nil
}

func (m *AdminService) DeleteProjectAttributes(ctx context.Context, request *admin.ProjectAttributesDeleteRequest) (
	*admin.ProjectAttributesDeleteResponse, error) {

	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ProjectAttributesDeleteResponse
	var err error
	m.Metrics.projectAttributesEndpointMetrics.delete.Time(func() {
		response, err = m.ResourceManager.DeleteProjectAttributes(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.projectAttributesEndpointMetrics.delete)
	}

	return response, nil
}

func (m *AdminService) ListMatchableAttributes(ctx context.Context, request *admin.ListMatchableAttributesRequest) (
	*admin.ListMatchableAttributesResponse, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ListMatchableAttributesResponse
	var err error
	m.Metrics.matchableAttributesEndpointMetrics.list.Time(func() {
		response, err = m.ResourceManager.ListAll(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.matchableAttributesEndpointMetrics.list)
	}

	return response, nil
}
