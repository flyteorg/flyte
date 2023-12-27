package shared

import (
	"context"
	"github.com/flyteorg/flyte/flytestdlib/database"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/spf13/cobra"
)

// NewMigrateCmd represents the migrate command
func NewMigrateCmd(migs []*gormigrate.Migration, databaseConfig *database.DbConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "This command will run all the migrations for the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			return database.Migrate(context.Background(), databaseConfig, migs)
		},
	}
}
