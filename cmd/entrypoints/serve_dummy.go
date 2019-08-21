package entrypoints

import (
	"context"
	"net"

	"github.com/lyft/datacatalog/pkg/config"
	"github.com/lyft/datacatalog/pkg/rpc/datacatalogservice"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var serveDummyCmd = &cobra.Command{
	Use:   "serve-dummy",
	Short: "Launches the Data Catalog server without any connections",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cfg := config.GetConfig()
		return serveDummy(ctx, cfg)
	},
}

func init() {
	RootCmd.AddCommand(serveDummyCmd)

	labeled.SetMetricKeys(contextutils.AppNameKey)
}

// Create and start the gRPC server and http healthcheck endpoint
func serveDummy(ctx context.Context, cfg *config.Config) error {
	// serve a http healthcheck endpoint
	go func() {
		err := serveHealthcheck(ctx, cfg)
		if err != nil {
			logger.Errorf(ctx, "Unable to serve http", cfg.GetGrpcHostAddress(), err)
		}
	}()

	grpcServer := newGRPCDummyServer(ctx)

	grpcListener, err := net.Listen("tcp", cfg.GetGrpcHostAddress())
	if err != nil {
		return err
	}

	logger.Infof(ctx, "Serving DataCatalog Insecure on port %v", cfg.GetGrpcHostAddress())
	return grpcServer.Serve(grpcListener)
}

// Creates a new GRPC Server with all the configuration
func newGRPCDummyServer(_ context.Context) *grpc.Server {
	grpcServer := grpc.NewServer()
	datacatalog.RegisterDataCatalogServer(grpcServer, &datacatalogservice.DataCatalogService{})
	return grpcServer
}
