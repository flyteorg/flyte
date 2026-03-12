package impl

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/flyteorg/flyte/v2/cache_service/repository/interfaces"
	"github.com/flyteorg/flyte/v2/cache_service/repository/models"
)

var _ interfaces.ReservationRepo = (*ReservationRepo)(nil)

type ReservationRepo struct {
	db *gorm.DB
}

func NewReservationRepo(db *gorm.DB) *ReservationRepo {
	return &ReservationRepo{db: db}
}

func (r *ReservationRepo) Get(ctx context.Context, key string) (*models.Reservation, error) {
	var reservation models.Reservation
	if err := r.db.WithContext(ctx).Where("key = ?", key).Take(&reservation).Error; err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (r *ReservationRepo) Create(ctx context.Context, reservation *models.Reservation) error {
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(reservation)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrDuplicatedKey
	}
	return nil
}

func (r *ReservationRepo) UpdateIfExpiredOrOwned(ctx context.Context, reservation *models.Reservation, now time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&models.Reservation{}).
		Where("key = ? AND (expires_at <= ? OR owner_id = ?)", reservation.Key, now, reservation.OwnerID).
		Updates(map[string]any{
			"owner_id":          reservation.OwnerID,
			"heartbeat_seconds": reservation.HeartbeatSeconds,
			"expires_at":        reservation.ExpiresAt,
			"updated_at":        now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrDuplicatedKey
	}
	return nil
}

func (r *ReservationRepo) DeleteByKeyAndOwner(ctx context.Context, key, ownerID string) error {
	result := r.db.WithContext(ctx).Delete(&models.Reservation{}, "key = ? AND owner_id = ?", key, ownerID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
