package get

import (
	"github.com/flyteorg/flytectl/cmd/config/subcommand"
	cmdcore "github.com/flyteorg/flytectl/cmd/core"

	"github.com/spf13/cobra"
)

// Long descriptions are whitespace sensitive when generating docs using sphinx.
const (
	getCmdShort = `Used for fetching various flyte resources including tasks/workflows/launchplans/executions/project.`
	getCmdLong  = `
Example get projects.
::

 bin/flytectl get project
`
)

// CreateGetCommand will return get command
func CreateGetCommand() *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: getCmdShort,
		Long:  getCmdLong,
	}

	getResourcesFuncs := map[string]cmdcore.CommandEntry{
		"project": {CmdFunc: getProjectsFunc, Aliases: []string{"projects"}, ProjectDomainNotRequired: true,
			Short: projectShort,
			Long:  projectLong},
		"task": {CmdFunc: getTaskFunc, Aliases: []string{"tasks"}, Short: taskShort,
			Long: taskLong, PFlagProvider: taskConfig},
		"workflow": {CmdFunc: getWorkflowFunc, Aliases: []string{"workflows"}, Short: workflowShort,
			Long: workflowLong},
		"launchplan": {CmdFunc: getLaunchPlanFunc, Aliases: []string{"launchplans"}, Short: launchPlanShort,
			Long: launchPlanLong, PFlagProvider: launchPlanConfig},
		"execution": {CmdFunc: getExecutionFunc, Aliases: []string{"executions"}, Short: executionShort,
			Long: executionLong},
		"task-resource-attribute": {CmdFunc: getTaskResourceAttributes, Aliases: []string{"task-resource-attributes"},
			Short: taskResourceAttributesShort,
			Long:  taskResourceAttributesLong, PFlagProvider: subcommand.DefaultTaskResourceFetchConfig},
	}

	cmdcore.AddCommands(getCmd, getResourcesFuncs)

	return getCmd
}
