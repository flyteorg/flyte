package single

import (
	"context"
	"github.com/flyteorg/flyteadmin/plugins"

	datacatalogConfig "github.com/flyteorg/datacatalog/pkg/config"
	datacatalogRepo "github.com/flyteorg/datacatalog/pkg/repositories"
	datacatalog "github.com/flyteorg/datacatalog/pkg/rpc/datacatalogservice"
	"github.com/flyteorg/flyteadmin/pkg/clusterresource"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	adminServer "github.com/flyteorg/flyteadmin/pkg/server"
	adminScheduler "github.com/flyteorg/flyteadmin/scheduler"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/k8s"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/pod"
	propellerEntrypoint "github.com/flyteorg/flytepropeller/pkg/controller"
	propellerConfig "github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	_ "github.com/golang/glog"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	_ "gorm.io/driver/postgres" // Required to import database driver.
)

func startDataCatalog(ctx context.Context) error {
	if err := datacatalogRepo.Migrate(ctx); err != nil {
		return err
	}
	cfg := datacatalogConfig.GetConfig()
	return datacatalog.ServeInsecure(ctx, cfg)
}

func startClusterResourceController(ctx context.Context) error {
	configuration := runtime.NewConfigurationProvider()
	scope := promutils.NewScope(configuration.ApplicationConfiguration().GetTopLevelConfig().MetricsScope).NewSubScope("clusterresource")
	clusterResourceController, err := clusterresource.NewClusterResourceControllerFromConfig(ctx, scope, configuration)
	if err != nil {
		return err
	}
	clusterResourceController.Run()
	logger.Infof(ctx, "ClusterResourceController started running successfully")
	return nil
}

func startAdmin(ctx context.Context) error {
	logger.Infof(ctx, "Running Database Migrations...")
	if err := adminServer.Migrate(ctx); err != nil {
		return err
	}
	logger.Infof(ctx, "Seeding default projects...")
	if err := adminServer.SeedProjects(ctx, []string{"flytesnacks"}); err != nil {
		return err
	}
	g := new(errgroup.Group)
	g.Go(func() error {
		logger.Infof(ctx, "Starting Scheduler...")
		return adminScheduler.StartScheduler(ctx)
	})
	logger.Infof(ctx, "Starting cluster resource controller...")
	g.Go(func() error {
		return startClusterResourceController(ctx)
	})
	g.Go(func() error {
		logger.Infof(ctx, "Starting Admin server...")
		registry := plugins.NewAtomicRegistry(plugins.NewRegistry())
		return adminServer.Serve(ctx, registry.Load(), GetConsoleHandlers())
	})
	return g.Wait()
}

func startPropeller(ctx context.Context) error {
	return propellerEntrypoint.StartController(ctx, propellerConfig.GetConfig(), "all")
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "This command will start Flyte cluster locally",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		g, childCtx := errgroup.WithContext(ctx)
		g.Go(func() error {
			err := startAdmin(childCtx)
			if err != nil {
				logger.Errorf(childCtx, "Failed to start Admin, err: %v", err)
				return err
			}
			return nil
		})

		g.Go(func() error {
			// labeled.SetMetricKeys(storage.FailureTypeLabel)
			err := startPropeller(childCtx)
			if err != nil {
				logger.Errorf(childCtx, "Failed to start Propeller, err: %v", err)
				return err
			}
			return nil
		})

		g.Go(func() error {
			err := startDataCatalog(childCtx)
			if err != nil {
				logger.Errorf(childCtx, "Failed to start Datacatalog, err: %v", err)
				return err
			}
			return nil
		})

		return g.Wait()
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
	// Set Keys
	// TODO: Do we really need this? I got the error when setting metric keys. "panic: cannot set metric keys more than once"
	//labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey,
	//	contextutils.ExecIDKey, contextutils.WorkflowIDKey, contextutils.NodeIDKey, contextutils.TaskIDKey,
	//	contextutils.TaskTypeKey, common.RuntimeTypeKey, common.RuntimeVersionKey)
}
