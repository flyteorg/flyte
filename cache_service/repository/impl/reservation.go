package impl

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	repositoryerrors "github.com/flyteorg/flyte/v2/cache_service/repository/errors"
	"github.com/flyteorg/flyte/v2/cache_service/repository/interfaces"
	"github.com/flyteorg/flyte/v2/cache_service/repository/models"
)

var _ interfaces.ReservationRepo = (*ReservationRepo)(nil)

type ReservationRepo struct {
	db *sqlx.DB
}

func NewReservationRepo(db *sqlx.DB) *ReservationRepo {
	return &ReservationRepo{db: db}
}

func (r *ReservationRepo) Get(ctx context.Context, key string) (*models.Reservation, error) {
	var reservation models.Reservation
	err := sqlx.GetContext(ctx, r.db, &reservation,
		"SELECT * FROM cache_service_reservations WHERE key = $1", key)
	if err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (r *ReservationRepo) Create(ctx context.Context, reservation *models.Reservation) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO cache_service_reservations (key, owner_id, heartbeat_seconds, expires_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		 ON CONFLICT (key) DO NOTHING`,
		reservation.Key, reservation.OwnerID, reservation.HeartbeatSeconds, reservation.ExpiresAt)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return repositoryerrors.ErrAlreadyExists
	}
	return nil
}

func (r *ReservationRepo) UpdateIfExpiredOrOwned(ctx context.Context, reservation *models.Reservation, now time.Time) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE cache_service_reservations
		 SET owner_id = $1, heartbeat_seconds = $2, expires_at = $3, updated_at = $4
		 WHERE key = $5 AND (expires_at <= $6 OR owner_id = $7)`,
		reservation.OwnerID, reservation.HeartbeatSeconds, reservation.ExpiresAt, now,
		reservation.Key, now, reservation.OwnerID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return repositoryerrors.ErrReservationNotClaimable
	}
	return nil
}

func (r *ReservationRepo) DeleteByKeyAndOwner(ctx context.Context, key, ownerID string) error {
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM cache_service_reservations WHERE key = $1 AND owner_id = $2",
		key, ownerID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
