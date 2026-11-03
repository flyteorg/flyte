package main

import (
	"context"
	"fmt"
	"os"

	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/runs"
	runsconfig "github.com/flyteorg/flyte/v2/runs/config"
)

func main() {
	a := &app.App{
		Name:  runs.ServiceName,
		Short: "Runs Service for Flyte",
		Setup: func(ctx context.Context, sc *app.SetupContext) error {
			cfg := runsconfig.GetConfig()
			sc.Host = cfg.Server.Host
			sc.Port = cfg.Server.Port

			if err := otelutils.RegisterProvidersWithContext(ctx, runs.ServiceName, otelutils.GetConfig()); err != nil {
				return fmt.Errorf("failed to initialize database telemetry: %w", err)
			}
			db, err := database.GetDB(ctx, database.GetConfig(), runs.ServiceName)
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			sc.DB = db

			labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
			dataStore, err := storage.NewDataStore(storage.GetConfig(), promutils.NewTestScope())
			if err != nil {
				return fmt.Errorf("failed to initialize storage: %w", err)
			}
			sc.DataStore = dataStore

			return runs.Setup(ctx, sc)
		},
	}
	if err := a.Run(); err != nil {
		os.Exit(1)
	}
}
