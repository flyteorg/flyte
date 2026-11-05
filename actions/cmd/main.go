package main

import (
	"context"
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes/scheme"

	"github.com/flyteorg/flyte/v2/actions"
	actionsconfig "github.com/flyteorg/flyte/v2/actions/config"
	actionsk8s "github.com/flyteorg/flyte/v2/actions/k8s"
	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

func main() {
	a := &app.App{
		Name:  "actions-service",
		Short: "Actions Service for Flyte",
		Setup: func(ctx context.Context, sc *app.SetupContext) error {
			cfg := actionsconfig.GetConfig()
			sc.Host = cfg.Server.Host
			sc.Port = cfg.Server.Port
			sc.Scope = promutils.NewScope("actions-service")

			if err := actionsk8s.InitScheme(); err != nil {
				return fmt.Errorf("failed to initialize Kubernetes scheme: %w", err)
			}
			k8sClient, k8sConfig, err := app.InitKubernetesClient(ctx, cfg.Kubernetes, scheme.Scheme)
			if err != nil {
				return fmt.Errorf("failed to initialize Kubernetes client: %w", err)
			}
			k8sCache, err := app.InitKubernetesCache(k8sConfig, scheme.Scheme)
			if err != nil {
				return err
			}
			sc.K8sClient = k8sClient
			sc.K8sConfig = k8sConfig
			sc.K8sCache = k8sCache
			sc.Namespace = cfg.Kubernetes.Namespace
			sc.AddWorker("kubernetes-cache", k8sCache.Start)

			return actions.Setup(ctx, sc)
		},
	}
	if err := a.Run(); err != nil {
		os.Exit(1)
	}
}
