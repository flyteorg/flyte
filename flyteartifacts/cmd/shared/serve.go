package shared

import (
	"context"
	"google.golang.org/grpc/reflection"
	"net"
	"net/http"

	sharedCfg "github.com/flyteorg/flyte/flyteartifacts/pkg/configuration/shared"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func NewServeCmd(commandName string, serverCfg sharedCfg.ServerConfiguration, grpcHook GrpcRegistrationHook, httpHook HttpRegistrationHook) *cobra.Command {
	// serveCmd represents the serve command
	return &cobra.Command{
		Use:   "serve",
		Short: "Launches the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			return ServeGateway(ctx, commandName, &serverCfg, grpcHook, httpHook)
		},
	}
}

// Creates a new gRPC Server with all the configuration
func newGRPCServer(ctx context.Context, serviceName string, serverCfg *sharedCfg.ServerConfiguration,
	grpcHook GrpcRegistrationHook, opts ...grpc.ServerOption) (*grpc.Server, error) {

	scope := promutils.NewScope(serverCfg.Metrics.MetricsScope)
	scope = scope.NewSubScope(serviceName)

	var grpcUnaryInterceptors = make([]grpc.UnaryServerInterceptor, 0)
	var streamServerInterceptors = make([]grpc.StreamServerInterceptor, 0)

	serverOpts := []grpc.ServerOption{
		grpc.StreamInterceptor(grpcPrometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpcMiddleware.ChainUnaryServer(
			grpcUnaryInterceptors...,
		)),
		grpc.ChainStreamInterceptor(
			streamServerInterceptors...,
		),
	}

	serverOpts = append(serverOpts, opts...)

	grpcServer := grpc.NewServer(serverOpts...)
	grpcPrometheus.Register(grpcServer)
	if grpcHook != nil {
		err := grpcHook(ctx, grpcServer, scope)
		if err != nil {
			return nil, err
		}
	}

	healthServer := health.NewServer()
	healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	if serverCfg.GrpcServerReflection {
		reflection.Register(grpcServer)
	}

	return grpcServer, nil
}

// ServeGateway launches the grpc and http servers.
func ServeGateway(ctx context.Context, serviceName string, serverCfg *sharedCfg.ServerConfiguration, grpcHook GrpcRegistrationHook, httpHook HttpRegistrationHook) error {

	if grpcHook != nil {
		if err := launchGrpcServer(ctx, serviceName, serverCfg, grpcHook); err != nil {
			return err
		}
	}

	if httpHook != nil {
		if err := launchHttpServer(ctx, serverCfg, httpHook); err != nil {
			return err
		}
	}

	for {
		<-ctx.Done()
		return nil
	}
}

// launchGrpcServer launches grpc server with server config and also registers the grpc hook for the service.
func launchGrpcServer(ctx context.Context, serviceName string, serverCfg *sharedCfg.ServerConfiguration, grpcHook GrpcRegistrationHook) error {
	var grpcServer *grpc.Server
	if serverCfg.Security.Secure {
		tlsCreds, err := credentials.NewServerTLSFromFile(serverCfg.Security.Ssl.CertificateFile, serverCfg.Security.Ssl.KeyFile)
		if err != nil {
			return err
		}

		grpcServer, err = newGRPCServer(ctx, serviceName, serverCfg, grpcHook, grpc.Creds(tlsCreds))
		if err != nil {
			return errors.Wrap(err, "failed to create secure GRPC server")
		}
	} else {
		var err error
		grpcServer, err = newGRPCServer(ctx, serviceName, serverCfg, grpcHook)
		if err != nil {
			return errors.Wrap(err, "failed to create GRPC server")
		}
	}

	return listenAndServe(ctx, grpcServer, serverCfg.GetGrpcHostAddress())
}

// listenAndServe on the grpcHost address and serve connections using the grpcServer
func listenAndServe(ctx context.Context, grpcServer *grpc.Server, grpcHost string) error {
	conn, err := net.Listen("tcp", grpcHost)
	if err != nil {
		panic(err)
	}

	go func() {
		if err := grpcServer.Serve(conn); err != nil {
			logger.Fatalf(ctx, "Failed to create GRPC Server, Err: ", err)
		}

		logger.Infof(ctx, "Serving GRPC Gateway server on: %v", grpcHost)
	}()
	return nil
}

// launchHttpServer launches an http server for converting REST calls to grpc internally
func launchHttpServer(ctx context.Context, cfg *sharedCfg.ServerConfiguration, httpHook HttpRegistrationHook) error {
	logger.Infof(ctx, "Starting HTTP/1 Gateway server on %s", cfg.GetHttpHostAddress())

	httpServer, serverMux, err := newHTTPServer(cfg.Security)
	if err != nil {
		return err
	}

	if err = registerHttpHook(ctx, serverMux, cfg.GetGrpcHostAddress(), uint32(cfg.GrpcMaxResponseStatusBytes), httpHook, promutils.NewScope(cfg.Metrics.MetricsScope)); err != nil {
		return err
	}

	handler := getHTTPHandler(httpServer, cfg.Security)

	go func() {
		err = http.ListenAndServe(cfg.GetHttpHostAddress(), handler)
		if err != nil {
			logger.Fatalf(ctx, "Failed to Start HTTP Server, Err: %v", err)
		}

		logger.Infof(ctx, "Serving HTTP/1 on: %v", cfg.GetHttpHostAddress())
	}()
	return nil
}

// getHTTPHandler gets the http handler for the configured security options
func getHTTPHandler(httpServer *http.ServeMux, _ sharedCfg.ServerSecurityOptions) http.Handler {
	// not really used yet (reserved for admin)
	var handler http.Handler = httpServer
	return handler
}

// registerHttpHook registers the http hook for a service which multiplexes to the grpc endpoint for that service.
func registerHttpHook(ctx context.Context, gwmux *runtime.ServeMux, grpcHost string, maxResponseStatusBytes uint32, httpHook HttpRegistrationHook, scope promutils.Scope) error {
	if httpHook == nil {
		return nil
	}
	grpcOptions := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithMaxHeaderListSize(maxResponseStatusBytes),
	}

	return httpHook(ctx, gwmux, grpcHost, grpcOptions, scope)
}

func healthCheckFunc(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// newHTTPServer creates a new http sever
func newHTTPServer(_ sharedCfg.ServerSecurityOptions) (*http.ServeMux, *runtime.ServeMux, error) {
	// Register the server that will serve HTTP/REST Traffic
	mux := http.NewServeMux()

	// Register healthcheck
	mux.HandleFunc("/healthcheck", healthCheckFunc)

	var gwmuxOptions = make([]runtime.ServeMuxOption, 0)
	// This option means that http requests are served with protobufs, instead of json. We always want this.
	gwmuxOptions = append(gwmuxOptions, runtime.WithMarshalerOption("application/octet-stream", &runtime.ProtoMarshaller{}))

	// Create the grpc-gateway server with the options specified
	gwmux := runtime.NewServeMux(gwmuxOptions...)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Debugf(ctx, "Running identity interceptor for http endpoint [%s]", r.URL.String())

		gwmux.ServeHTTP(w, r.WithContext(ctx))
	})
	return mux, gwmux, nil
}
