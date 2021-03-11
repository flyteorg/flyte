// Defines an event scheduler interface
package interfaces

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

type AddScheduleInput struct {
	// Defines the unique identifier associated with the schedule
	Identifier admin.NamedEntityIdentifier
	// Defines the schedule expression.
	ScheduleExpression admin.Schedule
	// Message payload encoded as an CloudWatch event rule InputTemplate.
	Payload *string
	// Optional: The application-wide prefix to be applied for schedule names.
	ScheduleNamePrefix string
}

type RemoveScheduleInput struct {
	// Defines the unique identifier associated with the schedule
	Identifier admin.NamedEntityIdentifier
	// Optional: The application-wide prefix to be applied for schedule names.
	ScheduleNamePrefix string
}

type EventScheduler interface {
	// Schedules an event.
	AddSchedule(ctx context.Context, input AddScheduleInput) error

	// Removes an existing schedule.
	RemoveSchedule(ctx context.Context, input RemoveScheduleInput) error
}
