package util

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

func TestCanUserOrgActivateLaunchPlans(t *testing.T) {
	repo := repositoryMocks.NewMockRepository()
	repo.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetCountCallback(func(ctx context.Context, input interfaces.CountResourceInput) (int64, error) {
		assert.Len(t, input.InlineFilters, 2)
		orgExpr, _ := input.InlineFilters[0].GetGormQueryExpr()
		activeExpr, _ := input.InlineFilters[1].GetGormQueryExpr()

		assert.Equal(t, orgExpr.Args, org)
		assert.Equal(t, orgExpr.Query, testutils.OrgQueryPattern)

		assert.Equal(t, activeExpr.Args, int32(admin.LaunchPlanState_ACTIVE))
		assert.Equal(t, activeExpr.Query, testutils.StateQueryPattern)

		return 2, nil
	})

	eligible, err := CanUserOrgActivateLaunchPlans(context.Background(), org, 3, repo)

	assert.NoError(t, err)
	assert.True(t, eligible)
}
