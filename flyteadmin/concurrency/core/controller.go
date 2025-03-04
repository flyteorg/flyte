package core

import (
	"context"
	"time"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

// ConcurrencyController is the main interface for controlling execution concurrency
// It manages the lifecycle of executions based on concurrency policies
type ConcurrencyController interface {
	// Initialize initializes the controller from a previous state if available
	Initialize(ctx context.Context) error

	// ProcessPendingExecutions processes all pending executions according to concurrency policies
	// It runs in a background process and is periodically scheduled
	ProcessPendingExecutions(ctx context.Context) error

	// CheckConcurrencyConstraints checks if an execution can run based on concurrency constraints
	// Returns true if the execution can proceed, false otherwise
	CheckConcurrencyConstraints(ctx context.Context, launchPlanID core.Identifier) (bool, error)

	// GetConcurrencyPolicy returns the concurrency policy for a given launch plan
	GetConcurrencyPolicy(ctx context.Context, launchPlanID core.Identifier) (*admin.SchedulerPolicy, error)

	// TrackExecution adds a running execution to the concurrency tracking
	TrackExecution(ctx context.Context, execution models.Execution, launchPlanID core.Identifier) error

	// UntrackExecution removes a completed execution from concurrency tracking
	UntrackExecution(ctx context.Context, executionID core.WorkflowExecutionIdentifier) error

	// GetExecutionCounts returns the current count of executions for a launch plan
	GetExecutionCounts(ctx context.Context, launchPlanID core.Identifier) (int, error)

	// Close gracefully shuts down the controller
	Close() error
}

// ConcurrencyMetrics defines metrics for the concurrency controller
type ConcurrencyMetrics interface {
	// RecordPendingExecutions records the number of pending executions
	RecordPendingExecutions(count int64)

	// RecordProcessedExecution records that an execution was processed
	RecordProcessedExecution(policy string, allowed bool)

	// RecordLatency records the latency of processing a batch of executions
	RecordLatency(operation string, startTime time.Time)
}
