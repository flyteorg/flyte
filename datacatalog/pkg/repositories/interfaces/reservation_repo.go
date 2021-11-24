package interfaces

import (
	"context"
	"time"

	"github.com/flyteorg/datacatalog/pkg/repositories/models"
)

// Interface to interact with Reservation Table
type ReservationRepo interface {

	// Create a new reservation if the reservation does not already exist
	Create(ctx context.Context, reservation models.Reservation, now time.Time) error

	// Delete a reservation if it exists
	Delete(ctx context.Context, reservation models.ReservationKey, ownerID string) error

	// Get reservation
	Get(ctx context.Context, reservationKey models.ReservationKey) (models.Reservation, error)

	// Update an existing reservation. If called by the current owner, we update the
	// expiresAt timestamp. If called by a new owner and the current reservation has
	// expired, we attempt to take over the reservation.
	Update(ctx context.Context, reservation models.Reservation, now time.Time) error
}
