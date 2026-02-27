package actions

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyte/v2/actions/config"
	"github.com/flyteorg/flyte/v2/actions/service"
	"github.com/flyteorg/flyte/v2/app"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// Setup registers the ActionsService handler on the SetupContext mux.
// Requires sc.K8sClient and sc.Namespace to be set.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()
	_ = cfg // will be used when the k8s client is implemented

	// TODO: initialize k8s client once actions/k8s package is implemented.
	// For now, this is a placeholder that registers the service with a nil client.
	_ = fmt.Sprintf("actions: initializing for namespace: %s", sc.Namespace)
	logger.Infof(ctx, "Actions service initializing for namespace: %s", sc.Namespace)

	actionsSvc := service.NewActionsService(nil)

	path, handler := actionsconnect.NewActionsServiceHandler(actionsSvc)
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted ActionsService at %s", path)

	return nil
}
