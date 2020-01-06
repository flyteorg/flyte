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

func (m *AdminService) GetNamedEntity(ctx context.Context, request *admin.NamedEntityGetRequest) (*admin.NamedEntity, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}

	var response *admin.NamedEntity
	var err error
	m.Metrics.namedEntityEndpointMetrics.get.Time(func() {
		response, err = m.NamedEntityManager.GetNamedEntity(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"GetNamedEntity",
		audit.ParametersFromNamedEntityIdentifierAndResource(request.Id, request.ResourceType),
		audit.ReadOnly,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.namedEntityEndpointMetrics.get)
	}
	m.Metrics.namedEntityEndpointMetrics.get.Success()
	return response, nil

}

func (m *AdminService) UpdateNamedEntity(ctx context.Context, request *admin.NamedEntityUpdateRequest) (
	*admin.NamedEntityUpdateResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}

	var response *admin.NamedEntityUpdateResponse
	var err error
	m.Metrics.namedEntityEndpointMetrics.update.Time(func() {
		response, err = m.NamedEntityManager.UpdateNamedEntity(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"UpdateNamedEntity",
		audit.ParametersFromNamedEntityIdentifierAndResource(request.Id, request.ResourceType),
		audit.ReadWrite,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.namedEntityEndpointMetrics.update)
	}
	m.Metrics.namedEntityEndpointMetrics.update.Success()
	return response, nil
}

func (m *AdminService) ListNamedEntities(ctx context.Context, request *admin.NamedEntityListRequest) (
	*admin.NamedEntityList, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}

	var response *admin.NamedEntityList
	var err error
	m.Metrics.namedEntityEndpointMetrics.list.Time(func() {
		response, err = m.NamedEntityManager.ListNamedEntities(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"ListNamedEntities",
		map[string]string{
			audit.Project:      request.Project,
			audit.Domain:       request.Domain,
			audit.ResourceType: request.ResourceType.String(),
		},
		audit.ReadOnly,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.namedEntityEndpointMetrics.list)
	}
	m.Metrics.namedEntityEndpointMetrics.list.Success()
	return response, nil
}
