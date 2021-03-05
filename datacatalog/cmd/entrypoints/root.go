package entrypoints

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/config/viper"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	cfgFile string

	configAccessor = viper.NewAccessor(config.Options{StrictMode: true})
)

func init() {
	// See https://gist.github.com/nak3/78a32817a8a3950ae48f239a44cd3663
	// allows `$ datacatalog --logtostderr` to work
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	// Add persistent flags - persistent flags persist through all sub-commands
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./datacatalog_config.yaml)")

	RootCmd.AddCommand(viper.GetConfigCommand())

	// Allow viper to read the value of the flags
	configAccessor.InitializePflags(RootCmd.PersistentFlags())

	err := flag.CommandLine.Parse([]string{})
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return nil
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "datacatalog",
	Short: "Launches datacatalog",
	Long: `
To get started run the serve subcommand which will start a server on localhost:8089:

    datacatalog serve
`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig(cmd.Flags())
	},
}

func initConfig(flags *pflag.FlagSet) error {
	configAccessor = viper.NewAccessor(config.Options{
		SearchPaths: []string{cfgFile, ".", "/etc/flyte/config", "$GOPATH/src/github.com/flyteorg/datacatalog"},
		StrictMode:  false,
	})

	fmt.Println("Using config file: ", configAccessor.ConfigFilesUsed())

	configAccessor.InitializePflags(flags)

	return configAccessor.UpdateConfig(context.TODO())
}
