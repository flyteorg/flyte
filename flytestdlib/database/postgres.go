package database

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const PqInvalidDBCode = "3D000"
const PqDbAlreadyExistsCode = "42P04"
const PgDuplicatedKey = "23505"

func IsPgErrorWithCode(err error, code string) bool {
	pgErr := &pgconn.PgError{}
	if errors.As(err, &pgErr) {
		// err chain does not contain a pgconn.PgError
		return pgErr.Code == code
	}

	return false
}
