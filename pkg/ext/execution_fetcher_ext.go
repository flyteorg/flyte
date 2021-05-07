package ext

import (
	"context"

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
