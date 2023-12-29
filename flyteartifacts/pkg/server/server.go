package server

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/blob"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/configuration"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/db"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/server/processor"
	admin2 "github.com/flyteorg/flyte/flyteidl/clients/go/admin"
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
	Metrics         ServiceMetrics
	Service         CoreService
	EventConsumer   processor.EventsProcessorInterface
	TriggerEngine   TriggerHandlerInterface
	eventsToTrigger chan artifact.Artifact
}

func (a *ArtifactService) CreateArtifact(ctx context.Context, req *artifact.CreateArtifactRequest) (*artifact.CreateArtifactResponse, error) {
	resp, err := a.Service.CreateArtifact(ctx, req)
	if err != nil {
		return resp, err
	}

	// gatepr add go func
	execIDs, err := a.TriggerEngine.EvaluateNewArtifact(ctx, resp.Artifact)
	if err != nil {
		logger.Warnf(ctx, "Failed to evaluate triggers for artifact: %v, err: %v", resp.Artifact, err)
	} else {
		logger.Infof(ctx, "Triggered %v executions", len(execIDs))
	}
	return resp, err
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

func (a *ArtifactService) SetExecutionInputs(ctx context.Context, req *artifact.ExecutionInputsRequest) (*artifact.ExecutionInputsResponse, error) {
	return a.Service.SetExecutionInputs(ctx, req)
}

func (a *ArtifactService) FindByWorkflowExec(ctx context.Context, req *artifact.FindByWorkflowExecRequest) (*artifact.SearchArtifactsResponse, error) {
	return a.Service.FindByWorkflowExec(ctx, req)
}
func (a *ArtifactService) runProcessor(ctx context.Context) {
	for {
		select {
		case art := <-a.eventsToTrigger:
			logger.Infof(ctx, "Received artifact: %v", art)
			execIDs, err := a.TriggerEngine.EvaluateNewArtifact(ctx, &art)
			if err != nil {
				logger.Warnf(ctx, "Failed to evaluate triggers for artifact: %v, err: %v", art, err)
			} else {
				logger.Infof(ctx, "Triggered %v executions", len(execIDs))
			}
		case <-ctx.Done():
			logger.Infof(ctx, "Stopping artifact processor")
			return
		}
	}
}

func NewArtifactService(ctx context.Context, scope promutils.Scope) *ArtifactService {
	cfg := configuration.GetApplicationConfig()
	fmt.Println(cfg)
	eventsCfg := configuration.GetEventsProcessorConfig()

	// channel that the event processor should use to pass created artifacts to, which this artifact server will
	// then call the trigger engine with.
	createdArtifacts := make(chan artifact.Artifact, 1000)

	storage := db.NewStorage(ctx, scope.NewSubScope("storage:rds"))
	blobStore := blob.NewArtifactBlobStore(ctx, scope.NewSubScope("storage:s3"))
	coreService := NewCoreService(storage, &blobStore, scope.NewSubScope("server"))
	triggerHandler, err := NewTriggerEngine(ctx, storage, &coreService, scope.NewSubScope("triggers"))
	if err != nil {
		logger.Errorf(ctx, "Failed to create Trigger engine, stopping server. Error: %v", err)
		panic(err)
	}

	adminClientCfg := admin2.GetConfig(ctx)
	clientSet, err := admin2.NewClientsetBuilder().WithConfig(adminClientCfg).Build(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to create Admin client set, stopping server. Error: %v", err)
		panic(err)
	}

	handler := processor.NewServiceCallHandler(ctx, &coreService, createdArtifacts, *clientSet)
	eventsReceiverAndHandler := processor.NewBackgroundProcessor(ctx, *eventsCfg, &coreService, createdArtifacts, scope.NewSubScope("events"))
	if eventsReceiverAndHandler != nil {
		go func() {
			logger.Info(ctx, "Starting Artifact service background processing...")
			eventsReceiverAndHandler.StartProcessing(ctx, handler)
		}()
	}

	as := &ArtifactService{
		Metrics:         InitMetrics(scope),
		Service:         coreService,
		EventConsumer:   eventsReceiverAndHandler,
		TriggerEngine:   &triggerHandler,
		eventsToTrigger: createdArtifacts,
	}

	go as.runProcessor(ctx)

	return as
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
