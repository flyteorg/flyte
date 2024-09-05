package adminservice

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

func (m *AdminService) GetConfiguration(ctx context.Context, request *admin.ConfigurationGetRequest) (
	*admin.ConfigurationGetResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}

	var response *admin.ConfigurationGetResponse
	var err error
	m.Metrics.configurationEndpointMetrics.get.Time(func() {
		response, err = m.ConfigurationManager.
			GetConfiguration(ctx, request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.configurationEndpointMetrics.get)
	}

	return response, nil
}

func (m *AdminService) UpdateConfiguration(ctx context.Context, request *admin.ConfigurationUpdateRequest) (
	*admin.ConfigurationUpdateResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}

	var response *admin.ConfigurationUpdateResponse
	var err error
	m.Metrics.configurationEndpointMetrics.update.Time(func() {
		response, err = m.ConfigurationManager.UpdateConfiguration(ctx, request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.configurationEndpointMetrics.update)
	}

	return response, nil
}
