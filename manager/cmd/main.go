package main

import (
	"context"
	"os"

	stdlibapp "github.com/flyteorg/flyte/v2/flytestdlib/app"
)

func main() {
	component := componentAll
	a := &stdlibapp.App{
		Name:  "flyte",
		Short: "Flyte service manager",
		Setup: func(ctx context.Context, sc *stdlibapp.SetupContext) error {
			return setupComponent(ctx, sc, component)
		},
	}

	cmd := a.Command()
	cmd.Flags().StringVar(&component, "component", componentAll, "component to run: all, runs, actions, events, secret, cache, app, dataproxy, or executor")
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
