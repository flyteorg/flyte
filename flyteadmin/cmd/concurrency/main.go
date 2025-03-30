package main

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/cmd/concurrency/entrypoints"
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flyte/flyteadmin/pkg/server"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

// Define a custom metrics initialization function since promutils.SetMetricKeys is undefined
func initMetrics() {
	// Initialize metrics scope with a custom registerer
	// This is a simplified version - in production, you would configure this more thoroughly
	_ = promutils.NewScope("flyteconcurrency")
}

func main() {
	ctx := context.Background()

	// Set up logging
	logger.SetConfig(&logger.Config{})

	// Initialize metrics
	appConfig := runtime.NewConfigurationProvider().ApplicationConfiguration().GetTopLevelConfig()
	server.SetMetricKeys(appConfig)

	// Execute the root command
	err := entrypoints.Execute()
	if err != nil {
		logger.Fatalf(ctx, "Failed to execute command: %v", err)
	}
}
