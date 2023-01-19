package impl

import (
	"context"
	"time"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"

	"github.com/flyteorg/datacatalog/pkg/errors"
	"github.com/flyteorg/datacatalog/pkg/manager/impl/validators"
	"github.com/flyteorg/datacatalog/pkg/manager/interfaces"
	"github.com/flyteorg/datacatalog/pkg/repositories"
	repoerrors "github.com/flyteorg/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/datacatalog/pkg/repositories/transformers"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

type reservationMetrics struct {
	scope                        promutils.Scope
	reservationAcquired          labeled.Counter
	reservationReleased          labeled.Counter
	reservationAlreadyInProgress labeled.Counter
	acquireReservationFailure    labeled.Counter
	acquireReservationsFailure   labeled.Counter
	releaseReservationFailure    labeled.Counter
	releaseReservationsFailure   labeled.Counter
	reservationDoesNotExist      labeled.Counter
}

type NowFunc func() time.Time

type reservationManager struct {
	repo                           repositories.RepositoryInterface
	heartbeatGracePeriodMultiplier time.Duration
	maxHeartbeatInterval           time.Duration
	now                            NowFunc
	systemMetrics                  reservationMetrics
}

// NewReservationManager creates a new reservation manager with the specified properties
func NewReservationManager(
	repo repositories.RepositoryInterface,
	heartbeatGracePeriodMultiplier time.Duration,
	maxHeartbeatInterval time.Duration,
	nowFunc NowFunc, // Easier to mock time.Time for testing
	reservationScope promutils.Scope,
) interfaces.ReservationManager {
	systemMetrics := reservationMetrics{
		scope: reservationScope,
		reservationAcquired: labeled.NewCounter(
			"reservation_acquired",
			"Number of times a reservation was acquired",
			reservationScope),
		reservationReleased: labeled.NewCounter(
			"reservation_released",
			"Number of times a reservation was released",
			reservationScope),
		reservationAlreadyInProgress: labeled.NewCounter(
			"reservation_already_in_progress",
			"Number of times we try of acquire a reservation but the reservation is in progress",
			reservationScope,
		),
		acquireReservationFailure: labeled.NewCounter(
			"acquire_reservation_failure",
			"Number of times we failed to acquire reservation",
			reservationScope,
		),
		acquireReservationsFailure: labeled.NewCounter(
			"acquire_reservations_failure",
			"Number of times we failed to acquire multiple reservations",
			reservationScope,
		),
		releaseReservationFailure: labeled.NewCounter(
			"release_reservation_failure",
			"Number of times we failed to release a reservation",
			reservationScope,
		),
		releaseReservationsFailure: labeled.NewCounter(
			"release_reservations_failure",
			"Number of times we failed to release multiple reservations",
			reservationScope,
		),
		reservationDoesNotExist: labeled.NewCounter(
			"reservation_does_not_exist",
			"Number of times we attempt to modify a reservation that does not exist",
			reservationScope,
		),
	}

	return &reservationManager{
		repo:                           repo,
		heartbeatGracePeriodMultiplier: heartbeatGracePeriodMultiplier,
		maxHeartbeatInterval:           maxHeartbeatInterval,
		now:                            nowFunc,
		systemMetrics:                  systemMetrics,
	}
}

// GetOrExtendReservation attempts to acquire a reservation for the specified artifact. If there is no active
// reservation, successfully acquire it. If you are the owner of the active reservation, extend it. If another owner,
// return the existing reservation.
func (r *reservationManager) GetOrExtendReservation(ctx context.Context, request *datacatalog.GetOrExtendReservationRequest) (*datacatalog.GetOrExtendReservationResponse, error) {
	if err := validators.ValidateGetOrExtendReservationRequest(request); err != nil {
		r.systemMetrics.acquireReservationFailure.Inc(ctx)
		return nil, err
	}

	reservationID := request.ReservationId

	// Use minimum of maxHeartbeatInterval and requested heartbeat interval
	heartbeatInterval := r.maxHeartbeatInterval
	requestHeartbeatInterval := request.GetHeartbeatInterval()
	if requestHeartbeatInterval != nil && requestHeartbeatInterval.AsDuration() < heartbeatInterval {
		heartbeatInterval = requestHeartbeatInterval.AsDuration()
	}

	reservation, err := r.tryAcquireReservation(ctx, reservationID, request.OwnerId, heartbeatInterval)
	if err != nil {
		r.systemMetrics.acquireReservationFailure.Inc(ctx)
		return nil, err
	}

	return &datacatalog.GetOrExtendReservationResponse{
		Reservation: &reservation,
	}, nil
}

// GetOrExtendReservations attempts to get or extend reservations for multiple artifacts in a single operation.
func (r *reservationManager) GetOrExtendReservations(ctx context.Context, request *datacatalog.GetOrExtendReservationsRequest) (*datacatalog.GetOrExtendReservationsResponse, error) {
	if err := validators.ValidateGetOrExtendReservationsRequest(request); err != nil {
		r.systemMetrics.acquireReservationsFailure.Inc(ctx)
		return nil, err
	}

	var reservations []*datacatalog.Reservation
	for _, req := range request.GetReservations() {
		// Use minimum of maxHeartbeatInterval and requested heartbeat interval
		heartbeatInterval := r.maxHeartbeatInterval
		requestHeartbeatInterval := req.GetHeartbeatInterval()
		if requestHeartbeatInterval != nil && requestHeartbeatInterval.AsDuration() < heartbeatInterval {
			heartbeatInterval = requestHeartbeatInterval.AsDuration()
		}

		reservation, err := r.tryAcquireReservation(ctx, req.ReservationId, req.OwnerId, heartbeatInterval)
		if err != nil {
			r.systemMetrics.acquireReservationsFailure.Inc(ctx)
			return nil, err
		}

		reservations = append(reservations, &reservation)
	}

	return &datacatalog.GetOrExtendReservationsResponse{
		Reservations: reservations,
	}, nil
}

// tryAcquireReservation will fetch the reservation first and only create/update
// the reservation if it does not exist or has expired.
// This is an optimization to reduce the number of writes to db. We always need
// to do a GET here because we want to know who owns the reservation
// and show it to users on the UI. However, the reservation is held by a single
// task most of the times and there is no need to do a write.
func (r *reservationManager) tryAcquireReservation(ctx context.Context, reservationID *datacatalog.ReservationID, ownerID string, heartbeatInterval time.Duration) (datacatalog.Reservation, error) {
	repo := r.repo.ReservationRepo()
	reservationKey := transformers.FromReservationID(reservationID)
	repoReservation, err := repo.Get(ctx, reservationKey)

	reservationExists := true
	if err != nil {
		if errors.IsDoesNotExistError(err) {
			// Reservation does not exist yet so let's create one
			reservationExists = false
		} else {
			return datacatalog.Reservation{}, err
		}
	}

	now := r.now()
	newRepoReservation := models.Reservation{
		ReservationKey: reservationKey,
		OwnerID:        ownerID,
		ExpiresAt:      now.Add(heartbeatInterval * r.heartbeatGracePeriodMultiplier),
	}

	// Conditional upsert on reservation. Race conditions are handled
	// within the reservation repository Create and Update function calls.
	var repoErr error
	if !reservationExists {
		repoErr = repo.Create(ctx, newRepoReservation, now)
	} else if repoReservation.ExpiresAt.Before(now) || repoReservation.OwnerID == ownerID {
		repoErr = repo.Update(ctx, newRepoReservation, now)
	} else {
		logger.Debugf(ctx, "Reservation: %+v is held by %s", reservationKey, repoReservation.OwnerID)

		reservation, err := transformers.CreateReservation(&repoReservation, heartbeatInterval)
		if err != nil {
			return reservation, err
		}

		r.systemMetrics.reservationAlreadyInProgress.Inc(ctx)
		return reservation, nil
	}

	if repoErr != nil {
		if repoErr.Error() == repoerrors.AlreadyExists {
			// Looks like someone else tried to obtain the reservation
			// at the same time and they won. Let's find out who won.
			rsv1, err := repo.Get(ctx, reservationKey)
			if err != nil {
				return datacatalog.Reservation{}, err
			}

			reservation, err := transformers.CreateReservation(&rsv1, heartbeatInterval)
			if err != nil {
				return reservation, err
			}

			r.systemMetrics.reservationAlreadyInProgress.Inc(ctx)
			return reservation, nil
		}

		return datacatalog.Reservation{}, repoErr
	}

	// Reservation has been acquired or extended without error
	reservation, err := transformers.CreateReservation(&newRepoReservation, heartbeatInterval)
	if err != nil {
		return reservation, err
	}

	r.systemMetrics.reservationAlreadyInProgress.Inc(ctx)
	return reservation, nil
}

// ReleaseReservation releases an active reservation with the specified owner. If one does not exist, gracefully return.
func (r *reservationManager) ReleaseReservation(ctx context.Context, request *datacatalog.ReleaseReservationRequest) (*datacatalog.ReleaseReservationResponse, error) {
	if err := validators.ValidateReleaseReservationRequest(request); err != nil {
		return nil, err
	}

	if err := r.releaseReservation(ctx, request.ReservationId, request.OwnerId); err != nil {
		r.systemMetrics.releaseReservationFailure.Inc(ctx)
		return nil, err
	}

	return &datacatalog.ReleaseReservationResponse{}, nil
}

// ReleaseReservations releases reservations for multiple artifacts in a single operation.
// This is an idempotent operation, releasing reservations multiple times or trying to release an unknown reservation
// will not result in an error being returned.
func (r *reservationManager) ReleaseReservations(ctx context.Context, request *datacatalog.ReleaseReservationsRequest) (*datacatalog.ReleaseReservationResponse, error) {
	if err := validators.ValidateReleaseReservationsRequest(request); err != nil {
		return nil, err
	}

	for _, req := range request.GetReservations() {
		if err := r.releaseReservation(ctx, req.ReservationId, req.OwnerId); err != nil {
			r.systemMetrics.releaseReservationsFailure.Inc(ctx)
			return nil, err
		}
	}

	return &datacatalog.ReleaseReservationResponse{}, nil
}

// releaseReservation performs the actual reservation release operation, deleting the respective object from
// datacatalog's database, thus freeing the associated artifact for other entities. If the specified reservation was not
// found, no error will be returned.
func (r *reservationManager) releaseReservation(ctx context.Context, reservationID *datacatalog.ReservationID, ownerID string) error {
	repo := r.repo.ReservationRepo()
	reservationKey := transformers.FromReservationID(reservationID)

	err := repo.Delete(ctx, reservationKey, ownerID)
	if err != nil {
		if errors.IsDoesNotExistError(err) {
			logger.Warnf(ctx, "Reservation does not exist id: %+v, err %v", reservationID, err)
			r.systemMetrics.reservationDoesNotExist.Inc(ctx)
			return nil
		}

		logger.Errorf(ctx, "Failed to release reservation: %+v, err: %v", reservationKey, err)
		return err
	}

	r.systemMetrics.reservationReleased.Inc(ctx)
	return nil
}
