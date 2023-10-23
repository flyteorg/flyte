package ext

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytectl/pkg/filters"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

// FetchAllVerOfWorkflow fetches all the versions for give workflow name
func (a *AdminFetcherExtClient) FetchAllVerOfWorkflow(ctx context.Context, workflowName, project, domain string, filter filters.Filters) ([]*admin.Workflow, error) {
	tranformFilters, err := filters.BuildResourceListRequestWithName(filter, project, domain, workflowName)
	if err != nil {
		return nil, err
	}
	wList, err := a.AdminServiceClient().ListWorkflows(ctx, tranformFilters)
	if err != nil {
		return nil, err
	}
	if len(wList.Workflows) == 0 {
		return nil, fmt.Errorf("no workflow retrieved for %v", workflowName)
	}
	return wList.Workflows, nil
}

// FetchAllWorkflows fetches all workflows in project domain
func (a *AdminFetcherExtClient) FetchAllWorkflows(ctx context.Context, project, domain string, filter filters.Filters) ([]*admin.NamedEntity, error) {
	tranformFilters, err := filters.BuildNamedEntityListRequest(filter, project, domain, core.ResourceType_WORKFLOW)
	if err != nil {
		return nil, err
	}
	wList, err := a.AdminServiceClient().ListNamedEntities(ctx, tranformFilters)
	if err != nil {
		return nil, err
	}
	if len(wList.Entities) == 0 {
		return nil, fmt.Errorf("no workflow retrieved for %v project %v domain", project, domain)
	}
	return wList.Entities, nil
}

// FetchWorkflowLatestVersion fetches latest version for given workflow name
func (a *AdminFetcherExtClient) FetchWorkflowLatestVersion(ctx context.Context, name, project, domain string, filter filters.Filters) (*admin.Workflow, error) {
	// Fetch the latest version of the workflow.
	wVersions, err := a.FetchAllVerOfWorkflow(ctx, name, project, domain, filter)
	if err != nil {
		return nil, err
	}
	return a.FetchWorkflowVersion(ctx, name, wVersions[0].Id.Version, project, domain)
}

// FetchWorkflowVersion fetches particular version of workflow
func (a *AdminFetcherExtClient) FetchWorkflowVersion(ctx context.Context, name, version, project, domain string) (*admin.Workflow, error) {
	lp, err := a.AdminServiceClient().GetWorkflow(ctx, &admin.ObjectGetRequest{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_WORKFLOW,
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
