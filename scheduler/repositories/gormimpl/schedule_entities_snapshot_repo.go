package gormimpl

import (
	"context"
	"errors"

	flyteSchedulerDbErrors "github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	interfaces2 "github.com/flyteorg/flyteadmin/scheduler/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flytestdlib/promutils"
	"gorm.io/gorm"
)

// ScheduleEntitiesSnapshotRepo Implementation of ScheduleEntitiesSnapshotRepoInterface.
type ScheduleEntitiesSnapshotRepo struct {
	db               *gorm.DB
	errorTransformer flyteSchedulerDbErrors.ErrorTransformer
	metrics          gormMetrics
}

// TODO : always overwrite the exisiting snapshot instead of creating new rows
func (r *ScheduleEntitiesSnapshotRepo) Write(ctx context.Context, input models.ScheduleEntitiesSnapshot) error {
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Omit("id").Create(&input)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (r *ScheduleEntitiesSnapshotRepo) Read(ctx context.Context) (models.ScheduleEntitiesSnapshot, error) {
	var schedulableEntitiesSnapshot models.ScheduleEntitiesSnapshot
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Last(&schedulableEntitiesSnapshot)
	timer.Stop()

	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return models.ScheduleEntitiesSnapshot{}, flyteSchedulerDbErrors.GetSingletonMissingEntityError("schedule_entities_snapshots")
		}
		return models.ScheduleEntitiesSnapshot{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return schedulableEntitiesSnapshot, nil
}

// NewScheduleEntitiesSnapshotRepo Returns an instance of ScheduleEntitiesSnapshotRepoInterface
func NewScheduleEntitiesSnapshotRepo(
	db *gorm.DB, errorTransformer flyteSchedulerDbErrors.ErrorTransformer, scope promutils.Scope) interfaces2.ScheduleEntitiesSnapShotRepoInterface {
	metrics := newMetrics(scope)
	return &ScheduleEntitiesSnapshotRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
