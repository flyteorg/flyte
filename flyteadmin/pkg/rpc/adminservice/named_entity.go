package adminservice

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

func (m *AdminService) GetNamedEntity(ctx context.Context, request *admin.NamedEntityGetRequest) (*admin.NamedEntity, error) {
	var response *admin.NamedEntity
	var err error
	m.Metrics.namedEntityEndpointMetrics.get.Time(func() {
		response, err = m.NamedEntityManager.GetNamedEntity(ctx, request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.namedEntityEndpointMetrics.get)
	}
	m.Metrics.namedEntityEndpointMetrics.get.Success()
	return response, nil

}

func (m *AdminService) UpdateNamedEntity(ctx context.Context, request *admin.NamedEntityUpdateRequest) (
	*admin.NamedEntityUpdateResponse, error) {
	var response *admin.NamedEntityUpdateResponse
	var err error
	m.Metrics.namedEntityEndpointMetrics.update.Time(func() {
		response, err = m.NamedEntityManager.UpdateNamedEntity(ctx, request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.namedEntityEndpointMetrics.update)
	}
	m.Metrics.namedEntityEndpointMetrics.update.Success()
	return response, nil
}

func (m *AdminService) ListNamedEntities(ctx context.Context, request *admin.NamedEntityListRequest) (
	*admin.NamedEntityList, error) {
	var response *admin.NamedEntityList
	var err error
	m.Metrics.namedEntityEndpointMetrics.list.Time(func() {
		response, err = m.NamedEntityManager.ListNamedEntities(ctx, request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.namedEntityEndpointMetrics.list)
	}
	m.Metrics.namedEntityEndpointMetrics.list.Success()
	return response, nil
}
