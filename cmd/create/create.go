package create

import (
	"github.com/flyteorg/flytectl/cmd/config/subcommand/project"
	cmdcore "github.com/flyteorg/flytectl/cmd/core"

	"github.com/spf13/cobra"
)

// Long descriptions are whitespace sensitive when generating docs using Sphinx.
const (
	createCmdShort = `Creates various Flyte resources such as tasks, workflows, launch plans, executions, and projects.`
	createCmdLong  = `
Create Flyte resource; if a project:
::

 flytectl create project --file project.yaml 
`
)

// RemoteCreateCommand will return create Flyte resource commands
func RemoteCreateCommand() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: createCmdShort,
		Long:  createCmdLong,
	}
	createResourcesFuncs := map[string]cmdcore.CommandEntry{
		"project": {CmdFunc: createProjectsCommand, Aliases: []string{"projects"}, ProjectDomainNotRequired: true, PFlagProvider: project.DefaultProjectConfig, Short: projectShort,
			Long: projectLong},
		"execution": {CmdFunc: createExecutionCommand, Aliases: []string{"executions"}, ProjectDomainNotRequired: false, PFlagProvider: executionConfig, Short: executionShort,
			Long: executionLong},
	}
	cmdcore.AddCommands(createCmd, createResourcesFuncs)
	return createCmd
}
