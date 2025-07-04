package cacheservice

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/flyteorg/flyte/cacheservice/pkg/config"
	"github.com/flyteorg/flyte/cacheservice/pkg/manager/impl"
	"github.com/flyteorg/flyte/cacheservice/pkg/manager/interfaces"
	"github.com/flyteorg/flyte/cacheservice/pkg/repositories"
	"github.com/flyteorg/flyte/cacheservice/pkg/runtime"
	"github.com/flyteorg/flyte/flyteadmin/plugins"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/cacheservice"
	cacheserviceV2 "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/cacheservice/v2"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/grpcutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type CacheService struct {
	CacheManager interfaces.CacheManager
}

func (s *CacheService) Get(ctx context.Context, request *cacheservice.GetCacheRequest) (*cacheservice.GetCacheResponse, error) {
	return s.CacheManager.Get(ctx, request)
}

func (s *CacheService) Put(ctx context.Context, request *cacheservice.PutCacheRequest) (*cacheservice.PutCacheResponse, error) {
	return s.CacheManager.Put(ctx, request)
}

func (s *CacheService) Delete(ctx context.Context, request *cacheservice.DeleteCacheRequest) (*cacheservice.DeleteCacheResponse, error) {
	return s.CacheManager.Delete(ctx, request)
}

func (s *CacheService) GetOrExtendReservation(ctx context.Context, request *cacheservice.GetOrExtendReservationRequest) (*cacheservice.GetOrExtendReservationResponse, error) {
	return s.CacheManager.GetOrExtendReservation(ctx, request, time.Now())
}

func (s *CacheService) ReleaseReservation(ctx context.Context, request *cacheservice.ReleaseReservationRequest) (*cacheservice.ReleaseReservationResponse, error) {
	return s.CacheManager.ReleaseReservation(ctx, request)
}

type V2CacheService struct {
	CacheManager interfaces.CacheManager
}

func updateCacheKey(key string, id *cacheserviceV2.Identifier) string {
	return fmt.Sprintf("%s-%s-%s-%s", id.GetOrg(), id.GetProject(), id.GetDomain(), key)
}

func (s *V2CacheService) Get(ctx context.Context, request *cacheserviceV2.GetCacheRequest) (*cacheservice.GetCacheResponse, error) {
	v1Request := request.GetBaseRequest()
	v1Request.Key = updateCacheKey(v1Request.GetKey(), request.GetIdentifier())

	return s.CacheManager.Get(ctx, v1Request)
}

func (s *V2CacheService) Put(ctx context.Context, request *cacheserviceV2.PutCacheRequest) (*cacheservice.PutCacheResponse, error) {
	v1Request := request.GetBaseRequest()
	v1Request.Key = updateCacheKey(v1Request.GetKey(), request.GetIdentifier())

	return s.CacheManager.Put(ctx, v1Request)
}

func (s *V2CacheService) Delete(ctx context.Context, request *cacheserviceV2.DeleteCacheRequest) (*cacheservice.DeleteCacheResponse, error) {
	v1Request := request.GetBaseRequest()
	v1Request.Key = updateCacheKey(v1Request.GetKey(), request.GetIdentifier())

	return s.CacheManager.Delete(ctx, v1Request)
}

func (s *V2CacheService) GetOrExtendReservation(ctx context.Context, request *cacheserviceV2.GetOrExtendReservationRequest) (*cacheservice.GetOrExtendReservationResponse, error) {
	v1Request := request.GetBaseRequest()
	v1Request.Key = updateCacheKey(v1Request.GetKey(), request.GetIdentifier())

	return s.CacheManager.GetOrExtendReservation(ctx, v1Request, time.Now())
}

func (s *V2CacheService) ReleaseReservation(ctx context.Context, request *cacheserviceV2.ReleaseReservationRequest) (*cacheservice.ReleaseReservationResponse, error) {
	v1Request := request.GetBaseRequest()
	v1Request.Key = updateCacheKey(v1Request.GetKey(), request.GetIdentifier())

	return s.CacheManager.ReleaseReservation(ctx, v1Request)
}

func NewCacheServiceServers() (*CacheService, *V2CacheService) {
	configProvider := runtime.NewConfigurationProvider()
	cacheServiceConfig := configProvider.ApplicationConfiguration().GetCacheServiceConfig()
	cacheServiceScope := promutils.NewScope(cacheServiceConfig.MetricsScope).NewSubScope("cacheservice")
	ctx := contextutils.WithAppName(context.Background(), "cacheservice")

	defer func() {
		if err := recover(); err != nil {
			cacheServiceScope.MustNewCounter("initialization_panic",
				"panics encountered initializing the cache service").Inc()
			logger.Fatalf(context.Background(), fmt.Sprintf("caught panic: %v [%+v]", err, string(debug.Stack())))
		}
	}()

	storeConfig := storage.GetConfig()
	dataStorageClient, err := storage.NewDataStore(storeConfig, cacheServiceScope.NewSubScope("storage"))
	if err != nil {
		logger.Errorf(ctx, "Failed to create DataStore %v, err %v", storeConfig, err)
		panic(err)
	}
	logger.Infof(ctx, "Created data storage.")

	baseStorageReference := dataStorageClient.GetBaseContainerFQN(ctx)
	storagePrefix, err := dataStorageClient.ConstructReference(ctx, baseStorageReference, cacheServiceConfig.StoragePrefix)
	if err != nil {
		logger.Errorf(ctx, "Failed to create prefix %v, err %v", cacheServiceConfig.StoragePrefix, err)
		panic(err)
	}

	outputStore := impl.NewCacheOutputBlobStore(dataStorageClient, storagePrefix)
	// TODO - @pvditt this should prolly not be handled here
	pgDbConfigValues := configProvider.ApplicationConfiguration().GetDbConfig()
	repos := repositories.GetRepositories(ctx, cacheServiceConfig, *pgDbConfigValues, cacheServiceScope)

	cacheManager := impl.NewCacheManager(outputStore, repos.CachedOutputRepo(), repos.ReservationRepo(), cacheServiceConfig.MaxInlineSizeBytes, cacheServiceScope.NewSubScope("cache"),
		time.Duration(cacheServiceConfig.HeartbeatGracePeriodMultiplier), cacheServiceConfig.MaxReservationHeartbeat.Duration)

	v1Service := &CacheService{
		CacheManager: cacheManager,
	}

	v2Service := &V2CacheService{
		CacheManager: cacheManager,
	}

	return v1Service, v2Service
}

// ServeInsecure creates and starts the gRPC server
func ServeInsecure(ctx context.Context, pluginRegistry *plugins.Registry, cfg *config.Config) error {
	grpcListener, err := net.Listen("tcp", cfg.GetGrpcHostAddress())
	if err != nil {
		return err
	}

	grpcServer := newGRPCServer(ctx, pluginRegistry, cfg)

	logger.Infof(ctx, "Serving CacheService Insecure on port %v", config.GetConfig().GetGrpcHostAddress())
	return grpcServer.Serve(grpcListener)
}

// Creates a new GRPC Server with all the configuration
func newGRPCServer(ctx context.Context, pluginRegistry *plugins.Registry, cfg *config.Config) *grpc.Server {
	tracerProvider := otelutils.GetTracerProvider(otelutils.CacheServiceServerTracer)

	otelUnaryServerInterceptor := otelgrpc.UnaryServerInterceptor(
		otelgrpc.WithTracerProvider(tracerProvider),
		otelgrpc.WithPropagators(propagation.TraceContext{}),
	)

	srvMetrics := grpcutils.GrpcServerMetrics()

	interceptors := []grpc.UnaryServerInterceptor{
		srvMetrics.UnaryServerInterceptor(),
		otelUnaryServerInterceptor,
	}
	middlewareInterceptors := plugins.Get[grpc.UnaryServerInterceptor](pluginRegistry, plugins.PluginIDUnaryServiceMiddleware)
	if middlewareInterceptors != nil {
		logger.Infof(ctx, "Creating gRPC server with interceptors")
		interceptors = append(interceptors, middlewareInterceptors)
	}
	chainedUnaryInterceptors := grpcmiddleware.ChainUnaryServer(interceptors...)

	serverOpts := []grpc.ServerOption{
		grpc.StreamInterceptor(srvMetrics.StreamServerInterceptor()),
		grpc.UnaryInterceptor(chainedUnaryInterceptors),
	}

	grpcServer := grpc.NewServer(serverOpts...)

	v1Server, v2Server := NewCacheServiceServers()
	cacheservice.RegisterCacheServiceServer(grpcServer, v1Server)
	cacheserviceV2.RegisterCacheServiceServer(grpcServer, v2Server)

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	if cfg.GrpcServerReflection {
		reflection.Register(grpcServer)
	}
	return grpcServer
}

// ServeHTTPHealthCheck create a http healthcheck endpoint
func ServeHTTPHealthCheck(ctx context.Context, cfg *config.Config) error {
	mux := http.NewServeMux()

	// Register Health check
	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	logger.Infof(ctx, "Serving CacheService http on port %v", cfg.GetHTTPHostAddress())

	server := &http.Server{
		Addr:              cfg.GetHTTPHostAddress(),
		Handler:           mux,
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeoutSeconds) * time.Second,
	}
	return server.ListenAndServe()
}

// ServeDummy creates and starts the gRPC dummy server and http healthcheck endpoint
func ServeDummy(ctx context.Context, cfg *config.Config) error {
	grpcServer := newGRPCDummyServer(ctx, cfg)

	grpcListener, err := net.Listen("tcp", cfg.GetGrpcHostAddress())
	if err != nil {
		return err
	}

	logger.Infof(ctx, "Serving CacheService Insecure on port %v", cfg.GetGrpcHostAddress())
	return grpcServer.Serve(grpcListener)
}

// Creates a new GRPC Server with all the configuration
func newGRPCDummyServer(_ context.Context, cfg *config.Config) *grpc.Server {
	tracerProvider := otelutils.GetTracerProvider(otelutils.CacheServiceClientTracer)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			otelgrpc.UnaryServerInterceptor(
				otelgrpc.WithTracerProvider(tracerProvider),
				otelgrpc.WithPropagators(propagation.TraceContext{}),
			),
		),
	)
	cacheservice.RegisterCacheServiceServer(grpcServer, &CacheService{})
	cacheserviceV2.RegisterCacheServiceServer(grpcServer, &V2CacheService{})
	if cfg.GrpcServerReflection {
		reflection.Register(grpcServer)
	}
	return grpcServer
}
