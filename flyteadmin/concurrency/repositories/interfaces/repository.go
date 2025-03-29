package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

// ConcurrencyRepoInterface defines the interface for accessing concurrency-related data
type ConcurrencyRepoInterface interface {
	// GetPendingExecutions retrieves all pending executions ordered by creation time
	GetPendingExecutions(ctx context.Context) ([]models.Execution, error)

	// GetRunningExecutionsForLaunchPlan retrieves all non-terminal executions for a launch plan
	GetRunningExecutionsForLaunchPlan(ctx context.Context, launchPlanID core.Identifier) ([]models.Execution, error)

	// GetRunningExecutionsCount gets the count of non-terminal executions for a launch plan
	GetRunningExecutionsCount(ctx context.Context, launchPlanID core.Identifier) (int, error)

	// GetActiveLaunchPlanWithConcurrency gets a launch plan with concurrency settings
	GetActiveLaunchPlanWithConcurrency(ctx context.Context, launchPlanID core.Identifier) (*models.LaunchPlan, error)

	// UpdateExecutionPhase updates the phase of an execution
	UpdateExecutionPhase(ctx context.Context, executionID core.WorkflowExecutionIdentifier, phase core.WorkflowExecution_Phase) error

	// GetOldestRunningExecution gets the oldest running execution for a launch plan
	GetOldestRunningExecution(ctx context.Context, launchPlanID core.Identifier) (*models.Execution, error)

	// GetAllLaunchPlansWithConcurrency gets all launch plans with concurrency settings
	GetAllLaunchPlansWithConcurrency(ctx context.Context) ([]models.LaunchPlan, error)

	// GetExecutionByID returns an execution when given an executionID
	GetExecutionByID(ctx context.Context, executionID core.WorkflowExecutionIdentifier) (*models.Execution, error)

	// GetLaunchPlanByID retrieves a launch plan by its database ID and returns its core.Identifier
	GetLaunchPlanByID(ctx context.Context, launchPlanID uint) (*core.Identifier, error)
}
