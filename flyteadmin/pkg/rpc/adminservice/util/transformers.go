package util

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
)

// TransformAndRecordError transforms errors to grpc-compatible error types and optionally truncates it if necessary.
func TransformAndRecordError(err error, metrics *RequestMetrics) error {
	errMsg := err.Error()
	shouldTruncate := false
	if len(errMsg) > common.MaxResponseStatusBytes {
		errMsg = errMsg[:common.MaxResponseStatusBytes]
		shouldTruncate = true
	}

	adminErr, isAdminErr := err.(errors.FlyteAdminError)
	grpcStatus, isStatus := status.FromError(err)
	switch {
	case isAdminErr:
		if shouldTruncate {
			adminErr = errors.NewFlyteAdminError(adminErr.Code(), errMsg)
		}
	case isStatus:
		adminErr = errors.NewFlyteAdminError(grpcStatus.Code(), errMsg)
	default:
		adminErr = errors.NewFlyteAdminError(codes.Internal, errMsg)
	}

	metrics.Record(adminErr.Code())
	return adminErr
}
