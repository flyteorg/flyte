package cmd

import (
	"context"

	"github.com/lyft/flytectl/cmd/get"

	"github.com/lyft/flytectl/cmd/config"
	stdConfig "github.com/lyft/flytestdlib/config"
	"github.com/lyft/flytestdlib/config/viper"
	"github.com/spf13/cobra"
)

var (
	cfgFile        string
	configAccessor = viper.NewAccessor(stdConfig.Options{StrictMode: true})
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		PersistentPreRunE: initConfig,
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/config.yaml)")

	configAccessor.InitializePflags(rootCmd.PersistentFlags())

	// Due to https://github.com/lyft/flyte/issues/341, project flag will have to be specified as
	// --root.project, this adds a convenience on top to allow --project to be used
	rootCmd.PersistentFlags().StringVarP(&(config.GetConfig().Project), "project", "p", "", "Specifies the Flyte project.")
	rootCmd.PersistentFlags().StringVarP(&(config.GetConfig().Domain), "domain", "d", "", "Specifies the Flyte project's domain.")
	rootCmd.PersistentFlags().StringVarP(&(config.GetConfig().Output), "output", "o", "table", "Specifies the output type")

	rootCmd.AddCommand(viper.GetConfigCommand())
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(get.CreateGetCommand())
	config.GetConfig()

	return rootCmd
}

func initConfig(_ *cobra.Command, _ []string) error {
	configAccessor = viper.NewAccessor(stdConfig.Options{
		StrictMode:  true,
		SearchPaths: []string{cfgFile},
	})

	err := configAccessor.UpdateConfig(context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func ExecuteCmd() error {
	return newRootCmd().Execute()
}
