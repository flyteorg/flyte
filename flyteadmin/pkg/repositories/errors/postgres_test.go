package errors

import (
	"errors"
	"testing"

	"github.com/jackc/pgconn"
	pgxPgconn "github.com/jackc/pgx/v5/pgconn"
	"github.com/magiconair/properties/assert"
	"google.golang.org/grpc/codes"

	mockScope "github.com/flyteorg/flyte/flytestdlib/promutils"
)

func TestToFlyteAdminError_InvalidPqError(t *testing.T) {
	err := errors.New("foo")
	transformedErr := NewPostgresErrorTransformer(mockScope.NewTestScope()).ToFlyteAdminError(err)
	assert.Equal(t, codes.Internal, transformedErr.Code())
	assert.Equal(t, "unexpected error type for: foo", transformedErr.Error())
}

func TestToFlyteAdminError_UniqueConstraintViolation(t *testing.T) {
	err := &pgconn.PgError{
		Code:    "23505",
		Message: "message",
	}
	transformedErr := NewPostgresErrorTransformer(mockScope.NewTestScope()).ToFlyteAdminError(err)
	assert.Equal(t, codes.AlreadyExists, transformedErr.Code())
	assert.Equal(t, "value with matching already exists (message)",
		transformedErr.Error())

	err2 := &pgxPgconn.PgError{
		Code:    "23505",
		Message: "message",
	}
	transformedErr = NewPostgresErrorTransformer(mockScope.NewTestScope()).ToFlyteAdminError(err2)
	assert.Equal(t, codes.AlreadyExists, transformedErr.Code())
	assert.Equal(t, "value with matching already exists (message)",
		transformedErr.Error())

}

func TestToFlyteAdminError_UnrecognizedPostgresError(t *testing.T) {
	err := &pgconn.PgError{
		Code:    "foo",
		Message: "message",
	}
	transformedErr := NewPostgresErrorTransformer(mockScope.NewTestScope()).ToFlyteAdminError(err)
	assert.Equal(t, codes.Unknown, transformedErr.Code())
	assert.Equal(t, "failed database operation with message",
		transformedErr.Error())

	err2 := &pgxPgconn.PgError{
		Code:    "foo",
		Message: "message",
	}
	transformedErr = NewPostgresErrorTransformer(mockScope.NewTestScope()).ToFlyteAdminError(err2)
	assert.Equal(t, codes.Unknown, transformedErr.Code())
	assert.Equal(t, "failed database operation with message",
		transformedErr.Error())
}
