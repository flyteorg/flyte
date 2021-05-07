package ext

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

// FetchAllVerOfLP fetches all the versions for give launch plan name
func (a *AdminFetcherExtClient) FetchAllVerOfLP(ctx context.Context, lpName, project, domain string) ([]*admin.LaunchPlan, error) {
	tList, err := a.AdminServiceClient().ListLaunchPlans(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: project,
			Domain:  domain,
			Name:    lpName,
		},
		SortBy: &admin.Sort{
			Key:       "created_at",
			Direction: admin.Sort_DESCENDING,
		},
		Limit: 100,
	})
	if err != nil {
		return nil, err
	}
	if len(tList.LaunchPlans) == 0 {
		return nil, fmt.Errorf("no launchplans retrieved for %v", lpName)
	}
	return tList.LaunchPlans, nil
}

// FetchLPLatestVersion fetches latest version for give launch plan name
func (a *AdminFetcherExtClient) FetchLPLatestVersion(ctx context.Context, name, project, domain string) (*admin.LaunchPlan, error) {
	// Fetch the latest version of the task.
	lpVersions, err := a.FetchAllVerOfLP(ctx, name, project, domain)
	if err != nil {
		return nil, err
	}
	lp := lpVersions[0]
	return lp, nil
}

// FetchLPVersion fetches particular version of launch plan
func (a *AdminFetcherExtClient) FetchLPVersion(ctx context.Context, name, version, project, domain string) (*admin.LaunchPlan, error) {
	lp, err := a.AdminServiceClient().GetLaunchPlan(ctx, &admin.ObjectGetRequest{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_LAUNCH_PLAN,
			Project:      project,
			Domain:       domain,
			Name:         name,
			Version:      version,
		},
	})
	if err != nil {
		return nil, err
	}
	return lp, nil
}
