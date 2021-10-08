package entrypoints

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/flyteorg/flyteadmin/pkg/common"
	repositoryCommonConfig "github.com/flyteorg/flyteadmin/pkg/repositories/config"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flyteadmin/scheduler"
	schdulerRepoConfig "github.com/flyteorg/flyteadmin/scheduler/repositories"
	"github.com/flyteorg/flyteidl/clients/go/admin"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/profutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"

	_ "github.com/jinzhu/gorm/dialects/postgres" // Required to import database driver.
	"github.com/spf13/cobra"
)

var schedulerRunCmd = &cobra.Command{
	Use:   "run",
	Short: "This command will start the flyte native scheduler and periodically get new schedules from the db for scheduling",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		configuration := runtime.NewConfigurationProvider()
		applicationConfiguration := configuration.ApplicationConfiguration().GetTopLevelConfig()
		schedulerConfiguration := configuration.ApplicationConfiguration().GetSchedulerConfig()

		// Define the schedulerScope for prometheus metrics
		schedulerScope := promutils.NewScope(applicationConfiguration.MetricsScope).NewSubScope("flytescheduler")
		schedulerPanics := schedulerScope.MustNewCounter("initialization_panic",
			"panics encountered initializing the flyte native scheduler")

		defer func() {
			if err := recover(); err != nil {
				schedulerPanics.Inc()
				logger.Fatalf(ctx, fmt.Sprintf("caught panic: %v [%+v]", err, string(debug.Stack())))
			}
		}()

		dbConfigValues := configuration.ApplicationConfiguration().GetDbConfig()
		dbConfig := repositoryCommonConfig.NewDbConfig(dbConfigValues)
		db := schdulerRepoConfig.GetRepository(
			schdulerRepoConfig.POSTGRES, dbConfig, schedulerScope.NewSubScope("database"))

		clientSet, err := admin.ClientSetBuilder().WithConfig(admin.GetConfig(ctx)).Build(ctx)
		if err != nil {
			logger.Fatalf(ctx, "Flyte native scheduler failed to start due to %v", err)
			return err
		}
		adminServiceClient := clientSet.AdminClient()

		scheduleExecutor := scheduler.NewScheduledExecutor(db,
			configuration.ApplicationConfiguration().GetSchedulerConfig().GetWorkflowExecutorConfig(), schedulerScope, adminServiceClient)

		logger.Info(ctx, "Successfully initialized a native flyte scheduler")

		// Serve profiling endpoints.
		go func() {
			err := profutils.StartProfilingServerWithDefaultHandlers(
				ctx, schedulerConfiguration.ProfilerPort.Port, nil)
			if err != nil {
				logger.Panicf(ctx, "Failed to Start profiling and Metrics server. Error, %v", err)
			}
		}()

		err = scheduleExecutor.Run(ctx)
		if err != nil {
			logger.Fatalf(ctx, "Flyte native scheduler failed to start due to %v", err)
			return err
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(schedulerRunCmd)

	// Set Keys
	labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey,
		contextutils.ExecIDKey, contextutils.WorkflowIDKey, contextutils.NodeIDKey, contextutils.TaskIDKey,
		contextutils.TaskTypeKey, common.RuntimeTypeKey, common.RuntimeVersionKey)
}
