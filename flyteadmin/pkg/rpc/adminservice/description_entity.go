package adminservice

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

func (m *AdminService) GetDescriptionEntity(ctx context.Context, request *admin.ObjectGetRequest) (*admin.DescriptionEntity, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	// NOTE: When the Get HTTP endpoint is called the resource type is implicit (from the URL) so we must add it
	// to the request.
	if request.Id != nil && request.Id.ResourceType == core.ResourceType_UNSPECIFIED {
		logger.Infof(ctx, "Adding resource type for unspecified value in request: [%+v]", request)
		request.Id.ResourceType = core.ResourceType_TASK
	}
	var response *admin.DescriptionEntity
	var err error
	m.Metrics.descriptionEntityMetrics.get.Time(func() {
		response, err = m.DescriptionEntityManager.GetDescriptionEntity(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.descriptionEntityMetrics.get)
	}
	m.Metrics.descriptionEntityMetrics.get.Success()
	return response, nil
}

func (m *AdminService) ListDescriptionEntities(ctx context.Context, request *admin.DescriptionEntityListRequest) (*admin.DescriptionEntityList, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.DescriptionEntityList
	var err error
	m.Metrics.descriptionEntityMetrics.list.Time(func() {
		response, err = m.DescriptionEntityManager.ListDescriptionEntity(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.descriptionEntityMetrics.list)
	}
	m.Metrics.descriptionEntityMetrics.list.Success()
	return response, nil
}
