package entrypoints

import (
	"context"

	errors2 "github.com/pkg/errors"
	"github.com/spf13/cobra"
	_ "gorm.io/driver/postgres" // Required to import database driver.

	"github.com/flyteorg/flyte/flyteadmin/pkg/clusterresource"
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/profutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

var parentClusterResourceCmd = &cobra.Command{
	Use:   "clusterresource",
	Short: "This command administers the ClusterResourceController. Please choose a subcommand.",
}

var controllerRunCmd = &cobra.Command{
	Use:   "run",
	Short: "This command will start a cluster resource controller to periodically sync cluster resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		configuration := runtime.NewConfigurationProvider()
		scope := promutils.NewScope(configuration.ApplicationConfiguration().GetTopLevelConfig().MetricsScope).NewSubScope("clusterresource")
		clusterResourceController, err := clusterresource.NewClusterResourceControllerFromConfig(ctx, scope, configuration)
		if err != nil {
			return err
		}

		// Serve profiling endpoints.
		cfg := runtime.NewConfigurationProvider()
		go func() {
			err := profutils.StartProfilingServerWithDefaultHandlers(
				ctx, cfg.ApplicationConfiguration().GetTopLevelConfig().GetProfilerPort(), nil)
			if err != nil {
				logger.Panicf(ctx, "Failed to Start profiling and Metrics server. Error, %v", err)
			}
		}()

		clusterResourceController.Run()
		logger.Infof(ctx, "ClusterResourceController started running successfully")
		return nil
	},
}

var controllerSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "This command will sync cluster resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		configuration := runtime.NewConfigurationProvider()
		scope := promutils.NewScope(configuration.ApplicationConfiguration().GetTopLevelConfig().MetricsScope).NewSubScope("clusterresource")
		clusterResourceController, err := clusterresource.NewClusterResourceControllerFromConfig(ctx, scope, configuration)
		if err != nil {
			return err
		}
		err = clusterResourceController.Sync(ctx)
		if err != nil {
			return errors2.Wrap(err, "Failed to sync cluster resources ")
		}
		logger.Infof(ctx, "ClusterResourceController synced successfully")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(parentClusterResourceCmd)
	parentClusterResourceCmd.AddCommand(controllerRunCmd)
	parentClusterResourceCmd.AddCommand(controllerSyncCmd)
}
