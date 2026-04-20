package interfaces

import (
	"context"
	"time"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// ActionRepo defines the interface for actions/runs data access
type ActionRepo interface {
	// Run operations
	GetRun(ctx context.Context, runID *common.RunIdentifier) (*models.Run, error)
	// AbortRun marks only the root action as ABORTED and sets abort_requested_at on it.
	// K8s cascades CRD deletion to child actions via OwnerReferences; the action service
	// informer handles marking them ABORTED in DB when their CRDs are deleted.
	AbortRun(ctx context.Context, runID *common.RunIdentifier, reason string, abortedBy *common.EnrichedIdentity) error

	// Action operations
	CreateAction(ctx context.Context, action *models.Action, updateTriggeredAt bool) (*models.Action, error)
	InsertEvents(ctx context.Context, events []*models.ActionEvent) error
	ListEvents(ctx context.Context, actionID *common.ActionIdentifier, limit int) ([]*models.ActionEvent, error)
	ListEventsSince(ctx context.Context, actionID *common.ActionIdentifier, attempt uint32, since time.Time, offset, limit int) ([]*models.ActionEvent, error)
	GetLatestEventByAttempt(ctx context.Context, actionID *common.ActionIdentifier, attempt uint32) (*models.ActionEvent, error)
	GetAction(ctx context.Context, actionID *common.ActionIdentifier) (*models.Action, error)
	ListActions(ctx context.Context, input ListResourceInput) ([]*models.Action, error)
	UpdateActionPhase(ctx context.Context, actionID *common.ActionIdentifier, phase common.ActionPhase, attempts uint32, cacheStatus core.CatalogCacheStatus, endTime *time.Time) error
	// AbortAction marks only the targeted action as ABORTED and sets abort_requested_at.
	// K8s cascades CRD deletion to descendants via OwnerReferences.
	AbortAction(ctx context.Context, actionID *common.ActionIdentifier, reason string, abortedBy *common.EnrichedIdentity) error

	// Abort reconciliation — used by the background AbortReconciler.
	ListPendingAborts(ctx context.Context) ([]*models.Action, error)
	MarkAbortAttempt(ctx context.Context, actionID *common.ActionIdentifier) (attemptCount int, err error)
	ClearAbortRequest(ctx context.Context, actionID *common.ActionIdentifier) error

	// Watch operations (for streaming)
	WatchRunUpdates(ctx context.Context, runID *common.RunIdentifier, updates chan<- *models.Run, errs chan<- error)
	WatchAllRunUpdates(ctx context.Context, updates chan<- *models.Run, errs chan<- error)
	WatchAllActionUpdates(ctx context.Context, runID *common.RunIdentifier, updates chan<- *models.Action, errs chan<- error)
	WatchActionUpdates(ctx context.Context, actionID *common.ActionIdentifier, updates chan<- *models.Action, errs chan<- error)

	// State operations
	UpdateActionState(ctx context.Context, actionID *common.ActionIdentifier, state string) error
	GetActionState(ctx context.Context, actionID *common.ActionIdentifier) (string, error)

	// Event notification (for state updates)
	NotifyStateUpdate(ctx context.Context, actionID *common.ActionIdentifier) error
	WatchStateUpdates(ctx context.Context, updates chan<- *common.ActionIdentifier, errs chan<- error)

	// Aggregation operations
	ListRootActions(ctx context.Context, project, domain string, startDate, endDate *time.Time, limit int) ([]*models.Action, error)
}
