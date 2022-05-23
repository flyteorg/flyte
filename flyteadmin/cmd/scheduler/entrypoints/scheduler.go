package entrypoints

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/server"

	"github.com/flyteorg/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flyteadmin/scheduler"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/profutils"
	"github.com/spf13/cobra"
	_ "gorm.io/driver/postgres" // Required to import database driver.
)

var schedulerRunCmd = &cobra.Command{
	Use:   "run",
	Short: "This command will start the flyte native scheduler and periodically get new schedules from the db for scheduling",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		schedulerConfiguration := runtime.NewConfigurationProvider().ApplicationConfiguration().GetSchedulerConfig()
		// Serve profiling endpoints.
		go func() {
			err := profutils.StartProfilingServerWithDefaultHandlers(
				ctx, schedulerConfiguration.ProfilerPort.Port, nil)
			if err != nil {
				logger.Panicf(ctx, "Failed to Start profiling and Metrics server. Error, %v", err)
			}
		}()

		configuration := runtime.NewConfigurationProvider()
		applicationConfiguration := configuration.ApplicationConfiguration().GetTopLevelConfig()
		server.SetMetricKeys(applicationConfiguration)

		return scheduler.StartScheduler(ctx)
	},
}

func init() {
	RootCmd.AddCommand(schedulerRunCmd)
}
