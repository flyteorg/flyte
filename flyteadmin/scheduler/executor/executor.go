package executor

import (
	"context"
	"time"

	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
)

// Executor allows the ability to create scheduled executions on admin
type Executor interface {
	// Execute sends a scheduled execution request to admin
	Execute(ctx context.Context, scheduledTime time.Time, s models.SchedulableEntity) error
}
