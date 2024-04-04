package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	cacheErr "github.com/flyteorg/flyte/cacheservice/pkg/errors"
	errors2 "github.com/flyteorg/flyte/cacheservice/pkg/repositories/errors"
	"github.com/flyteorg/flyte/cacheservice/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/cacheservice/pkg/repositories/models"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

var (
	_ interfaces.CachedOutputRepo = &postgresCachedOutputRepo{}
)

type postgresCachedOutputRepo struct {
	db               *gorm.DB
	errorTransformer errors2.ErrorTransformer
	repoMetrics      gormMetrics
}

func (c postgresCachedOutputRepo) Get(ctx context.Context, key string) (*models.CachedOutput, error) {
	timer := c.repoMetrics.GetDuration.Start(ctx)
	defer timer.Stop()

	var cachedOutput models.CachedOutput
	result := c.db.WithContext(ctx).Where(&models.CachedOutput{
		BaseModel: models.BaseModel{
			ID: key,
		},
	}).Take(&cachedOutput)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, cacheErr.NewNotFoundError("output", key)
		}
		return nil, result.Error
	}
	return &cachedOutput, nil
}

func (c postgresCachedOutputRepo) Put(ctx context.Context, key string, cachedOutput *models.CachedOutput) error {
	timer := c.repoMetrics.PutDuration.Start(ctx)
	defer timer.Stop()

	tx := c.db.WithContext(ctx).Begin()

	tx = tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&cachedOutput)

	if tx.Error != nil {
		tx.Rollback()
		return c.errorTransformer.ToCacheServiceError(tx.Error)
	}

	tx = tx.Commit()
	if tx.Error != nil {
		return c.errorTransformer.ToCacheServiceError(tx.Error)
	}

	return nil
}

func (c postgresCachedOutputRepo) Delete(ctx context.Context, key string) error {
	timer := c.repoMetrics.DeleteDuration.Start(ctx)
	defer timer.Stop()

	var cachedOutput models.CachedOutput

	result := c.db.WithContext(ctx).Where(&models.CachedOutput{
		BaseModel: models.BaseModel{
			ID: key,
		},
	}).Delete(&cachedOutput)
	if result.Error != nil {
		return c.errorTransformer.ToCacheServiceError(result.Error)
	}

	if result.RowsAffected == 0 {
		return cacheErr.NewNotFoundError("output", key)
	}

	return nil
}

func NewPostgresOutputRepo(db *gorm.DB, scope promutils.Scope) interfaces.CachedOutputRepo {
	return &postgresCachedOutputRepo{
		db:               db,
		errorTransformer: errors2.NewPostgresErrorTransformer(),
		repoMetrics:      newPostgresRepoMetrics(scope.NewSubScope("cached_output")),
	}
}
