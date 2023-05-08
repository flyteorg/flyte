//go:build !race
// +build !race

package core

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/time/rate"

	adminModels "github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flyteadmin/scheduler/executor/mocks"
	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flyteadmin/scheduler/snapshoter"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/promutils"
)

var scheduleCron models.SchedulableEntity
var scheduleFixed models.SchedulableEntity
var scheduleCronDeactivated models.SchedulableEntity
var scheduleFixedDeactivated models.SchedulableEntity
var scheduleNonExistentDeActivated models.SchedulableEntity

func setup(t *testing.T, subscope string, useUtcTz bool) *GoCronScheduler {
	var schedules []models.SchedulableEntity
	True := true
	False := false
	scheduleCron = models.SchedulableEntity{
		BaseModel: adminModels.BaseModel{
			UpdatedAt: time.Now(),
		},
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: "project",
			Domain:  "domain",
			Name:    "cron1",
			Version: "version1",
		},
		CronExpression: "0 19 * * *",
		Active:         &True,
	}
	scheduleFixed = models.SchedulableEntity{
		BaseModel: adminModels.BaseModel{
			UpdatedAt: time.Now(),
		},
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: "project",
			Domain:  "domain",
			Name:    "fixed1",
			Version: "version1",
		},
		FixedRateValue: 1,
		Unit:           admin.FixedRateUnit_HOUR,
		Active:         &True,
	}
	scheduleCronDeactivated = models.SchedulableEntity{
		BaseModel: adminModels.BaseModel{
			UpdatedAt: time.Now(),
		},
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: "project",
			Domain:  "domain",
			Name:    "cron1",
			Version: "version1",
		},
		CronExpression: "0 19 * * *",
		Active:         &False,
	}
	scheduleFixedDeactivated = models.SchedulableEntity{
		BaseModel: adminModels.BaseModel{
			UpdatedAt: time.Now(),
		},
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: "project",
			Domain:  "domain",
			Name:    "fixed1",
			Version: "version1",
		},
		FixedRateValue: 1,
		Unit:           admin.FixedRateUnit_HOUR,
		Active:         &False,
	}
	scheduleNonExistentDeActivated = models.SchedulableEntity{
		BaseModel: adminModels.BaseModel{
			UpdatedAt: time.Now(),
		},
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: "project",
			Domain:  "domain",
			Name:    "cron3",
			Version: "version3",
		},
		CronExpression: "0 11 * * *",
		Active:         &False,
	}
	schedules = append(schedules, scheduleCron)
	schedules = append(schedules, scheduleFixed)
	schedules = append(schedules, scheduleCronDeactivated)
	schedules = append(schedules, scheduleFixedDeactivated)
	schedules = append(schedules, scheduleNonExistentDeActivated)
	return setupWithSchedules(t, subscope, schedules, useUtcTz)
}

func setupWithSchedules(t *testing.T, subscope string, schedules []models.SchedulableEntity, useUtcTz bool) *GoCronScheduler {
	configuration := runtime.NewConfigurationProvider()
	applicationConfiguration := configuration.ApplicationConfiguration().GetTopLevelConfig()
	schedulerScope := promutils.NewScope(applicationConfiguration.MetricsScope).NewSubScope(subscope)
	rateLimiter := rate.NewLimiter(1, 10)
	executor := new(mocks.Executor)
	snapshot := &snapshoter.SnapshotV1{}
	executor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)
	g := NewGoCronScheduler(context.Background(), schedules, schedulerScope, snapshot, rateLimiter, executor, useUtcTz)
	goCronScheduler, ok := g.(*GoCronScheduler)
	goCronScheduler.UpdateSchedules(context.Background(), schedules)
	assert.True(t, ok)
	goCronScheduler.BootStrapSchedulesFromSnapShot(context.Background(), schedules, snapshot)
	goCronScheduler.CatchupAll(context.Background(), time.Now())
	return goCronScheduler
}

func TestUseUTCTz(t *testing.T) {
	t.Run("use local timezone", func(t *testing.T) {
		g := setup(t, "use_local_tz", false)
		loc := g.cron.Location()
		assert.NotNil(t, loc)
		assert.Equal(t, time.Local, loc)
	})
	t.Run("use utc timezone", func(t *testing.T) {
		g := setup(t, "use_utc_tz", true)
		loc := g.cron.Location()
		assert.NotNil(t, loc)
		assert.Equal(t, time.UTC, loc)
	})
}
func TestCalculateSnapshot(t *testing.T) {
	t.Run("empty snapshot", func(t *testing.T) {
		ctx := context.Background()
		g := setupWithSchedules(t, "empty_snapshot", nil, false)
		snapshot := g.CalculateSnapshot(ctx)
		assert.NotNil(t, snapshot)
		assert.True(t, snapshot.IsEmpty())
	})
	t.Run("non empty snapshot", func(t *testing.T) {
		ctx := context.Background()
		g := setup(t, "non_empty_snapshot", false)
		g.jobStore.Range(func(key, value interface{}) bool {
			currTime := time.Now()
			job := value.(*GoCronJob)
			job.lastExecTime = &currTime
			return true
		})
		snapshot := g.CalculateSnapshot(ctx)
		assert.NotNil(t, snapshot)
		assert.False(t, snapshot.IsEmpty())
	})
}

func TestGetTimedFuncWithSchedule(t *testing.T) {
	type test struct {
		input models.SchedulableEntity
		scope string
		want  error
	}

	t.Run("happy case", func(t *testing.T) {
		tests := []test{
			{input: scheduleCron, scope: "happy_case_cron", want: nil},
			{input: scheduleFixed, scope: "happy_case_fixed", want: nil},
		}
		for _, tc := range tests {
			ctx := context.Background()
			g := setup(t, tc.scope, false)
			timeFunc := g.GetTimedFuncWithSchedule()
			assert.NotNil(t, timeFunc)
			err := timeFunc(ctx, tc.input, time.Now())
			assert.Equal(t, tc.want, err)
		}
	})
	t.Run("failure case", func(t *testing.T) {
		ctx := context.Background()
		g := setup(t, "failure_case", false)
		executor := new(mocks.Executor)
		executor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failure case"))
		g.executor = executor
		timeFunc := g.GetTimedFuncWithSchedule()
		assert.NotNil(t, timeFunc)
		err := timeFunc(ctx, scheduleCron, time.Now())
		assert.NotNil(t, err)
		assert.Equal(t, "failure case", err.Error())
	})
}

func TestGetCronScheduledTime(t *testing.T) {
	fromTime := time.Date(2022, time.January, 27, 19, 0, 0, 0, time.UTC)
	nextTime, err := getCronScheduledTime("0 19 * * *", fromTime)
	assert.Nil(t, err)
	expectedNextTime := time.Date(2022, time.January, 28, 19, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedNextTime, nextTime)
}

func TestGetCatchUpTimes(t *testing.T) {
	t.Run("to time before scheduled time", func(t *testing.T) {
		s := models.SchedulableEntity{
			CronExpression: "0 19 * * *",
		}
		from := time.Date(2022, time.January, 27, 19, 0, 0, 0, time.UTC)
		to := time.Date(2022, time.January, 28, 00, 51, 6, 0, time.UTC)
		catchupTimes, err := GetCatchUpTimes(s, from, to)
		assert.Nil(t, err)
		assert.Nil(t, catchupTimes)
	})
	t.Run("fixed rate catch up times", func(t *testing.T) {
		s := models.SchedulableEntity{
			FixedRateValue: 1,
			Unit:           admin.FixedRateUnit_HOUR,
		}
		from := time.Date(2022, time.January, 27, 19, 0, 0, 0, time.UTC)
		to := time.Date(2022, time.January, 28, 19, 0, 0, 0, time.UTC)
		catchupTimes, err := GetCatchUpTimes(s, from, to)
		assert.Nil(t, err)
		assert.NotNil(t, catchupTimes)
		assert.Equal(t, 24, len(catchupTimes))
	})
	t.Run("to time equal scheduled time", func(t *testing.T) {
		s := models.SchedulableEntity{
			CronExpression: "0 19 * * *",
		}
		from := time.Date(2022, time.January, 27, 19, 0, 0, 0, time.UTC)
		to := time.Date(2022, time.January, 28, 19, 0, 0, 0, time.UTC)
		catchupTimes, err := GetCatchUpTimes(s, from, to)
		assert.Nil(t, err)
		assert.NotNil(t, catchupTimes)
		assert.Equal(t, 1, len(catchupTimes))
		expectedNextTime := time.Date(2022, time.January, 28, 19, 0, 0, 0, time.UTC)
		assert.Equal(t, expectedNextTime, catchupTimes[0])
	})
	t.Run("to time after scheduled time", func(t *testing.T) {
		s := models.SchedulableEntity{
			CronExpression: "0 19 * * *",
		}
		from := time.Date(2022, time.January, 27, 19, 0, 0, 0, time.UTC)
		to := time.Date(2022, time.January, 28, 19, 50, 0, 0, time.UTC)
		catchupTimes, err := GetCatchUpTimes(s, from, to)
		assert.Nil(t, err)
		assert.NotNil(t, catchupTimes)
		assert.Equal(t, 1, len(catchupTimes))
		expectedNextTime := time.Date(2022, time.January, 28, 19, 0, 0, 0, time.UTC)
		assert.Equal(t, expectedNextTime, catchupTimes[0])
	})
	t.Run("invalid cron", func(t *testing.T) {
		s := models.SchedulableEntity{
			CronExpression: "0 19 * *",
		}
		from := time.Date(2022, time.January, 27, 19, 0, 0, 0, time.UTC)
		to := time.Date(2022, time.January, 28, 00, 51, 6, 0, time.UTC)
		_, err := GetCatchUpTimes(s, from, to)
		assert.NotNil(t, err)
	})
}

func TestGetFixedRateDurationFromSchedule(t *testing.T) {
	t.Run("minute duration", func(t *testing.T) {
		d := time.Duration(1)
		duration, err := getFixedRateDurationFromSchedule(admin.FixedRateUnit_MINUTE, 1)
		assert.Nil(t, err)
		assert.Equal(t, d*time.Minute, duration)
	})
	t.Run("hour duration", func(t *testing.T) {
		d := time.Duration(1)
		duration, err := getFixedRateDurationFromSchedule(admin.FixedRateUnit_HOUR, 1)
		assert.Nil(t, err)
		assert.Equal(t, d*time.Hour, duration)
	})
	t.Run("day duration", func(t *testing.T) {
		d := time.Duration(1)
		duration, err := getFixedRateDurationFromSchedule(admin.FixedRateUnit_DAY, 1)
		assert.Nil(t, err)
		assert.Equal(t, d*time.Hour*24, duration)
	})
	t.Run("invalid", func(t *testing.T) {
		_, err := getFixedRateDurationFromSchedule(100, 1)
		assert.NotNil(t, err)
	})
}

func TestCatchUpAllSchedule(t *testing.T) {
	ctx := context.Background()
	g := setup(t, "catch_up_all_schedules", false)
	toTime := time.Date(2022, time.January, 29, 0, 0, 0, 0, time.UTC)
	catchupSuccess := g.CatchupAll(ctx, toTime)
	assert.True(t, catchupSuccess)
}
