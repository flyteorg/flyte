package entrypoints

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/flyteorg/flyte/cacheservice/pkg/config"
	"github.com/flyteorg/flyte/cacheservice/pkg/rpc/cacheservice"
	"github.com/flyteorg/flyte/cacheservice/pkg/runtime"
	"github.com/flyteorg/flyte/flyteadmin/plugins"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/flytestdlib/profutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

var pluginRegistryStore = plugins.NewAtomicRegistry(plugins.NewRegistry())

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Launches the Cache Service server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cfg := config.GetConfig()

		// serve a http healthcheck endpoint
		go func() {
			err := cacheservice.ServeHTTPHealthCheck(ctx, cfg)
			if err != nil {
				logger.Errorf(ctx, "Unable to serve http", config.GetConfig().GetHTTPHostAddress(), err)
			}
		}()

		// Serve profiling endpoint.
		cacheServiceConfig := runtime.NewConfigurationProvider().ApplicationConfiguration().GetCacheServiceConfig()
		go func() {
			err := profutils.StartProfilingServerWithDefaultHandlers(
				context.Background(), cacheServiceConfig.ProfilerPort, nil)
			if err != nil {
				logger.Panicf(context.Background(), "Failed to Start profiling and Metrics server. Error, %v", err)
			}
		}()

		labeled.SetMetricKeys(contextutils.AppNameKey)

		// register otel tracer providers
		for _, serviceName := range []string{otelutils.CacheServiceGormTracer, otelutils.CacheServiceServerTracer} {
			if err := otelutils.RegisterTracerProvider(serviceName, otelutils.GetConfig()); err != nil {
				logger.Errorf(ctx, "Failed to create otel tracer provider. %v", err)
				return err
			}
		}

		return cacheservice.ServeInsecure(ctx, pluginRegistryStore.Load(), cfg)
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
}
