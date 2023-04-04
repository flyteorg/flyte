package ext

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytectl/pkg/filters"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

func (a *AdminFetcherExtClient) FetchAllVerOfTask(ctx context.Context, name, project, domain string, filter filters.Filters) ([]*admin.Task, error) {
	transformFilters, err := filters.BuildResourceListRequestWithName(filter, project, domain, name)
	if err != nil {
		return nil, err
	}
	tList, err := a.AdminServiceClient().ListTasks(ctx, transformFilters)
	if err != nil {
		return nil, err
	}
	if len(tList.Tasks) == 0 {
		return nil, fmt.Errorf("no tasks retrieved for %v", name)
	}
	return tList.Tasks, nil
}

func (a *AdminFetcherExtClient) FetchTaskLatestVersion(ctx context.Context, name, project, domain string, filter filters.Filters) (*admin.Task, error) {
	var t *admin.Task
	var err error
	// Fetch the latest version of the task.
	var taskVersions []*admin.Task
	taskVersions, err = a.FetchAllVerOfTask(ctx, name, project, domain, filter)
	if err != nil {
		return nil, err
	}
	t = taskVersions[0]
	return t, nil
}

func (a *AdminFetcherExtClient) FetchTaskVersion(ctx context.Context, name, version, project, domain string) (*admin.Task, error) {
	t, err := a.AdminServiceClient().GetTask(ctx, &admin.ObjectGetRequest{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Project:      project,
			Domain:       domain,
			Name:         name,
			Version:      version,
		},
	})
	if err != nil {
		return nil, err
	}
	return t, nil
}
