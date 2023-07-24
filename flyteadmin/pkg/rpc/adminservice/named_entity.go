package adminservice

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (m *AdminService) GetNamedEntity(ctx context.Context, request *admin.NamedEntityGetRequest) (*admin.NamedEntity, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}

	var response *admin.NamedEntity
	var err error
	m.Metrics.namedEntityEndpointMetrics.get.Time(func() {
		response, err = m.NamedEntityManager.GetNamedEntity(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.namedEntityEndpointMetrics.get)
	}
	m.Metrics.namedEntityEndpointMetrics.get.Success()
	return response, nil

}

func (m *AdminService) UpdateNamedEntity(ctx context.Context, request *admin.NamedEntityUpdateRequest) (
	*admin.NamedEntityUpdateResponse, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}

	var response *admin.NamedEntityUpdateResponse
	var err error
	m.Metrics.namedEntityEndpointMetrics.update.Time(func() {
		response, err = m.NamedEntityManager.UpdateNamedEntity(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.namedEntityEndpointMetrics.update)
	}
	m.Metrics.namedEntityEndpointMetrics.update.Success()
	return response, nil
}

func (m *AdminService) ListNamedEntities(ctx context.Context, request *admin.NamedEntityListRequest) (
	*admin.NamedEntityList, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}

	var response *admin.NamedEntityList
	var err error
	m.Metrics.namedEntityEndpointMetrics.list.Time(func() {
		response, err = m.NamedEntityManager.ListNamedEntities(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.namedEntityEndpointMetrics.list)
	}
	m.Metrics.namedEntityEndpointMetrics.list.Success()
	return response, nil
}
