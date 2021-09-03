package repositories

import (
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/scheduler/repositories/gormimpl"
	"github.com/flyteorg/flyteadmin/scheduler/repositories/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/jinzhu/gorm"
)

type PostgresRepo struct {
	schedulableEntityRepo        interfaces.SchedulableEntityRepoInterface
	scheduleEntitiesSnapshotRepo interfaces.ScheduleEntitiesSnapShotRepoInterface
}

func (p *PostgresRepo) SchedulableEntityRepo() interfaces.SchedulableEntityRepoInterface {
	return p.schedulableEntityRepo
}

func (p *PostgresRepo) ScheduleEntitiesSnapshotRepo() interfaces.ScheduleEntitiesSnapShotRepoInterface {
	return p.scheduleEntitiesSnapshotRepo
}

func NewPostgresRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) SchedulerRepoInterface {
	return &PostgresRepo{
		schedulableEntityRepo:        gormimpl.NewSchedulableEntityRepo(db, errorTransformer, scope.NewSubScope("schedulable_entity")),
		scheduleEntitiesSnapshotRepo: gormimpl.NewScheduleEntitiesSnapshotRepo(db, errorTransformer, scope.NewSubScope("schedule_entities_snapshot")),
	}
}
