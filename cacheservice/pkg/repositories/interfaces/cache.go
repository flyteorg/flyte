package interfaces

import (
	"context"
	"time"

	"github.com/flyteorg/flyte/cacheservice/pkg/repositories/models"
)

//go:generate mockery -all -output=../mocks -case=underscore

type CachedOutputRepo interface {
	Get(ctx context.Context, key string) (*models.CachedOutput, error)
	Put(ctx context.Context, key string, cachedOutput *models.CachedOutput) error
	Delete(ctx context.Context, key string) error
}

type ReservationRepo interface {
	Create(ctx context.Context, reservation *models.CacheReservation, now time.Time) error
	Update(ctx context.Context, reservation *models.CacheReservation, now time.Time) error
	Get(ctx context.Context, key string) (*models.CacheReservation, error)
	Delete(ctx context.Context, key string, ownerID string) error
}

type RepositoryInterface interface {
	CachedOutputRepo() CachedOutputRepo
	ReservationRepo() ReservationRepo
}
