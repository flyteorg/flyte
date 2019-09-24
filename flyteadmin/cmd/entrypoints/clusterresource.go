package entrypoints

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/clusterresource"

	"github.com/lyft/flyteadmin/pkg/flytek8s"

	"github.com/lyft/flyteadmin/pkg/runtime"

	"github.com/lyft/flytestdlib/logger"

	_ "github.com/jinzhu/gorm/dialects/postgres" // Required to import database driver.
	"github.com/lyft/flyteadmin/pkg/config"
	"github.com/lyft/flyteadmin/pkg/repositories"
	repositoryConfig "github.com/lyft/flyteadmin/pkg/repositories/config"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/spf13/cobra"
)

var parentClusterResourceCmd = &cobra.Command{
	Use:   "clusterresource",
	Short: "This command administers the ClusterResourceController. Please choose a subcommand.",
}

func GetLocalDbConfig() repositoryConfig.DbConfig {
	return repositoryConfig.DbConfig{
		Host:   "localhost",
		Port:   5432,
		DbName: "postgres",
		User:   "postgres",
	}
}

var controllerRunCmd = &cobra.Command{
	Use:   "run",
	Short: "This command will start a cluster resource controller to periodically sync cluster resources",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		configuration := runtime.NewConfigurationProvider()
		scope := promutils.NewScope(configuration.ApplicationConfiguration().GetTopLevelConfig().MetricsScope).NewSubScope("clusterresource")
		dbConfigValues := configuration.ApplicationConfiguration().GetDbConfig()
		dbConfig := repositoryConfig.DbConfig{
			Host:         dbConfigValues.Host,
			Port:         dbConfigValues.Port,
			DbName:       dbConfigValues.DbName,
			User:         dbConfigValues.User,
			Password:     dbConfigValues.Password,
			ExtraOptions: dbConfigValues.ExtraOptions,
		}
		db := repositories.GetRepository(
			repositories.POSTGRES, dbConfig, scope.NewSubScope("database"))

		cfg := config.GetConfig()
		kubeClient, err := flytek8s.NewKubeClient(cfg.KubeConfig, cfg.Master, configuration.ClusterConfiguration())
		if err != nil {
			scope.NewSubScope("flytekubeconfig").MustNewCounter(
				"kubeconfig_get_error",
				"count of errors encountered fetching and initializing kube config").Inc()
			logger.Fatalf(ctx, "Failed to initialize kubeClient: %+v", err)
		}

		clusterResourceController := clusterresource.NewClusterResourceController(db, kubeClient, scope)
		clusterResourceController.Run()
		logger.Infof(ctx, "ClusterResourceController started successfully")
	},
}

var controllerSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "This command will sync cluster resources",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		configuration := runtime.NewConfigurationProvider()
		scope := promutils.NewScope(configuration.ApplicationConfiguration().GetTopLevelConfig().MetricsScope).NewSubScope("clusterresource")
		dbConfigValues := configuration.ApplicationConfiguration().GetDbConfig()
		dbConfig := repositoryConfig.DbConfig{
			Host:         dbConfigValues.Host,
			Port:         dbConfigValues.Port,
			DbName:       dbConfigValues.DbName,
			User:         dbConfigValues.User,
			Password:     dbConfigValues.Password,
			ExtraOptions: dbConfigValues.ExtraOptions,
		}
		db := repositories.GetRepository(
			repositories.POSTGRES, dbConfig, scope.NewSubScope("database"))

		cfg := config.GetConfig()
		kubeClient, err := flytek8s.NewKubeClient(cfg.KubeConfig, cfg.Master, configuration.ClusterConfiguration())
		if err != nil {
			scope.NewSubScope("flytekubeconfig").MustNewCounter(
				"kubeconfig_get_error",
				"count of errors encountered fetching and initializing kube config").Inc()
			logger.Fatalf(ctx, "Failed to initialize kubeClient: %+v", err)
		}

		clusterResourceController := clusterresource.NewClusterResourceController(db, kubeClient, scope)
		err = clusterResourceController.Sync(ctx)
		if err != nil {
			logger.Fatalf(ctx, "Failed to sync cluster resources [%+v]", err)
		}
		logger.Infof(ctx, "ClusterResourceController started successfully")
	},
}

func init() {
	RootCmd.AddCommand(parentClusterResourceCmd)
	parentClusterResourceCmd.AddCommand(controllerRunCmd)
	parentClusterResourceCmd.AddCommand(controllerSyncCmd)
}
