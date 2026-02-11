package runs

import (
	"context"
	"fmt"
	"net/http"

	"github.com/flyteorg/flyte/v2/app"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task/taskconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/runs/config"
	"github.com/flyteorg/flyte/v2/runs/migrations"
	"github.com/flyteorg/flyte/v2/runs/repository"
	"github.com/flyteorg/flyte/v2/runs/service"
	statek8s "github.com/flyteorg/flyte/v2/state/k8s"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// Setup registers Run and Task service handlers on the SetupContext mux.
// Requires sc.DB, sc.DataStore, and sc.K8sClient to be set.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()

	if err := migrations.RunMigrations(sc.DB); err != nil {
		return fmt.Errorf("runs: failed to run migrations: %w", err)
	}

	repo := repository.NewRepository(sc.DB)

	// In unified mode, intra-service calls go through the same mux.
	queueURL := cfg.QueueServiceURL
	if sc.BaseURL != "" {
		queueURL = sc.BaseURL
	}
	queueClient := workflowconnect.NewQueueServiceClient(
		http.DefaultClient,
		queueURL,
	)

	// Create a state client for watching TaskAction CRs
	if err := statek8s.InitScheme(); err != nil {
		return fmt.Errorf("runs: failed to initialize k8s scheme: %w", err)
	}
	stateClient := statek8s.NewStateClient(sc.K8sClient, sc.Namespace, cfg.WatchBufferSize)
	if err := stateClient.StartWatching(ctx); err != nil {
		return fmt.Errorf("runs: failed to start TaskAction watcher: %w", err)
	}

	runsSvc := service.NewRunService(repo, queueClient, cfg.StoragePrefix, sc.DataStore, stateClient)
	taskSvc := service.NewTaskService(repo)

	runsPath, runsHandler := workflowconnect.NewRunServiceHandler(runsSvc)
	sc.Mux.Handle(runsPath, runsHandler)
	logger.Infof(ctx, "Mounted RunService at %s", runsPath)

	taskPath, taskHandler := taskconnect.NewTaskServiceHandler(taskSvc)
	sc.Mux.Handle(taskPath, taskHandler)
	logger.Infof(ctx, "Mounted TaskService at %s", taskPath)

	sc.AddReadyCheck(func(r *http.Request) error {
		sqlDB, err := sc.DB.DB()
		if err != nil {
			return fmt.Errorf("database connection error: %w", err)
		}
		if err := sqlDB.Ping(); err != nil {
			return fmt.Errorf("database ping failed: %w", err)
		}
		return nil
	})

	return nil
}
