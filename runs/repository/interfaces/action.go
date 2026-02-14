package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// ActionRepo defines the interface for actions/runs data access
type ActionRepo interface {
	// Run operations
	CreateRun(ctx context.Context, run *models.Run) (*models.Run, error)
	GetRun(ctx context.Context, runID *common.RunIdentifier) (*models.Run, error)
	ListRuns(ctx context.Context, req *workflow.ListRunsRequest) ([]*models.Run, string, error)
	AbortRun(ctx context.Context, runID *common.RunIdentifier, reason string, abortedBy *common.EnrichedIdentity) error

	// Action operations
	CreateAction(ctx context.Context, runID uint, actionSpec *workflow.ActionSpec) (*models.Action, error)
	GetAction(ctx context.Context, actionID *common.ActionIdentifier) (*models.Action, error)
	ListActions(ctx context.Context, runID *common.RunIdentifier, limit int, token string) ([]*models.Action, string, error)
	UpdateActionPhase(ctx context.Context, actionID *common.ActionIdentifier, phase string, startTime, endTime *string) error
	AbortAction(ctx context.Context, actionID *common.ActionIdentifier, reason string, abortedBy *common.EnrichedIdentity) error

	// Watch operations (for streaming)
	WatchRunUpdates(ctx context.Context, runID *common.RunIdentifier, updates chan<- *models.Run, errs chan<- error)
	WatchAllRunUpdates(ctx context.Context, updates chan<- *models.Run, errs chan<- error)
	WatchActionUpdates(ctx context.Context, runID *common.RunIdentifier, updates chan<- *models.Action, errs chan<- error)

	// State operations
	UpdateActionState(ctx context.Context, actionID *common.ActionIdentifier, state string) error
	GetActionState(ctx context.Context, actionID *common.ActionIdentifier) (string, error)

	// Event notification (for state updates)
	NotifyStateUpdate(ctx context.Context, actionID *common.ActionIdentifier) error
	WatchStateUpdates(ctx context.Context, updates chan<- *common.ActionIdentifier, errs chan<- error)
}
