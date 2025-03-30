package main

import (
	"github.com/flyteorg/flyte/flyteadmin/cmd/concurrency/entrypoints"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

func main() {
	// Set up logging and metrics
	logger.SetConfig(&logger.Config{})
	promutils.SetMetricKeys(promutils.RegistererName("flyteconcurrency"))

	// Execute the root command
	err := entrypoints.Execute()
	if err != nil {
		logger.Fatal(err)
	}
}
