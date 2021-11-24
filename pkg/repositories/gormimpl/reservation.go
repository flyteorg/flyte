package gormimpl

import (
	"context"

	datacatalog_error "github.com/flyteorg/datacatalog/pkg/errors"
	"google.golang.org/grpc/codes"

	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"

	errors2 "github.com/flyteorg/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/interfaces"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flytestdlib/promutils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type reservationRepo struct {
	db               *gorm.DB
	repoMetrics      gormMetrics
	errorTransformer errors2.ErrorTransformer
}

// NewReservationRepo creates a reservationRepo
func NewReservationRepo(db *gorm.DB, errorTransformer errors2.ErrorTransformer, scope promutils.Scope) interfaces.ReservationRepo {
	return &reservationRepo{
		db:               db,
		errorTransformer: errorTransformer,
		repoMetrics:      newGormMetrics(scope),
	}
}

func (r *reservationRepo) Create(ctx context.Context, reservation models.Reservation, now time.Time) error {
	timer := r.repoMetrics.CreateDuration.Start(ctx)
	defer timer.Stop()

	result := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&reservation)
	if result.Error != nil {
		return r.errorTransformer.ToDataCatalogError(result.Error)
	}

	if result.RowsAffected == 0 {
		return datacatalog_error.NewDataCatalogError(codes.FailedPrecondition, errors2.AlreadyExists)
	}

	return nil
}

func (r *reservationRepo) Delete(ctx context.Context, reservationKey models.ReservationKey, ownerID string) error {
	timer := r.repoMetrics.DeleteDuration.Start(ctx)
	defer timer.Stop()

	var reservation models.Reservation

	result := r.db.Where(&models.Reservation{
		ReservationKey: reservationKey,
		OwnerID:        ownerID,
	}).Delete(&reservation)
	if result.Error != nil {
		return r.errorTransformer.ToDataCatalogError(result.Error)
	}

	if result.RowsAffected == 0 {
		return errors2.GetMissingEntityError("Reservation",
			&datacatalog.ReservationID{
				DatasetId: &datacatalog.DatasetID{
					Project: reservationKey.DatasetProject,
					Domain:  reservationKey.DatasetDomain,
					Name:    reservationKey.DatasetName,
					Version: reservationKey.DatasetVersion,
				},
				TagName: reservationKey.TagName,
			})
	}

	return nil
}

func (r *reservationRepo) Get(ctx context.Context, reservationKey models.ReservationKey) (models.Reservation, error) {
	timer := r.repoMetrics.GetDuration.Start(ctx)
	defer timer.Stop()

	var reservation models.Reservation

	result := r.db.Where(&models.Reservation{
		ReservationKey: reservationKey,
	}).Take(&reservation)

	if result.Error != nil {
		return reservation, r.errorTransformer.ToDataCatalogError(result.Error)
	}

	return reservation, nil
}

func (r *reservationRepo) Update(ctx context.Context, reservation models.Reservation, now time.Time) error {
	timer := r.repoMetrics.UpdateDuration.Start(ctx)
	defer timer.Stop()

	result := r.db.Model(&models.Reservation{
		ReservationKey: reservation.ReservationKey,
	}).Where("expires_at<=? OR owner_id=?", now, reservation.OwnerID).Updates(reservation)
	if result.Error != nil {
		return r.errorTransformer.ToDataCatalogError(result.Error)
	}

	if result.RowsAffected == 0 {
		return datacatalog_error.NewDataCatalogError(codes.FailedPrecondition, errors2.AlreadyExists)
	}

	return nil
}
