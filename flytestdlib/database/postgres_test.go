package database

import (
	"errors"
	"net"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

type wrappedError struct {
	err error
}

func (e *wrappedError) Error() string {
	return e.err.Error()
}

func (e *wrappedError) Unwrap() error {
	return e.err
}

func TestIsInvalidDBPgError(t *testing.T) {
	// wrap error with wrappedError when testing to ensure the function checks the whole error chain

	testCases := []struct {
		Name           string
		Err            error
		ExpectedResult bool
	}{
		{
			Name:           "nil error",
			Err:            nil,
			ExpectedResult: false,
		},
		{
			Name:           "not a PgError",
			Err:            &wrappedError{err: &net.OpError{Op: "connect", Err: errors.New("connection refused")}},
			ExpectedResult: false,
		},
		{
			Name:           "PgError but not invalid DB",
			Err:            &wrappedError{&pgconn.PgError{Severity: "FATAL", Message: "out of memory", Code: "53200"}},
			ExpectedResult: false,
		},
		{
			Name:           "PgError and is invalid DB",
			Err:            &wrappedError{&pgconn.PgError{Severity: "FATAL", Message: "database \"flyte\" does not exist", Code: "3D000"}},
			ExpectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.ExpectedResult, IsPgErrorWithCode(tc.Err, PqInvalidDBCode))
		})
	}
}

func TestIsPgDbAlreadyExistsError(t *testing.T) {
	// wrap error with wrappedError when testing to ensure the function checks the whole error chain

	testCases := []struct {
		Name           string
		Err            error
		ExpectedResult bool
	}{
		{
			Name:           "nil error",
			Err:            nil,
			ExpectedResult: false,
		},
		{
			Name:           "not a PgError",
			Err:            &wrappedError{err: &net.OpError{Op: "connect", Err: errors.New("connection refused")}},
			ExpectedResult: false,
		},
		{
			Name:           "PgError but not already exists",
			Err:            &wrappedError{&pgconn.PgError{Severity: "FATAL", Message: "out of memory", Code: "53200"}},
			ExpectedResult: false,
		},
		{
			Name:           "PgError and is already exists",
			Err:            &wrappedError{&pgconn.PgError{Severity: "FATAL", Message: "database \"flyte\" does not exist", Code: "42P04"}},
			ExpectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.ExpectedResult, IsPgErrorWithCode(tc.Err, PqDbAlreadyExistsCode))
		})
	}
}
