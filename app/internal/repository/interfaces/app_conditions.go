package interfaces

import (
	"context"

	flyteapp "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app"
)

// AppConditionsRepo persists and retrieves the condition history for apps.
// Each app has at most one row; conditions are stored as a serialized proto array.
type AppConditionsRepo interface {
	// AppendCondition appends cond to the stored conditions for the given app,
	// trimming the oldest entries if the total exceeds maxConditions.
	// Creates the row if it does not exist.
	AppendCondition(ctx context.Context, appID *flyteapp.Identifier, cond *flyteapp.Condition, maxConditions int) error

	// GetConditions returns the stored conditions for the given app.
	// Returns nil (not an error) if no conditions have been recorded yet.
	GetConditions(ctx context.Context, appID *flyteapp.Identifier) ([]*flyteapp.Condition, error)

	// DeleteConditions removes the conditions row for the given app.
	// No-ops if the row does not exist.
	DeleteConditions(ctx context.Context, appID *flyteapp.Identifier) error
}
