package entrypoints

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

const (
	healthCheckSuccess = "Health check passed, Flyteadmin is up and running"
	healthCheckError   = "health check failed with status %v"
)

var preCheckRunCmd = &cobra.Command{
	Use:   "precheck",
	Short: "This command will check pre requirement for scheduler",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		clientSet, err := admin.ClientSetBuilder().WithConfig(admin.GetConfig(ctx)).Build(ctx)

		if err != nil {
			logger.Errorf(ctx, "Flyte native scheduler precheck failed due to %v\n", err)
			return err
		}

		healthCheckResponse, err := clientSet.HealthServiceClient().Check(ctx,
			&service.HealthRequest{Service: "flyteadmin"})
		if err != nil {
			return fmt.Errorf("failed to perform flyteadmin health check: %v", err)
		}
		if healthCheckResponse.GetStatus() != service.HealthResponse_SERVING {
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
