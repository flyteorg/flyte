// No-op event event scheduler for use in development.
package noop

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/async/schedule/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
)

type EventScheduler struct{}

func (s *EventScheduler) CreateScheduleInput(ctx context.Context, appConfig *runtimeInterfaces.SchedulerConfig, identifier core.Identifier, schedule *admin.Schedule) (interfaces.AddScheduleInput, error) {
	panic("implement me")
}

func (s *EventScheduler) AddSchedule(ctx context.Context, input interfaces.AddScheduleInput) error {
	logger.Debugf(ctx, "Received call to add schedule [%+v]", input)
	logger.Debug(ctx, "Not scheduling anything")
	return nil
}

func (s *EventScheduler) RemoveSchedule(ctx context.Context, input interfaces.RemoveScheduleInput) error {
	logger.Debugf(ctx, "Received call to remove schedule [%+v]", input.Identifier)
	logger.Debug(ctx, "Not scheduling anything")
	return nil
}

func NewNoopEventScheduler() interfaces.EventScheduler {
	return &EventScheduler{}
}
