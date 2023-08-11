package util

import (
	"context"
	"errors"
	"strings"
	"testing"

	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyteadmin/pkg/common"
	adminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
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

func TestTransformError_GRPCError(t *testing.T) {
	err := status.Error(codes.InvalidArgument, strings.Repeat("X", common.MaxResponseStatusBytes+1))

	transformedError := TransformAndRecordError(err, &testRequestMetrics)

	transormerStatus, ok := status.FromError(transformedError)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, transormerStatus.Code())
	assert.Len(t, transormerStatus.Message(), common.MaxResponseStatusBytes)
}

func TestTransformError_BasicError(t *testing.T) {
	err := errors.New("some error")

	transformedError := TransformAndRecordError(err, &testRequestMetrics)

	transormerStatus, ok := status.FromError(transformedError)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, transormerStatus.Code())
}

func TestTruncateErrorMessage(t *testing.T) {
	err := adminErrors.NewFlyteAdminError(codes.InvalidArgument, strings.Repeat("X", common.MaxResponseStatusBytes+1))

	transformedError := TransformAndRecordError(err, &testRequestMetrics)

	transormerStatus, ok := status.FromError(transformedError)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, transormerStatus.Code())
	assert.Len(t, transormerStatus.Message(), common.MaxResponseStatusBytes)
}
