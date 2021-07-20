package recovery

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
)

//go:generate mockery -name Client -output=mocks -case=underscore

type Client interface {
	RecoverNodeExecution(ctx context.Context, execID *core.WorkflowExecutionIdentifier, id *core.NodeExecutionIdentifier) (*admin.NodeExecution, error)
	RecoverNodeExecutionData(ctx context.Context, execID *core.WorkflowExecutionIdentifier, id *core.NodeExecutionIdentifier) (*admin.NodeExecutionGetDataResponse, error)
}

type recoveryClient struct {
	adminClient service.AdminServiceClient
}

func (c *recoveryClient) RecoverNodeExecution(ctx context.Context, execID *core.WorkflowExecutionIdentifier, nodeID *core.NodeExecutionIdentifier) (*admin.NodeExecution, error) {
	origNodeID := &core.NodeExecutionIdentifier{
		ExecutionId: execID,
		NodeId:      nodeID.NodeId,
	}
	return c.adminClient.GetNodeExecution(ctx, &admin.NodeExecutionGetRequest{
		Id: origNodeID,
	})
}

func (c *recoveryClient) RecoverNodeExecutionData(ctx context.Context, execID *core.WorkflowExecutionIdentifier, nodeID *core.NodeExecutionIdentifier) (*admin.NodeExecutionGetDataResponse, error) {
	origNodeID := &core.NodeExecutionIdentifier{
		ExecutionId: execID,
		NodeId:      nodeID.NodeId,
	}
	return c.adminClient.GetNodeExecutionData(ctx, &admin.NodeExecutionGetDataRequest{
		Id: origNodeID,
	})
}

func NewClient(adminClient service.AdminServiceClient) Client {
	return &recoveryClient{
		adminClient: adminClient,
	}
}
