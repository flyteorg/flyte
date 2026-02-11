package main

import (
	"context"
	"fmt"
	"os"

	"github.com/flyteorg/flyte/v2/app"
	"github.com/flyteorg/flyte/v2/queue"
	queueconfig "github.com/flyteorg/flyte/v2/queue/config"
)

func main() {
	a := &app.App{
		Name:  "queue-service",
		Short: "Queue Service for Flyte",
		Setup: func(ctx context.Context, sc *app.SetupContext) error {
			cfg := queueconfig.GetConfig()
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

			return queue.Setup(ctx, sc)
		},
	}
	if err := a.Run(); err != nil {
		os.Exit(1)
	}
}
