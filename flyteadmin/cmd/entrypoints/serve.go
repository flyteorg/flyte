package entrypoints

import (
	"context"
	_ "net/http/pprof" // Required to serve application.

	"github.com/spf13/cobra"

	runtimeConfig "github.com/flyteorg/flyte/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flyte/flyteadmin/pkg/server"
	"github.com/flyteorg/flyte/flyteadmin/plugins"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/flytestdlib/profutils"
)

var pluginRegistryStore = plugins.NewAtomicRegistry(plugins.NewRegistry())

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Launches the Flyte admin server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		// Serve profiling endpoints.
		cfg := runtimeConfig.NewConfigurationProvider()
		go func() {
			err := profutils.StartProfilingServerWithDefaultHandlers(
				ctx, cfg.ApplicationConfiguration().GetTopLevelConfig().GetProfilerPort(), nil)
			if err != nil {
				logger.Panicf(ctx, "Failed to Start profiling and Metrics server. Error, %v", err)
			}
		}()
		server.SetMetricKeys(cfg.ApplicationConfiguration().GetTopLevelConfig())

		// register otel tracer providers
		for _, serviceName := range []string{otelutils.AdminGormTracer, otelutils.AdminServerTracer, otelutils.BlobstoreClientTracer} {
			if err := otelutils.RegisterTracerProviderWithContext(ctx, serviceName, otelutils.GetConfig()); err != nil {
				logger.Errorf(ctx, "Failed to create otel tracer provider. %v", err)
				return err
			}
		}

		return server.Serve(ctx, pluginRegistryStore.Load(), nil)
	},
}

func init() {
	// Command information
	RootCmd.AddCommand(serveCmd)
	RootCmd.AddCommand(secretsCmd)
}
