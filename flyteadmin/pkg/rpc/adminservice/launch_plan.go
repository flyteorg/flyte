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

func (m *AdminService) CreateLaunchPlan(
	ctx context.Context, request *admin.LaunchPlanCreateRequest) (*admin.LaunchPlanCreateResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.LaunchPlanCreateResponse
	var err error
	m.Metrics.launchPlanEndpointMetrics.create.Time(func() {
		response, err = m.LaunchPlanManager.CreateLaunchPlan(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.launchPlanEndpointMetrics.create)
	}
	m.Metrics.launchPlanEndpointMetrics.create.Success()
	return response, nil
}

func (m *AdminService) GetLaunchPlan(ctx context.Context, request *admin.ObjectGetRequest) (*admin.LaunchPlan, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	// NOTE: When the Get HTTP endpoint is called the resource type is implicit (from the URL) so we must add it
	// to the request.
	if request.Id != nil && request.Id.ResourceType == core.ResourceType_UNSPECIFIED {
		logger.Infof(ctx, "Adding resource type for unspecified value in request: [%+v]", request)
		request.Id.ResourceType = core.ResourceType_LAUNCH_PLAN
	}
	var response *admin.LaunchPlan
	var err error
	m.Metrics.launchPlanEndpointMetrics.get.Time(func() {
		response, err = m.LaunchPlanManager.GetLaunchPlan(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.launchPlanEndpointMetrics.get)
	}
	m.Metrics.launchPlanEndpointMetrics.get.Success()
	return response, nil

}

func (m *AdminService) GetActiveLaunchPlan(ctx context.Context, request *admin.ActiveLaunchPlanRequest) (*admin.LaunchPlan, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.LaunchPlan
	var err error
	m.Metrics.launchPlanEndpointMetrics.getActive.Time(func() {
		response, err = m.LaunchPlanManager.GetActiveLaunchPlan(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.launchPlanEndpointMetrics.getActive)
	}
	m.Metrics.launchPlanEndpointMetrics.getActive.Success()
	return response, nil
}

func (m *AdminService) UpdateLaunchPlan(ctx context.Context, request *admin.LaunchPlanUpdateRequest) (
	*admin.LaunchPlanUpdateResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	// NOTE: When the Get HTTP endpoint is called the resource type is implicit (from the URL) so we must add it
	// to the request.
	if request.Id != nil && request.Id.ResourceType == core.ResourceType_UNSPECIFIED {
		logger.Infof(ctx, "Adding resource type for unspecified value in request: [%+v]", request)
		request.Id.ResourceType = core.ResourceType_LAUNCH_PLAN
	}
	var response *admin.LaunchPlanUpdateResponse
	var err error
	m.Metrics.launchPlanEndpointMetrics.update.Time(func() {
		response, err = m.LaunchPlanManager.UpdateLaunchPlan(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.launchPlanEndpointMetrics.update)
	}
	m.Metrics.launchPlanEndpointMetrics.update.Success()
	return response, nil
}

func (m *AdminService) ListLaunchPlans(ctx context.Context, request *admin.ResourceListRequest) (
	*admin.LaunchPlanList, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Empty request.  Please rephrase.")
	}
	var response *admin.LaunchPlanList
	var err error
	m.Metrics.launchPlanEndpointMetrics.list.Time(func() {
		response, err = m.LaunchPlanManager.ListLaunchPlans(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.launchPlanEndpointMetrics.list)
	}

	m.Metrics.launchPlanEndpointMetrics.list.Success()
	return response, nil
}

func (m *AdminService) ListActiveLaunchPlans(ctx context.Context, request *admin.ActiveLaunchPlanListRequest) (
	*admin.LaunchPlanList, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Empty request.  Please rephrase.")
	}
	var response *admin.LaunchPlanList
	var err error
	m.Metrics.launchPlanEndpointMetrics.listActive.Time(func() {
		response, err = m.LaunchPlanManager.ListActiveLaunchPlans(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.launchPlanEndpointMetrics.listActive)
	}

	m.Metrics.launchPlanEndpointMetrics.listActive.Success()
	return response, nil
}

func (m *AdminService) ListLaunchPlanIds(ctx context.Context, request *admin.NamedEntityIdentifierListRequest) (
	*admin.NamedEntityIdentifierList, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Empty request.  Please rephrase.")
	}

	var response *admin.NamedEntityIdentifierList
	var err error
	m.Metrics.launchPlanEndpointMetrics.listIds.Time(func() {
		response, err = m.LaunchPlanManager.ListLaunchPlanIds(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.launchPlanEndpointMetrics.listIds)
	}

	m.Metrics.launchPlanEndpointMetrics.listIds.Success()
	return response, nil
}
