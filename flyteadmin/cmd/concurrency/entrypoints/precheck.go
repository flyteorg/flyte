package entrypoints

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

const (
	healthCheckSuccess = "Health check passed, Flyteadmin is up and running"
	healthCheckError   = "health check failed with status %v"
)

var preCheckRunCmd = &cobra.Command{
	Use:   "precheck",
	Short: "This command will check prerequisites for the concurrency controller",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		clientSet, err := admin.ClientSetBuilder().WithConfig(admin.GetConfig(ctx)).Build(ctx)
		if err != nil {
			logger.Errorf(ctx, "Flyte concurrency controller precheck failed due to %v\n", err)
			return err
		}

		healthCheckResponse, err := clientSet.HealthServiceClient().Check(ctx,
			&grpc_health_v1.HealthCheckRequest{Service: "flyteadmin"})
		if err != nil {
			return err
		}
		if healthCheckResponse.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
			logger.Errorf(ctx, healthCheckError, healthCheckResponse.GetStatus())
			return fmt.Errorf(healthCheckError, healthCheckResponse.GetStatus())
		}
		logger.Infof(ctx, "Health check response is %v", healthCheckResponse)
		logger.Infof(ctx, healthCheckSuccess)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(preCheckRunCmd)
}
