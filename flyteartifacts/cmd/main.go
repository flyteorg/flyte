package main

import (
	"context"
	sharedCmd "github.com/flyteorg/flyte/flyteartifacts/cmd/shared"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/server"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"

	_ "net/http/pprof" // Required to serve application.
)

func main() {
	ctx := context.Background()
	logger.Infof(ctx, "Beginning Flyte Artifacts Service")
	rootCmd := sharedCmd.NewRootCmd("artifacts", server.GrpcRegistrationHook, server.HttpRegistrationHook)
	migs := server.GetMigrations(ctx)
	dbConfig := server.GetDbConfig()
	rootCmd.AddCommand(sharedCmd.NewMigrateCmd(migs, dbConfig))
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		panic(err)
	}
}

func init() {
	// Set Keys
	labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey,
		contextutils.DomainKey, storage.FailureTypeLabel)
}
