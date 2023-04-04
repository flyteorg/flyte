package entrypoints

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/config/viper"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	cfgFile        string
	configAccessor = viper.NewAccessor(config.Options{})
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "flytescheduler",
	Short: "Flyte native scheduler to run cron and fixed rate scheduled workflows",
	Long: `
Use the run subcommand which will start the scheduler by connecting to DB containing schedules

    flytescheduler run --config  flyteadmin_config.yaml  --admin.endpoint dns:///localhost:8089 --admin.insecure
`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig(cmd.Flags())
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func init() {
	// allows `$ flytescheduler --logtostderr` to work
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	// Add persistent flags - persistent flags persist through all sub-commands
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./flyteadmin_config.yaml)")

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
