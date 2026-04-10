package errors

import (
	"database/sql"
	stderrors "errors"
)

var ErrReservationNotClaimable = stderrors.New("reservation not claimable")

// ErrAlreadyExists is returned when an insert violates a unique constraint.
var ErrAlreadyExists = stderrors.New("already exists")

func IsNotFound(err error) bool {
	return stderrors.Is(err, sql.ErrNoRows)
}

func IsAlreadyExists(err error) bool {
	return stderrors.Is(err, ErrAlreadyExists)
}

func IsReservationNotClaimable(err error) bool {
	return stderrors.Is(err, ErrReservationNotClaimable)
}
