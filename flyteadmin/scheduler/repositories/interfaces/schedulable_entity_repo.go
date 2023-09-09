package interfaces

import (
	"context"

	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
)

//go:generate mockery -name=SchedulableEntityRepoInterface -output=../mocks -case=underscore

// SchedulableEntityRepoInterface : An Interface for interacting with the schedulable entity in the database
type SchedulableEntityRepoInterface interface {

	// Create a schedulable entity in the database store
	Create(ctx context.Context, input models.SchedulableEntity) error

	// Activate a schedulable entity in the database store.
	Activate(ctx context.Context, input models.SchedulableEntity) error

	// Deactivate a schedulable entity in the database store.
	Deactivate(ctx context.Context, ID models.SchedulableEntityKey) error

	// Get a schedulable entity from the database store using the schedulable entity id.
	Get(ctx context.Context, ID models.SchedulableEntityKey) (models.SchedulableEntity, error)

	// GetAll Gets all the active schedulable entities from the db
	GetAll(ctx context.Context) ([]models.SchedulableEntity, error)
}
