package errors

import (
	"errors"
	"testing"

	mockScope "github.com/flyteorg/flytestdlib/promutils"

	flyteAdminError "github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/lib/pq"
	"github.com/magiconair/properties/assert"
	"google.golang.org/grpc/codes"
)

func TestToFlyteAdminError_InvalidPqError(t *testing.T) {
	err := errors.New("foo")
	transformedErr := NewPostgresErrorTransformer(mockScope.NewTestScope()).ToFlyteAdminError(err)
	assert.Equal(t, codes.Internal, transformedErr.(flyteAdminError.FlyteAdminError).Code())
	assert.Equal(t, "unexpected error type for: foo", transformedErr.(flyteAdminError.FlyteAdminError).Error())
}

func TestToFlyteAdminError_UniqueConstraintViolation(t *testing.T) {
	err := &pq.Error{
		Code:       "23505",
		Constraint: "constraint",
		Message:    "message",
	}
	transformedErr := NewPostgresErrorTransformer(mockScope.NewTestScope()).ToFlyteAdminError(err)
	assert.Equal(t, codes.AlreadyExists, transformedErr.(flyteAdminError.FlyteAdminError).Code())
	assert.Equal(t, "value with matching constraint already exists (message)",
		transformedErr.(flyteAdminError.FlyteAdminError).Error())
}

func TestToFlyteAdminError_UnrecognizedPostgresError(t *testing.T) {
	err := &pq.Error{
		Code:    "foo",
		Message: "message",
	}
	transformedErr := NewPostgresErrorTransformer(mockScope.NewTestScope()).ToFlyteAdminError(err)
	assert.Equal(t, codes.Unknown, transformedErr.(flyteAdminError.FlyteAdminError).Code())
	assert.Equal(t, "failed database operation with message",
		transformedErr.(flyteAdminError.FlyteAdminError).Error())
}
