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
	// Recreates a previously-run workflow execution that will point to the original execution so that propeller will
	// only start executing from the last known failure point. Propeller can recover individual workflow execution nodes
	// which previously succeeded based on the recovery (original) workflow execution id.
	RecoverExecution(ctx context.Context, request admin.ExecutionRecoverRequest, requestedAt time.Time) (
		*admin.ExecutionCreateResponse, error)
	CreateWorkflowEvent(ctx context.Context, request admin.WorkflowExecutionEventRequest) (
		*admin.WorkflowExecutionEventResponse, error)
	GetExecution(ctx context.Context, request admin.WorkflowExecutionGetRequest) (*admin.Execution, error)
	UpdateExecution(ctx context.Context, request admin.ExecutionUpdateRequest, requestedAt time.Time) (
		*admin.ExecutionUpdateResponse, error)
	GetExecutionData(ctx context.Context, request admin.WorkflowExecutionGetDataRequest) (
		*admin.WorkflowExecutionGetDataResponse, error)
	ListExecutions(ctx context.Context, request admin.ResourceListRequest) (*admin.ExecutionList, error)
	TerminateExecution(
		ctx context.Context, request admin.ExecutionTerminateRequest) (*admin.ExecutionTerminateResponse, error)
}
