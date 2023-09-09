package single

import (
	"context"
	"net/http"

	datacatalogConfig "github.com/flyteorg/datacatalog/pkg/config"
	datacatalogRepo "github.com/flyteorg/datacatalog/pkg/repositories"
	datacatalog "github.com/flyteorg/datacatalog/pkg/rpc/datacatalogservice"
	"github.com/flyteorg/flyteadmin/pkg/clusterresource"
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/plugins"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	adminServer "github.com/flyteorg/flyteadmin/pkg/server"
	adminScheduler "github.com/flyteorg/flyteadmin/scheduler"
	propellerEntrypoint "github.com/flyteorg/flytepropeller/pkg/controller"
	propellerConfig "github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flytepropeller/pkg/signals"
	webhookEntrypoint "github.com/flyteorg/flytepropeller/pkg/webhook"
	webhookConfig "github.com/flyteorg/flytepropeller/pkg/webhook/config"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/profutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flytestdlib/storage"
	_ "github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	_ "gorm.io/driver/postgres" // Required to import database driver.
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
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
	projects := []string{"flytesnacks"}
	if len(cfg.SeedProjects) != 0 {
		projects = cfg.SeedProjects
	}
	logger.Infof(ctx, "Seeding default projects...",projects)
	if err := adminServer.SeedProjects(ctx, projects); err != nil {
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
	propellerCfg := propellerConfig.GetConfig()
	propellerScope := promutils.NewScope(propellerConfig.GetConfig().MetricsPrefix).NewSubScope("propeller").NewSubScope(propellerCfg.LimitNamespace)
	limitNamespace := ""
	if propellerCfg.LimitNamespace != defaultNamespace {
		limitNamespace = propellerCfg.LimitNamespace
	}

	options := manager.Options{
		Namespace:  limitNamespace,
		SyncPeriod: &propellerCfg.DownstreamEval.Duration,
		NewClient: func(cache cache.Cache, config *rest.Config, options client.Options, uncachedObjects ...client.Object) (client.Client, error) {
			return executors.NewFallbackClientBuilder(propellerScope.NewSubScope("kube")).Build(cache, config, options)
		},
		CertDir: webhookConfig.GetConfig().CertDir,
		Port:    webhookConfig.GetConfig().ListenPort,
	}

	mgr, err := propellerEntrypoint.CreateControllerManager(ctx, propellerCfg, options)
	if err != nil {
		logger.Errorf(ctx, "Failed to create controller manager. %v", err)
		return err
	}
	g, childCtx := errgroup.WithContext(ctx)

	if !cfg.DisableWebhook {
		g.Go(func() error {
			logger.Infof(childCtx, "Starting to initialize certificate...")
			err := webhookEntrypoint.InitCerts(childCtx, propellerCfg, webhookConfig.GetConfig())
			if err != nil {
				logger.Errorf(childCtx, "Failed to initialize certificates for Secrets Webhook. %v", err)
				return err
			}
			logger.Infof(childCtx, "Starting Webhook server...")
			return webhookEntrypoint.Run(signals.SetupSignalHandler(childCtx), propellerCfg, webhookConfig.GetConfig(), defaultNamespace, &propellerScope, mgr)
		})
	}

	if !cfg.Disabled {
		g.Go(func() error {
			logger.Infof(childCtx, "Starting Flyte Propeller...")
			return propellerEntrypoint.StartController(childCtx, propellerCfg, defaultNamespace, mgr, &propellerScope)
		})
	}

	if !cfg.DisableWebhook || !cfg.Disabled {
		handlers := map[string]http.Handler{
			"/k8smetrics": promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{
				ErrorHandling: promhttp.HTTPErrorOnError,
			}),
		}

		g.Go(func() error {
			err := profutils.StartProfilingServerWithDefaultHandlers(childCtx, propellerCfg.ProfilerPort.Port, handlers)
			if err != nil {
				logger.Fatalf(childCtx, "Failed to Start profiling and metrics server. Error: %v", err)
			}
			return err
		})

		g.Go(func() error {
			err := propellerEntrypoint.StartControllerManager(childCtx, mgr)
			if err != nil {
				logger.Fatalf(childCtx, "Failed to start controller manager. Error: %v", err)
			}
			return err
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
					logger.Panicf(childCtx, "Failed to start Admin, err: %v", err)
					return err
				}
				return nil
			})
		}

		if !cfg.Propeller.Disabled {
			g.Go(func() error {
				err := startPropeller(childCtx, cfg.Propeller)
				if err != nil {
					logger.Panicf(childCtx, "Failed to start Propeller, err: %v", err)
					return err
				}
				return nil
			})
		}

		if !cfg.DataCatalog.Disabled {
			g.Go(func() error {
				err := startDataCatalog(childCtx, cfg.DataCatalog)
				if err != nil {
					logger.Panicf(childCtx, "Failed to start Datacatalog, err: %v", err)
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
