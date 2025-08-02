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

	adminModels "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flyte/flyteadmin/scheduler/executor/mocks"
	"github.com/flyteorg/flyte/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/scheduler/snapshoter"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
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
	executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(nil)
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
		executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failure case"))
		g.executor = executor
		timeFunc := g.GetTimedFuncWithSchedule()
		assert.NotNil(t, timeFunc)
		err := timeFunc(ctx, scheduleCron, time.Now())
		assert.NotNil(t, err)
		assert.Equal(t, "failure case", err.Error())
	})
}

func TestGetCronScheduledTime(t *testing.T) {
	tests := []struct {
		name           string
		cronExpression string
		fromTime       time.Time
		expectError    bool
		errorContains  string
		expectedTime   time.Time
	}{
		{
			name:           "valid cron expression",
			cronExpression: "0 19 * * *",
			fromTime:       time.Date(2022, time.January, 27, 19, 0, 0, 0, time.UTC),
			expectError:    false,
			expectedTime:   time.Date(2022, time.January, 28, 19, 0, 0, 0, time.UTC),
		},
		{
			name:           "invalid cron expression with February 31st should return error",
			cronExpression: "0 0 31 2 *", // February 31st - invalid
			fromTime:       time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			expectError:    true,
			errorContains:  "invalid crontab configuration",
		},
		{
			name:           "invalid cron expression with April 31st should return error",
			cronExpression: "0 0 31 4 *", // April 31st - invalid
			fromTime:       time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			expectError:    true,
			errorContains:  "invalid crontab configuration",
		},
		{
			name:           "invalid cron expression with malformed syntax should return error",
			cronExpression: "invalid cron expression",
			fromTime:       time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextTime, err := getCronScheduledTime(tt.cronExpression, tt.fromTime)

			if tt.expectError {
				assert.NotNil(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.True(t, nextTime.IsZero())
				t.Logf("Got expected error: %v", err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedTime, nextTime)
			}
		})
	}
}

func TestGetCatchUpTimesWithInvalidCronExpression(t *testing.T) {
	tests := []struct {
		name           string
		cronExpression string
		fromTime       time.Time
		toTime         time.Time
		expectError    bool
		errorContains  string
		expectCatchup  bool
	}{
		{
			name:           "invalid cron expression with February 31st should return error",
			cronExpression: "0 0 31 2 *", // February 31st - invalid
			fromTime:       time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			toTime:         time.Date(2023, time.December, 31, 23, 59, 59, 0, time.UTC),
			expectError:    true,
			errorContains:  "invalid crontab configuration",
			expectCatchup:  false,
		},
		{
			name:           "invalid cron expression with April 31st should return error",
			cronExpression: "0 0 31 4 *", // April 31st - invalid
			fromTime:       time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			toTime:         time.Date(2023, time.December, 31, 23, 59, 59, 0, time.UTC),
			expectError:    true,
			errorContains:  "invalid crontab configuration",
			expectCatchup:  false,
		},
		{
			name:           "valid cron expression should work normally",
			cronExpression: "0 0 15 2 *", // February 15th - valid
			fromTime:       time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			toTime:         time.Date(2023, time.December, 31, 23, 59, 59, 0, time.UTC),
			expectError:    false,
			expectCatchup:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := models.SchedulableEntity{
				CronExpression: tt.cronExpression,
			}

			catchupTimes, err := GetCatchUpTimes(s, tt.fromTime, tt.toTime)

			if tt.expectError {
				assert.NotNil(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, catchupTimes)
				t.Logf("Got expected error: %v", err)
			} else {
				assert.Nil(t, err)
				if tt.expectCatchup {
					assert.True(t, len(catchupTimes) > 0, "Should get valid catchup times")
					t.Logf("Got %d valid catchup times", len(catchupTimes))
				} else {
					assert.Nil(t, catchupTimes)
				}
			}
		})
	}
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

func TestGoCronScheduler_BootStrapSchedulesFromSnapShot(t *testing.T) {
	g := setupWithSchedules(t, "testing", []models.SchedulableEntity{}, true)
	True := true
	False := false
	scheduleActive1 := models.SchedulableEntity{
		BaseModel: adminModels.BaseModel{
			UpdatedAt: time.Date(1000, time.October, 19, 10, 0, 0, 0, time.UTC),
		},
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: "project",
			Domain:  "domain",
			Name:    "schedule_active_1",
			Version: "version1",
		},
		CronExpression: "0 19 * * *",
		Active:         &True,
	}
	scheduleActive2 := models.SchedulableEntity{
		BaseModel: adminModels.BaseModel{
			UpdatedAt: time.Date(2000, time.November, 19, 10, 0, 0, 0, time.UTC),
		},
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: "project",
			Domain:  "domain",
			Name:    "schedule_active_2",
			Version: "version1",
		},
		CronExpression: "0 19 * * *",
		Active:         &True,
	}
	scheduleInactive := models.SchedulableEntity{
		BaseModel: adminModels.BaseModel{
			UpdatedAt: time.Date(3000, time.December, 19, 10, 0, 0, 0, time.UTC),
		},
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: "project",
			Domain:  "domain",
			Name:    "cron3",
			Version: "version1",
		},
		CronExpression: "0 19 * * *",
		Active:         &False,
	}

	schedule1SnapshotTime := time.Date(5000, time.December, 19, 10, 0, 0, 0, time.UTC)
	schedule2SnapshotTime := time.Date(6000, time.December, 19, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name                 string
		schedules            []models.SchedulableEntity
		snapshoter           snapshoter.Snapshot
		expectedCatchUpTimes map[string]*time.Time
	}{
		{
			name:                 "two active",
			schedules:            []models.SchedulableEntity{scheduleActive1, scheduleActive2},
			snapshoter:           &snapshoter.SnapshotV1{},
			expectedCatchUpTimes: map[string]*time.Time{"11407394263542327059": &scheduleActive1.UpdatedAt, "1420107156943834850": &scheduleActive2.UpdatedAt},
		},
		{
			name:                 "two active one inactive",
			schedules:            []models.SchedulableEntity{scheduleActive1, scheduleActive2, scheduleInactive},
			snapshoter:           &snapshoter.SnapshotV1{},
			expectedCatchUpTimes: map[string]*time.Time{"11407394263542327059": &scheduleActive1.UpdatedAt, "1420107156943834850": &scheduleActive2.UpdatedAt},
		},
		{
			name:      "two active one inactive with snapshot populated",
			schedules: []models.SchedulableEntity{scheduleActive1, scheduleActive2, scheduleInactive},
			snapshoter: &snapshoter.SnapshotV1{
				LastTimes: map[string]*time.Time{
					"11407394263542327059": &schedule1SnapshotTime,
					"1420107156943834850":  &schedule2SnapshotTime,
				},
			},
			expectedCatchUpTimes: map[string]*time.Time{"11407394263542327059": &schedule1SnapshotTime, "1420107156943834850": &schedule2SnapshotTime},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g.BootStrapSchedulesFromSnapShot(context.Background(), tt.schedules, tt.snapshoter)
			g.jobStore.Range(func(key, value interface{}) bool {
				jobID := key.(string)
				job := value.(*GoCronJob)
				if !*job.schedule.Active {
					return true
				}
				assert.Equal(t, job.catchupFromTime, tt.expectedCatchUpTimes[jobID])
				return true
			})
			for _, schedule := range tt.schedules {
				g.DeScheduleJob(context.TODO(), schedule)
			}
		})
	}
}
