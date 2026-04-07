package main

import (
	"context"
	"os"

	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/events"
	eventsconfig "github.com/flyteorg/flyte/v2/events/config"
)

func main() {
	a := &app.App{
		Name:  "events-service",
		Short: "Events Service for Flyte",
		Setup: func(ctx context.Context, sc *app.SetupContext) error {
			cfg := eventsconfig.GetConfig()
			sc.Host = cfg.Server.Host
			sc.Port = cfg.Server.Port

			return events.Setup(ctx, sc)
		},
	}
	if err := a.Run(); err != nil {
		os.Exit(1)
	}
}
