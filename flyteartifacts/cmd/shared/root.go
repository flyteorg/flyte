package shared

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/flyteorg/flyte/flyteartifacts/pkg/configuration"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/config/viper"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/profutils"
)

var (
	cfgFile        string
	configAccessor = viper.NewAccessor(config.Options{})
)

// NewRootCmd represents the base command when called without any subcommands
func NewRootCmd(rootUse string, grpcHook GrpcRegistrationHook, httpHook HttpRegistrationHook) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   rootUse,
		Short: "Short description",
		Long:  "Long description to be filled in later",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := initConfig(cmd, args)
			if err != nil {
				return err
			}

			go func() {
				ctx := context.Background()
				metricsCfg := configuration.GetApplicationConfig().ArtifactServerConfig.Metrics
				err := profutils.StartProfilingServerWithDefaultHandlers(ctx,
					metricsCfg.Port.Port, nil)
				if err != nil {
					logger.Panicf(ctx, "Failed to Start profiling and metrics server. Error: %v", err)
				}
			}()

			return nil
		},
	}

	initSubCommands(rootCmd, grpcHook, httpHook)
	return rootCmd
}

func initConfig(cmd *cobra.Command, _ []string) error {
	configAccessor = viper.NewAccessor(config.Options{
		SearchPaths: []string{cfgFile, ".", "/etc/flyte/config", "$GOPATH/src/github.com/flyteorg/flyte/flyteartifacts"},
		StrictMode:  false,
	})

	fmt.Println("Using config file: ", configAccessor.ConfigFilesUsed())

	// persistent flags were initially bound to the root command so we must bind to the same command to avoid
	// overriding those initial ones. We need to traverse up to the root command and initialize pflags for that.
	rootCmd := cmd
	for rootCmd.Parent() != nil {
		rootCmd = rootCmd.Parent()
	}

	configAccessor.InitializePflags(rootCmd.PersistentFlags())

	return configAccessor.UpdateConfig(context.TODO())
}

func initSubCommands(rootCmd *cobra.Command, grpcHook GrpcRegistrationHook, httpHook HttpRegistrationHook) {
	// allows ` --logtostderr` to work
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	// Add persistent flags - persistent flags persist through all sub-commands
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "artifact_config.yaml", "config file (default is ./artifact_config.yaml)")

	rootCmd.AddCommand(viper.GetConfigCommand())
	cfg := configuration.GetApplicationConfig()
	rootCmd.AddCommand(NewServeCmd(rootCmd.Use, cfg.ArtifactServerConfig, grpcHook, httpHook))

	// Allow viper to read the value of the flags
	configAccessor.InitializePflags(rootCmd.PersistentFlags())

	err := flag.CommandLine.Parse([]string{})
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
