package entrypoints

import (
	"context"

	"github.com/flyteorg/flyteadmin/plugins"

	"github.com/flyteorg/flytestdlib/profutils"

	_ "net/http/pprof" // Required to serve application.

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/server"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/spf13/cobra"

	runtimeConfig "github.com/flyteorg/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
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

		return server.Serve(ctx, pluginRegistryStore.Load(), nil)
	},
}

func init() {
	// Command information
	RootCmd.AddCommand(serveCmd)
	RootCmd.AddCommand(secretsCmd)

	// Set Keys
	labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey,
		contextutils.ExecIDKey, contextutils.WorkflowIDKey, contextutils.NodeIDKey, contextutils.TaskIDKey,
		contextutils.TaskTypeKey, contextutils.PhaseKey, contextutils.LaunchPlanIDKey, common.RuntimeTypeKey,
		common.RuntimeVersionKey)
}
