package internal

import (
	"context"
	"fmt"
	"net/http"

	stdlibapp "github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"

	appconfig "github.com/flyteorg/flyte/v2/app/internal/config"
	appk8s "github.com/flyteorg/flyte/v2/app/internal/k8s"
	"github.com/flyteorg/flyte/v2/app/internal/migrations"
	repoimpl "github.com/flyteorg/flyte/v2/app/internal/repository/impl"
	"github.com/flyteorg/flyte/v2/app/internal/service"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app/appconnect"
)

// Setup registers the InternalAppService handler on the SetupContext mux.
// It is mounted at /internal<path> to avoid collision with the control plane
// AppService, which shares the same proto service definition.
func Setup(ctx context.Context, sc *stdlibapp.SetupContext, cfg *appconfig.InternalAppConfig) error {
	if !cfg.Enabled {
		logger.Infof(ctx, "InternalAppService disabled (apps.enabled=false), skipping setup")
		return nil
	}

	if err := stdlibapp.InitAppScheme(); err != nil {
		return fmt.Errorf("internalapp: failed to register Knative scheme: %w", err)
	}

	if err := migrations.RunMigrations(ctx, sc.DB); err != nil {
		return fmt.Errorf("internalapp: failed to run migrations: %w", err)
	}

	conditionRepo := repoimpl.NewAppConditionsRepo(sc.DB)
	appK8sClient := appk8s.NewAppK8sClient(sc.K8sClient, sc.K8sCache, cfg, conditionRepo)
	internalAppSvc := service.NewInternalAppService(appK8sClient, conditionRepo)

	if err := appK8sClient.StartWatching(ctx); err != nil {
		return fmt.Errorf("internalapp: failed to start KService watcher: %w", err)
	}
	sc.AddWorker("app-kservice-watcher", func(ctx context.Context) error {
		<-ctx.Done()
		appK8sClient.StopWatching()
		return nil
	})

	path, handler := appconnect.NewAppServiceHandler(internalAppSvc)
	sc.Mux.Handle("/internal"+path, http.StripPrefix("/internal", handler))
	logger.Infof(ctx, "Mounted InternalAppService at /internal%s", path)

	return nil
}
