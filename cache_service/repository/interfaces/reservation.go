package interfaces

import (
	"context"
	"time"

	"github.com/flyteorg/flyte/v2/cache_service/repository/models"
)

type ReservationRepo interface {
	Get(ctx context.Context, key string) (*models.Reservation, error)
	Create(ctx context.Context, reservation *models.Reservation) error
	UpdateIfExpiredOrOwned(ctx context.Context, reservation *models.Reservation, now time.Time) error
	DeleteByKeyAndOwner(ctx context.Context, key, ownerID string) error
}
