package main

import (
	"context"
	"fmt"
	"os"

	"github.com/flyteorg/flyte/v2/app"
	"github.com/flyteorg/flyte/v2/executor"
	executorconfig "github.com/flyteorg/flyte/v2/executor/pkg/config"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	a := &app.App{
		Name:  "executor",
		Short: "Executor controller manager for Flyte TaskActions",
		Setup: func(ctx context.Context, sc *app.SetupContext) error {
			cfg := executorconfig.GetConfig()

			// Executor doesn't serve HTTP — it uses controller-runtime's own
			// health probe port. Set a dummy port so the app skeleton starts
			// its HTTP server on a non-conflicting address (or 0 to disable).
			sc.Port = 0

			k8sConfig := ctrl.GetConfigOrDie()
			sc.K8sConfig = k8sConfig

			if err := executor.Setup(ctx, sc); err != nil {
				return fmt.Errorf("executor setup failed: %w", err)
			}

			_ = cfg // config is read inside executor.Setup's worker
			return nil
		},
	}
	if err := a.Run(); err != nil {
		os.Exit(1)
	}
}

