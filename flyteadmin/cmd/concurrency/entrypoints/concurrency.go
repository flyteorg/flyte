package entrypoints

import (
	"context"

	"github.com/spf13/cobra"
	_ "gorm.io/driver/postgres" // Required to import database driver.

	"github.com/flyteorg/flyte/flyteadmin/concurrency"
	"github.com/flyteorg/flyte/flyteadmin/concurrency/executor"
	"github.com/flyteorg/flyte/flyteadmin/concurrency/informer"
	concurrencyRepo "github.com/flyteorg/flyte/flyteadmin/concurrency/repositories/gorm"
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flyte/flyteadmin/pkg/server"
	"github.com/flyteorg/flyte/flyteidl/clients/go/admin"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/profutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

var concurrencyRunCmd = &cobra.Command{
	Use:   "run",
	Short: "This command will start the flyte concurrency controller to manage execution concurrency",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		configuration := runtime.NewConfigurationProvider()
		concurrencyConfig := configuration.ApplicationConfiguration().GetConcurrencyConfig()

		// Serve profiling endpoints.
		go func() {
			err := profutils.StartProfilingServerWithDefaultHandlers(
				ctx, concurrencyConfig.ProfilerPort.Port, nil)
			if err != nil {
				logger.Panicf(ctx, "Failed to Start profiling and Metrics server. Error: %v", err)
			}
		}()

		// Setup metrics
		applicationConfiguration := configuration.ApplicationConfiguration().GetTopLevelConfig()
		server.SetMetricKeys(applicationConfiguration)
		scope := promutils.NewScope(applicationConfiguration.MetricsScope)
		concurrencyScope := scope.NewSubScope("concurrency")

		// Create client for Admin service
		adminClient, err := admin.ClientSetBuilder().
			WithConfig(admin.GetConfig(ctx)).
			Build(ctx)
		if err != nil {
			logger.Errorf(ctx, "Failed to build Admin client: %v", err)
			return err
		}

		// Initialize database
		db, err := server.GetDB(ctx, configuration.ApplicationConfiguration())
		if err != nil {
			logger.Errorf(ctx, "Failed to establish database connection: %v", err)
			return err
		}

		// Create concurrency repository
		repo := concurrencyRepo.NewGormConcurrencyRepo(
			db,
			server.GetRepoConfig(configuration.ApplicationConfiguration()),
			concurrencyScope,
		)

		// Create launch plan informer
		launchPlanInformer := informer.NewLaunchPlanInformer(
			repo,
			concurrencyConfig.RefreshInterval.Duration,
			concurrencyScope,
		)

		// Create workflow executor
		workflowExecutor := executor.NewWorkflowExecutor(
			repo,
			adminClient.AdminClient,
			concurrencyScope,
		)

		// Create concurrency controller
		controller := concurrency.NewConcurrencyController(
			repo,
			workflowExecutor,
			launchPlanInformer,
			concurrencyConfig.ProcessingInterval.Duration,
			concurrencyConfig.Workers,
			concurrencyConfig.MaxRetries,
			concurrencyScope,
		)

		// Initialize and start the controller
		err = controller.Initialize(ctx)
		if err != nil {
			logger.Errorf(ctx, "Failed to initialize concurrency controller: %v", err)
			return err
		}

		logger.Infof(ctx, "Concurrency controller started successfully")

		// Block indefinitely
		select {}
	},
}

func init() {
	RootCmd.AddCommand(concurrencyRunCmd)
}
