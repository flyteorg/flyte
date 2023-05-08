package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flyteorg/flyteadmin/scheduler/executor"
	"github.com/flyteorg/flyteadmin/scheduler/identifier"
	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flyteadmin/scheduler/snapshoter"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"golang.org/x/time/rate"
)

// goCronMetrics mertrics recorded for go cron.
type goCronMetrics struct {
	Scope                     promutils.Scope
	JobFuncPanicCounter       prometheus.Counter
	JobScheduledFailedCounter prometheus.Counter
	CatchupErrCounter         prometheus.Counter
}

// GoCronScheduler this provides a scheduler functionality using the https://github.com/robfig/cron library.
type GoCronScheduler struct {
	cron        *cron.Cron
	jobStore    sync.Map
	metrics     goCronMetrics
	rateLimiter *rate.Limiter
	executor    executor.Executor
	snapshot    snapshoter.Snapshot
}

// GetTimedFuncWithSchedule returns the job function with scheduled time parameter
func (g *GoCronScheduler) GetTimedFuncWithSchedule() TimedFuncWithSchedule {
	return func(jobCtx context.Context, schedule models.SchedulableEntity, scheduleTime time.Time) error {
		_ = g.rateLimiter.Wait(jobCtx)
		err := g.executor.Execute(jobCtx, scheduleTime, schedule)
		if err != nil {
			logger.Errorf(jobCtx, "unable to fire the schedule %+v at %v time due to %v", schedule, scheduleTime,
				err)
		}
		return err
	}
}

// BootStrapSchedulesFromSnapShot allows to initialize the scheduler from a previous snapshot of the schedule executions
func (g *GoCronScheduler) BootStrapSchedulesFromSnapShot(ctx context.Context, schedules []models.SchedulableEntity,
	snapshot snapshoter.Snapshot) {
	for _, s := range schedules {
		if *s.Active {
			funcRef := g.GetTimedFuncWithSchedule()
			nameOfSchedule := identifier.GetScheduleName(ctx, s)
			// Initialize the lastExectime as the updatedAt time
			// Assumption here that schedule was activated and that the 0th execution of the schedule
			// which will be used as a reference
			lastExecTime := &s.UpdatedAt

			fromSnapshot := snapshot.GetLastExecutionTime(nameOfSchedule)
			// Use the latest time if available in the snapshot
			if fromSnapshot != nil && fromSnapshot.After(s.UpdatedAt) {
				lastExecTime = fromSnapshot
			}
			err := g.ScheduleJob(ctx, s, funcRef, lastExecTime)
			if err != nil {
				g.metrics.JobScheduledFailedCounter.Inc()
				logger.Errorf(ctx, "unable to register the schedule %+v due to %v", s, err)
			}
		}
	}
}

// UpdateSchedules updates all the schedules in the schedulers job store
func (g *GoCronScheduler) UpdateSchedules(ctx context.Context, schedules []models.SchedulableEntity) {
	for _, s := range schedules {
		// Schedule or Deschedule job from the scheduler based on the activation status
		if !*s.Active {
			g.DeScheduleJob(ctx, s)
		} else {
			// Get the TimedFuncWithSchedule
			funcRef := g.GetTimedFuncWithSchedule()
			err := g.ScheduleJob(ctx, s, funcRef, nil)
			if err != nil {
				g.metrics.JobScheduledFailedCounter.Inc()
				logger.Errorf(ctx, "unable to register the schedule %+v due to %v", s, err)
			}
		}
	} // Done iterating over all the read schedules
}

// CalculateSnapshot creates a snapshot of the existing state of the schedules run by the scheduler which can be used in case of failure.
func (g *GoCronScheduler) CalculateSnapshot(ctx context.Context) snapshoter.Snapshot {
	snapshot := g.snapshot.Create()
	g.jobStore.Range(func(key, value interface{}) bool {
		job := value.(*GoCronJob)
		scheduleIdentifier := key.(string)
		if job.lastExecTime != nil {
			snapshot.UpdateLastExecutionTime(scheduleIdentifier, job.lastExecTime)
		}
		return true
	})
	return snapshot
}

// ScheduleJob allows to schedule a job using the implemented scheduler
func (g *GoCronScheduler) ScheduleJob(ctx context.Context, schedule models.SchedulableEntity,
	funcWithSchedule TimedFuncWithSchedule, lastExecTime *time.Time) error {

	nameOfSchedule := identifier.GetScheduleName(ctx, schedule)

	if _, ok := g.jobStore.Load(nameOfSchedule); ok {
		logger.Debugf(ctx, "Job already exists in the map for name %v with schedule %+v",
			nameOfSchedule, schedule)
		return nil
	}

	// Update the catchupFrom time as the lastExecTime.
	// Here lastExecTime is passed to this function only from BootStrapSchedulesFromSnapShot which is during bootup
	// Once initialized we wont be changing the catchupTime until the next boot
	job := &GoCronJob{nameOfSchedule: nameOfSchedule, schedule: schedule, funcWithSchedule: funcWithSchedule,
		catchupFromTime: lastExecTime, lastExecTime: lastExecTime, ctx: ctx}

	// Define the timed job function to be used for the callback at the scheduled time
	//jobFunc := job.GetTimedFunc(ctx, g.metrics)

	if len(job.schedule.CronExpression) > 0 {
		err := g.AddCronJob(ctx, job)
		if err != nil {
			logger.Errorf(ctx, "failed to add cron schedule %+v due to %v", schedule, err)
			return err
		}
	} else {
		err := g.AddFixedIntervalJob(ctx, job)
		if err != nil {
			logger.Errorf(ctx, "failed to add fixed rate schedule %+v due to %v", schedule, err)
			return err
		}
	}
	// Store only if there are no errors.
	g.jobStore.Store(nameOfSchedule, job)
	return nil
}

// DeScheduleJob allows to remove a scheduled job using the implemented scheduler
func (g *GoCronScheduler) DeScheduleJob(ctx context.Context, schedule models.SchedulableEntity) {
	nameOfSchedule := identifier.GetScheduleName(ctx, schedule)
	if _, ok := g.jobStore.Load(nameOfSchedule); !ok {
		logger.Debugf(ctx, "Job doesn't exists in the map for name %v with schedule %+v "+
			" and hence already removed", nameOfSchedule, schedule)
		return
	}
	val, _ := g.jobStore.Load(nameOfSchedule)
	jobWrapper := val.(*GoCronJob)

	s := jobWrapper.schedule
	if len(s.CronExpression) > 0 {
		g.RemoveCronJob(ctx, jobWrapper)
	} else {
		g.RemoveFixedIntervalJob(ctx, jobWrapper)
	}

	// Delete it from the job store
	g.jobStore.Delete(nameOfSchedule)
}

// CatchupAll catches up all the schedules from the jobStore to until time
func (g *GoCronScheduler) CatchupAll(ctx context.Context, until time.Time) bool {
	failed := false
	g.jobStore.Range(func(key, value interface{}) bool {
		job := value.(*GoCronJob)
		var fromTime *time.Time
		if !*job.schedule.Active {
			logger.Debugf(ctx, "schedule %+v was inactive during catchup", job.schedule)
			return true
		}

		fromTime = job.catchupFromTime
		if fromTime != nil {
			logger.Infof(ctx, "catching up schedule %+v from %v to %v", job.schedule, fromTime, until)
			err := g.CatchUpSingleSchedule(ctx, job.schedule, *fromTime, until)
			if err != nil {
				// stop the iteration since one of the catchups failed
				failed = true
				return false
			}
			logger.Infof(ctx, "caught up successfully on the schedule %+v from %v to %v", job.schedule, fromTime, until)
		}
		return true
	})
	return !failed
}

// CatchUpSingleSchedule catches up the schedule s from fromTime to toTime
func (g *GoCronScheduler) CatchUpSingleSchedule(ctx context.Context, s models.SchedulableEntity, fromTime time.Time, toTime time.Time) error {
	var catchUpTimes []time.Time
	var err error
	catchUpTimes, err = GetCatchUpTimes(s, fromTime, toTime)
	if err != nil {
		return err
	}
	var catchupTime time.Time
	for _, catchupTime = range catchUpTimes {
		_ = g.rateLimiter.Wait(ctx)
		err := g.executor.Execute(ctx, catchupTime, s)
		if err != nil {
			g.metrics.CatchupErrCounter.Inc()
			logger.Errorf(ctx, "unable to fire the schedule %+v at %v time due to %v", s, catchupTime, err)
			return err
		}
	}
	return nil
}

// GetCatchUpTimes find list of timestamps to be caught up on for schedule s from fromTime to toTime
func GetCatchUpTimes(s models.SchedulableEntity, from time.Time, to time.Time) ([]time.Time, error) {
	var scheduledTimes []time.Time
	currFrom := from
	for currFrom.Before(to) {
		scheduledTime, err := GetScheduledTime(s, currFrom)
		if err != nil {
			return nil, err
		}
		if scheduledTime.After(to) {
			break
		}
		scheduledTimes = append(scheduledTimes, scheduledTime)
		currFrom = scheduledTime
	}
	return scheduledTimes, nil
}

// GetScheduledTime find next schedule time for both cron and fixed rate scheduled entity given the fromTime
func GetScheduledTime(s models.SchedulableEntity, fromTime time.Time) (time.Time, error) {
	if len(s.CronExpression) > 0 {
		return getCronScheduledTime(s.CronExpression, fromTime)
	}
	return getFixedIntervalScheduledTime(s.Unit, s.FixedRateValue, fromTime)
}

func getCronScheduledTime(cronString string, fromTime time.Time) (time.Time, error) {
	sched, err := cron.ParseStandard(cronString)
	if err != nil {
		return time.Time{}, err
	}
	return sched.Next(fromTime), nil
}

func getFixedIntervalScheduledTime(unit admin.FixedRateUnit, fixedRateValue uint32, fromTime time.Time) (time.Time, error) {
	d, err := getFixedRateDurationFromSchedule(unit, fixedRateValue)
	if err != nil {
		return time.Time{}, err
	}
	fixedRateSchedule := cron.ConstantDelaySchedule{Delay: d}
	return fixedRateSchedule.Next(fromTime), nil
}

// AddFixedIntervalJob adds the fixes interval job to the job store.
func (g *GoCronScheduler) AddFixedIntervalJob(ctx context.Context, job *GoCronJob) error {
	d, err := getFixedRateDurationFromSchedule(job.schedule.Unit, job.schedule.FixedRateValue)
	if err != nil {
		return err
	}

	//nolint
	var jobFunc cron.TimedFuncJob
	jobFunc = job.Run

	var lastTime time.Time
	if job.lastExecTime != nil {
		lastTime = *job.lastExecTime
	}
	entryID := g.cron.ScheduleTimedJob(cron.ConstantDelaySchedule{Delay: d}, jobFunc, lastTime)
	// Update the enttry id in the job which is handle to be used for removal
	job.entryID = entryID
	logger.Infof(ctx, "successfully added the fixed rate schedule %s to the scheduler for schedule %+v",
		job.nameOfSchedule, job.schedule)

	return nil
}

// RemoveFixedIntervalJob removes the fixes interval job from the job store.
func (g *GoCronScheduler) RemoveFixedIntervalJob(ctx context.Context, job *GoCronJob) {
	g.cron.Remove(job.entryID)
	logger.Infof(ctx, "successfully removed the schedule %s from scheduler for schedule %+v",
		job.nameOfSchedule, job.schedule)
}

// AddCronJob adds the job from the cron store
func (g *GoCronScheduler) AddCronJob(ctx context.Context, job *GoCronJob) error {
	//nolint
	var jobFunc cron.TimedFuncJob
	jobFunc = job.Run

	entryID, err := g.cron.AddTimedJob(job.schedule.CronExpression, jobFunc)
	// Update the enttry id in the job which is handle to be used for removal
	job.entryID = entryID
	if err == nil {
		logger.Infof(ctx, "successfully added the schedule %s to the scheduler for schedule %+v",
			job.nameOfSchedule, job.schedule)
	}
	return err
}

// RemoveCronJob removes the job from the cron store
func (g *GoCronScheduler) RemoveCronJob(ctx context.Context, job *GoCronJob) {
	g.cron.Remove(job.entryID)
	logger.Infof(ctx, "successfully removed the schedule %s from scheduler for schedue %+v",
		job.nameOfSchedule, job.schedule)

}

func getFixedRateDurationFromSchedule(unit admin.FixedRateUnit, fixedRateValue uint32) (time.Duration, error) {
	d := time.Duration(fixedRateValue)
	switch unit {
	case admin.FixedRateUnit_MINUTE:
		d = d * time.Minute
	case admin.FixedRateUnit_HOUR:
		d = d * time.Hour
	case admin.FixedRateUnit_DAY:
		d = d * time.Hour * 24
	default:
		return -1, fmt.Errorf("unsupported unit %v for fixed rate scheduling ", unit)
	}
	return d, nil
}

func NewGoCronScheduler(ctx context.Context, schedules []models.SchedulableEntity, scope promutils.Scope,
	snapshot snapshoter.Snapshot, rateLimiter *rate.Limiter, executor executor.Executor, useUtcTz bool) Scheduler {
	// Create the new cron scheduler and start it off
	var opts []cron.Option
	if useUtcTz {
		opts = append(opts, cron.WithLocation(time.UTC))
	}
	c := cron.New(opts...)
	c.Start()
	scheduler := &GoCronScheduler{
		cron:        c,
		jobStore:    sync.Map{},
		metrics:     getCronMetrics(scope),
		rateLimiter: rateLimiter,
		executor:    executor,
		snapshot:    snapshot,
	}
	scheduler.BootStrapSchedulesFromSnapShot(ctx, schedules, snapshot)
	return scheduler
}

func getCronMetrics(scope promutils.Scope) goCronMetrics {
	return goCronMetrics{
		Scope: scope,
		JobFuncPanicCounter: scope.MustNewCounter("job_func_panic_counter",
			"count of crashes for the job functions executed by the scheduler"),
		JobScheduledFailedCounter: scope.MustNewCounter("job_schedule_failed_counter",
			"count of scheduling failures by the scheduler"),
		CatchupErrCounter: scope.MustNewCounter("catchup_error_counter",
			"count of unsuccessful attempts to catchup on the schedules"),
	}
}
