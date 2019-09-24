// No-op event event scheduler for use in development.
package noop

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/async/schedule/interfaces"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/lyft/flytestdlib/logger"
)

type EventScheduler struct{}

func (s *EventScheduler) AddSchedule(ctx context.Context, input interfaces.AddScheduleInput) error {
	logger.Debugf(ctx, "Received call to add schedule [%+v]", input)
	logger.Debug(ctx, "Not scheduling anything")
	return nil
}

func (s *EventScheduler) RemoveSchedule(ctx context.Context, identifier admin.NamedEntityIdentifier) error {
	logger.Debugf(ctx, "Received call to remove schedule [%+v]", identifier)
	logger.Debug(ctx, "Not scheduling anything")
	return nil
}

func NewNoopEventScheduler() interfaces.EventScheduler {
	return &EventScheduler{}
}
