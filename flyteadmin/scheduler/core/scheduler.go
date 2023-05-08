package core

import (
	"context"
	"time"

	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flyteadmin/scheduler/snapshoter"
)

type TimedFuncWithSchedule func(ctx context.Context, s models.SchedulableEntity, t time.Time) error

// Scheduler is the main scheduler interfaces for scheduling/descheduling jobs, updating the schedules,
// calculating snapshot of the schedules , bootstrapping the scheduler from the snapshot as well as the catcup functionality
type Scheduler interface {
	// ScheduleJob allows to schedule a job using the implemented scheduler
	ScheduleJob(ctx context.Context, s models.SchedulableEntity, f TimedFuncWithSchedule, lastExecTime *time.Time) error
	// DeScheduleJob allows to remove a scheduled job using the implemented scheduler
	DeScheduleJob(ctx context.Context, s models.SchedulableEntity)
	// BootStrapSchedulesFromSnapShot allows to initialize the scheduler from a previous snapshot of the schedule executions
	BootStrapSchedulesFromSnapShot(ctx context.Context, schedules []models.SchedulableEntity, snapshot snapshoter.Snapshot)
	// UpdateSchedules updates all the schedules in the schedulers job store
	UpdateSchedules(ctx context.Context, s []models.SchedulableEntity)
	// CalculateSnapshot creates a snapshot of the existing state of the schedules run by the scheduler which can be used in case of failure.
	CalculateSnapshot(ctx context.Context) snapshoter.Snapshot
	// CatchupAll catches up all the schedules in the schedulers job store to the until time
	CatchupAll(ctx context.Context, until time.Time) bool
}
