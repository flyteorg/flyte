package interfaces

import (
	"context"

	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
)

//go:generate mockery -name=ScheduleEntitiesSnapShotRepoInterface -output=../mocks -case=underscore

// ScheduleEntitiesSnapShotRepoInterface : An Interface for interacting with the snapshot of schedulable entities in the database
type ScheduleEntitiesSnapShotRepoInterface interface {

	// Create/ Update the snapshot in the  database store
	Write(ctx context.Context, input models.ScheduleEntitiesSnapshot) error

	// Get the latest snapshot from the database store.
	Read(ctx context.Context) (models.ScheduleEntitiesSnapshot, error)
}
