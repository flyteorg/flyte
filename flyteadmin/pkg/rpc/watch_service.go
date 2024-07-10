package rpc

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/config"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	runtimeIfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/watch"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type WatchService struct {
	manager interfaces.WatchInterface
}

func (e *WatchService) WatchExecutionStatusUpdates(req *watch.WatchExecutionStatusUpdatesRequest, srv service.WatchService_WatchExecutionStatusUpdatesServer) error {
	return e.manager.WatchExecutionStatusUpdates(req, srv)
}

func NewWatchService(ctx context.Context, cfg *config.ServerConfig, scope promutils.Scope, conf runtimeIfaces.Configuration) *WatchService {
	databaseConfig := conf.ApplicationConfiguration().GetDbConfig()
	logConfig := logger.GetConfig()

	db, err := repositories.GetDB(ctx, databaseConfig, logConfig)
	if err != nil {
		logger.Fatal(ctx, err)
	}
	dbScope := scope.NewSubScope("database")
	repo := repositories.NewGormRepo(
		db, errors.NewPostgresErrorTransformer(scope.NewSubScope("errors")), dbScope)

	eventManager := impl.NewWatchManager(cfg, repo, scope.NewSubScope("watch_manager"))
	return &WatchService{manager: eventManager}
}
