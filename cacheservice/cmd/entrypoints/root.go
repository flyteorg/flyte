package entrypoints

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/flyteorg/flyte/flyteadmin/plugins"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/config/viper"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

var (
	cfgFile        string
	configAccessor = viper.NewAccessor(config.Options{})
)

func init() {
	// See https://gist.github.com/nak3/78a32817a8a3950ae48f239a44cd3663
	// allows `$ cacheservice --logtostderr` to work
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	// Add persistent flags - persistent flags persist through all sub-commands
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./cacheservice_config.yaml)")

	RootCmd.AddCommand(viper.GetConfigCommand())

	// Allow viper to read the value of the flags
	configAccessor.InitializePflags(RootCmd.PersistentFlags())

	err := flag.CommandLine.Parse([]string{})
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func Execute(pluginRegistry *plugins.Registry) error {
	pluginRegistryStore.Store(pluginRegistry)
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return nil
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "cacheservice",
	Short: "Launches cacheservice",
	Long: `
To get started run the serve subcommand which will start a server on localhost:8091:

	cacheservice serve
`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig(cmd.Flags())
	},
}

func initConfig(flags *pflag.FlagSet) error {
	configAccessor = viper.NewAccessor(config.Options{
		SearchPaths: []string{cfgFile, ".", "/etc/flyte/config", "$GOPATH/src/github.com/flyteorg/flyte/cacheservice"},
		StrictMode:  false,
	})

	logger.Infof(context.TODO(), "Using config file: %v", configAccessor.ConfigFilesUsed())

	configAccessor.InitializePflags(flags)

	return configAccessor.UpdateConfig(context.TODO())
}
