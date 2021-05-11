package delete

import (
	"github.com/flyteorg/flytectl/cmd/config/subcommand"
	cmdcore "github.com/flyteorg/flytectl/cmd/core"

	"github.com/spf13/cobra"
)

// Long descriptions are whitespace sensitive when generating docs using sphinx.
const (
	deleteCmdShort = `Used for terminating/deleting various flyte resources including tasks/workflows/launchplans/executions/project.`
	deleteCmdLong  = `
Example Delete executions.
::

 bin/flytectl delete execution kxd1i72850  -d development  -p flytesnacks
`
)

// RemoteDeleteCommand will return delete command
func RemoteDeleteCommand() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: deleteCmdShort,
		Long:  deleteCmdLong,
	}
	terminateResourcesFuncs := map[string]cmdcore.CommandEntry{
		"execution": {CmdFunc: terminateExecutionFunc, Aliases: []string{"executions"}, Short: execCmdShort,
			Long: execCmdLong},
		"task-resource-attribute": {CmdFunc: deleteTaskResourceAttributes, Aliases: []string{"task-resource-attributes"},
			Short: taskResourceAttributesShort,
			Long:  taskResourceAttributesLong, PFlagProvider: subcommand.DefaultTaskResourceDelConfig, ProjectDomainNotRequired: true},
	}
	cmdcore.AddCommands(deleteCmd, terminateResourcesFuncs)
	return deleteCmd
}
