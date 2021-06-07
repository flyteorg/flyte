package ext

import (
	"context"

	"github.com/flyteorg/flytectl/pkg/filters"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

func (a *AdminFetcherExtClient) FetchExecution(ctx context.Context, name, project, domain string) (*admin.Execution, error) {
	e, err := a.AdminServiceClient().GetExecution(ctx, &admin.WorkflowExecutionGetRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: project,
			Domain:  domain,
			Name:    name,
		},
	})
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (a *AdminFetcherExtClient) ListExecution(ctx context.Context, project, domain string, filter filters.Filters) (*admin.ExecutionList, error) {
	transformFilters, err := filters.BuildResourceListRequestWithName(filter, project, domain, "")
	if err != nil {
		return nil, err
	}
	e, err := a.AdminServiceClient().ListExecutions(ctx, transformFilters)
	if err != nil {
		return nil, err
	}
	return e, nil
}
