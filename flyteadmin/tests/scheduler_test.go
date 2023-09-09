//go:build integration
// +build integration

package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/time/rate"

	adminModels "github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	scheduler "github.com/flyteorg/flyteadmin/scheduler/core"
	"github.com/flyteorg/flyteadmin/scheduler/executor/mocks"
	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flyteadmin/scheduler/snapshoter"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/promutils"
)

func TestScheduleJob(t *testing.T) {
	ctx := context.Background()
	True := true
	now := time.Now()
	scheduleFixed := models.SchedulableEntity{
		BaseModel: adminModels.BaseModel{
			UpdatedAt: now,
		},
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: "project",
			Domain:  "domain",
			Name:    "fixed1",
			Version: "version1",
		},
		FixedRateValue: 1,
		Unit:           admin.FixedRateUnit_MINUTE,
		Active:         &True,
	}

	c := cron.New()
	configuration := runtime.NewConfigurationProvider()
	applicationConfiguration := configuration.ApplicationConfiguration().GetTopLevelConfig()
	schedulerScope := promutils.NewScope(applicationConfiguration.MetricsScope).NewSubScope("schedule_job")
	rateLimiter := rate.NewLimiter(1, 10)
	executor := new(mocks.Executor)
	snapshot := &snapshoter.SnapshotV1{}
	executor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)
	g := scheduler.NewGoCronScheduler(context.Background(), []models.SchedulableEntity{}, schedulerScope, snapshot, rateLimiter, executor, false)
	c.Start()

	tests := []struct {
		testName      string
		lastExecTime  *time.Time
		assertionFunc func(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool
	}{
		{testName: "without_schedule_time", lastExecTime: nil},
		{testName: "using_schedule_time", lastExecTime: &now},
	}
	wg := &sync.WaitGroup{}
	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			wg.Add(1)
			timedFuncWithSchedule := func(jobCtx context.Context, schedule models.SchedulableEntity, scheduleTime time.Time) error {
				if scheduleTime.IsZero() {
					assert.NotEqual(t, now, scheduleTime)
				} else {
					assert.WithinDuration(t, now, scheduleTime, time.Minute*2)
				}
				wg.Done()
				return nil
			}
			err := g.ScheduleJob(ctx, scheduleFixed, timedFuncWithSchedule, tc.lastExecTime)
			assert.NoError(t, err)
		})
	}

	select {
	case <-time.After(time.Minute * 2):
		assert.Fail(t, "timed job didn't get triggered")
	case <-wait(wg):
		c.Stop()
		break
	}
}

func wait(wg *sync.WaitGroup) chan bool {
	ch := make(chan bool)
	go func() {
		wg.Wait()
		ch <- true
	}()
	return ch
}
