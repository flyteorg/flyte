package interfaces

import (
	"context"
	"time"

	"github.com/flyteorg/flyte/cacheservice/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/cacheservice"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

//go:generate mockery -all -output=../mocks -case=underscore

type CacheManager interface {
	Get(ctx context.Context, request *cacheservice.GetCacheRequest) (*cacheservice.GetCacheResponse, error)
	Put(ctx context.Context, request *cacheservice.PutCacheRequest) (*cacheservice.PutCacheResponse, error)
	Delete(ctx context.Context, request *cacheservice.DeleteCacheRequest) (*cacheservice.DeleteCacheResponse, error)
	GetOrExtendReservation(ctx context.Context, request *cacheservice.GetOrExtendReservationRequest, now time.Time) (*cacheservice.GetOrExtendReservationResponse, error)
	ReleaseReservation(ctx context.Context, request *cacheservice.ReleaseReservationRequest) (*cacheservice.ReleaseReservationResponse, error)
}

type CacheOutputBlobStore interface {
	Create(ctx context.Context, key string, output *core.LiteralMap) (string, error)
	Delete(ctx context.Context, uri string) error
}

type CacheDataStoreClient interface {
	Get(ctx context.Context, key string) (*models.CachedOutput, error)
	Put(ctx context.Context, key string, cachedOutput *models.CachedOutput) error
	Delete(ctx context.Context, key string) error
}

type ReservationDataStoreClient interface {
	Create(ctx context.Context, reservation *models.Reservation, now time.Time) error
	Update(ctx context.Context, reservation *models.Reservation, now time.Time) error
	Get(ctx context.Context, key string) (*models.Reservation, error)
	Delete(ctx context.Context, key string, ownerID string) error
}
