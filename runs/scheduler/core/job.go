package core

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// Executor fires a single scheduled run.
type Executor interface {
	Execute(ctx context.Context, t *models.Trigger, scheduledAt time.Time) error
}

// GoCronJob wraps a *models.Trigger and fires the executor at each scheduled time.
// It implements cron.TimedJob so the scheduler passes the exact scheduled time to Run.
type GoCronJob struct {
	trigger  *models.Trigger
	executor Executor
	ctx      context.Context
	entryID  cron.EntryID
}

// NewGoCronJob constructs a GoCronJob.
func NewGoCronJob(ctx context.Context, t *models.Trigger, executor Executor) *GoCronJob {
	return &GoCronJob{
		trigger:  t,
		executor: executor,
		ctx:      ctx,
	}
}

// Run is called by the cron scheduler with the exact scheduled trigger time.
func (j *GoCronJob) Run(scheduledAt time.Time) {
	if err := j.executor.Execute(j.ctx, j.trigger, scheduledAt.UTC()); err != nil {
		logger.Errorf(j.ctx, "scheduler: failed to fire trigger %s/%s/%s at %s: %v",
			j.trigger.Project, j.trigger.Domain, j.trigger.Name, scheduledAt.Format(time.RFC3339), err)
	}
}

// TriggerKey returns a stable string key for this job (used as the cron job name).
func TriggerKey(t *models.Trigger) string {
	return t.Project + "/" + t.Domain + "/" + t.TaskName + "/" + t.Name
}
