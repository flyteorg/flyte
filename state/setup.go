package state

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyte/v2/app"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/state/config"
	"github.com/flyteorg/flyte/v2/state/k8s"
	"github.com/flyteorg/flyte/v2/state/service"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// Setup registers the State service handler on the SetupContext mux.
// Requires sc.K8sClient and sc.Namespace to be set.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()

	if err := k8s.InitScheme(); err != nil {
		return fmt.Errorf("state: failed to initialize scheme: %w", err)
	}

	stateClient := k8s.NewStateClient(sc.K8sClient, sc.Namespace, cfg.WatchBufferSize)
	logger.Infof(ctx, "State K8s client initialized for namespace: %s", sc.Namespace)

	if err := stateClient.StartWatching(ctx); err != nil {
		return fmt.Errorf("state: failed to start TaskAction watcher: %w", err)
	}

	stateSvc := service.NewStateService(stateClient)

	path, handler := workflowconnect.NewStateServiceHandler(stateSvc)
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted StateService at %s (K8s-based)", path)

	return nil
}
