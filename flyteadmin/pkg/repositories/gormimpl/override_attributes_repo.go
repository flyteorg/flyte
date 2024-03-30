package gormimpl

import (
	"context"
	adminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"gorm.io/gorm"
)

type OverrideAttributesRepo struct {
	db               *gorm.DB
	errorTransformer adminErrors.ErrorTransformer
	metrics          gormMetrics
}

func (r *OverrideAttributesRepo) GetActive(ctx context.Context) (models.OverrideAttributes, error) {
	var overrideAttributes models.OverrideAttributes
	err := r.db.Where(&models.OverrideAttributes{
		Active: true,
	}).First(&overrideAttributes).Error
	if err != nil {
		return models.OverrideAttributes{}, r.errorTransformer.ToFlyteAdminError(err)
	}
	return overrideAttributes, nil
}
