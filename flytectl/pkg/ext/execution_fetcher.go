package ext

import (
	"context"

	"github.com/flyteorg/flyte/flytectl/pkg/filters"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

const (
	executionPageSize = 100
	// Offset-based pages need a stable order even when execution timestamps tie.
	// https://github.com/flyteorg/flyte/blob/ebb536a047facf32efb28cc02555a013874a0170/flyteadmin/pkg/repositories/models/base_model.go#L10-L17
	executionRowID = "id"
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
	req := &admin.NodeExecutionListRequest{
		WorkflowExecutionId: &core.WorkflowExecutionIdentifier{
			Project: project,
			Domain:  domain,
			Name:    name,
		},
		UniqueParentId: uniqueParentID,
		Limit:          executionPageSize,
		SortBy:         &admin.Sort{Key: executionRowID, Direction: admin.Sort_ASCENDING},
	}
	var nodes admin.NodeExecutionList
	for {
		page, err := a.AdminServiceClient().ListNodeExecutions(ctx, req)
		if err != nil {
			return nil, err
		}
		nodes.NodeExecutions = append(nodes.NodeExecutions, page.GetNodeExecutions()...)
		req.Token = page.GetToken()
		if req.GetToken() == "" {
			return &nodes, nil
		}
	}
}

func (a *AdminFetcherExtClient) FetchTaskExecutionsOnNode(ctx context.Context, nodeID, execName, project, domain string) (*admin.TaskExecutionList, error) {
	req := &admin.TaskExecutionListRequest{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: nodeID,
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: project,
				Domain:  domain,
				Name:    execName,
			},
		},
		Limit:  executionPageSize,
		SortBy: &admin.Sort{Key: executionRowID, Direction: admin.Sort_ASCENDING},
	}
	var tasks admin.TaskExecutionList
	for {
		page, err := a.AdminServiceClient().ListTaskExecutions(ctx, req)
		if err != nil {
			return nil, err
		}
		tasks.TaskExecutions = append(tasks.TaskExecutions, page.GetTaskExecutions()...)
		req.Token = page.GetToken()
		if req.GetToken() == "" {
			return &tasks, nil
		}
	}
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
