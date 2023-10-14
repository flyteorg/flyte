package entrypoints

import (
	"context"
	runtimeConfig "github.com/flyteorg/flyte/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flyte/flyteadmin/pkg/server"
	"github.com/flyteorg/flyte/flyteadmin/plugins"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/profutils"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/spf13/cobra"
	_ "net/http/pprof" // Required to serve application.
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
		if cfg.ApplicationConfiguration().GetTopLevelConfig().EnableGrpcHistograms {
			grpcPrometheus.EnableHandlingTimeHistogram()
		}

		return server.Serve(ctx, pluginRegistryStore.Load(), nil)
	},
}

func init() {
	// Command information
	RootCmd.AddCommand(serveCmd)
	RootCmd.AddCommand(secretsCmd)
}
