package gormimpl

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flytestdlib/promutils"

	"gorm.io/gorm"
)

type ExecutionEventRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	metrics          gormMetrics
}

func (r *ExecutionEventRepo) Create(ctx context.Context, input models.ExecutionEvent) error {
	timer := r.metrics.CreateDuration.Start()
	tx := r.db.Omit("id").Create(&input)
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
