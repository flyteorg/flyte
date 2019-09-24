package util

import (
	"context"
	"errors"
	"testing"

	adminErrors "github.com/lyft/flyteadmin/pkg/errors"
	mockScope "github.com/lyft/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var testRequestMetrics = NewRequestMetrics(mockScope.NewTestScope(), "foo")

func TestTransformError_FlyteAdminError(t *testing.T) {
	invalidArgError := adminErrors.NewFlyteAdminError(codes.InvalidArgument, "invalid arg")
	transformedError := TransformAndRecordError(invalidArgError, &testRequestMetrics)
	transormerStatus, ok := status.FromError(transformedError)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, transormerStatus.Code())
}

func TestTransformError_FlyteAdminErrorWithDetails(t *testing.T) {
	terminalStateError := adminErrors.NewAlreadyInTerminalStateError(context.Background(), "terminal state", "curPhase")
	transformedError := TransformAndRecordError(terminalStateError, &testRequestMetrics)
	transormerStatus, ok := status.FromError(transformedError)
	assert.True(t, ok)
	assert.Equal(t, codes.FailedPrecondition, transormerStatus.Code())
	assert.Equal(t, 1, len(transormerStatus.Details()))
}

func TestTransformError_BasicError(t *testing.T) {
	err := errors.New("some error")
	transformedError := TransformAndRecordError(err, &testRequestMetrics)
	transormerStatus, ok := status.FromError(transformedError)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, transormerStatus.Code())
}
