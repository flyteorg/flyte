package errors

import (
	stderrors "errors"

	"gorm.io/gorm"
)

var ErrReservationNotClaimable = stderrors.New("reservation not claimable")

func IsNotFound(err error) bool {
	return stderrors.Is(err, gorm.ErrRecordNotFound)
}

func IsAlreadyExists(err error) bool {
	return stderrors.Is(err, gorm.ErrDuplicatedKey)
}

func IsReservationNotClaimable(err error) bool {
	return stderrors.Is(err, ErrReservationNotClaimable)
}
