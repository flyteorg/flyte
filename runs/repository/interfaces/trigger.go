package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// TriggerRepo manages the mutable triggers table and the append-only trigger_revisions table.
// Mirrors the Action (latest state) + ActionEvent (history) pattern used elsewhere in OSS.
type TriggerRepo interface {
	// SaveTrigger inserts or updates the latest trigger state and appends a revision row.
	// expectedRevision is the caller's current LatestRevision for optimistic locking.
	// Pass 0 when creating a brand-new trigger.
	SaveTrigger(ctx context.Context, trigger *models.Trigger, expectedRevision uint64) (*models.Trigger, error)

	// GetTrigger returns the latest (non-deleted) trigger row by composite key.
	GetTrigger(ctx context.Context, key TriggerNameKey) (*models.Trigger, error)

	// GetTriggerRevision returns a specific immutable revision row.
	GetTriggerRevision(ctx context.Context, project, domain, taskName, name string, revision uint64) (*models.TriggerRevision, error)

	// ListTriggers returns the latest (non-deleted) trigger rows matching the input.
	ListTriggers(ctx context.Context, input ListResourceInput) ([]*models.Trigger, error)

	// ListTriggerRevisions returns revision history for a given trigger, ordered by revision DESC.
	ListTriggerRevisions(ctx context.Context, project, domain, taskName, name string, input ListResourceInput) ([]*models.TriggerRevision, error)

	// UpdateTriggers activates or deactivates the given triggers and appends revision rows.
	UpdateTriggers(ctx context.Context, keys []TriggerNameKey, active bool) error

	// DeleteTriggers soft-deletes the given triggers and appends revision rows.
	DeleteTriggers(ctx context.Context, keys []TriggerNameKey) error

}

// TriggerNameKey is a lightweight identity tuple used for batch operations.
type TriggerNameKey struct {
	Project  string
	Domain   string
	TaskName string
	Name     string
}

// NewTriggerNameKey constructs a TriggerNameKey.
func NewTriggerNameKey(project, domain, taskName, name string) TriggerNameKey {
	return TriggerNameKey{project, domain, taskName, name}
}
