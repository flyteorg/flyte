package entrypoints

import (
	"context"
	"fmt"
	"time"

	"github.com/flyteorg/flytestdlib/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/avast/retry-go"
	adminClient "github.com/flyteorg/flyteidl/clients/go/admin"
	"github.com/pkg/errors"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/spf13/cobra"
)

const (
	timeout            = 30 * time.Second
	timeoutError       = "timeout: failed to connect service %q within %v"
	connectionError    = "error: failed to connect service at %q: %+v"
	deadlineError      = "timeout: health rpc did not complete within %v"
	healthCheckError   = "Health check failed with status %v"
	healthCheckSuccess = "Health check passed, Flyteadmin is up and running"
)

var preCheckRunCmd = &cobra.Command{
	Use:   "precheck",
	Short: "This command will check pre requirement for scheduler",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := []grpc.DialOption{
			grpc.WithUserAgent("grpc_health_probe"),
			grpc.WithBlock(),
			grpc.WithInsecure(),
		}
		ctx := context.Background()
		config := adminClient.GetConfig(ctx)

		err := retry.Do(
			func() error {
				dialCtx, dialCancel := context.WithTimeout(ctx, timeout)
				defer dialCancel()
				conn, err := grpc.DialContext(dialCtx, config.Endpoint.String(), opts...)
				if err != nil {
					if err == context.DeadlineExceeded {
						logger.Errorf(ctx, timeoutError, config.Endpoint.String(), timeout)
						return errors.New(fmt.Sprintf(timeoutError, config.Endpoint.String(), timeout))
					}
					logger.Errorf(ctx, connectionError, config.Endpoint.String(), err)
					return errors.New(fmt.Sprintf(connectionError, config.Endpoint.String(), err))
				}
				rpcCtx := metadata.NewOutgoingContext(ctx, metadata.MD{})
				resp, err := healthpb.NewHealthClient(conn).Check(rpcCtx,
					&healthpb.HealthCheckRequest{
						Service: "",
					})
				if err != nil {
					if stat, ok := status.FromError(err); ok && stat.Code() == codes.Unimplemented {
						return retry.Unrecoverable(err)
					} else if stat, ok := status.FromError(err); ok && stat.Code() == codes.DeadlineExceeded {
						logger.Errorf(ctx, deadlineError, timeout)
						return errors.New(fmt.Sprintf(deadlineError, timeout))
					}
					return err
				}
				if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
					logger.Errorf(ctx, healthCheckError, resp.GetStatus())
					return errors.New(fmt.Sprintf(healthCheckError, resp.GetStatus()))
				}
				return nil
			},
			retry.Delay(retry.BackOffDelay(10, nil, &retry.Config{})),
		)
		if err != nil {
			return err
		}

		logger.Printf(ctx, healthCheckSuccess)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(preCheckRunCmd)
}
