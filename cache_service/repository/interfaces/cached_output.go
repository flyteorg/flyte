package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/v2/cache_service/repository/models"
)

type CachedOutputRepo interface {
	Get(ctx context.Context, key string) (*models.CachedOutput, error)
	Put(ctx context.Context, output *models.CachedOutput) error
	Delete(ctx context.Context, key string) error
}
