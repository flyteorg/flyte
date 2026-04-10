package main

import (
	"context"
	"fmt"
	"os"

	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/dataproxy"
	"github.com/flyteorg/flyte/v2/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
)

func main() {
	a := &app.App{
		Name:  "dataproxy-service",
		Short: "Data Proxy Service for Flyte",
		Setup: func(ctx context.Context, sc *app.SetupContext) error {
			sc.Host = "0.0.0.0"
			sc.Port = 8088

			labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
			dataStore, err := storage.NewDataStore(storage.GetConfig(), promutils.NewTestScope())
			if err != nil {
				return fmt.Errorf("failed to initialize storage: %w", err)
			}
			sc.DataStore = dataStore

			return dataproxy.Setup(ctx, sc)
		},
	}
	if err := a.Run(); err != nil {
		os.Exit(1)
	}
}
