package entrypoints

import (
	"context"
	"fmt"
	"html"
	"net"
	"net/http"

	"github.com/lyft/datacatalog/pkg/config"
	"github.com/lyft/datacatalog/pkg/rpc/datacatalogservice"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Launches the Data Catalog server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cfg := config.GetConfig()

		// serve a http healthcheck endpoint
		go func() {
			err := serveHealthcheck(ctx, cfg)
			if err != nil {
				logger.Errorf(ctx, "Unable to serve http", config.GetConfig().GetGrpcHostAddress(), err)
			}
		}()

		return serveInsecure(ctx, cfg)
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)

	labeled.SetMetricKeys(contextutils.AppNameKey)
}

// Create and start the gRPC server
func serveInsecure(ctx context.Context, cfg *config.Config) error {
	grpcServer := newGRPCServer(ctx)

	grpcListener, err := net.Listen("tcp", cfg.GetGrpcHostAddress())
	if err != nil {
		return err
	}

	logger.Infof(ctx, "Serving DataCatalog Insecure on port %v", config.GetConfig().GetGrpcHostAddress())
	return grpcServer.Serve(grpcListener)
}

// Creates a new GRPC Server with all the configuration
func newGRPCServer(_ context.Context) *grpc.Server {
	grpcServer := grpc.NewServer()
	datacatalog.RegisterDataCatalogServer(grpcServer, datacatalogservice.NewDataCatalogService())
	return grpcServer
}

func serveHealthcheck(ctx context.Context, cfg *config.Config) error {
	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Healthcheck success on %v", html.EscapeString(r.URL.Path))
	})

	logger.Infof(ctx, "Serving DataCatalog http on port %v", cfg.GetHTTPHostAddress())
	return http.ListenAndServe(cfg.GetHTTPHostAddress(), nil)
}
