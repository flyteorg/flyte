// Defines an event scheduler interface
package interfaces

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

type AddScheduleInput struct {
	// Defines the unique identifier associated with the schedule
	Identifier admin.NamedEntityIdentifier
	// Defines the schedule expression.
	ScheduleExpression admin.Schedule
	// Message payload encoded as an CloudWatch event rule InputTemplate.
	Payload *string
}

type EventScheduler interface {
	// Schedules an event.
	AddSchedule(ctx context.Context, input AddScheduleInput) error

	// Removes an existing schedule.
	RemoveSchedule(ctx context.Context, identifier admin.NamedEntityIdentifier) error
}
