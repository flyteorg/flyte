package core

import (
	"context"
	"fmt"
	"runtime/pprof"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/flyteorg/flyte/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// GoCronJob this provides a wrapper around the go cron libraries job function.
type GoCronJob struct {
	ctx              context.Context
	nameOfSchedule   string
	schedule         models.SchedulableEntity
	funcWithSchedule TimedFuncWithSchedule
	lastExecTime     *time.Time
	catchupFromTime  *time.Time
	entryID          cron.EntryID
}

func (g *GoCronJob) Run(t time.Time) {
	// Create job function label to be used for creating the child context
	jobFuncLabel := fmt.Sprintf("jobfunc-%v", g.nameOfSchedule)
	jobFuncCtxWithLabel := contextutils.WithGoroutineLabel(g.ctx, jobFuncLabel)
	// TODO : add panic counter metric

	pprof.SetGoroutineLabels(jobFuncCtxWithLabel)
	if err := g.funcWithSchedule(jobFuncCtxWithLabel, g.schedule, t); err != nil {
		logger.Errorf(jobFuncCtxWithLabel, "Got error while scheduling %v", err)
	}
	// Update the lastExecTime only if new trigger time t is after lastExecTime.
	if g.lastExecTime == nil || g.lastExecTime.Before(t) {
		g.lastExecTime = &t
	}
}
