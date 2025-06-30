package get

import (
	"context"
	"fmt"

	cmdcore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

const (
	healthDesc = "Gets health of the service."
)

func HealthCheck(ctx context.Context, args []string, cmdCtx cmdcore.CommandContext) error {
	if len(args) == 0 {
		logger.Infof(ctx, "No arguments provided, using default service name: flyteadmin")
		args = []string{"flyteadmin"}
	}
	resp, err := cmdCtx.ClientSet().HealthServiceClient().Check(ctx, &service.HealthRequest{Service: args[0]})
	if err != nil {
		return fmt.Errorf("failed to get health: %v", err)
	}
	fmt.Println(marshalOptions.Format(resp))
	return nil
}
