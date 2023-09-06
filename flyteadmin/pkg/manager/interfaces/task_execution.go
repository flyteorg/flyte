package interfaces

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

// Interface for managing Flyte Workflow TaskExecutions
type TaskExecutionInterface interface {
	CreateTaskExecutionEvent(ctx context.Context, request admin.TaskExecutionEventRequest) (
		*admin.TaskExecutionEventResponse, error)
	GetTaskExecution(ctx context.Context, request admin.TaskExecutionGetRequest) (*admin.TaskExecution, error)
	ListTaskExecutions(ctx context.Context, request admin.TaskExecutionListRequest) (*admin.TaskExecutionList, error)
	GetTaskExecutionData(
		ctx context.Context, request admin.TaskExecutionGetDataRequest) (*admin.TaskExecutionGetDataResponse, error)
}
