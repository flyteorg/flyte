package server

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/blob"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/configuration"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/db"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/server/processor"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyte/flytestdlib/database"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	_ "net/http/pprof" // Required to serve application.
)

type ArtifactService struct {
	artifact.UnimplementedArtifactRegistryServer
	Metrics       ServiceMetrics
	Service       CoreService
	EventConsumer processor.EventsProcessorInterface
}

func (a *ArtifactService) CreateArtifact(ctx context.Context, req *artifact.CreateArtifactRequest) (*artifact.CreateArtifactResponse, error) {
	return a.Service.CreateArtifact(ctx, req)
}

func (a *ArtifactService) GetArtifact(ctx context.Context, req *artifact.GetArtifactRequest) (*artifact.GetArtifactResponse, error) {
	return a.Service.GetArtifact(ctx, req)
}

func (a *ArtifactService) SearchArtifacts(ctx context.Context, req *artifact.SearchArtifactsRequest) (*artifact.SearchArtifactsResponse, error) {
	return a.Service.SearchArtifacts(ctx, req)
}

func (a *ArtifactService) CreateTrigger(ctx context.Context, req *artifact.CreateTriggerRequest) (*artifact.CreateTriggerResponse, error) {
	return a.Service.CreateTrigger(ctx, req)
}

func (a *ArtifactService) DeleteTrigger(ctx context.Context, req *artifact.DeleteTriggerRequest) (*artifact.DeleteTriggerResponse, error) {
	return a.Service.DeleteTrigger(ctx, req)
}

func (a *ArtifactService) AddTag(ctx context.Context, req *artifact.AddTagRequest) (*artifact.AddTagResponse, error) {
	return a.Service.AddTag(ctx, req)
}

func (a *ArtifactService) RegisterProducer(ctx context.Context, req *artifact.RegisterProducerRequest) (*artifact.RegisterResponse, error) {
	return a.Service.RegisterProducer(ctx, req)
}

func (a *ArtifactService) RegisterConsumer(ctx context.Context, req *artifact.RegisterConsumerRequest) (*artifact.RegisterResponse, error) {
	return a.Service.RegisterConsumer(ctx, req)
}

func NewArtifactService(ctx context.Context, scope promutils.Scope) *ArtifactService {
	cfg := configuration.GetApplicationConfig()
	fmt.Println(cfg)
	eventsCfg := configuration.GetEventsProcessorConfig()

	storage := db.NewStorage(ctx, scope.NewSubScope("storage:rds"))
	blobStore := blob.NewArtifactBlobStore(ctx, scope.NewSubScope("storage:s3"))
	coreService := NewCoreService(storage, &blobStore, scope.NewSubScope("server"))
	eventsReceiverAndHandler := processor.NewBackgroundProcessor(ctx, *eventsCfg, &coreService, scope.NewSubScope("events"))
	if eventsReceiverAndHandler != nil {
		go func() {
			logger.Info(ctx, "Starting Artifact service background processing...")
			eventsReceiverAndHandler.StartProcessing(ctx)
		}()
	}

	return &ArtifactService{
		Metrics:       InitMetrics(scope),
		Service:       coreService,
		EventConsumer: eventsReceiverAndHandler,
	}
}

func HttpRegistrationHook(ctx context.Context, gwmux *runtime.ServeMux, grpcAddress string, grpcConnectionOpts []grpc.DialOption, _ promutils.Scope) error {
	err := artifact.RegisterArtifactRegistryHandlerFromEndpoint(ctx, gwmux, grpcAddress, grpcConnectionOpts)
	if err != nil {
		return errors.Wrap(err, "error registering execution service")
	}
	return nil
}

func GrpcRegistrationHook(ctx context.Context, server *grpc.Server, scope promutils.Scope) error {
	serviceImpl := NewArtifactService(ctx, scope)
	artifact.RegisterArtifactRegistryServer(server, serviceImpl)

	return nil
}

// GetMigrations should be hidden behind the storage interface in the future.
func GetMigrations(_ context.Context) []*gormigrate.Migration {
	return db.Migrations
}
func GetDbConfig() *database.DbConfig {
	cfg := configuration.ApplicationConfig.GetConfig().(*configuration.ApplicationConfiguration)
	return &cfg.ArtifactDatabaseConfig
}
