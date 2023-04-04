package core

import (
	"context"
	"fmt"
	"runtime/pprof"
	"time"

	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/robfig/cron/v3"
)

// GoCronJob this provides a wrapper around the go cron libraries job function.
type GoCronJob struct {
	ctx              context.Context
	nameOfSchedule   string
	schedule         models.SchedulableEntity
	funcWithSchedule TimedFuncWithSchedule
	lastTime         *time.Time
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
	// Update the lastTime only if new trigger time t is after lastTime.
	if g.lastTime == nil || g.lastTime.Before(t) {
		g.lastTime = &t
	}
}
