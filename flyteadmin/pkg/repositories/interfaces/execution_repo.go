package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
)

//go:generate mockery -name=ExecutionRepoInterface -output=../mocks -case=underscore

// Defines the interface for interacting with workflow execution models.
type ExecutionRepoInterface interface {
	// Inserts a workflow execution model into the database store.
	Create(ctx context.Context, input models.Execution, executionTagModel []*models.ExecutionTag) error
	// This updates only an existing execution model with all non-empty fields in the input.
	Update(ctx context.Context, execution models.Execution) error
	// Returns a matching execution if it exists.
	Get(ctx context.Context, input Identifier) (models.Execution, error)
	// Returns executions matching query parameters. A limit must be provided for the results page size.
	List(ctx context.Context, input ListResourceInput) (ExecutionCollectionOutput, error)
	// Returns count of executions matching query parameters.
	Count(ctx context.Context, input CountResourceInput) (int64, error)
	// Returns count of executions matching query parameters, grouped by phase.
	CountByPhase(ctx context.Context, input CountResourceInput) (ExecutionCountsByPhaseOutput, error)

	// FindNextStatusUpdatesCheckpoint returns last auto-increment ID of an execution in a contiguous chunk of terminal executions
	FindNextStatusUpdatesCheckpoint(ctx context.Context, cluster string, checkpoint uint) (uint, error)

	// FindStatusUpdates returns status updates after a given checkpoint
	FindStatusUpdates(ctx context.Context, cluster string, checkpoint uint, limit, offset int) ([]ExecutionStatus, error)
}

// Response format for a query on workflows.
type ExecutionCollectionOutput struct {
	Executions []models.Execution
}

type ExecutionCountsByPhase struct {
	Phase string
	Count int64
}

type ExecutionCountsByPhaseOutput []ExecutionCountsByPhase

type ExecutionStatus struct {
	models.ExecutionKey
	ID      uint
	Closure []byte
}
