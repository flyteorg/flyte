package repository

import (
	"errors"

	"gorm.io/gorm"
)

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func IsAlreadyExists(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey)
}
