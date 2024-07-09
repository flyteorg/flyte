package adminservice

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

func (m *AdminService) RegisterProject(ctx context.Context, request *admin.ProjectRegisterRequest) (
	*admin.ProjectRegisterResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ProjectRegisterResponse
	var err error
	m.Metrics.projectEndpointMetrics.register.Time(func() {
		response, err = m.ProjectManager.CreateProject(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.projectEndpointMetrics.register)
	}

	return response, nil
}

func (m *AdminService) ListProjects(ctx context.Context, request *admin.ProjectListRequest) (*admin.Projects, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.Projects
	var err error
	m.Metrics.projectEndpointMetrics.list.Time(func() {
		response, err = m.ProjectManager.ListProjects(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.projectEndpointMetrics.list)
	}

	m.Metrics.projectEndpointMetrics.list.Success()
	return response, nil
}

func (m *AdminService) UpdateProject(ctx context.Context, request *admin.Project) (
	*admin.ProjectUpdateResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ProjectUpdateResponse
	var err error
	m.Metrics.projectEndpointMetrics.register.Time(func() {
		response, err = m.ProjectManager.UpdateProject(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.projectEndpointMetrics.update)
	}

	return response, nil
}

func (m *AdminService) GetProject(ctx context.Context, request *admin.ProjectGetRequest) (*admin.Project, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.Project
	var err error
	m.Metrics.projectEndpointMetrics.get.Time(func() {
		response, err = m.ProjectManager.GetProject(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.projectEndpointMetrics.get)
	}

	m.Metrics.projectEndpointMetrics.get.Success()
	return response, nil
}

func (m *AdminService) GetDomains(ctx context.Context, request *admin.GetDomainRequest) (*admin.GetDomainsResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.GetDomainsResponse
	m.Metrics.domainEndpointMetrics.get.Time(func() {
		response = m.ProjectManager.GetDomains(ctx, *request)
	})

	m.Metrics.domainEndpointMetrics.get.Success()
	return response, nil
}
