package main

import (
	"context"
	"fmt"
	"os"

	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/cache_service"
	cacheserviceconfig "github.com/flyteorg/flyte/v2/cache_service/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
)

func main() {
	a := &app.App{
		Name:  "cache-service",
		Short: "Cache Service for Flyte",
		Setup: func(ctx context.Context, sc *app.SetupContext) error {
			cfg := cacheserviceconfig.GetConfig()
			sc.Host = cfg.Server.Host
			sc.Port = cfg.Server.Port

			db, err := database.GetDB(ctx, database.GetConfig())
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			sc.DB = db

			labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
			dataStore, err := storage.NewDataStore(storage.GetConfig(), promutils.NewScope("cache-service"))
			if err != nil {
				return fmt.Errorf("failed to initialize storage: %w", err)
			}
			sc.DataStore = dataStore

			return cache_service.Setup(ctx, sc)
		},
	}

	if err := a.Run(); err != nil {
		os.Exit(1)
	}
}
