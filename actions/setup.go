package actions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/flyteorg/flyte/v2/actions/config"
	actionsk8s "github.com/flyteorg/flyte/v2/actions/k8s"
	"github.com/flyteorg/flyte/v2/actions/service"
	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

// Setup registers the ActionsService handler on the SetupContext mux.
// Requires sc.K8sClient and sc.Namespace to be set.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()

	if err := actionsk8s.InitScheme(); err != nil {
		return fmt.Errorf("actions: failed to initialize scheme: %w", err)
	}

	runServiceURL := cfg.RunServiceURL
	if sc.BaseURL != "" {
		runServiceURL = sc.BaseURL
	}
	runClient := workflowconnect.NewInternalRunServiceClient(http.DefaultClient, runServiceURL)

	actionsClient := actionsk8s.NewActionsClient(
		sc.K8sClient,
		sc.K8sCache,
		cfg.WatchBufferSize,
		cfg.WatchWorkers,
		runClient,
		cfg.RecordFilterSize,
		sc.Scope,
	)
	logger.Infof(ctx, "Actions K8s client initialized")

	if err := actionsClient.StartWatching(ctx); err != nil {
		return fmt.Errorf("actions: failed to start TaskAction watcher: %w", err)
	}
	sc.AddWorker("actions-watcher", func(ctx context.Context) error {
		<-ctx.Done()
		actionsClient.StopWatching()
		return nil
	})

	actionsSvc := service.NewActionsService(actionsClient)

	path, handler := actionsconnect.NewActionsServiceHandler(actionsSvc)
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted ActionsService at %s", path)

	return nil
}
