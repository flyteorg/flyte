package main

import (
	"flag"
	"fmt"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/configuration"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/server"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/config/viper"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"os"

	"context"

	_ "net/http/pprof" // Required to serve application.
)

var (
	cfgFile        string
	configAccessor = viper.NewAccessor(config.Options{})
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Launches the Flyte artifacts server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cfg := configuration.ApplicationConfig.GetConfig().(*configuration.ApplicationConfiguration)
		fmt.Printf("cfg: [%+v]\n", cfg)
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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig(cmd.Flags())
	},
}

func init() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	// Add persistent flags - persistent flags persist through all sub-commands
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./artifact_config.yaml)")

	// Allow viper to read the value of the flags
	configAccessor.InitializePflags(RootCmd.PersistentFlags())

	// Command information
	RootCmd.AddCommand(serveCmd)

	err := flag.CommandLine.Parse([]string{})
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

}

func initConfig(flags *pflag.FlagSet) error {
	configAccessor = viper.NewAccessor(config.Options{
		SearchPaths: []string{cfgFile, "./artifact_config.yaml", ".", "/etc/flyte/config", "$GOPATH/src/github.com/flyteorg/flyte/flyteartifacts"},
		StrictMode:  false,
	})

	logger.Infof(context.TODO(), "Using config file: %v", configAccessor.ConfigFilesUsed())

	configAccessor.InitializePflags(flags)

	err := flag.CommandLine.Parse([]string{})
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	return configAccessor.UpdateConfig(context.TODO())
}

func main() {
	glog.V(2).Info("Beginning Flyte Artifacts Service")
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		panic(err)
	}
}
