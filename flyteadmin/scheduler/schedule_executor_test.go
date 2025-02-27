//go:build !race
// +build !race

package scheduler

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/mocks"
	adminModels "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	schedMocks "github.com/flyteorg/flyte/flyteadmin/scheduler/repositories/mocks"
	"github.com/flyteorg/flyte/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/scheduler/snapshoter"
	adminMocks "github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

var schedules []models.SchedulableEntity
var db repositoryInterfaces.Repository

func setupScheduleExecutor(t *testing.T, s string) ScheduledExecutor {
	db = mocks.NewMockRepository()
	var scope = promutils.NewScope(s)
	scheduleExecutorConfig := runtimeInterfaces.WorkflowExecutorConfig{
		FlyteWorkflowExecutorConfig: &runtimeInterfaces.FlyteWorkflowExecutorConfig{
			AdminRateLimit: &runtimeInterfaces.AdminRateLimit{
				Tps:   100,
				Burst: 10,
			},
		},
	}
	var bytesArray []byte
	f := bytes.NewBuffer(bytesArray)
	writer := snapshoter.VersionedSnapshot{}
	snapshot := &snapshoter.SnapshotV1{
		LastTimes: map[string]*time.Time{},
	}
	err := writer.WriteSnapshot(f, snapshot)
	assert.Nil(t, err)
	mockAdminClient := new(adminMocks.AdminServiceClient)
	snapshotRepo := db.ScheduleEntitiesSnapshotRepo().(*schedMocks.ScheduleEntitiesSnapShotRepoInterface)
	snapshotModel := models.ScheduleEntitiesSnapshot{
		BaseModel: adminModels.BaseModel{
			ID:        17,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Snapshot: f.Bytes(),
	}
	snapshotRepo.EXPECT().Read(mock.Anything).Return(snapshotModel, nil)
	snapshotRepo.EXPECT().Write(mock.Anything, mock.Anything).Return(nil)
	mockAdminClient.EXPECT().CreateExecution(context.Background(), mock.Anything).
		Return(&admin.ExecutionCreateResponse{}, nil)
	return NewScheduledExecutor(db, scheduleExecutorConfig,
		scope, mockAdminClient)
}

func TestSuccessfulSchedulerExec(t *testing.T) {
	t.Run("add cron schedule", func(t *testing.T) {
		scheduleExecutor := setupScheduleExecutor(t, "cron")
		scheduleEntitiesRepo := db.SchedulableEntityRepo().(*schedMocks.SchedulableEntityRepoInterface)
		activeV2 := true
		createAt := time.Now()
		schedules = append(schedules, models.SchedulableEntity{
			BaseModel: adminModels.BaseModel{
				ID:        1,
				CreatedAt: createAt,
				UpdatedAt: time.Now(),
			},
			SchedulableEntityKey: models.SchedulableEntityKey{
				Project: "project",
				Domain:  "domain",
				Name:    "cron_schedule",
				Version: "v2",
			},
			CronExpression:      "@every 1s",
			KickoffTimeInputArg: "kickoff_time",
			Active:              &activeV2,
		})

		scheduleEntitiesRepo.EXPECT().GetAll(mock.Anything).Return(schedules, nil)
		go func() {
			err := scheduleExecutor.Run(context.Background())
			assert.Nil(t, err)
		}()
		time.Sleep(10 * time.Second)
		scheduleEntitiesRepo = db.SchedulableEntityRepo().(*schedMocks.SchedulableEntityRepoInterface)
		activeV2 = false
		schedules = nil
		schedules = append(schedules, models.SchedulableEntity{
			BaseModel: adminModels.BaseModel{
				ID:        1,
				CreatedAt: createAt,
				UpdatedAt: time.Now(),
			},
			SchedulableEntityKey: models.SchedulableEntityKey{
				Project: "project",
				Domain:  "domain",
				Name:    "cron_schedule",
				Version: "v2",
			},
			CronExpression:      "@every 1s",
			KickoffTimeInputArg: "kickoff_time",
			Active:              &activeV2,
		})
		scheduleEntitiesRepo.EXPECT().GetAll(mock.Anything).Return(schedules, nil)
		time.Sleep(30 * time.Second)
	})

	t.Run("add fixed rate schedule", func(t *testing.T) {
		scheduleExecutor := setupScheduleExecutor(t, "fixed")
		scheduleEntitiesRepo := db.SchedulableEntityRepo().(*schedMocks.SchedulableEntityRepoInterface)
		activeV2 := true
		createAt := time.Now()
		schedules = append(schedules, models.SchedulableEntity{
			BaseModel: adminModels.BaseModel{
				ID:        1,
				CreatedAt: createAt,
				UpdatedAt: time.Now(),
			},
			SchedulableEntityKey: models.SchedulableEntityKey{
				Project: "project",
				Domain:  "domain",
				Name:    "fixed_rate_schedule",
				Version: "v2",
			},
			FixedRateValue:      1,
			Unit:                admin.FixedRateUnit_MINUTE,
			KickoffTimeInputArg: "kickoff_time",
			Active:              &activeV2,
		})
		scheduleEntitiesRepo.EXPECT().GetAll(mock.Anything).Return(schedules, nil)

		go func() {
			err := scheduleExecutor.Run(context.Background())
			assert.Nil(t, err)
		}()
		time.Sleep(10 * time.Second)
		scheduleEntitiesRepo = db.SchedulableEntityRepo().(*schedMocks.SchedulableEntityRepoInterface)
		activeV2 = false
		schedules = nil
		schedules = append(schedules, models.SchedulableEntity{
			BaseModel: adminModels.BaseModel{
				ID:        1,
				CreatedAt: createAt,
				UpdatedAt: time.Now(),
			},
			SchedulableEntityKey: models.SchedulableEntityKey{
				Project: "project",
				Domain:  "domain",
				Name:    "fixed_rate_schedule",
				Version: "v2",
			},
			FixedRateValue:      1,
			Unit:                admin.FixedRateUnit_MINUTE,
			KickoffTimeInputArg: "kickoff_time",
			Active:              &activeV2,
		})
		scheduleEntitiesRepo.EXPECT().GetAll(mock.Anything).Return(schedules, nil)
		time.Sleep(30 * time.Second)
	})

	t.Run("unable to read schedules", func(t *testing.T) {
		scheduleExecutor := setupScheduleExecutor(t, "unable_read_schedules")
		scheduleEntitiesRepo := db.SchedulableEntityRepo().(*schedMocks.SchedulableEntityRepoInterface)
		scheduleEntitiesRepo.EXPECT().GetAll(mock.Anything).Return(nil, fmt.Errorf("unable to read schedules"))

		go func() {
			err := scheduleExecutor.Run(context.Background())
			assert.NotNil(t, err)
		}()
	})
}
