package shared

import (
	stdlibLogger "github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/spf13/cobra"
)
\
// NewMigrateCmd represents the migrate command
func NewMigrateCmd(migs []*gormigrate.Migration) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "This command will run all the migrations for the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			return migrions.Migrate(confi.GetDBConfig(), stdlibLogger.GetConfig(), promutils.NewScope("dbmigrate"), migrs)
		},
	}
}
