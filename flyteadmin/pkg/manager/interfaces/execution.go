package interfaces

import (
	"context"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

// Interface for managing Flyte Workflow Executions
type ExecutionInterface interface {
	CreateExecution(ctx context.Context, request admin.ExecutionCreateRequest, requestedAt time.Time) (
		*admin.ExecutionCreateResponse, error)
	RelaunchExecution(ctx context.Context, request admin.ExecutionRelaunchRequest, requestedAt time.Time) (
		*admin.ExecutionCreateResponse, error)
	CreateWorkflowEvent(ctx context.Context, request admin.WorkflowExecutionEventRequest) (
		*admin.WorkflowExecutionEventResponse, error)
	GetExecution(ctx context.Context, request admin.WorkflowExecutionGetRequest) (*admin.Execution, error)
	GetExecutionData(ctx context.Context, request admin.WorkflowExecutionGetDataRequest) (
		*admin.WorkflowExecutionGetDataResponse, error)
	ListExecutions(ctx context.Context, request admin.ResourceListRequest) (*admin.ExecutionList, error)
	TerminateExecution(
		ctx context.Context, request admin.ExecutionTerminateRequest) (*admin.ExecutionTerminateResponse, error)
}
