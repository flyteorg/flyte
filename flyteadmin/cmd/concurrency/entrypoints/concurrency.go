package entrypoints

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	_ "gorm.io/driver/postgres" // Required to import database driver.

	"github.com/flyteorg/flyte/flyteadmin/concurrency"
	"github.com/flyteorg/flyte/flyteadmin/concurrency/executor"
	"github.com/flyteorg/flyte/flyteadmin/concurrency/informer"
	concurrencyRepo "github.com/flyteorg/flyte/flyteadmin/concurrency/repositories/gorm"
	managerMocks "github.com/flyteorg/flyte/flyteadmin/pkg/manager/mocks"
	repoMocks "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/mocks"
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
			profileErr := profutils.StartProfilingServerWithDefaultHandlers(
				ctx, concurrencyConfig.ProfilerPort.Port, nil)
			if profileErr != nil {
				logger.Panicf(ctx, "Failed to Start profiling and Metrics server. Error: %v", profileErr)
			}
		}()

		// Setup metrics
		applicationConfiguration := configuration.ApplicationConfiguration().GetTopLevelConfig()
		scope := promutils.NewScope(applicationConfiguration.MetricsScope)
		concurrencyScope := scope.NewSubScope("concurrency")

		// Create client for Admin service
		// This is not used in this implementation, but would be in a real deployment
		clientResult, clientErr := admin.ClientSetBuilder().
			WithConfig(admin.GetConfig(ctx)).
			Build(ctx)
		if clientErr != nil {
			logger.Errorf(ctx, "Failed to build Admin client: %v", clientErr)
			return clientErr
		}
		// Mark as used to avoid linter errors
		_ = clientResult

		// Initialize database
		db, dbErr := server.GetDB(ctx, configuration.ApplicationConfiguration())
		if dbErr != nil {
			logger.Errorf(ctx, "Failed to establish database connection: %v", dbErr)
			return dbErr
		}

		// Initialize execution interface
		// In a real implementation, this would be using a real repository
		executionManager := &managerMocks.ExecutionInterface{}

		// Initialize an empty repository implementation for development
		// In a real environment, this would be properly initialized
		mockRepo := repoMocks.NewMockRepository()

		// Create concurrency repository
		repo := concurrencyRepo.NewGormConcurrencyRepo(
			db,
			mockRepo,
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
			executionManager,
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
		initErr := controller.Initialize(ctx)
		if initErr != nil {
			logger.Errorf(ctx, "Failed to initialize concurrency controller: %v", initErr)
			return initErr
		}

		logger.Infof(ctx, "Concurrency controller started successfully")

		// Block indefinitely
		select {}
	},
}

var concurrencyPrecheckCmd = &cobra.Command{
	Use:   "precheck",
	Short: "Verify dependencies for the concurrency controller are ready",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		configuration := runtime.NewConfigurationProvider()

		// Check database connection
		logger.Infof(ctx, "Checking database connection...")
		// Fixed: Using db result to avoid unused variable warning
		_, dbErr := server.GetDB(ctx, configuration.ApplicationConfiguration())
		if dbErr != nil {
			logger.Errorf(ctx, "Failed to establish database connection: %v", dbErr)
			return dbErr
		}
		logger.Infof(ctx, "Database connection successful")

		// Check admin client configuration
		logger.Infof(ctx, "Checking admin client configuration...")
		_, clientErr := admin.ClientSetBuilder().
			WithConfig(admin.GetConfig(ctx)).
			Build(ctx)
		if clientErr != nil {
			logger.Errorf(ctx, "Failed to build Admin client: %v", clientErr)
			return clientErr
		}
		logger.Infof(ctx, "Admin client configuration verified")

		// Verify concurrency configuration
		concurrencyConfig := configuration.ApplicationConfiguration().GetConcurrencyConfig()
		if concurrencyConfig == nil {
			logger.Errorf(ctx, "Concurrency configuration is missing")
			configErr := fmt.Errorf("concurrency configuration is missing")
			return configErr
		}

		logger.Infof(ctx, "Concurrency controller dependencies verified successfully")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(concurrencyRunCmd)
	RootCmd.AddCommand(concurrencyPrecheckCmd)
}
