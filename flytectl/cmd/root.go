package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyteorg/flytectl/cmd/compile"
	"github.com/flyteorg/flytectl/cmd/config"
	configuration "github.com/flyteorg/flytectl/cmd/configuration"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/cmd/create"
	"github.com/flyteorg/flytectl/cmd/delete"
	"github.com/flyteorg/flytectl/cmd/demo"
	"github.com/flyteorg/flytectl/cmd/get"
	"github.com/flyteorg/flytectl/cmd/register"
	"github.com/flyteorg/flytectl/cmd/sandbox"
	"github.com/flyteorg/flytectl/cmd/update"
	"github.com/flyteorg/flytectl/cmd/upgrade"
	"github.com/flyteorg/flytectl/cmd/version"
	f "github.com/flyteorg/flytectl/pkg/filesystemutils"
	"github.com/flyteorg/flytectl/pkg/printer"
	stdConfig "github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/config/viper"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	cfgFile        string
	configAccessor = viper.NewAccessor(stdConfig.Options{StrictMode: true})
)

const (
	configFileDir  = ".flyte"
	configFileName = "config.yaml"
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		PersistentPreRunE: initConfig,
		Long:              "Flytectl is a CLI tool written in Go to interact with the FlyteAdmin service.",
		Short:             "Flytectl CLI tool",
		Use:               "flytectl",
		DisableAutoGenTag: true,
	}

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.flyte/config.yaml)")

	configAccessor.InitializePflags(rootCmd.PersistentFlags())

	// Due to https://github.com/flyteorg/flyte/issues/341, project flag will have to be specified as
	// --root.project, this adds a convenience on top to allow --project to be used
	rootCmd.PersistentFlags().StringVarP(&(config.GetConfig().Project), "project", "p", "", "Specifies the Flyte project.")
	rootCmd.PersistentFlags().StringVarP(&(config.GetConfig().Domain), "domain", "d", "", "Specifies the Flyte project's domain.")
	rootCmd.PersistentFlags().StringVarP(&(config.GetConfig().Output), "output", "o", printer.OutputFormatTABLE.String(), fmt.Sprintf("Specifies the output type - supported formats %s. NOTE: dot, doturl are only supported for Workflow", printer.OutputFormats()))

	rootCmd.AddCommand(get.CreateGetCommand())
	compileCmd := compile.CreateCompileCommand()
	cmdCore.AddCommands(rootCmd, compileCmd)
	rootCmd.AddCommand(create.RemoteCreateCommand())
	rootCmd.AddCommand(update.CreateUpdateCommand())
	rootCmd.AddCommand(register.RemoteRegisterCommand())
	rootCmd.AddCommand(delete.RemoteDeleteCommand())
	rootCmd.AddCommand(sandbox.CreateSandboxCommand())
	rootCmd.AddCommand(demo.CreateDemoCommand())
	rootCmd.AddCommand(configuration.CreateConfigCommand())
	rootCmd.AddCommand(completionCmd)
	// Added version command
	versionCmd := version.GetVersionCommand(rootCmd)
	cmdCore.AddCommands(rootCmd, versionCmd)

	// Added upgrade command
	upgradeCmd := upgrade.SelfUpgrade(rootCmd)
	cmdCore.AddCommands(rootCmd, upgradeCmd)

	config.GetConfig()

	return rootCmd
}

func initConfig(cmd *cobra.Command, _ []string) error {
	configFile := f.FilePathJoin(f.UserHomeDir(), configFileDir, configFileName)
	// TODO: Move flyteconfig env variable logic in flytestdlib
	if len(os.Getenv("FLYTECTL_CONFIG")) > 0 {
		configFile = os.Getenv("FLYTECTL_CONFIG")
	}

	if len(cfgFile) > 0 {
		configFile = cfgFile
	}

	configAccessor = viper.NewAccessor(stdConfig.Options{
		StrictMode:  true,
		SearchPaths: []string{configFile},
	})

	// persistent flags were initially bound to the root command so we must bind to the same command to avoid
	// overriding those initial ones. We need to traverse up to the root command and initialize pflags for that.
	rootCmd := cmd
	for rootCmd.Parent() != nil {
		rootCmd = rootCmd.Parent()
	}

	configAccessor.InitializePflags(rootCmd.PersistentFlags())

	err := configAccessor.UpdateConfig(context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func GenerateDocs() error {
	rootCmd := newRootCmd()
	err := GenReSTTree(rootCmd, "gen")
	if err != nil {
		logrus.Fatal(err)
		return err
	}
	return nil
}

func GenReSTTree(cmd *cobra.Command, dir string) error {
	emptyStr := func(s string) string { return "" }
	// Sphinx cross-referencing format
	linkHandler := func(name, ref string) string {
		return fmt.Sprintf(":doc:`%s`", ref)
	}
	return doc.GenReSTTreeCustom(cmd, dir, emptyStr, linkHandler)
}

func ExecuteCmd() error {
	return newRootCmd().Execute()
}
