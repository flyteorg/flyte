package service

import (
	"context"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// QueueClientInterface defines the interface for queue operations
// This allows for easier testing and mocking
type QueueClientInterface interface {
	EnqueueAction(ctx context.Context, req *workflow.EnqueueActionRequest) error
	AbortQueuedRun(ctx context.Context, runID *common.RunIdentifier, reason *string) error
	AbortQueuedAction(ctx context.Context, actionID *common.ActionIdentifier, reason *string) error
}
