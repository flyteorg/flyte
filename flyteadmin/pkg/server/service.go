package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/flyteorg/flyte/flyteadmin/auth"
	"github.com/flyteorg/flyte/flyteadmin/auth/authzserver"
	authConfig "github.com/flyteorg/flyte/flyteadmin/auth/config"
	"github.com/flyteorg/flyte/flyteadmin/auth/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/dataproxy"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/config"
	"github.com/flyteorg/flyte/flyteadmin/pkg/rpc"
	"github.com/flyteorg/flyte/flyteadmin/pkg/rpc/adminservice"
	"github.com/flyteorg/flyte/flyteadmin/pkg/rpc/adminservice/middleware"
	runtime2 "github.com/flyteorg/flyte/flyteadmin/pkg/runtime"
	runtimeIfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/plugins"
	"github.com/flyteorg/flyte/flyteidl/clients/go/assets"
	grpcService "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/gateway/flyteidl/service"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/secretmanager"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var defaultCorsHeaders = []string{"Content-Type"}

// Serve starts a server and blocks the calling goroutine
func Serve(ctx context.Context, pluginRegistry *plugins.Registry, additionalHandlers map[string]func(http.ResponseWriter, *http.Request)) error {
	serverConfig := config.GetConfig()
	configuration := runtime2.NewConfigurationProvider()
	adminScope := promutils.NewScope(configuration.ApplicationConfiguration().GetTopLevelConfig().GetMetricsScope()).NewSubScope("admin")

	if serverConfig.Security.Secure {
		return serveGatewaySecure(ctx, pluginRegistry, serverConfig, authConfig.GetConfig(), storage.GetConfig(), additionalHandlers, adminScope)
	}

	return serveGatewayInsecure(ctx, pluginRegistry, serverConfig, authConfig.GetConfig(), storage.GetConfig(), additionalHandlers, adminScope)
}

func SetMetricKeys(appConfig *runtimeIfaces.ApplicationConfig) {
	var keys = make([]contextutils.Key, 0, len(appConfig.MetricKeys))
	for _, keyName := range appConfig.MetricKeys {
		key := contextutils.Key(keyName)
		keys = append(keys, key)
	}
	logger.Infof(context.TODO(), "setting metrics keys to %+v", keys)
	if len(keys) > 0 {
		labeled.SetMetricKeys(keys...)
	}
}

// Creates a new gRPC Server with all the configuration
func newGRPCServer(ctx context.Context, pluginRegistry *plugins.Registry, cfg *config.ServerConfig,
	storageCfg *storage.Config, authCtx interfaces.AuthenticationContext,
	scope promutils.Scope, sm core.SecretManager, opts ...grpc.ServerOption) (*grpc.Server, error) {

	logger.Infof(ctx, "Registering default middleware with blanket auth validation")
	pluginRegistry.RegisterDefault(plugins.PluginIDUnaryServiceMiddleware, grpcmiddleware.ChainUnaryServer(
		RequestIDInterceptor, auth.BlanketAuthorization, auth.ExecutionUserIdentifierInterceptor))

	if cfg.GrpcConfig.EnableGrpcLatencyMetrics {
		logger.Debugf(ctx, "enabling grpc histogram metrics")
		grpcprometheus.EnableHandlingTimeHistogram()
	}

	// Not yet implemented for streaming
	tracerProvider := otelutils.GetTracerProvider(otelutils.AdminServerTracer)
	otelUnaryServerInterceptor := otelgrpc.UnaryServerInterceptor(
		otelgrpc.WithTracerProvider(tracerProvider),
		otelgrpc.WithPropagators(propagation.TraceContext{}),
	)

	adminScope := scope.NewSubScope("admin")
	recoveryInterceptor := middleware.NewRecoveryInterceptor(adminScope)

	var chainedUnaryInterceptors grpc.UnaryServerInterceptor
	if cfg.Security.UseAuth {
		logger.Infof(ctx, "Creating gRPC server with authentication")
		middlewareInterceptors := plugins.Get[grpc.UnaryServerInterceptor](pluginRegistry, plugins.PluginIDUnaryServiceMiddleware)
		chainedUnaryInterceptors = grpcmiddleware.ChainUnaryServer(
			// recovery interceptor should always be first in order to handle any panics in the middleware or server
			recoveryInterceptor.UnaryServerInterceptor(),
			grpcrecovery.UnaryServerInterceptor(),
			grpcprometheus.UnaryServerInterceptor,
			otelUnaryServerInterceptor,
			auth.GetAuthenticationCustomMetadataInterceptor(authCtx),
			grpcauth.UnaryServerInterceptor(auth.GetAuthenticationInterceptor(authCtx)),
			auth.AuthenticationLoggingInterceptor,
			middlewareInterceptors,
		)
	} else {
		logger.Infof(ctx, "Creating gRPC server without authentication")
		chainedUnaryInterceptors = grpcmiddleware.ChainUnaryServer(
			// recovery interceptor should always be first in order to handle any panics in the middleware or server
			recoveryInterceptor.UnaryServerInterceptor(),
			grpcprometheus.UnaryServerInterceptor,
			otelUnaryServerInterceptor,
		)
	}

	chainedStreamInterceptors := grpcmiddleware.ChainStreamServer(
		// recovery interceptor should always be first in order to handle any panics in the middleware or server
		recoveryInterceptor.StreamServerInterceptor(),
		grpcprometheus.StreamServerInterceptor,
	)

	serverOpts := []grpc.ServerOption{
		// recovery interceptor should always be first in order to handle any panics in the middleware or server
		grpc.StreamInterceptor(chainedStreamInterceptors),
		grpc.UnaryInterceptor(chainedUnaryInterceptors),
	}
	if cfg.GrpcConfig.MaxMessageSizeBytes > 0 {
		serverOpts = append(serverOpts, grpc.MaxRecvMsgSize(cfg.GrpcConfig.MaxMessageSizeBytes), grpc.MaxSendMsgSize(cfg.GrpcConfig.MaxMessageSizeBytes))
	}
	serverOpts = append(serverOpts, opts...)
	grpcServer := grpc.NewServer(serverOpts...)
	grpcprometheus.Register(grpcServer)
	dataStorageClient, err := storage.NewDataStore(storageCfg, scope.NewSubScope("storage"))
	if err != nil {
		logger.Error(ctx, "Failed to initialize storage config")
		panic(err)
	}

	configuration := runtime2.NewConfigurationProvider()
	adminServer := adminservice.NewAdminServer(ctx, pluginRegistry, configuration, cfg.KubeConfig, cfg.Master, dataStorageClient, adminScope, sm)
	grpcService.RegisterAdminServiceServer(grpcServer, adminServer)
	if cfg.Security.UseAuth {
		grpcService.RegisterAuthMetadataServiceServer(grpcServer, authCtx.AuthMetadataService())
		grpcService.RegisterIdentityServiceServer(grpcServer, authCtx.IdentityService())
	}

	dataProxySvc, err := dataproxy.NewService(cfg.DataProxy, adminServer.NodeExecutionManager, dataStorageClient, adminServer.TaskExecutionManager)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize dataProxy service. Error: %w", err)
	}

	pluginRegistry.RegisterDefault(plugins.PluginIDDataProxy, dataProxySvc)
	grpcService.RegisterDataProxyServiceServer(grpcServer, plugins.Get[grpcService.DataProxyServiceServer](pluginRegistry, plugins.PluginIDDataProxy))

	grpcService.RegisterSignalServiceServer(grpcServer, rpc.NewSignalServer(ctx, configuration, scope.NewSubScope("signal")))

	additionalService := plugins.Get[common.RegisterAdditionalGRPCService](pluginRegistry, plugins.PluginIDAdditionalGRPCService)
	if additionalService != nil {
		if err := additionalService(ctx, grpcServer); err != nil {
			return nil, err
		}
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
		// TODO: find a better way to point to admin.swagger.json
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(assets.AdminSwaggerFile)
		if err != nil {
			logger.Errorf(ctx, "failed to write openAPI information, error: %s", err.Error())
		}
	}
}

func healthCheckFunc(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func newHTTPServer(ctx context.Context, pluginRegistry *plugins.Registry, cfg *config.ServerConfig, _ *authConfig.Config, authCtx interfaces.AuthenticationContext,
	additionalHandlers map[string]func(http.ResponseWriter, *http.Request),
	grpcAddress string, grpcConnectionOpts ...grpc.DialOption) (*http.ServeMux, error) {

	// Register the server that will serve HTTP/REST Traffic
	mux := http.NewServeMux()

	// Add any additional handlers that have been passed in for the main HTTP server
	for p, f := range additionalHandlers {
		mux.HandleFunc(p, f)
	}

	// Register healthcheck
	mux.HandleFunc("/healthcheck", healthCheckFunc)

	// Register OpenAPI endpoint
	// This endpoint will serve the OpenAPI2 spec generated by the swagger protoc plugin, and bundled by go-bindata
	mux.HandleFunc("/api/v1/openapi", GetHandleOpenapiSpec(ctx))

	var gwmuxOptions = make([]runtime.ServeMuxOption, 0)
	// This option means that http requests are served with protobufs, instead of json. We always want this.
	gwmuxOptions = append(gwmuxOptions, runtime.WithMarshalerOption("application/octet-stream", &runtime.ProtoMarshaller{}))
	// grpc-gateway v2 switched the marshaller used to encode JSON messages in 2.5.0. This changed the
	// default encoding from snake_case (the v1 behavior) to lowerCamelCase, which is the case recommended
	// by protobuf. However the protobuf docs do mention that JSON printers may provide a way to use
	// the proto names as field names instead. This option in grpc-gateway v2 does just that,
	// by setting a custom marshaler. We are enabling this narrowly however, by applying it only for
	// the application/json content type.
	gwmuxOptions = append(gwmuxOptions, runtime.WithMarshalerOption("application/json", &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames:     true,
			EmitUnpopulated:   true,
			EmitDefaultValues: true,
		},
	}))

	// This option sets subject in the user info response
	gwmuxOptions = append(gwmuxOptions, runtime.WithForwardResponseOption(auth.GetUserInfoForwardResponseHandler()))

	// Use custom header matcher to allow additional headers to be passed through
	gwmuxOptions = append(gwmuxOptions, runtime.WithIncomingHeaderMatcher(auth.GetCustomHeaderMatcher(pluginRegistry)))

	if cfg.Security.UseAuth {
		// Add HTTP handlers for OIDC endpoints
		auth.RegisterHandlers(ctx, mux, authCtx, pluginRegistry)

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

	err := service.RegisterAdminServiceHandlerFromEndpoint(ctx, gwmux, grpcAddress, grpcConnectionOpts)
	if err != nil {
		return nil, errors.Wrap(err, "error registering admin service")
	}

	err = service.RegisterAuthMetadataServiceHandlerFromEndpoint(ctx, gwmux, grpcAddress, grpcConnectionOpts)
	if err != nil {
		return nil, errors.Wrap(err, "error registering auth service")
	}

	err = service.RegisterIdentityServiceHandlerFromEndpoint(ctx, gwmux, grpcAddress, grpcConnectionOpts)
	if err != nil {
		return nil, errors.Wrap(err, "error registering identity service")
	}

	err = service.RegisterDataProxyServiceHandlerFromEndpoint(ctx, gwmux, grpcAddress, grpcConnectionOpts)
	if err != nil {
		return nil, errors.Wrap(err, "error registering data proxy service")
	}

	err = service.RegisterSignalServiceHandlerFromEndpoint(ctx, gwmux, grpcAddress, grpcConnectionOpts)
	if err != nil {
		return nil, errors.Wrap(err, "error registering signal service")
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := GetOrGenerateRequestIDForRequest(r)
		gwmux.ServeHTTP(w, r.WithContext(ctx))
	})

	return mux, nil
}

// RequestIDInterceptor is a server interceptor that sets the request id on the context for any incoming calls.
func RequestIDInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(GetOrGenerateRequestIDForGRPC(ctx), req)
}

// GetOrGenerateRequestIDForGRPC returns a context with request id set from the context or from grpc metadata if it exists,
// otherwise it generates a new one.
func GetOrGenerateRequestIDForGRPC(ctx context.Context) context.Context {
	if ctx.Value(contextutils.RequestIDKey) != nil {
		return ctx
	} else if md, exists := metadata.FromIncomingContext(ctx); exists && len(md.Get(contextutils.RequestIDKey.String())) > 0 {
		return contextutils.WithRequestID(ctx, md.Get(contextutils.RequestIDKey.String())[0])
	} else {
		return contextutils.WithRequestID(ctx, generateRequestID())
	}
}

// GetOrGenerateRequestIDForRequest returns a context with request id set from the context or from metadata if it exists,
// otherwise it generates a new one.
func GetOrGenerateRequestIDForRequest(req *http.Request) context.Context {
	ctx := req.Context()
	if ctx.Value(contextutils.RequestIDKey) != nil {
		return ctx
	} else if md, exists := metadata.FromIncomingContext(ctx); exists && len(md.Get(contextutils.RequestIDKey.String())) > 0 {
		return contextutils.WithRequestID(ctx, md.Get(contextutils.RequestIDKey.String())[0])
	} else if req.Header != nil && req.Header.Get(contextutils.RequestIDKey.String()) != "" {
		return contextutils.WithRequestID(ctx, req.Header.Get(contextutils.RequestIDKey.String()))
	} else {
		return contextutils.WithRequestID(ctx, generateRequestID())
	}
}

func generateRequestID() string {
	return "a-" + rand.String(20)
}

func serveGatewayInsecure(ctx context.Context, pluginRegistry *plugins.Registry, cfg *config.ServerConfig,
	authCfg *authConfig.Config, storageConfig *storage.Config,
	additionalHandlers map[string]func(http.ResponseWriter, *http.Request),
	scope promutils.Scope) error {
	logger.Infof(ctx, "Serving Flyte Admin Insecure")

	// This will parse configuration and create the necessary objects for dealing with auth
	var authCtx interfaces.AuthenticationContext
	var err error

	sm := secretmanager.NewFileEnvSecretManager(secretmanager.GetConfig())

	// This code is here to support authentication without SSL. This setup supports a network topology where
	// Envoy does the SSL termination. The final hop is made over localhost only on a trusted machine.
	// Warning: Running authentication without SSL in any other topology is a severe security flaw.
	// See the auth.Config object for additional settings as well.
	if cfg.Security.UseAuth {

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

	grpcServer, err := newGRPCServer(ctx, pluginRegistry, cfg, storageConfig, authCtx, scope, sm)
	if err != nil {
		return fmt.Errorf("failed to create a newGRPCServer. Error: %w", err)
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
	httpServer, err := newHTTPServer(ctx, pluginRegistry, cfg, authCfg, authCtx, additionalHandlers, cfg.GetGrpcHostAddress(), grpcOptions...)
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

	server := &http.Server{
		Addr:              cfg.GetHostAddress(),
		Handler:           handler,
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeoutSeconds) * time.Second,
	}

	err = server.ListenAndServe()
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

func serveGatewaySecure(ctx context.Context, pluginRegistry *plugins.Registry, cfg *config.ServerConfig, authCfg *authConfig.Config,
	storageCfg *storage.Config,
	additionalHandlers map[string]func(http.ResponseWriter, *http.Request),
	scope promutils.Scope) error {
	certPool, cert, err := GetSslCredentials(ctx, cfg.Security.Ssl.CertificateFile, cfg.Security.Ssl.KeyFile)
	sm := secretmanager.NewFileEnvSecretManager(secretmanager.GetConfig())

	if err != nil {
		return err
	}
	// This will parse configuration and create the necessary objects for dealing with auth
	var authCtx interfaces.AuthenticationContext
	if cfg.Security.UseAuth {
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

	grpcServer, err := newGRPCServer(ctx, pluginRegistry, cfg, storageCfg, authCtx, scope, sm, grpc.Creds(credentials.NewServerTLSFromCert(cert)))
	if err != nil {
		return fmt.Errorf("failed to create a newGRPCServer. Error: %w", err)
	}

	// Whatever certificate is used, pass it along for easier development
	// #nosec G402
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
	httpServer, err := newHTTPServer(ctx, pluginRegistry, cfg, authCfg, authCtx, additionalHandlers, cfg.GetHostAddress(), serverOpts...)
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
		// #nosec G402
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*cert},
			NextProtos:   []string{"h2"},
		},
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeoutSeconds) * time.Second,
	}

	err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))

	if err != nil {
		return errors.Wrapf(err, "failed to Start HTTP/2 Server")
	}
	return nil
}
