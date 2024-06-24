package entrypoints

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/flyteorg/flyte/datacatalog/pkg/config"
	"github.com/flyteorg/flyte/datacatalog/pkg/rpc/datacatalogservice"
	"github.com/flyteorg/flyte/datacatalog/pkg/runtime"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/flytestdlib/profutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Launches the Data Catalog server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cfg := config.GetConfig()

		// serve a http healthcheck endpoint
		go func() {
			err := datacatalogservice.ServeHTTPHealthCheck(ctx, cfg)
			if err != nil {
				logger.Errorf(ctx, "Unable to serve http", config.GetConfig().GetHTTPHostAddress(), err)
			}
		}()

		// Serve profiling endpoint.
		dataCatalogConfig := runtime.NewConfigurationProvider().ApplicationConfiguration().GetDataCatalogConfig()
		go func() {
			err := profutils.StartProfilingServerWithDefaultHandlers(
				context.Background(), dataCatalogConfig.ProfilerPort, nil)
			if err != nil {
				logger.Panicf(context.Background(), "Failed to Start profiling and Metrics server. Error, %v", err)
			}
		}()

		// Set Keys
		labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)

		// register otel tracer providers
		for _, serviceName := range []string{otelutils.DataCatalogGormTracer, otelutils.DataCatalogServerTracer} {
			if err := otelutils.RegisterTracerProviderWithContext(ctx, serviceName, otelutils.GetConfig()); err != nil {
				logger.Errorf(ctx, "Failed to create otel tracer provider. %v", err)
				return err
			}
		}

		return datacatalogservice.ServeInsecure(ctx, cfg)
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
}
