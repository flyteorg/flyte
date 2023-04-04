package interfaces

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

//go:generate mockery -name=SignalRepoInterface -output=../mocks -case=underscore

// Defines the interface for interacting with signal models.
type SignalRepoInterface interface {
	// Get retrieves a signal model from the database store.
	Get(ctx context.Context, input models.SignalKey) (models.Signal, error)
	// GetOrCreate inserts a signal model into the database store or returns one if it already exists.
	GetOrCreate(ctx context.Context, input *models.Signal) error
	// List all signals that match the input values.
	List(ctx context.Context, input ListResourceInput) ([]models.Signal, error)
	// Update sets the value on a signal in the database store.
	Update(ctx context.Context, input models.SignalKey, value []byte) error
}

type GetSignalInput struct {
	SignalID core.SignalIdentifier
}
