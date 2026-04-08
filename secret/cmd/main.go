package main

import (
	"context"
	"fmt"
	"os"

	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/secret"
	secretconfig "github.com/flyteorg/flyte/v2/secret/config"
)

func main() {
	a := &app.App{
		Name:  "secret-service",
		Short: "Secret Service for Flyte",
		Setup: func(ctx context.Context, sc *app.SetupContext) error {
			cfg := secretconfig.GetConfig()
			sc.Host = cfg.Server.Host
			sc.Port = cfg.Server.Port

			k8sClient, _, err := app.InitKubernetesClient(ctx, app.K8sConfig{
				KubeConfig: cfg.Kubernetes.KubeConfig,
				Namespace:  cfg.Kubernetes.Namespace,
				QPS:        cfg.Kubernetes.QPS,
				Burst:      cfg.Kubernetes.Burst,
				Timeout:    cfg.Kubernetes.Timeout,
			}, nil)
			if err != nil {
				return fmt.Errorf("failed to initialize Kubernetes client: %w", err)
			}
			sc.K8sClient = k8sClient
			sc.Namespace = cfg.Kubernetes.Namespace

			return secret.Setup(ctx, sc)
		},
	}
	if err := a.Run(); err != nil {
		os.Exit(1)
	}
}
