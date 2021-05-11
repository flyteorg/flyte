package ext

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

// FetchAllVerOfWorkflow fetches all the versions for give workflow name
func (a *AdminFetcherExtClient) FetchAllVerOfWorkflow(ctx context.Context, workflowName, project, domain string) ([]*admin.Workflow, error) {
	wList, err := a.AdminServiceClient().ListWorkflows(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: project,
			Domain:  domain,
			Name:    workflowName,
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
	if len(wList.Workflows) == 0 {
		return nil, fmt.Errorf("no workflow retrieved for %v", workflowName)
	}
	return wList.Workflows, nil
}

// FetchWorkflowLatestVersion fetches latest version for given workflow name
func (a *AdminFetcherExtClient) FetchWorkflowLatestVersion(ctx context.Context, name, project, domain string) (*admin.Workflow, error) {
	// Fetch the latest version of the workflow.
	wVersions, err := a.FetchAllVerOfWorkflow(ctx, name, project, domain)
	if err != nil {
		return nil, err
	}
	w := wVersions[0]
	return w, nil
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
