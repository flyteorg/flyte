package entrypoints

import (
	"context"
	"crypto/tls"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/secretmanager"

	authConfig "github.com/flyteorg/flyteadmin/auth/config"

	"github.com/flyteorg/flyteadmin/auth/authzserver"

	"github.com/gorilla/handlers"

	"github.com/flyteorg/flyteadmin/auth"
	"github.com/flyteorg/flyteadmin/auth/interfaces"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	"net"
	"net/http"
	_ "net/http/pprof" // Required to serve application.
	"strings"

	"github.com/flyteorg/flyteadmin/pkg/server"
	"github.com/pkg/errors"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/flyteorg/flyteadmin/pkg/common"
	flyteService "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/flyteorg/flyteadmin/pkg/config"
	"github.com/flyteorg/flyteadmin/pkg/rpc/adminservice"
	"github.com/spf13/cobra"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var defaultCorsHeaders = []string{"Content-Type"}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Launches the Flyte admin server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		serverConfig := config.GetConfig()

		if serverConfig.Security.Secure {
			return serveGatewaySecure(ctx, serverConfig, authConfig.GetConfig())
		}

		return serveGatewayInsecure(ctx, serverConfig, authConfig.GetConfig())
	},
}

func init() {
	// Command information
	RootCmd.AddCommand(serveCmd)
	RootCmd.AddCommand(secretsCmd)

	// Set Keys
	labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey,
		contextutils.ExecIDKey, contextutils.WorkflowIDKey, contextutils.NodeIDKey, contextutils.TaskIDKey,
		contextutils.TaskTypeKey, common.RuntimeTypeKey, common.RuntimeVersionKey)
}

func blanketAuthorization(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
	resp interface{}, err error) {

	identityContext := auth.IdentityContextFromContext(ctx)
	if identityContext.IsEmpty() {
		return handler(ctx, req)
	}

	if !identityContext.Scopes().Has(auth.ScopeAll) {
		return nil, status.Errorf(codes.Unauthenticated, "authenticated user doesn't have required scope")
	}

	return handler(ctx, req)
}

// Creates a new gRPC Server with all the configuration
func newGRPCServer(ctx context.Context, cfg *config.ServerConfig, authCtx interfaces.AuthenticationContext,
	opts ...grpc.ServerOption) (*grpc.Server, error) {
	// Not yet implemented for streaming
	var chainedUnaryInterceptors grpc.UnaryServerInterceptor
	if cfg.Security.UseAuth {
		logger.Infof(ctx, "Creating gRPC server with authentication")
		chainedUnaryInterceptors = grpc_middleware.ChainUnaryServer(grpcPrometheus.UnaryServerInterceptor,
			auth.GetAuthenticationCustomMetadataInterceptor(authCtx),
			grpcauth.UnaryServerInterceptor(auth.GetAuthenticationInterceptor(authCtx)),
			auth.AuthenticationLoggingInterceptor,
			blanketAuthorization,
		)
	} else {
		logger.Infof(ctx, "Creating gRPC server without authentication")
		chainedUnaryInterceptors = grpc_middleware.ChainUnaryServer(grpcPrometheus.UnaryServerInterceptor)
	}

	serverOpts := []grpc.ServerOption{
		grpc.StreamInterceptor(grpcPrometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(chainedUnaryInterceptors),
	}
	if cfg.GrpcConfig.MaxMessageSizeBytes > 0 {
		serverOpts = append(serverOpts, grpc.MaxRecvMsgSize(cfg.GrpcConfig.MaxMessageSizeBytes))
	}
	serverOpts = append(serverOpts, opts...)
	grpcServer := grpc.NewServer(serverOpts...)
	grpcPrometheus.Register(grpcServer)
	flyteService.RegisterAdminServiceServer(grpcServer, adminservice.NewAdminServer(ctx, cfg.KubeConfig, cfg.Master))
	if cfg.Security.UseAuth {
		flyteService.RegisterAuthMetadataServiceServer(grpcServer, authCtx.AuthMetadataService())
		flyteService.RegisterIdentityServiceServer(grpcServer, authCtx.IdentityService())
	}

	healthServer := health.NewServer()
	healthServer.SetServingStatus("flyteadmin", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	if cfg.GrpcConfig.ServerReflection || cfg.GrpcServerReflection {
		reflection.Register(grpcServer)
	}
	return grpcServer, nil
}

func GetHandleOpenapiSpec(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		swaggerBytes, err := flyteService.Asset("admin.swagger.json")
		if err != nil {
			logger.Warningf(ctx, "Err %v", err)
			w.WriteHeader(http.StatusFailedDependency)
		} else {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write(swaggerBytes)
			if err != nil {
				logger.Errorf(ctx, "failed to write openAPI information, error: %s", err.Error())
			}
		}
	}
}

func healthCheckFunc(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func newHTTPServer(ctx context.Context, cfg *config.ServerConfig, authCfg *authConfig.Config, authCtx interfaces.AuthenticationContext,
	grpcAddress string, grpcConnectionOpts ...grpc.DialOption) (*http.ServeMux, error) {

	// Register the server that will serve HTTP/REST Traffic
	mux := http.NewServeMux()

	// Register healthcheck
	mux.HandleFunc("/healthcheck", healthCheckFunc)

	// Register OpenAPI endpoint
	// This endpoint will serve the OpenAPI2 spec generated by the swagger protoc plugin, and bundled by go-bindata
	mux.HandleFunc("/api/v1/openapi", GetHandleOpenapiSpec(ctx))

	var gwmuxOptions = make([]runtime.ServeMuxOption, 0)
	// This option means that http requests are served with protobufs, instead of json. We always want this.
	gwmuxOptions = append(gwmuxOptions, runtime.WithMarshalerOption("application/octet-stream", &runtime.ProtoMarshaller{}))

	if cfg.Security.UseAuth {
		// Add HTTP handlers for OIDC endpoints
		auth.RegisterHandlers(ctx, mux, authCtx)

		// Add HTTP handlers for OAuth2 endpoints
		authzserver.RegisterHandlers(mux, authCtx)

		// This option translates HTTP authorization data (cookies) into a gRPC metadata field
		gwmuxOptions = append(gwmuxOptions, runtime.WithMetadata(auth.GetHTTPRequestCookieToMetadataHandler(authCtx)))

		// In an attempt to be able to selectively enforce whether or not authentication is required, we're going to tag
		// the requests that come from the HTTP gateway. See the enforceHttp/Grpc options for more information.
		gwmuxOptions = append(gwmuxOptions, runtime.WithMetadata(auth.GetHTTPMetadataTaggingHandler()))
	}

	// Create the grpc-gateway server with the options specified
	gwmux := runtime.NewServeMux(gwmuxOptions...)

	err := flyteService.RegisterAdminServiceHandlerFromEndpoint(ctx, gwmux, grpcAddress, grpcConnectionOpts)
	if err != nil {
		return nil, errors.Wrap(err, "error registering admin service")
	}

	err = flyteService.RegisterAuthMetadataServiceHandlerFromEndpoint(ctx, gwmux, grpcAddress, grpcConnectionOpts)
	if err != nil {
		return nil, errors.Wrap(err, "error registering auth service")
	}

	err = flyteService.RegisterIdentityServiceHandlerFromEndpoint(ctx, gwmux, grpcAddress, grpcConnectionOpts)
	if err != nil {
		return nil, errors.Wrap(err, "error registering identity service")
	}

	mux.Handle("/", gwmux)

	return mux, nil
}

func serveGatewayInsecure(ctx context.Context, cfg *config.ServerConfig, authCfg *authConfig.Config) error {
	logger.Infof(ctx, "Serving Flyte Admin Insecure")

	// This will parse configuration and create the necessary objects for dealing with auth
	var authCtx interfaces.AuthenticationContext
	var err error
	// This code is here to support authentication without SSL. This setup supports a network topology where
	// Envoy does the SSL termination. The final hop is made over localhost only on a trusted machine.
	// Warning: Running authentication without SSL in any other topology is a severe security flaw.
	// See the auth.Config object for additional settings as well.
	if cfg.Security.UseAuth {
		sm := secretmanager.NewFileEnvSecretManager(secretmanager.GetConfig())
		var oauth2Provider interfaces.OAuth2Provider
		var oauth2ResourceServer interfaces.OAuth2ResourceServer
		if authCfg.AppAuth.AuthServerType == authConfig.AuthorizationServerTypeSelf {
			oauth2Provider, err = authzserver.NewProvider(ctx, authCfg.AppAuth.SelfAuthServer, sm)
			if err != nil {
				logger.Errorf(ctx, "Error creating authorization server %s", err)
				return err
			}

			oauth2ResourceServer = oauth2Provider
		} else {
			oauth2ResourceServer, err = authzserver.NewOAuth2ResourceServer(ctx, authCfg.AppAuth.ExternalAuthServer, authCfg.UserAuth.OpenID.BaseURL)
			if err != nil {
				logger.Errorf(ctx, "Error creating resource server %s", err)
				return err
			}
		}

		oauth2MetadataProvider := authzserver.NewService(authCfg)
		oidcUserInfoProvider := auth.NewUserInfoProvider()

		authCtx, err = auth.NewAuthenticationContext(ctx, sm, oauth2Provider, oauth2ResourceServer, oauth2MetadataProvider, oidcUserInfoProvider, authCfg)
		if err != nil {
			logger.Errorf(ctx, "Error creating auth context %s", err)
			return err
		}
	}

	grpcServer, err := newGRPCServer(ctx, cfg, authCtx)
	if err != nil {
		return errors.Wrap(err, "failed to create GRPC server")
	}

	logger.Infof(ctx, "Serving GRPC Traffic on: %s", cfg.GetGrpcHostAddress())
	lis, err := net.Listen("tcp", cfg.GetGrpcHostAddress())
	if err != nil {
		return errors.Wrapf(err, "failed to listen on GRPC port: %s", cfg.GetGrpcHostAddress())
	}

	go func() {
		err := grpcServer.Serve(lis)
		logger.Fatalf(ctx, "Failed to create GRPC Server, Err: ", err)
	}()

	logger.Infof(ctx, "Starting HTTP/1 Gateway server on %s", cfg.GetHostAddress())
	grpcOptions := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithMaxHeaderListSize(common.MaxResponseStatusBytes),
	}
	if cfg.GrpcConfig.MaxMessageSizeBytes > 0 {
		grpcOptions = append(grpcOptions,
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(cfg.GrpcConfig.MaxMessageSizeBytes)))
	}
	httpServer, err := newHTTPServer(ctx, cfg, authCfg, authCtx, cfg.GetGrpcHostAddress(), grpcOptions...)
	if err != nil {
		return err
	}

	var handler http.Handler
	if cfg.Security.AllowCors {
		handler = handlers.CORS(
			handlers.AllowCredentials(),
			handlers.AllowedOrigins(cfg.Security.AllowedOrigins),
			handlers.AllowedHeaders(append(defaultCorsHeaders, cfg.Security.AllowedHeaders...)),
			handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "HEAD", "PUT", "PATCH"}),
		)(httpServer)
	} else {
		handler = httpServer
	}

	err = http.ListenAndServe(cfg.GetHostAddress(), handler)
	if err != nil {
		return errors.Wrapf(err, "failed to Start HTTP Server")
	}

	return nil
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise.
// See https://github.com/philips/grpc-gateway-example/blob/master/cmd/serve.go for reference
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a partial recreation of gRPC's internal checks
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func serveGatewaySecure(ctx context.Context, cfg *config.ServerConfig, authCfg *authConfig.Config) error {
	certPool, cert, err := server.GetSslCredentials(ctx, cfg.Security.Ssl.CertificateFile, cfg.Security.Ssl.KeyFile)
	if err != nil {
		return err
	}
	// This will parse configuration and create the necessary objects for dealing with auth
	var authCtx interfaces.AuthenticationContext
	if cfg.Security.UseAuth {
		sm := secretmanager.NewFileEnvSecretManager(secretmanager.GetConfig())
		var oauth2Provider interfaces.OAuth2Provider
		var oauth2ResourceServer interfaces.OAuth2ResourceServer
		if authCfg.AppAuth.AuthServerType == authConfig.AuthorizationServerTypeSelf {
			oauth2Provider, err = authzserver.NewProvider(ctx, authCfg.AppAuth.SelfAuthServer, sm)
			if err != nil {
				logger.Errorf(ctx, "Error creating authorization server %s", err)
				return err
			}

			oauth2ResourceServer = oauth2Provider
		} else {
			oauth2ResourceServer, err = authzserver.NewOAuth2ResourceServer(ctx, authCfg.AppAuth.ExternalAuthServer, authCfg.UserAuth.OpenID.BaseURL)
			if err != nil {
				logger.Errorf(ctx, "Error creating resource server %s", err)
				return err
			}
		}

		oauth2MetadataProvider := authzserver.NewService(authCfg)
		oidcUserInfoProvider := auth.NewUserInfoProvider()

		authCtx, err = auth.NewAuthenticationContext(ctx, sm, oauth2Provider, oauth2ResourceServer, oauth2MetadataProvider, oidcUserInfoProvider, authCfg)
		if err != nil {
			logger.Errorf(ctx, "Error creating auth context %s", err)
			return err
		}
	}

	grpcServer, err := newGRPCServer(ctx, cfg, authCtx,
		grpc.Creds(credentials.NewServerTLSFromCert(cert)))
	if err != nil {
		return errors.Wrap(err, "failed to create GRPC server")
	}

	// Whatever certificate is used, pass it along for easier development
	dialCreds := credentials.NewTLS(&tls.Config{
		ServerName: cfg.GetHostAddress(),
		RootCAs:    certPool,
	})
	serverOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(dialCreds),
	}
	if cfg.GrpcConfig.MaxMessageSizeBytes > 0 {
		serverOpts = append(serverOpts,
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(cfg.GrpcConfig.MaxMessageSizeBytes)))
	}
	httpServer, err := newHTTPServer(ctx, cfg, authCfg, authCtx, cfg.GetHostAddress(), serverOpts...)
	if err != nil {
		return err
	}

	conn, err := net.Listen("tcp", cfg.GetHostAddress())
	if err != nil {
		panic(err)
	}

	srv := &http.Server{
		Addr:    cfg.GetHostAddress(),
		Handler: grpcHandlerFunc(grpcServer, httpServer),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*cert},
			NextProtos:   []string{"h2"},
		},
	}

	err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))

	if err != nil {
		return errors.Wrapf(err, "failed to Start HTTP/2 Server")
	}
	return nil
}
