package gormimpl

import (
	"context"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type ExecutionEventRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	metrics          gormMetrics
}

func (r *ExecutionEventRepo) Create(ctx context.Context, input models.ExecutionEvent) error {
	if err := util.FilterResourceMutation(ctx, input.Project, input.Domain); err != nil {
		return err
	}
	timer := r.metrics.CreateDuration.Start()
	tx := r.db.WithContext(ctx).Omit("id").Create(&input)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

// Returns an instance of ExecutionRepoInterface
func NewExecutionEventRepo(
	db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) interfaces.ExecutionEventRepoInterface {
	metrics := newMetrics(scope)
	return &ExecutionEventRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
