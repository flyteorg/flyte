package gormimpl

import (
	"context"
	adminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"gorm.io/gorm"
)

type OverrideAttributesRepo struct {
	db               *gorm.DB
	errorTransformer adminErrors.ErrorTransformer
	metrics          gormMetrics
}

func (r *OverrideAttributesRepo) GetActive(ctx context.Context) (models.OverrideAttributes, error) {
	var overrideAttributes models.OverrideAttributes
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Where(&models.OverrideAttributes{
		Active: true,
	}).First(&overrideAttributes)
	timer.Stop()
	// TODO: what if the table is empty?
	if tx.Error != nil {
		return models.OverrideAttributes{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return overrideAttributes, nil
}

func (r *OverrideAttributesRepo) EraseActive(ctx context.Context) error {
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Model(&models.OverrideAttributes{}).Where(&models.OverrideAttributes{
		Active: true,
	}).Update("active", false)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (r *OverrideAttributesRepo) Create(ctx context.Context, input models.OverrideAttributes) error {
	timer := r.metrics.CreateDuration.Start()
	tx := r.db.Create(&input)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

// Returns an instance of OverrideAttributesRepoInterface
func NewOverrideAttributesRepo(
	db *gorm.DB, errorTransformer adminErrors.ErrorTransformer, scope promutils.Scope) interfaces.OverrideAttributesInterface {
	metrics := newMetrics(scope)
	return &OverrideAttributesRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
