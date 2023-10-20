package main

import (
	"context"
	sharedCmd "github.com/flyteorg/flyte/flyteartifacts/cmd/shared"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/server"
	"github.com/flyteorg/flyte/flytestdlib/logger"

	_ "net/http/pprof" // Required to serve application.
)

func main() {
	ctx := context.Background()
	logger.Infof(ctx, "Beginning Flyte Artifacts Service")
	rootCmd := sharedCmd.NewRootCmd("artifacts", server.GrpcRegistrationHook, server.HttpRegistrationHook)
	migs := server.GetMigrations(ctx)
	initializationSql := "create extension if not exists hstore;"
	dbConfig := server.GetDbConfig()
	rootCmd.AddCommand(sharedCmd.NewMigrateCmd(migs, dbConfig, initializationSql))
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		panic(err)
	}
}
