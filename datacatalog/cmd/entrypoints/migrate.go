package entrypoints

import (
	"github.com/flyteorg/datacatalog/pkg/repositories"

	"context"

	"github.com/spf13/cobra"
)

var parentMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "This command controls migration behavior for the Flyte Catalog database. Please choose a subcommand.",
}

// This runs all the migrations
var migrateCmd = &cobra.Command{
	Use:   "run",
	Short: "This command will run all the migrations for the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return repositories.Migrate(ctx)
	},
}

func init() {
	RootCmd.AddCommand(parentMigrateCmd)
	parentMigrateCmd.AddCommand(migrateCmd)
}
