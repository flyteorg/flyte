package repository

import (
	"context"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// Repository defines the interface for queue data access
type Repository interface {
	// EnqueueAction persists a new action to the queue
	EnqueueAction(ctx context.Context, req *workflow.EnqueueActionRequest) error

	// AbortQueuedRun marks all actions in a run as aborted
	AbortQueuedRun(ctx context.Context, runID *common.RunIdentifier, reason string) error

	// AbortQueuedAction marks a specific action as aborted
	AbortQueuedAction(ctx context.Context, actionID *common.ActionIdentifier, reason string) error

	// GetQueuedActions retrieves queued actions ready for processing
	GetQueuedActions(ctx context.Context, limit int) ([]*QueuedAction, error)

	// MarkAsProcessing marks an action as being processed
	MarkAsProcessing(ctx context.Context, actionID *common.ActionIdentifier) error

	// MarkAsCompleted marks an action as completed
	MarkAsCompleted(ctx context.Context, actionID *common.ActionIdentifier) error

	// MarkAsFailed marks an action as failed with an error message
	MarkAsFailed(ctx context.Context, actionID *common.ActionIdentifier, errorMsg string) error
}
