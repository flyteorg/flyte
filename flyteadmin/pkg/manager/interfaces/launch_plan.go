package interfaces

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

// Interface for managing Flyte Launch Plans
type LaunchPlanInterface interface {
	// Interface to create Launch Plans based on the request.
	CreateLaunchPlan(ctx context.Context, request admin.LaunchPlanCreateRequest) (
		*admin.LaunchPlanCreateResponse, error)
	UpdateLaunchPlan(ctx context.Context, request admin.LaunchPlanUpdateRequest) (
		*admin.LaunchPlanUpdateResponse, error)
	GetLaunchPlan(ctx context.Context, request admin.ObjectGetRequest) (
		*admin.LaunchPlan, error)
	GetActiveLaunchPlan(ctx context.Context, request admin.ActiveLaunchPlanRequest) (
		*admin.LaunchPlan, error)
	ListLaunchPlans(ctx context.Context, request admin.ResourceListRequest) (
		*admin.LaunchPlanList, error)
	ListActiveLaunchPlans(ctx context.Context, request admin.ActiveLaunchPlanListRequest) (
		*admin.LaunchPlanList, error)
	ListLaunchPlanIds(ctx context.Context, request admin.NamedEntityIdentifierListRequest) (
		*admin.NamedEntityIdentifierList, error)
}
