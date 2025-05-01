package datacatalogservice

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/flyteorg/flyte/datacatalog/pkg/config"
	"github.com/flyteorg/flyte/datacatalog/pkg/manager/impl"
	"github.com/flyteorg/flyte/datacatalog/pkg/manager/interfaces"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories"
	"github.com/flyteorg/flyte/datacatalog/pkg/runtime"
	catalog "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type DataCatalogService struct {
	DatasetManager     interfaces.DatasetManager
	ArtifactManager    interfaces.ArtifactManager
	TagManager         interfaces.TagManager
	ReservationManager interfaces.ReservationManager
}

func (s *DataCatalogService) CreateDataset(ctx context.Context, request *catalog.CreateDatasetRequest) (*catalog.CreateDatasetResponse, error) {
	return s.DatasetManager.CreateDataset(ctx, request)
}

func (s *DataCatalogService) CreateArtifact(ctx context.Context, request *catalog.CreateArtifactRequest) (*catalog.CreateArtifactResponse, error) {
	return s.ArtifactManager.CreateArtifact(ctx, request)
}

func (s *DataCatalogService) GetDataset(ctx context.Context, request *catalog.GetDatasetRequest) (*catalog.GetDatasetResponse, error) {
	return s.DatasetManager.GetDataset(ctx, request)
}

func (s *DataCatalogService) GetArtifact(ctx context.Context, request *catalog.GetArtifactRequest) (*catalog.GetArtifactResponse, error) {
	return s.ArtifactManager.GetArtifact(ctx, request)
}

func (s *DataCatalogService) ListArtifacts(ctx context.Context, request *catalog.ListArtifactsRequest) (*catalog.ListArtifactsResponse, error) {
	return s.ArtifactManager.ListArtifacts(ctx, request)
}

func (s *DataCatalogService) AddTag(ctx context.Context, request *catalog.AddTagRequest) (*catalog.AddTagResponse, error) {
	return s.TagManager.AddTag(ctx, request)
}

func (s *DataCatalogService) ListDatasets(ctx context.Context, request *catalog.ListDatasetsRequest) (*catalog.ListDatasetsResponse, error) {
	return s.DatasetManager.ListDatasets(ctx, request)
}

func (s *DataCatalogService) UpdateArtifact(ctx context.Context, request *catalog.UpdateArtifactRequest) (*catalog.UpdateArtifactResponse, error) {
	return s.ArtifactManager.UpdateArtifact(ctx, request)
}

func (s *DataCatalogService) GetOrExtendReservation(ctx context.Context, request *catalog.GetOrExtendReservationRequest) (*catalog.GetOrExtendReservationResponse, error) {
	return s.ReservationManager.GetOrExtendReservation(ctx, request)
}

func (s *DataCatalogService) ReleaseReservation(ctx context.Context, request *catalog.ReleaseReservationRequest) (*catalog.ReleaseReservationResponse, error) {
	return s.ReservationManager.ReleaseReservation(ctx, request)
}

func NewDataCatalogService() *DataCatalogService {
	configProvider := runtime.NewConfigurationProvider()
	dataCatalogConfig := configProvider.ApplicationConfiguration().GetDataCatalogConfig()
	catalogScope := promutils.NewScope(dataCatalogConfig.MetricsScope).NewSubScope("datacatalog")
	ctx := contextutils.WithAppName(context.Background(), "datacatalog")

	defer func() {
		if err := recover(); err != nil {
			catalogScope.MustNewCounter("initialization_panic",
				"panics encountered initializing the datacatalog service").Inc()
			logger.Fatalf(context.Background(), fmt.Sprintf("caught panic: %v [%+v]", err, string(debug.Stack())))
		}
	}()

	storeConfig := storage.GetConfig()
	dataStorageClient, err := storage.NewDataStore(storeConfig, catalogScope.NewSubScope("storage"))
	if err != nil {
		logger.Errorf(ctx, "Failed to create DataStore %v, err %v", storeConfig, err)
		panic(err)
	}
	logger.Infof(ctx, "Created data storage.")

	baseStorageReference := dataStorageClient.GetBaseContainerFQN(ctx)
	storagePrefix, err := dataStorageClient.ConstructReference(ctx, baseStorageReference, dataCatalogConfig.StoragePrefix)
	if err != nil {
		logger.Errorf(ctx, "Failed to create prefix %v, err %v", dataCatalogConfig.StoragePrefix, err)
		panic(err)
	}

	dbConfigValues := configProvider.ApplicationConfiguration().GetDbConfig()
	repos := repositories.GetRepository(ctx, repositories.POSTGRES, *dbConfigValues, catalogScope)
	logger.Infof(ctx, "Created DB connection.")

	return &DataCatalogService{
		DatasetManager:  impl.NewDatasetManager(repos, dataStorageClient, catalogScope.NewSubScope("dataset")),
		ArtifactManager: impl.NewArtifactManager(repos, dataStorageClient, storagePrefix, catalogScope.NewSubScope("artifact")),
		TagManager:      impl.NewTagManager(repos, dataStorageClient, catalogScope.NewSubScope("tag")),
		ReservationManager: impl.NewReservationManager(repos, time.Duration(dataCatalogConfig.HeartbeatGracePeriodMultiplier), dataCatalogConfig.MaxReservationHeartbeat.Duration, time.Now,
			catalogScope.NewSubScope("reservation")),
	}
}

// Create and start the gRPC server
func ServeInsecure(ctx context.Context, cfg *config.Config) error {
	grpcServer := newGRPCServer(ctx, cfg)

	grpcListener, err := net.Listen("tcp", cfg.GetGrpcHostAddress())
	if err != nil {
		return err
	}

	logger.Infof(ctx, "Serving DataCatalog Insecure on port %v", config.GetConfig().GetGrpcHostAddress())
	return grpcServer.Serve(grpcListener)
}

// Creates a new GRPC Server with all the configuration
func newGRPCServer(_ context.Context, cfg *config.Config) *grpc.Server {
	tracerProvider := otelutils.GetTracerProvider(otelutils.DataCatalogServerTracer)
	opts := []grpc.ServerOption{grpc.UnaryInterceptor(
		otelgrpc.UnaryServerInterceptor(
			otelgrpc.WithTracerProvider(tracerProvider),
			otelgrpc.WithPropagators(propagation.TraceContext{}),
		),
	)}
	if cfg.GrpcMaxRecvMsgSizeMBs > 0 {
		opts = append(opts, grpc.MaxRecvMsgSize(cfg.GrpcMaxRecvMsgSizeMBs*1024*1024))
	}
	grpcServer := grpc.NewServer(opts...)
	catalog.RegisterDataCatalogServer(grpcServer, NewDataCatalogService())

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

	logger.Infof(ctx, "Serving DataCatalog http on port %v", cfg.GetHTTPHostAddress())

	server := &http.Server{
		Addr:              cfg.GetHTTPHostAddress(),
		Handler:           mux,
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeoutSeconds) * time.Second,
	}
	return server.ListenAndServe()
}

// Create and start the gRPC server and http healthcheck endpoint
func Serve(ctx context.Context, cfg *config.Config) error {
	grpcServer := newGRPCDummyServer(ctx, cfg)

	grpcListener, err := net.Listen("tcp", cfg.GetGrpcHostAddress())
	if err != nil {
		return err
	}

	logger.Infof(ctx, "Serving DataCatalog Insecure on port %v", cfg.GetGrpcHostAddress())
	return grpcServer.Serve(grpcListener)
}

// Creates a new GRPC Server with all the configuration
func newGRPCDummyServer(_ context.Context, cfg *config.Config) *grpc.Server {
	tracerProvider := otelutils.GetTracerProvider(otelutils.DataCatalogServerTracer)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			otelgrpc.UnaryServerInterceptor(
				otelgrpc.WithTracerProvider(tracerProvider),
				otelgrpc.WithPropagators(propagation.TraceContext{}),
			),
		),
	)
	catalog.RegisterDataCatalogServer(grpcServer, &DataCatalogService{})
	if cfg.GrpcServerReflection {
		reflection.Register(grpcServer)
	}
	return grpcServer
}
