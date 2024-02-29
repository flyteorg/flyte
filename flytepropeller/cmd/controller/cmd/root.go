// Package cmd contains commands for FlytePropeller controller.
package cmd

import (
	"context"
	"flag"
	"net/http"
	"os"
	"runtime"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/flyteorg/flyte/flytepropeller/pkg/controller"
	config2 "github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/signals"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/config/viper"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/flytestdlib/profutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/version"
)

const (
	defaultNamespace = "all"
	appName          = "flytepropeller"
)

var (
	cfgFile        string
	configAccessor = viper.NewAccessor(config.Options{StrictMode: true})
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "flyte-propeller",
	Short: "Operator for running Flyte Workflows",
	Long: `Flyte Propeller runs a workflow to completion by recursing through the nodes, 
			handling their tasks to completion and propagating their status upstream.`,
	PersistentPreRunE: initConfig,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return executeRootCmd(ctx, config2.GetConfig())
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	version.LogBuildInformation(appName)
	logger.Infof(context.TODO(), "Detected: %d CPU's\n", runtime.NumCPU())
	if err := rootCmd.Execute(); err != nil {
		logger.Error(context.TODO(), err)
		os.Exit(1)
	}
}

func init() {
	// allows `$ flytepropeller --logtostderr` to work
	klog.InitFlags(flag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	err := flag.CommandLine.Parse([]string{})
	if err != nil {
		logAndExit(err)
	}

	// Here you will define your flags and configuration settings. Cobra supports persistent flags, which, if defined
	// here, will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/config.yaml)")

	configAccessor.InitializePflags(rootCmd.PersistentFlags())

	rootCmd.AddCommand(viper.GetConfigCommand())
}

func initConfig(cmd *cobra.Command, _ []string) error {
	configAccessor = viper.NewAccessor(config.Options{
		StrictMode:  false,
		SearchPaths: []string{cfgFile},
	})

	configAccessor.InitializePflags(cmd.PersistentFlags())

	err := configAccessor.UpdateConfig(context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func logAndExit(err error) {
	logger.Error(context.Background(), err)
	os.Exit(-1)
}

func executeRootCmd(baseCtx context.Context, cfg *config2.Config) error {
	// set up signals so we handle the first shutdown signal gracefully
	ctx := signals.SetupSignalHandler(baseCtx)

	// set metric keys
	keys := contextutils.MetricKeysFromStrings(cfg.MetricKeys)
	logger.Infof(context.TODO(), "setting metrics keys to %+v", keys)
	if len(keys) > 0 {
		labeled.SetMetricKeys(keys...)
	}

	// register opentelementry tracer providers
	for _, serviceName := range []string{otelutils.AdminClientTracer, otelutils.BlobstoreClientTracer,
		otelutils.DataCatalogClientTracer, otelutils.FlytePropellerTracer, otelutils.K8sClientTracer} {
		if err := otelutils.RegisterTracerProvider(serviceName, otelutils.GetConfig()); err != nil {
			logger.Errorf(ctx, "Failed to create otel tracer provider. %v", err)
			return err
		}
	}

	// Add the propeller subscope because the MetricsPrefix only has "flyte:" to get uniform collection of metrics.
	propellerScope := promutils.NewScope(cfg.MetricsPrefix).NewSubScope("propeller").NewSubScope(cfg.LimitNamespace)
	limitNamespace := ""
	var namespaceConfigs map[string]cache.Config
	if cfg.LimitNamespace != defaultNamespace {
		limitNamespace = cfg.LimitNamespace
		namespaceConfigs = map[string]cache.Config{
			limitNamespace: {},
		}
	}

	options := manager.Options{
		Cache: cache.Options{
			SyncPeriod:        &cfg.DownstreamEval.Duration,
			DefaultNamespaces: namespaceConfigs,
		},
		NewCache:  executors.NewCache,
		NewClient: executors.BuildNewClientFunc(propellerScope),
		Metrics: metricsserver.Options{
			// Disable metrics serving
			BindAddress: "0",
		},
	}

	mgr, err := controller.CreateControllerManager(ctx, cfg, options)
	if err != nil {
		logger.Fatalf(ctx, "Failed to create controller manager. Error: %v", err)
		return err
	}

	handlers := map[string]http.Handler{
		"/k8smetrics": promhttp.HandlerFor(metrics.Registry,
			promhttp.HandlerOpts{
				ErrorHandling: promhttp.HTTPErrorOnError,
			},
		),
	}

	g, childCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		err := profutils.StartProfilingServerWithDefaultHandlers(childCtx, cfg.ProfilerPort.Port, handlers)
		if err != nil {
			logger.Fatalf(childCtx, "Failed to Start profiling and metrics server. Error: %v", err)
		}
		return err
	})

	g.Go(func() error {
		err := controller.StartControllerManager(childCtx, mgr)
		if err != nil {
			logger.Fatalf(childCtx, "Failed to start controller manager. Error: %v", err)
		}
		return err
	})

	g.Go(func() error {
		err := controller.StartController(childCtx, cfg, defaultNamespace, mgr, &propellerScope)
		if err != nil {
			logger.Fatalf(childCtx, "Failed to start controller. Error: %v", err)
		}
		return err
	})

	return g.Wait()
}
