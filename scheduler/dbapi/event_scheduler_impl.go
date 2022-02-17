package dbapi

import (
	"context"
	"fmt"

	repositoryInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"

	"github.com/flyteorg/flyteadmin/pkg/async/schedule/interfaces"
	scheduleInterfaces "github.com/flyteorg/flyteadmin/pkg/async/schedule/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
)

// eventScheduler used for saving the scheduler entries after launch plans are enabled or disabled.
type eventScheduler struct {
	db repositoryInterfaces.Repository
}

func (s *eventScheduler) CreateScheduleInput(ctx context.Context, appConfig *runtimeInterfaces.SchedulerConfig,
	identifier core.Identifier, schedule *admin.Schedule) (interfaces.AddScheduleInput, error) {

	addScheduleInput := scheduleInterfaces.AddScheduleInput{
		Identifier:         identifier,
		ScheduleExpression: *schedule,
	}
	return addScheduleInput, nil
}

func (s *eventScheduler) AddSchedule(ctx context.Context, input interfaces.AddScheduleInput) error {
	logger.Infof(ctx, "Received call to add schedule [%+v]", input)
	var cronString string
	var fixedRateValue uint32
	var fixedRateUnit admin.FixedRateUnit
	switch v := input.ScheduleExpression.GetScheduleExpression().(type) {
	case *admin.Schedule_Rate:
		fixedRateValue = v.Rate.Value
		fixedRateUnit = v.Rate.Unit
	case *admin.Schedule_CronSchedule:
		cronString = v.CronSchedule.Schedule
	default:
		return fmt.Errorf("failed adding schedule for unknown schedule expression type %v", v)
	}
	active := true
	modelInput := models.SchedulableEntity{
		CronExpression:      cronString,
		FixedRateValue:      fixedRateValue,
		Unit:                fixedRateUnit,
		KickoffTimeInputArg: input.ScheduleExpression.KickoffTimeInputArg,
		Active:              &active,
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: input.Identifier.Project,
			Domain:  input.Identifier.Domain,
			Name:    input.Identifier.Name,
			Version: input.Identifier.Version,
		},
	}
	err := s.db.SchedulableEntityRepo().Activate(ctx, modelInput)
	if err != nil {
		return err
	}
	logger.Infof(ctx, "Activated scheduled entity for %v ", input)
	return nil
}

func (s *eventScheduler) RemoveSchedule(ctx context.Context, input interfaces.RemoveScheduleInput) error {
	logger.Infof(ctx, "Received call to remove schedule [%+v]. Will deactivate it in the scheduler", input.Identifier)

	err := s.db.SchedulableEntityRepo().Deactivate(ctx, models.SchedulableEntityKey{
		Project: input.Identifier.Project,
		Domain:  input.Identifier.Domain,
		Name:    input.Identifier.Name,
		Version: input.Identifier.Version,
	})

	if err != nil {
		return err
	}
	logger.Infof(ctx, "Deactivated the schedule %v in the scheduler", input)
	return nil
}

func New(db repositoryInterfaces.Repository) interfaces.EventScheduler {
	return &eventScheduler{db: db}
}
