package executor

import (
	"context"
	"time"

	"github.com/flyteorg/flyte/flyteadmin/scheduler/repositories/models"
)

//go:generate mockery --name Executor --output=mocks --case=underscore --with-expecter

// Executor allows the ability to create scheduled executions on admin
type Executor interface {
	// Execute sends a scheduled execution request to admin
	Execute(ctx context.Context, scheduledTime time.Time, s models.SchedulableEntity) error
}
