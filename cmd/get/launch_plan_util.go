package get

import (
	"context"
	"fmt"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

// Reads the launchplan config to drive fetching the correct launch plans.
func FetchLPForName(ctx context.Context, name string, project string, domain string, cmdCtx cmdCore.CommandContext) ([]*admin.LaunchPlan, error) {
	var launchPlans []*admin.LaunchPlan
	var lp *admin.LaunchPlan
	var err error
	if launchPlanConfig.Latest {
		if lp, err = FetchLPLatestVersion(ctx, name, project, domain, cmdCtx); err != nil {
			return nil, err
		}
		launchPlans = append(launchPlans, lp)
	} else if launchPlanConfig.Version != "" {
		if lp, err = FetchLPVersion(ctx, name, launchPlanConfig.Version, project, domain, cmdCtx); err != nil {
			return nil, err
		}
		launchPlans = append(launchPlans, lp)
	} else {
		launchPlans, err = FetchAllVerOfLP(ctx, name, project, domain, cmdCtx)
		if err != nil {
			return nil, err
		}
	}
	if launchPlanConfig.ExecFile != "" {
		// There would be atleast one launchplan object when code reaches here and hence the length assertion is not required.
		lp = launchPlans[0]
		// Only write the first task from the tasks object.
		if err = CreateAndWriteExecConfigForWorkflow(lp, launchPlanConfig.ExecFile); err != nil {
			return nil, err
		}
	}
	return launchPlans, nil
}

func FetchAllVerOfLP(ctx context.Context, lpName string, project string, domain string, cmdCtx cmdCore.CommandContext) ([]*admin.LaunchPlan, error) {
	tList, err := cmdCtx.AdminClient().ListLaunchPlans(ctx, &admin.ResourceListRequest{
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

func FetchLPLatestVersion(ctx context.Context, name string, project string, domain string, cmdCtx cmdCore.CommandContext) (*admin.LaunchPlan, error) {
	// Fetch the latest version of the task.
	lpVersions, err := FetchAllVerOfLP(ctx, name, project, domain, cmdCtx)
	if err != nil {
		return nil, err
	}
	lp := lpVersions[0]
	return lp, nil
}

func FetchLPVersion(ctx context.Context, name string, version string, project string, domain string, cmdCtx cmdCore.CommandContext) (*admin.LaunchPlan, error) {
	lp, err := cmdCtx.AdminClient().GetLaunchPlan(ctx, &admin.ObjectGetRequest{
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
