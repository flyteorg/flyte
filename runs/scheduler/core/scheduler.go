package core

import (
	"context"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// GoCronScheduler manages a set of cron jobs, one per active trigger.
// It is safe for concurrent use.
type GoCronScheduler struct {
	cron     *cron.Cron
	executor Executor

	mu   sync.Mutex
	jobs map[string]*GoCronJob // TriggerKey -> job
}

// NewGoCronScheduler constructs a GoCronScheduler.
func NewGoCronScheduler(executor Executor) *GoCronScheduler {
	c := cron.New(cron.WithLocation(time.UTC), cron.WithLogger(cron.DiscardLogger))
	return &GoCronScheduler{
		cron:     c,
		executor: executor,
		jobs:     make(map[string]*GoCronJob),
	}
}

// Start starts the underlying cron scheduler. It is non-blocking.
func (s *GoCronScheduler) Start() {
	s.cron.Start()
}

// Stop gracefully stops the cron scheduler, waiting for running jobs to finish.
func (s *GoCronScheduler) Stop() context.Context {
	return s.cron.Stop()
}

// UpdateSchedules reconciles the set of running cron jobs with the supplied
// active triggers. New triggers are added; triggers no longer present are removed;
// triggers whose spec changed are replaced.
func (s *GoCronScheduler) UpdateSchedules(ctx context.Context, triggers []*models.Trigger) {
	desired := make(map[string]*models.Trigger, len(triggers))
	for _, t := range triggers {
		sched, err := ParseSchedule(t)
		if err != nil {
			logger.Errorf(ctx, "scheduler: invalid schedule for trigger %s: %v", TriggerKey(t), err)
			continue
		}
		if sched == nil {
			continue
		}
		desired[TriggerKey(t)] = t
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove jobs no longer desired.
	for key, job := range s.jobs {
		if _, ok := desired[key]; !ok {
			s.cron.Remove(job.entryID)
			delete(s.jobs, key)
			logger.Debugf(ctx, "scheduler: removed job %s", key)
		}
	}

	// Add new jobs; skip unchanged ones (same LatestRevision = same spec).
	for key, t := range desired {
		if existing, exists := s.jobs[key]; exists {
			if existing.trigger.LatestRevision == t.LatestRevision {
				continue
			}
			// Spec changed — remove old job before re-adding.
			s.cron.Remove(existing.entryID)
			delete(s.jobs, key)
		}

		sched, _ := ParseSchedule(t)
		startTime, err := StartTime(t)
		if err != nil {
			logger.Errorf(ctx, "scheduler: failed to compute start time for %s: %v", key, err)
			continue
		}
		job := NewGoCronJob(ctx, t, s.executor)
		job.entryID = s.cron.ScheduleTimedJob(sched, job, startTime)
		s.jobs[key] = job
		logger.Debugf(ctx, "scheduler: registered job %s", key)
	}
}

// CatchupAll fires catchup runs for all triggers that have missed scheduled times.
// It processes at most maxRunsPerLoop total runs across all triggers.
func (s *GoCronScheduler) CatchupAll(
	ctx context.Context,
	triggers []*models.Trigger,
	now time.Time,
	maxRunsPerLoop int,
) {
	fired := 0
	for _, t := range triggers {
		if fired >= maxRunsPerLoop {
			break
		}

		times, err := GetCatchUpTimes(t, now)
		if err != nil {
			logger.Errorf(ctx, "scheduler: catchup error for trigger %s: %v", TriggerKey(t), err)
			continue
		}

		for _, scheduledAt := range times {
			if fired >= maxRunsPerLoop {
				break
			}
			if err := s.executor.Execute(ctx, t, scheduledAt.UTC()); err != nil {
				logger.Errorf(ctx, "scheduler: catchup execute failed for %s at %s: %v",
					TriggerKey(t), scheduledAt.Format(time.RFC3339), err)
			}
			fired++
		}
	}
}
