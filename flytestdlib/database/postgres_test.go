package database

import (
	"context"
	"errors"
	"io/ioutil"
	"net"
	"os"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestResolvePassword(t *testing.T) {
	password := "123abc"
	tmpFile, err := os.CreateTemp("", "prefix")
	if err != nil {
		t.Errorf("Couldn't open temp file: %v", err)
	}
	defer tmpFile.Close()
	if _, err = tmpFile.WriteString(password); err != nil {
		t.Errorf("Couldn't write to temp file: %v", err)
	}
	resolvedPassword := resolvePassword(context.TODO(), "", tmpFile.Name())
	assert.Equal(t, resolvedPassword, password)
}

func TestGetPostgresDsn(t *testing.T) {
	pgConfig := PostgresConfig{
		Host:         "localhost",
		Port:         5432,
		DbName:       "postgres",
		User:         "postgres",
		ExtraOptions: "sslmode=disable",
	}
	t.Run("no password", func(t *testing.T) {
		dsn := getPostgresDsn(context.TODO(), pgConfig)
		assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres sslmode=disable", dsn)
	})
	t.Run("with password", func(t *testing.T) {
		pgConfig.Password = "pass"
		dsn := getPostgresDsn(context.TODO(), pgConfig)
		assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres password=pass sslmode=disable", dsn)

	})
	t.Run("with password, no extra", func(t *testing.T) {
		pgConfig.Password = "pass"
		pgConfig.ExtraOptions = ""
		dsn := getPostgresDsn(context.TODO(), pgConfig)
		assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres password=pass ", dsn)
	})
	t.Run("with password path", func(t *testing.T) {
		password := "123abc"
		tmpFile, err := ioutil.TempFile("", "prefix")
		if err != nil {
			t.Errorf("Couldn't open temp file: %v", err)
		}
		defer tmpFile.Close()
		if _, err = tmpFile.WriteString(password); err != nil {
			t.Errorf("Couldn't write to temp file: %v", err)
		}
		pgConfig.PasswordPath = tmpFile.Name()
		dsn := getPostgresDsn(context.TODO(), pgConfig)
		assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres password=123abc ", dsn)
	})
}

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
		tc := tc

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
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.ExpectedResult, IsPgErrorWithCode(tc.Err, PqDbAlreadyExistsCode))
		})
	}
}
