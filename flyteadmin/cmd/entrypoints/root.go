package entrypoints

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/flyteorg/flyteadmin/plugins"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/config/viper"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	cfgFile        string
	kubeMasterURL  string
	configAccessor = viper.NewAccessor(config.Options{})
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "flyteadmin",
	Short: "Fill in later",
	Long: `
To get started run the serve subcommand which will start a server on localhost:8088:

    flyteadmin serve

Then you can hit it with the client:

    flyteadmin adminservice foo bar baz

Or over HTTP 1.1 with curl:
    curl -X POST http://localhost:8088/api/v1/projects'
`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig(cmd.Flags())
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(pluginRegistry *plugins.Registry) error {
	pluginRegistryStore.Store(pluginRegistry)
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func init() {
	// allows `$ flyteadmin --logtostderr` to work
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	// Add persistent flags - persistent flags persist through all sub-commands
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./flyteadmin_config.yaml)")
	RootCmd.PersistentFlags().StringVar(&kubeMasterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")

	RootCmd.AddCommand(viper.GetConfigCommand())

	// Allow viper to read the value of the flags
	configAccessor.InitializePflags(RootCmd.PersistentFlags())

	err := flag.CommandLine.Parse([]string{})
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func initConfig(flags *pflag.FlagSet) error {
	configAccessor = viper.NewAccessor(config.Options{
		SearchPaths: []string{cfgFile, ".", "/etc/flyte/config", "$GOPATH/src/github.com/flyteorg/flyteadmin"},
		StrictMode:  false,
	})

	logger.Infof(context.TODO(), "Using config file: %v", configAccessor.ConfigFilesUsed())

	configAccessor.InitializePflags(flags)

	return configAccessor.UpdateConfig(context.TODO())
}
