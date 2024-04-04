package postgres

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	cacheErr "github.com/flyteorg/flyte/cacheservice/pkg/errors"
	errors2 "github.com/flyteorg/flyte/cacheservice/pkg/repositories/errors"
	"github.com/flyteorg/flyte/cacheservice/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/cacheservice/pkg/repositories/models"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

var (
	_ interfaces.ReservationRepo = &reservationRepo{}
)

type reservationRepo struct {
	db               *gorm.DB
	errorTransformer errors2.ErrorTransformer
	repoMetrics      gormMetrics
}

func (r reservationRepo) Create(ctx context.Context, reservation *models.CacheReservation, now time.Time) error {
	timer := r.repoMetrics.CreateReservationDuration.Start(ctx)
	defer timer.Stop()

	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&reservation)
	if result.Error != nil {
		return r.errorTransformer.ToCacheServiceError(result.Error)
	}

	if result.RowsAffected == 0 {
		return cacheErr.NewCacheServiceErrorf(codes.AlreadyExists, "reservation with key %s already exists", reservation.Key)
	}

	return nil
}

func (r reservationRepo) Update(ctx context.Context, reservation *models.CacheReservation, now time.Time) error {
	timer := r.repoMetrics.UpdateReservationDuration.Start(ctx)
	defer timer.Stop()

	result := r.db.WithContext(ctx).Model(&models.CacheReservation{
		Key: reservation.Key,
	}).Where("expires_at<=? OR owner_id=?", now, reservation.OwnerID).Updates(reservation)
	if result.Error != nil {
		return r.errorTransformer.ToCacheServiceError(result.Error)
	}

	if result.RowsAffected == 0 {
		return cacheErr.NewCacheServiceErrorf(codes.AlreadyExists, "reservation with key %s already exists", reservation.Key)
	}

	return nil
}

func (r reservationRepo) Get(ctx context.Context, key string) (*models.CacheReservation, error) {
	timer := r.repoMetrics.GetDuration.Start(ctx)
	defer timer.Stop()

	var reservation models.CacheReservation
	result := r.db.WithContext(ctx).Where(&models.CacheReservation{
		Key: key,
	}).Take(&reservation)

	if result.Error != nil {
		return &reservation, r.errorTransformer.ToCacheServiceError(result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, cacheErr.NewNotFoundError("reservation", key)
	}

	return &reservation, nil
}

func (r reservationRepo) Delete(ctx context.Context, key string, ownerID string) error {
	timer := r.repoMetrics.DeleteReservationDuration.Start(ctx)
	defer timer.Stop()

	var reservation models.CacheReservation
	result := r.db.WithContext(ctx).Where(&models.CacheReservation{
		Key:     key,
		OwnerID: ownerID,
	}).Delete(&reservation)
	if result.Error != nil {
		return r.errorTransformer.ToCacheServiceError(result.Error)
	}

	if result.RowsAffected == 0 {
		return cacheErr.NewNotFoundError("reservation", key)
	}

	return nil
}

func NewReservationRepo(db *gorm.DB, scope promutils.Scope) interfaces.ReservationRepo {
	return &reservationRepo{
		db:               db,
		errorTransformer: errors2.NewPostgresErrorTransformer(),
		repoMetrics:      newPostgresRepoMetrics(scope.NewSubScope("reservation")),
	}
}
