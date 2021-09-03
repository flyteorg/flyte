// Defines an event scheduler interface
package interfaces

import (
	"context"

	appInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

type AddScheduleInput struct {
	// Defines the unique identifier associated with the schedule
	Identifier core.Identifier
	// Defines the schedule expression.
	ScheduleExpression admin.Schedule
	// Message payload encoded as an CloudWatch event rule InputTemplate.
	Payload *string
	// Optional: The application-wide prefix to be applied for schedule names.
	ScheduleNamePrefix string
}

type RemoveScheduleInput struct {
	// Defines the unique identifier associated with the schedule
	Identifier core.Identifier
	// Optional: The application-wide prefix to be applied for schedule names.
	ScheduleNamePrefix string
}

type EventScheduler interface {
	// Schedules an event.
	AddSchedule(ctx context.Context, input AddScheduleInput) error

	// CreateScheduleInput using the scheduler config and launch plan identifier and schedule
	CreateScheduleInput(ctx context.Context, appConfig *appInterfaces.SchedulerConfig, identifier core.Identifier,
		schedule *admin.Schedule) (AddScheduleInput, error)

	// Removes an existing schedule.
	RemoveSchedule(ctx context.Context, input RemoveScheduleInput) error
}
