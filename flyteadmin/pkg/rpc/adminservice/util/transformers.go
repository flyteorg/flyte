package util

import (
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"

	"google.golang.org/grpc/codes"
)

// Transforms errors to grpc-compatible error types and optionally truncates it if necessary.
func TransformAndRecordError(err error, metrics *RequestMetrics) error {
	var errorMessage = err.Error()
	concatenateErrMessage := false
	if len(errorMessage) > common.MaxResponseStatusBytes {
		errorMessage = err.Error()[:common.MaxResponseStatusBytes]
		concatenateErrMessage = true
	}
	if flyteAdminError, ok := err.(errors.FlyteAdminError); !ok {
		err = errors.NewFlyteAdminError(codes.Internal, errorMessage)
	} else if concatenateErrMessage {
		err = errors.NewFlyteAdminError(flyteAdminError.Code(), errorMessage)
	}
	metrics.Record(err.(errors.FlyteAdminError).Code())
	return err
}
