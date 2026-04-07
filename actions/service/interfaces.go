package service

import (
	"context"

	"github.com/flyteorg/flyte/v2/actions/k8s"
	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// ActionsClientInterface defines the interface for actions operations.
// This combines the responsibilities of the queue and state clients.
type ActionsClientInterface interface {
	// Enqueue creates a TaskAction CR in Kubernetes.
	Enqueue(ctx context.Context, action *actions.Action, runSpec *task.RunSpec) error

	// GetState retrieves the state JSON for a TaskAction.
	GetState(ctx context.Context, actionID *common.ActionIdentifier) (string, error)

	// PutState updates the state and status of a TaskAction.
	PutState(ctx context.Context, actionID *common.ActionIdentifier, attempt uint32, status *workflow.ActionStatus, stateJSON string) error

	// AbortAction aborts a queued or running action, cascading to descendants.
	AbortAction(ctx context.Context, actionID *common.ActionIdentifier, reason *string) error

	// ListChildActions lists all TaskActions that are children of the given parent action.
	ListChildActions(ctx context.Context, parentActionID *common.ActionIdentifier) ([]*executorv1.TaskAction, error)

	// Subscribe creates a new subscription channel for action updates for the given parent action name.
	Subscribe(parentActionName string) chan *k8s.ActionUpdate

	// Unsubscribe removes the given channel from the subscription list for the parent action name.
	Unsubscribe(parentActionName string, ch chan *k8s.ActionUpdate)

	// StartWatching starts watching TaskAction resources.
	StartWatching(ctx context.Context) error

	// StopWatching stops the TaskAction watcher.
	StopWatching()
}
