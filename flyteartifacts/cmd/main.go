package main

import (
	"fmt"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/server"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"context"

	_ "net/http/pprof" // Required to serve application.
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Launches the Flyte artifacts server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		opts := make([]grpc.ServerOption, 0)
		return server.Serve(ctx, opts...)
	},
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "artifacts",
	Short: "Fill in later",
	Long: `
To get started run the serve subcommand which will start a server on localhost:50051
`,
}

func init() {
	// Command information
	RootCmd.AddCommand(serveCmd)
}

func main() {
	glog.V(2).Info("Beginning Flyte Artifacts Service")
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		panic(err)
	}
}
