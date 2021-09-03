package core

import (
	"context"

	"github.com/flyteorg/flyteadmin/scheduler/repositories"
	"github.com/flyteorg/flytestdlib/logger"
)

// Updater this updates the scheduler with the latest state of schedules from the DB.
type Updater struct {
	db        repositories.SchedulerRepoInterface
	scheduler Scheduler
}

func (u Updater) UpdateGoCronSchedules(ctx context.Context) {
	schedules, err := u.db.SchedulableEntityRepo().GetAll(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to fetch the schedules in this round due to %v", err)
		return
	}
	u.scheduler.UpdateSchedules(ctx, schedules)
}

func NewUpdater(db repositories.SchedulerRepoInterface,
	scheduler Scheduler) Updater {
	return Updater{db: db, scheduler: scheduler}
}
