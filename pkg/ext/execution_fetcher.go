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

func (a *AdminFetcherExtClient) FetchNodeExecutionData(ctx context.Context, nodeID, execName, project, domain string) (*admin.NodeExecutionGetDataResponse, error) {
	ne, err := a.AdminServiceClient().GetNodeExecutionData(ctx, &admin.NodeExecutionGetDataRequest{
		Id: &core.NodeExecutionIdentifier{
			NodeId: nodeID,
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: project,
				Domain:  domain,
				Name:    execName,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return ne, nil
}

func (a *AdminFetcherExtClient) FetchNodeExecutionDetails(ctx context.Context, name, project, domain, uniqueParentID string) (*admin.NodeExecutionList, error) {
	ne, err := a.AdminServiceClient().ListNodeExecutions(ctx, &admin.NodeExecutionListRequest{
		WorkflowExecutionId: &core.WorkflowExecutionIdentifier{
			Project: project,
			Domain:  domain,
			Name:    name,
		},
		UniqueParentId: uniqueParentID,
		Limit:          100,
	})
	if err != nil {
		return nil, err
	}
	return ne, nil
}

func (a *AdminFetcherExtClient) FetchTaskExecutionsOnNode(ctx context.Context, nodeID, execName, project, domain string) (*admin.TaskExecutionList, error) {
	te, err := a.AdminServiceClient().ListTaskExecutions(ctx, &admin.TaskExecutionListRequest{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: nodeID,
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: project,
				Domain:  domain,
				Name:    execName,
			},
		},
		Limit: 100,
	})
	if err != nil {
		return nil, err
	}
	return te, nil
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
