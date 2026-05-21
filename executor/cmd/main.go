package main

import (
	"context"
	"fmt"
	"os"

	"github.com/flyteorg/flyte/v2/executor"
	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	a := &app.App{
		Name:  "executor",
		Short: "Executor controller manager for Flyte TaskActions",
		Setup: func(ctx context.Context, sc *app.SetupContext) error {
			// Executor doesn't serve HTTP. Use 0 to avoid binding an app HTTP port.
			sc.Port = 0
			sc.K8sConfig = ctrl.GetConfigOrDie()

			if err := executor.Setup(ctx, sc); err != nil {
				return fmt.Errorf("executor setup failed: %w", err)
			}
			return nil
		},
	}

	if err := a.Run(); err != nil {
		os.Exit(1)
	}
}
