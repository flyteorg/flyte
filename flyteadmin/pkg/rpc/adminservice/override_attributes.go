package adminservice

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (m *AdminService) GetOverrideAttributes(ctx context.Context, request *admin.OverrideAttributesGetRequest) (
	*admin.OverrideAttributesGetResponse, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}

	var response *admin.OverrideAttributesGetResponse
	var err error
	m.Metrics.overrideAttributesMetrics.get.Time(func() {
		response, err = m.OverrideAttributeManager.GetOverrideAttributes(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.overrideAttributesMetrics.get)
	}

	return response, nil
}

func (m *AdminService) UpdateOverrideAttributes(ctx context.Context, request *admin.OverrideAttributesUpdateRequest) (
	*admin.OverrideAttributesUpdateResponse, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}

	var response *admin.OverrideAttributesUpdateResponse
	var err error
	m.Metrics.overrideAttributesMetrics.update.Time(func() {
		response, err = m.OverrideAttributeManager.UpdateOverrideAttributes(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.overrideAttributesMetrics.update)
	}

	return response, nil
}
