package util

import (
	"github.com/lyft/flyteadmin/pkg/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Transforms errors to grpc-compatible error types.
func TransformAndRecordError(err error, metrics *RequestMetrics) error {
	switch err.(type) {
	case errors.FlyteAdminError:
		metrics.Record(err.(errors.FlyteAdminError).Code())
		return err
	default:
		metrics.Record(codes.Internal)
		return status.Error(codes.Internal, err.Error())
	}
}
