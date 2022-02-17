package entrypoints

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"

	"github.com/flyteorg/flyteadmin/pkg/clusterresource/impl"
	"github.com/flyteorg/flyteadmin/pkg/clusterresource/interfaces"
	execClusterIfaces "github.com/flyteorg/flyteadmin/pkg/executioncluster/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/resources"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/flyteorg/flyteidl/clients/go/admin"

	"github.com/flyteorg/flyteadmin/pkg/clusterresource"
	"github.com/flyteorg/flyteadmin/pkg/config"
	executioncluster "github.com/flyteorg/flyteadmin/pkg/executioncluster/impl"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/spf13/cobra"
	_ "gorm.io/driver/postgres" // Required to import database driver.
)

var parentClusterResourceCmd = &cobra.Command{
	Use:   "clusterresource",
	Short: "This command administers the ClusterResourceController. Please choose a subcommand.",
}

func getClusterResourceController(ctx context.Context, scope promutils.Scope, configuration runtimeInterfaces.Configuration) clusterresource.Controller {
	initializationErrorCounter := scope.MustNewCounter(
		"flyteclient_initialization_error",
		"count of errors encountered initializing a flyte client from kube config")
	var listTargetsProvider execClusterIfaces.ListTargetsInterface
	var err error
	if len(configuration.ClusterConfiguration().GetClusterConfigs()) == 0 {
		serverConfig := config.GetConfig()
		listTargetsProvider, err = executioncluster.NewInCluster(initializationErrorCounter, serverConfig.KubeConfig, serverConfig.Master)
	} else {
		listTargetsProvider, err = executioncluster.NewListTargets(initializationErrorCounter, executioncluster.NewExecutionTargetProvider(), configuration.ClusterConfiguration())
	}
	if err != nil {
		panic(err)
	}

	var adminDataProvider interfaces.FlyteAdminDataProvider
	if configuration.ClusterResourceConfiguration().IsStandaloneDeployment() {
		clientSet, err := admin.ClientSetBuilder().WithConfig(admin.GetConfig(ctx)).Build(ctx)
		if err != nil {
			panic(err)
		}
		adminDataProvider = impl.NewAdminServiceDataProvider(clientSet.AdminClient())
	} else {
		dbConfig := runtime.NewConfigurationProvider().ApplicationConfiguration().GetDbConfig()
		logConfig := logger.GetConfig()

		db, err := repositories.GetDB(ctx, dbConfig, logConfig)
		if err != nil {
			logger.Fatal(ctx, err)
		}
		dbScope := scope.NewSubScope("db")

		repo := repositories.NewGormRepo(
			db, errors.NewPostgresErrorTransformer(dbScope.NewSubScope("errors")), dbScope)

		adminDataProvider = impl.NewDatabaseAdminDataProvider(repo, configuration, resources.NewResourceManager(repo, configuration.ApplicationConfiguration()))
	}

	return clusterresource.NewClusterResourceController(adminDataProvider, listTargetsProvider, scope)
}

var controllerRunCmd = &cobra.Command{
	Use:   "run",
	Short: "This command will start a cluster resource controller to periodically sync cluster resources",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		configuration := runtime.NewConfigurationProvider()
		scope := promutils.NewScope(configuration.ApplicationConfiguration().GetTopLevelConfig().MetricsScope).NewSubScope("clusterresource")
		clusterResourceController := getClusterResourceController(ctx, scope, configuration)
		clusterResourceController.Run()
		logger.Infof(ctx, "ClusterResourceController started running successfully")
	},
}

var controllerSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "This command will sync cluster resources",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		configuration := runtime.NewConfigurationProvider()
		scope := promutils.NewScope(configuration.ApplicationConfiguration().GetTopLevelConfig().MetricsScope).NewSubScope("clusterresource")
		clusterResourceController := getClusterResourceController(ctx, scope, configuration)
		err := clusterResourceController.Sync(ctx)
		if err != nil {
			logger.Fatalf(ctx, "Failed to sync cluster resources [%+v]", err)
		}
		logger.Infof(ctx, "ClusterResourceController synced successfully")
	},
}

func init() {
	RootCmd.AddCommand(parentClusterResourceCmd)
	parentClusterResourceCmd.AddCommand(controllerRunCmd)
	parentClusterResourceCmd.AddCommand(controllerSyncCmd)
}
