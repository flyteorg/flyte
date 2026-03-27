package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/flyteorg/flyte/v2/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/runs/migrations"
)

func main() {
	a := &app.App{
		Name:  "runs-migrate",
		Short: "Run runs-service database migrations",
	}
	cmd := a.Command()
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		if err := logger.SetConfig(logger.GetConfig()); err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		if err := database.Migrate(ctx, database.GetConfig(), migrations.RunsMigrations); err != nil {
			return fmt.Errorf("failed to run runs migrations: %w", err)
		}
		return nil
	}

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
