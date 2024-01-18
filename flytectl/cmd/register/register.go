package register

import (
	rconfig "github.com/flyteorg/flytectl/cmd/config/subcommand/register"
	cmdcore "github.com/flyteorg/flytectl/cmd/core"

	"github.com/spf13/cobra"
)

// Long descriptions are whitespace sensitive when generating docs using sphinx.
const (
	registerCmdShort = "Registers tasks, workflows, and launch plans from a list of generated serialized files."
	registerCmdLong  = `
Take input files as serialized versions of the tasks/workflows/launchplans and register them with FlyteAdmin.
Currently, these input files are protobuf files generated as output from Flytekit serialize.
Project and Domain are mandatory fields to be passed for registration and an optional version which defaults to v1.
If the entities are already registered with Flyte for the same version, the registration would fail.
`
)

// RemoteRegisterCommand will return register command
func RemoteRegisterCommand() *cobra.Command {
	registerCmd := &cobra.Command{
		Use:   "register",
		Short: registerCmdShort,
		Long:  registerCmdLong,
	}
	registerResourcesFuncs := map[string]cmdcore.CommandEntry{
		"files": {CmdFunc: registerFromFilesFunc, Aliases: []string{"file"}, PFlagProvider: rconfig.DefaultFilesConfig,
			Short: registerFilesShort, Long: registerFilesLong},
		"examples": {CmdFunc: registerExamplesFunc, Aliases: []string{"example", "flytesnack", "flytesnacks"}, PFlagProvider: rconfig.DefaultFilesConfig,
			Short: registerExampleShort, Long: registerExampleLong},
	}
	cmdcore.AddCommands(registerCmd, registerResourcesFuncs)
	return registerCmd
}
