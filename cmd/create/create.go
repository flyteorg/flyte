package create

import (
	cmdcore "github.com/lyft/flytectl/cmd/core"
	"github.com/spf13/cobra"
)

// Long descriptions are whitespace sensitive when generating docs using sphinx.
const (
	createCmdShort = `Used for creating various flyte resources including tasks/workflows/launchplans/executions/project.`
	createCmdLong  = `
Example create.
::

 bin/flytectl create project --file project.yaml 
`
)

// CreateCommand will return create command
func RemoteCreateCommand() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: createCmdShort,
		Long:  createCmdLong,
	}
	createResourcesFuncs := map[string]cmdcore.CommandEntry{
		"project": {CmdFunc: createProjectsCommand, Aliases: []string{"projects"}, ProjectDomainNotRequired: true, PFlagProvider: projectConfig, Short: projectShort,
			Long: projectLong},
	}
	cmdcore.AddCommands(createCmd, createResourcesFuncs)
	return createCmd
}
