package service

import (
	"context"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/state/k8s"
)

// StateClientInterface defines the interface for state operations
// This allows for easier testing and mocking
type StateClientInterface interface {
	// GetState retrieves the state JSON for a TaskAction
	GetState(ctx context.Context, actionID *common.ActionIdentifier) (string, error)

	// PutState updates the state JSON for a TaskAction
	PutState(ctx context.Context, actionID *common.ActionIdentifier, stateJSON string) error

	// ListChildActions lists all TaskActions that are children of the given parent action
	ListChildActions(ctx context.Context, parentActionID *common.ActionIdentifier) ([]*executorv1.TaskAction, error)

	// GetTaskAction retrieves a specific TaskAction
	GetTaskAction(ctx context.Context, actionID *common.ActionIdentifier) (*executorv1.TaskAction, error)

	// Subscribe creates a new subscription channel for action updates
	Subscribe() chan *k8s.ActionUpdate

	// Unsubscribe removes a subscription channel
	Unsubscribe(ch chan *k8s.ActionUpdate)

	// StartWatching starts watching TaskAction resources
	StartWatching(ctx context.Context) error

	// StopWatching stops the TaskAction watcher
	StopWatching()
}
