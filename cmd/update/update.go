package update

import (
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/spf13/cobra"
)

// Long descriptions are whitespace sensitive when generating docs using sphinx.
const (
	updateUse     = "update"
	updateShort   = `Used for updating flyte resources eg: project.`
	updatecmdLong = `
Currently this command only provides subcommands to update project.
Takes input project which need to be archived or unarchived. Name of the project to be updated is mandatory field.
Example update project to activate it.
::

 bin/flytectl update project -p flytesnacks --activateProject
`
)

// CreateUpdateCommand will return update command
func CreateUpdateCommand() *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   updateUse,
		Short: updateShort,
		Long:  updatecmdLong,
	}
	updateResourcesFuncs := map[string]cmdCore.CommandEntry{
		"launchplan": {CmdFunc: updateLPFunc, Aliases: []string{}, ProjectDomainNotRequired: false, PFlagProvider: namedEntityConfig,
			Short: updateLPShort, Long: updateLPLong},
		"project": {CmdFunc: updateProjectsFunc, Aliases: []string{}, ProjectDomainNotRequired: true, PFlagProvider: projectConfig,
			Short: projectShort, Long: projectLong},
		"task": {CmdFunc: updateTaskFunc, Aliases: []string{}, ProjectDomainNotRequired: false, PFlagProvider: namedEntityConfig,
			Short: updateTaskShort, Long: updateTaskLong},
		"workflow": {CmdFunc: updateWorkflowFunc, Aliases: []string{}, ProjectDomainNotRequired: false, PFlagProvider: namedEntityConfig,
			Short: updateWorkflowShort, Long: updateWorkflowLong},
	}
	cmdCore.AddCommands(updateCmd, updateResourcesFuncs)
	return updateCmd
}
