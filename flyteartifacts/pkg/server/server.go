package server

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/configuration"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/db"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	_ "net/http/pprof" // Required to serve application.
)

type ArtifactService struct {
	artifact.UnimplementedArtifactRegistryServer
	StorageProvider StorageInterface
	Metrics         ServiceMetrics
	Service         CoreService
}

func NewArtifactService(ctx context.Context, scope promutils.Scope) *ArtifactService {
	cfg := configuration.ApplicationConfig.GetConfig().(*configuration.ApplicationConfiguration)
	fmt.Println(cfg)

	storage := db.NewStorage(ctx, scope.NewSubScope("storage:rds"))
	coreService := NewCoreService(storage, scope.NewSubScope("server"))

	return &ArtifactService{
		Metrics: InitMetrics(scope),
		Service: coreService,
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
func GetMigrations(ctx context.Context) []*gormigrate.Migration {
	return db.Migrations
}
