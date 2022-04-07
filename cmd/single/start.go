package single

import (
	"context"
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/plugins"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flytestdlib/storage"

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
	webhookEntrypoint "github.com/flyteorg/flytepropeller/pkg/webhook"
	webhookConfig "github.com/flyteorg/flytepropeller/pkg/webhook/config"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	_ "github.com/golang/glog"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	_ "gorm.io/driver/postgres" // Required to import database driver.
)

const defaultNamespace = "all"

func startDataCatalog(ctx context.Context, _ DataCatalog) error {
	if err := datacatalogRepo.Migrate(ctx); err != nil {
		return err
	}
	catalogCfg := datacatalogConfig.GetConfig()
	return datacatalog.ServeInsecure(ctx, catalogCfg)
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

func startAdmin(ctx context.Context, cfg Admin) error {
	logger.Infof(ctx, "Running Database Migrations...")
	if err := adminServer.Migrate(ctx); err != nil {
		return err
	}

	logger.Infof(ctx, "Seeding default projects...")
	if err := adminServer.SeedProjects(ctx, []string{"flytesnacks"}); err != nil {
		return err
	}

	g, childCtx := errgroup.WithContext(ctx)

	if !cfg.DisableScheduler {
		logger.Infof(ctx, "Starting Scheduler...")
		g.Go(func() error {
			return adminScheduler.StartScheduler(childCtx)
		})
	}

	if !cfg.DisableClusterResourceManager {
		logger.Infof(ctx, "Starting cluster resource controller...")
		g.Go(func() error {
			return startClusterResourceController(childCtx)
		})
	}

	if !cfg.Disabled {
		g.Go(func() error {
			logger.Infof(ctx, "Starting Admin server...")
			registry := plugins.NewAtomicRegistry(plugins.NewRegistry())
			return adminServer.Serve(childCtx, registry.Load(), GetConsoleHandlers())
		})
	}
	return g.Wait()
}

func startPropeller(ctx context.Context, cfg Propeller) error {
	g, childCtx := errgroup.WithContext(ctx)

	if !cfg.DisableWebhook {
		g.Go(func() error {
			err := webhookEntrypoint.InitCerts(childCtx, propellerConfig.GetConfig(), webhookConfig.GetConfig())
			if err != nil {
				return err
			}
			return webhookEntrypoint.Run(childCtx, propellerConfig.GetConfig(), webhookConfig.GetConfig(), defaultNamespace)
		})
	}

	if !cfg.Disabled {
		g.Go(func() error {
			return propellerEntrypoint.StartController(ctx, propellerConfig.GetConfig(), defaultNamespace)
		})
	}

	return g.Wait()
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "This command will start Flyte cluster locally",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		g, childCtx := errgroup.WithContext(ctx)
		cfg := GetConfig()

		if !cfg.Admin.Disabled {
			g.Go(func() error {
				err := startAdmin(childCtx, cfg.Admin)
				if err != nil {
					logger.Errorf(childCtx, "Failed to start Admin, err: %v", err)
					return err
				}
				return nil
			})
		}

		if !cfg.Propeller.Disabled {
			g.Go(func() error {
				err := startPropeller(childCtx, cfg.Propeller)
				if err != nil {
					logger.Errorf(childCtx, "Failed to start Propeller, err: %v", err)
					return err
				}
				return nil
			})
		}

		if !cfg.DataCatalog.Disabled {
			g.Go(func() error {
				err := startDataCatalog(childCtx, cfg.DataCatalog)
				if err != nil {
					logger.Errorf(childCtx, "Failed to start Datacatalog, err: %v", err)
					return err
				}
				return nil
			})
		}

		return g.Wait()
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
	// Set Keys
	labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey,
		contextutils.ExecIDKey, contextutils.WorkflowIDKey, contextutils.NodeIDKey, contextutils.TaskIDKey,
		contextutils.TaskTypeKey, common.RuntimeTypeKey, common.RuntimeVersionKey, storage.FailureTypeLabel)
}
