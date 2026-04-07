package core

import (
	"context"
	"time"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/runs/repository/impl"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// ScheduleSyncer periodically reloads active schedule triggers from the database
// and reconciles the running cron jobs via GoCronScheduler.
type ScheduleSyncer struct {
	triggerRepo    interfaces.TriggerRepo
	scheduler      *GoCronScheduler
	resyncInterval time.Duration
}

// NewScheduleSyncer constructs a ScheduleSyncer.
func NewScheduleSyncer(
	triggerRepo interfaces.TriggerRepo,
	scheduler *GoCronScheduler,
	resyncInterval time.Duration,
) *ScheduleSyncer {
	return &ScheduleSyncer{
		triggerRepo:    triggerRepo,
		scheduler:      scheduler,
		resyncInterval: resyncInterval,
	}
}

// Run starts the sync loop. It blocks until ctx is cancelled.
func (s *ScheduleSyncer) Run(ctx context.Context) error {
	// Run an immediate sync on startup.
	s.sync(ctx)

	ticker := time.NewTicker(s.resyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			s.sync(ctx)
		}
	}
}

func (s *ScheduleSyncer) sync(ctx context.Context) {
	triggers, err := s.listActiveScheduleTriggers(ctx)
	if err != nil {
		logger.Errorf(ctx, "scheduler: failed to list active triggers: %v", err)
		return
	}

	logger.Debugf(ctx, "scheduler: syncing %d active schedule triggers", len(triggers))
	s.scheduler.UpdateSchedules(ctx, triggers)
}

func (s *ScheduleSyncer) listActiveScheduleTriggers(ctx context.Context) ([]*models.Trigger, error) {
	return ListActiveScheduleTriggers(ctx, s.triggerRepo)
}

// ListActiveScheduleTriggers returns all non-deleted, active triggers with
// automation_type = TYPE_SCHEDULE.
func ListActiveScheduleTriggers(ctx context.Context, repo interfaces.TriggerRepo) ([]*models.Trigger, error) {
	activeFilter := impl.NewEqualFilter("active", true)
	scheduleFilter := impl.NewEqualFilter("automation_type", "TYPE_SCHEDULE")
	filter := activeFilter.And(scheduleFilter)

	return repo.ListTriggers(ctx, interfaces.ListResourceInput{
		Limit:  10000, // fetch all; cron jobs are cheap in-memory
		Filter: filter,
	})
}
