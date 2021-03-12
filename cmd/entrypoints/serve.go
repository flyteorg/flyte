package entrypoints

import (
	"context"
	"net"
	"net/http"

	"github.com/flyteorg/datacatalog/pkg/config"
	"github.com/flyteorg/datacatalog/pkg/rpc/datacatalogservice"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Launches the Data Catalog server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cfg := config.GetConfig()

		// serve a http healthcheck endpoint
		go func() {
			err := serveHTTPHealthcheck(ctx, cfg)
			if err != nil {
				logger.Errorf(ctx, "Unable to serve http", config.GetConfig().GetHTTPHostAddress(), err)
			}
		}()

		return serveInsecure(ctx, cfg)
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
}

// Create and start the gRPC server
func serveInsecure(ctx context.Context, cfg *config.Config) error {
	grpcServer := newGRPCServer(ctx, cfg)

	grpcListener, err := net.Listen("tcp", cfg.GetGrpcHostAddress())
	if err != nil {
		return err
	}

	logger.Infof(ctx, "Serving DataCatalog Insecure on port %v", config.GetConfig().GetGrpcHostAddress())
	return grpcServer.Serve(grpcListener)
}

// Creates a new GRPC Server with all the configuration
func newGRPCServer(_ context.Context, cfg *config.Config) *grpc.Server {
	grpcServer := grpc.NewServer()
	datacatalog.RegisterDataCatalogServer(grpcServer, datacatalogservice.NewDataCatalogService())

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	if cfg.GrpcServerReflection {
		reflection.Register(grpcServer)
	}
	return grpcServer
}

func serveHTTPHealthcheck(ctx context.Context, cfg *config.Config) error {
	mux := http.NewServeMux()

	// Register Healthcheck
	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	logger.Infof(ctx, "Serving DataCatalog http on port %v", cfg.GetHTTPHostAddress())
	return http.ListenAndServe(cfg.GetHTTPHostAddress(), mux)
}
