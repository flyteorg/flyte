package interceptors

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// NewErrorInterceptor returns a server interceptor that converts ordinary Go
// errors into Connect errors with an internal code and an empty message so
// implementation details aren't exposed to clients. Existing Connect errors,
// including wrapped ones, are returned unchanged.
func NewErrorInterceptor() connect.Interceptor {
	return errorInterceptor{}
}

type errorInterceptor struct{}

func (errorInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		response, err := next(ctx, req)
		return response, asConnectError(ctx, err)
	}
}

func (errorInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (errorInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return asConnectError(ctx, next(ctx, conn))
	}
}

func asConnectError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		logger.Warnf(ctx, "RPC returned error: %v", err)
		return err
	}

	logger.Errorf(ctx, "RPC returned unexpected error: %v", err)
	// Return an empty error as to not leak implementation details
	return connect.NewError(connect.CodeInternal, nil)
}
