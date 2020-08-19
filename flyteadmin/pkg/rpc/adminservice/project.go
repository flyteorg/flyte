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

func (m *AdminService) RegisterProject(ctx context.Context, request *admin.ProjectRegisterRequest) (
	*admin.ProjectRegisterResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ProjectRegisterResponse
	var err error
	m.Metrics.projectEndpointMetrics.register.Time(func() {
		response, err = m.ProjectManager.CreateProject(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"RegisterProject",
		map[string]string{
			audit.Project: request.Project.Id,
		},
		audit.ReadWrite,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.projectEndpointMetrics.register)
	}

	return response, nil
}

func (m *AdminService) ListProjects(ctx context.Context, request *admin.ProjectListRequest) (*admin.Projects, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.Projects
	var err error
	m.Metrics.projectEndpointMetrics.list.Time(func() {
		response, err = m.ProjectManager.ListProjects(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"ListProjects",
		map[string]string{},
		audit.ReadOnly,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.projectEndpointMetrics.list)
	}

	m.Metrics.projectEndpointMetrics.list.Success()
	return response, nil
}

func (m *AdminService) UpdateProject(ctx context.Context, request *admin.Project) (
	*admin.ProjectUpdateResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ProjectUpdateResponse
	var err error
	m.Metrics.projectEndpointMetrics.register.Time(func() {
		response, err = m.ProjectManager.UpdateProject(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"UpdateProject",
		map[string]string{
			audit.Project: request.Id,
		},
		audit.ReadWrite,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.projectEndpointMetrics.update)
	}

	return response, nil
}
