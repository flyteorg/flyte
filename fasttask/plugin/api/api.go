package api

import (
	"context"
	"time"

	coreIdl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
)

//go:generate mockery -all -case=underscore

type TaskStatus struct {
	Phase        core.Phase
	Reason       string
	TaskDuration time.Duration
}

// FastTaskService defines the interface for managing assignment and management of task executions
type FastTaskService interface {
	CheckStatus(ctx context.Context, taskID, queueID, workerID string) (TaskStatus, error)
	Cleanup(ctx context.Context, taskID, queueID, workerID string) error
	OfferOnQueue(ctx context.Context, execID *coreIdl.WorkflowExecutionIdentifier, queueID, taskID, namespace, workflowID string, cmd []string, envVars map[string]string) (string, error)
}
