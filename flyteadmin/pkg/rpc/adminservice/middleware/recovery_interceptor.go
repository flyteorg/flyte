package middleware

import (
	"context"
	"runtime/debug"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

// RecoveryInterceptor is a struct for creating gRPC interceptors that handle panics in go
type RecoveryInterceptor struct {
	panicCounter prometheus.Counter
}

// NewRecoveryInterceptor creates a new RecoveryInterceptor with metrics under the provided scope
func NewRecoveryInterceptor(adminScope promutils.Scope) *RecoveryInterceptor {
	panicCounter := adminScope.MustNewCounter("handler_panic", "panics encountered while handling gRPC requests")
	return &RecoveryInterceptor{
		panicCounter: panicCounter,
	}
}

// UnaryServerInterceptor returns a new unary server interceptor for panic recovery.
func (ri *RecoveryInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ any, err error) {

		defer func() {
			if r := recover(); r != nil {
				ri.panicCounter.Inc()
				logger.Errorf(ctx, "panic-ed for request: [%+v] to %s with err: %v with Stack: %v", req, info.FullMethod, r, string(debug.Stack()))
				// Return INTERNAL to client with no info as to not leak implementation details
				err = status.Errorf(codes.Internal, "")
			}
		}()

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor for panic recovery.
func (ri *RecoveryInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {

		defer func() {
			if r := recover(); r != nil {
				ri.panicCounter.Inc()
				logger.Errorf(stream.Context(), "panic-ed for stream to %s with err: %v with Stack: %v", info.FullMethod, r, string(debug.Stack()))
				// Return INTERNAL to client with no info as to not leak implementation details
				err = status.Errorf(codes.Internal, "")
			}
		}()

		return handler(srv, stream)
	}
}
