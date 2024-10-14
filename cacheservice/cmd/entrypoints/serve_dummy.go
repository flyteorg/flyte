package entrypoints

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/flyteorg/flyte/cacheservice/pkg/config"
	"github.com/flyteorg/flyte/cacheservice/pkg/rpc/cacheservice"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

var serveDummyCmd = &cobra.Command{
	Use:   "serve-dummy",
	Short: "Launches the CacheService server without any connections",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cfg := config.GetConfig()

		// serve a http healthcheck endpoint
		go func() {
			err := cacheservice.ServeHTTPHealthCheck(ctx, cfg)
			if err != nil {
				logger.Errorf(ctx, "Unable to serve http", cfg.GetGrpcHostAddress(), err)
			}
		}()

		return cacheservice.ServeDummy(ctx, cfg)
	},
}

func init() {
	RootCmd.AddCommand(serveDummyCmd)
}
