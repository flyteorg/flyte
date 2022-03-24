package server

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strings"

	"github.com/flyteorg/flyteadmin/auth"
	"github.com/flyteorg/flyteadmin/auth/authzserver"
	authConfig "github.com/flyteorg/flyteadmin/auth/config"
	"github.com/flyteorg/flyteadmin/auth/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/config"
	"github.com/flyteorg/flyteadmin/pkg/rpc/adminservice"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/secretmanager"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/gorilla/handlers"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var defaultCorsHeaders = []string{"Content-Type"}

// Serve starts a server and blocks the calling goroutine
func Serve(ctx context.Context, additionalHandlers map[string]func(http.ResponseWriter, *http.Request)) error {
	serverConfig := config.GetConfig()

	if serverConfig.Security.Secure {
		return serveGatewaySecure(ctx, serverConfig, authConfig.GetConfig(), additionalHandlers)
	}

	return serveGatewayInsecure(ctx, serverConfig, authConfig.GetConfig(), additionalHandlers)
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
	opts ...grpc.ServerOption) *grpc.Server {
	// Not yet implemented for streaming
	var chainedUnaryInterceptors grpc.UnaryServerInterceptor
	if cfg.Security.UseAuth {
		logger.Infof(ctx, "Creating gRPC server with authentication")
		chainedUnaryInterceptors = grpcmiddleware.ChainUnaryServer(grpcprometheus.UnaryServerInterceptor,
			auth.GetAuthenticationCustomMetadataInterceptor(authCtx),
			grpcauth.UnaryServerInterceptor(auth.GetAuthenticationInterceptor(authCtx)),
			auth.AuthenticationLoggingInterceptor,
			blanketAuthorization,
		)
	} else {
		logger.Infof(ctx, "Creating gRPC server without authentication")
		chainedUnaryInterceptors = grpcmiddleware.ChainUnaryServer(grpcprometheus.UnaryServerInterceptor)
	}

	serverOpts := []grpc.ServerOption{
		grpc.StreamInterceptor(grpcprometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(chainedUnaryInterceptors),
	}
	if cfg.GrpcConfig.MaxMessageSizeBytes > 0 {
		serverOpts = append(serverOpts, grpc.MaxRecvMsgSize(cfg.GrpcConfig.MaxMessageSizeBytes))
	}
	serverOpts = append(serverOpts, opts...)
	grpcServer := grpc.NewServer(serverOpts...)
	grpcprometheus.Register(grpcServer)
	service.RegisterAdminServiceServer(grpcServer, adminservice.NewAdminServer(ctx, cfg.KubeConfig, cfg.Master))
	if cfg.Security.UseAuth {
		service.RegisterAuthMetadataServiceServer(grpcServer, authCtx.AuthMetadataService())
		service.RegisterIdentityServiceServer(grpcServer, authCtx.IdentityService())
	}

	healthServer := health.NewServer()
	healthServer.SetServingStatus("flyteadmin", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	if cfg.GrpcConfig.ServerReflection || cfg.GrpcServerReflection {
		reflection.Register(grpcServer)
	}
	return grpcServer
}

func GetHandleOpenapiSpec(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		swaggerBytes, err := service.Asset("admin.swagger.json")
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

func healthCheckFunc(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func newHTTPServer(ctx context.Context, cfg *config.ServerConfig, _ *authConfig.Config, authCtx interfaces.AuthenticationContext,
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

	mux.Handle("/", gwmux)

	return mux, nil
}

func serveGatewayInsecure(ctx context.Context, cfg *config.ServerConfig, authCfg *authConfig.Config, additionalHandlers map[string]func(http.ResponseWriter, *http.Request)) error {
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

	grpcServer := newGRPCServer(ctx, cfg, authCtx)

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
	httpServer, err := newHTTPServer(ctx, cfg, authCfg, authCtx, additionalHandlers, cfg.GetGrpcHostAddress(), grpcOptions...)
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

func serveGatewaySecure(ctx context.Context, cfg *config.ServerConfig, authCfg *authConfig.Config, additionalHandlers map[string]func(http.ResponseWriter, *http.Request)) error {
	certPool, cert, err := GetSslCredentials(ctx, cfg.Security.Ssl.CertificateFile, cfg.Security.Ssl.KeyFile)
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

	grpcServer := newGRPCServer(ctx, cfg, authCtx, grpc.Creds(credentials.NewServerTLSFromCert(cert)))

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
	httpServer, err := newHTTPServer(ctx, cfg, authCfg, authCtx, additionalHandlers, cfg.GetHostAddress(), serverOpts...)
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
	}

	err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))

	if err != nil {
		return errors.Wrapf(err, "failed to Start HTTP/2 Server")
	}
	return nil
}
