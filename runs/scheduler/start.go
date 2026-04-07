package scheduler

import (
	"context"
	"time"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/runs/config"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/scheduler/core"
	"github.com/flyteorg/flyte/v2/runs/scheduler/executor"
)

// Start wires together the trigger scheduler and returns a worker function
// that blocks until ctx is cancelled. The caller should register it with
// sc.AddWorker("trigger-scheduler", ...).
func Start(
	ctx context.Context,
	triggerRepo interfaces.TriggerRepo,
	cfg config.TriggerSchedulerConfig,
	baseURL string,
) func(ctx context.Context) error {
	exec := executor.NewTriggerExecutor(executor.TriggerExecutorConfig{
		BaseURL: baseURL,
		QPS:     cfg.ExecutionQPS,
		Burst:   cfg.ExecutionBurst,
	})

	sched := core.NewGoCronScheduler(exec)
	sched.Start()

	syncer := core.NewScheduleSyncer(triggerRepo, sched, cfg.ResyncInterval)

	return func(ctx context.Context) error {
		logger.Infof(ctx, "trigger-scheduler: starting")

		// Bootstrap: load all active triggers and catch up missed runs before
		// starting the cron scheduler's steady-state loop.
		triggers, err := core.ListActiveScheduleTriggers(ctx, triggerRepo)
		if err != nil {
			return err
		}
		sched.UpdateSchedules(ctx, triggers)
		sched.CatchupAll(ctx, triggers, time.Now().UTC(), cfg.MaxCatchupRunsPerLoop)

		// Steady-state: periodic resync (UpdateSchedules only, no catchup).
		err = syncer.Run(ctx)

		// Stop the cron scheduler and wait for any in-flight jobs to finish.
		stopCtx := sched.Stop()
		<-stopCtx.Done()
		logger.Infof(ctx, "trigger-scheduler: stopped")
		return err
	}
}
