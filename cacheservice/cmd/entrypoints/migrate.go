package entrypoints

import (
	"github.com/spf13/cobra"
)

var parentMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "This command controls migration behavior for the Flyte cacheservice database. Please choose a subcommand.",
}

// This runs all the migrations. This is a placeholder for now as cache service does not have any migrations
var migrateCmd = &cobra.Command{
	Use:   "run",
	Short: "This command will run all the migrations for the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	RootCmd.AddCommand(parentMigrateCmd)
	parentMigrateCmd.AddCommand(migrateCmd)
}
