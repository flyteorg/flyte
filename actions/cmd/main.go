package main

import (
	"context"
	"fmt"
	"os"

	"github.com/flyteorg/flyte/v2/actions"
	actionsconfig "github.com/flyteorg/flyte/v2/actions/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/app"
)

func main() {
	a := &app.App{
		Name:  "actions-service",
		Short: "Actions Service for Flyte",
		Setup: func(ctx context.Context, sc *app.SetupContext) error {
			cfg := actionsconfig.GetConfig()
			sc.Host = cfg.Server.Host
			sc.Port = cfg.Server.Port

			k8sClient, _, err := app.InitKubernetesClient(ctx, app.K8sConfig{
				KubeConfig: cfg.Kubernetes.KubeConfig,
				Namespace:  cfg.Kubernetes.Namespace,
			}, nil)
			if err != nil {
				return fmt.Errorf("failed to initialize Kubernetes client: %w", err)
			}
			sc.K8sClient = k8sClient
			sc.Namespace = cfg.Kubernetes.Namespace

			return actions.Setup(ctx, sc)
		},
	}
	if err := a.Run(); err != nil {
		os.Exit(1)
	}
}
