package actions

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyte/v2/actions/config"
	"github.com/flyteorg/flyte/v2/actions/k8s"
	"github.com/flyteorg/flyte/v2/actions/service"
	"github.com/flyteorg/flyte/v2/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect"
)

// Setup registers the ActionsService handler on the SetupContext mux.
// Requires sc.K8sClient and sc.Namespace to be set.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()

	if err := k8s.InitScheme(); err != nil {
		return fmt.Errorf("actions: failed to initialize scheme: %w", err)
	}

	actionsClient := k8s.NewActionsClient(sc.K8sClient, sc.Namespace, cfg.WatchBufferSize)
	logger.Infof(ctx, "Actions K8s client initialized for namespace: %s", sc.Namespace)

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
