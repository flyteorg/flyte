package impl

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/flyteorg/flyte/v2/cache_service/repository/interfaces"
	"github.com/flyteorg/flyte/v2/cache_service/repository/models"
)

var _ interfaces.CachedOutputRepo = (*CachedOutputRepo)(nil)

type CachedOutputRepo struct {
	db *gorm.DB
}

func NewCachedOutputRepo(db *gorm.DB) *CachedOutputRepo {
	return &CachedOutputRepo{db: db}
}

func (r *CachedOutputRepo) Get(ctx context.Context, key string) (*models.CachedOutput, error) {
	var output models.CachedOutput
	if err := r.db.WithContext(ctx).Where("key = ?", key).Take(&output).Error; err != nil {
		return nil, err
	}
	return &output, nil
}

func (r *CachedOutputRepo) Put(ctx context.Context, output *models.CachedOutput) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(output).Error
}

func (r *CachedOutputRepo) Delete(ctx context.Context, key string) error {
	result := r.db.WithContext(ctx).Delete(&models.CachedOutput{}, "key = ?", key)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
