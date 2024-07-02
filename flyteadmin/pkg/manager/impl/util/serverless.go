package util

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	repoInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

func CanUserOrgActivateLaunchPlans(ctx context.Context, org string, maxActiveLaunchPlans int, repo repoInterfaces.Repository) (bool, error) {
	orgFilter, err := common.NewSingleValueFilter(common.LaunchPlan, common.Equal, shared.Org, org)
	if err != nil {
		return false, err
	}

	activeFilter, err := common.NewSingleValueFilter(common.LaunchPlan, common.Equal, shared.State, int32(admin.LaunchPlanState_ACTIVE))
	if err != nil {
		return false, err
	}
	filters := []common.InlineFilter{orgFilter, activeFilter}

	count, err := repo.LaunchPlanRepo().Count(ctx, repoInterfaces.CountResourceInput{
		InlineFilters: filters,
	})
	if err != nil {
		logger.Debugf(ctx, "failed to list count of active launch plans for org [%s] with err %v", org, err)
		return false, err
	}
	return count < int64(maxActiveLaunchPlans), nil
}
