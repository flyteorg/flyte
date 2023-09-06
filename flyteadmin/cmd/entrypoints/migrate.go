package entrypoints

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/server"

	"github.com/spf13/cobra"
	_ "gorm.io/driver/postgres" // Required to import database driver.
)

var parentMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "This command controls migration behavior for the Flyte admin database. Please choose a subcommand.",
}

// This runs all the migrations
var migrateCmd = &cobra.Command{
	Use:   "run",
	Short: "This command will run all the migrations for the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return server.Migrate(ctx)
	},
}

// Rollback the latest migration
var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "This command will rollback one migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return server.Rollback(ctx)
	},
}

// This seeds the database with project values
var seedProjectsCmd = &cobra.Command{
	Use:   "seed-projects",
	Short: "Seed projects in the database.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return server.SeedProjects(ctx, args)
	},
}

func init() {
	RootCmd.AddCommand(parentMigrateCmd)
	parentMigrateCmd.AddCommand(migrateCmd)
	parentMigrateCmd.AddCommand(rollbackCmd)
	parentMigrateCmd.AddCommand(seedProjectsCmd)
}
