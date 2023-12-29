package shared

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type GrpcRegistrationHook func(ctx context.Context, server *grpc.Server, scope promutils.Scope) error
type HttpRegistrationHook func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption, scope promutils.Scope) error
