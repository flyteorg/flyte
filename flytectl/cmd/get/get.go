package get

import (
	"github.com/flyteorg/flytectl/cmd/config/subcommand/clusterresourceattribute"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/execution"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/executionclusterlabel"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/executionqueueattribute"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/launchplan"
	pluginoverride "github.com/flyteorg/flytectl/cmd/config/subcommand/plugin_override"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/project"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/task"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/taskresourceattribute"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/workflow"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/workflowexecutionconfig"
	cmdcore "github.com/flyteorg/flytectl/cmd/core"

	"github.com/spf13/cobra"
)

// Long descriptions are whitespace sensitive when generating docs using sphinx.
const (
	getCmdShort = `Fetch various Flyte resources including tasks/workflows/launchplans/executions/project.`
	getCmdLong  = `
For project, it is:
::

 flytectl get project
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
			Long:  projectLong, PFlagProvider: project.DefaultConfig},
		"task": {CmdFunc: getTaskFunc, Aliases: []string{"tasks"}, Short: taskShort,
			Long: taskLong, PFlagProvider: task.DefaultConfig},
		"workflow": {CmdFunc: getWorkflowFunc, Aliases: []string{"workflows"}, Short: workflowShort,
			Long: workflowLong, PFlagProvider: workflow.DefaultConfig},
		"launchplan": {CmdFunc: getLaunchPlanFunc, Aliases: []string{"launchplans"}, Short: launchPlanShort,
			Long: launchPlanLong, PFlagProvider: launchplan.DefaultConfig},
		"execution": {CmdFunc: getExecutionFunc, Aliases: []string{"executions"}, Short: executionShort,
			Long: executionLong, PFlagProvider: execution.DefaultConfig},
		"task-resource-attribute": {CmdFunc: getTaskResourceAttributes, Aliases: []string{"task-resource-attributes"},
			Short: taskResourceAttributesShort,
			Long:  taskResourceAttributesLong, PFlagProvider: taskresourceattribute.DefaultFetchConfig},
		"cluster-resource-attribute": {CmdFunc: getClusterResourceAttributes, Aliases: []string{"cluster-resource-attributes"},
			Short: clusterResourceAttributesShort,
			Long:  clusterResourceAttributesLong, PFlagProvider: clusterresourceattribute.DefaultFetchConfig},
		"execution-queue-attribute": {CmdFunc: getExecutionQueueAttributes, Aliases: []string{"execution-queue-attributes"},
			Short: executionQueueAttributesShort,
			Long:  executionQueueAttributesLong, PFlagProvider: executionqueueattribute.DefaultFetchConfig},
		"execution-cluster-label": {CmdFunc: getExecutionClusterLabel, Aliases: []string{"execution-cluster-labels"},
			Short: executionClusterLabelShort,
			Long:  executionClusterLabelLong, PFlagProvider: executionclusterlabel.DefaultFetchConfig},
		"plugin-override": {CmdFunc: getPluginOverridesFunc, Aliases: []string{"plugin-overrides"},
			Short: pluginOverrideShort,
			Long:  pluginOverrideLong, PFlagProvider: pluginoverride.DefaultFetchConfig},
		"workflow-execution-config": {CmdFunc: getWorkflowExecutionConfigFunc, Aliases: []string{"workflow-execution-config"},
			Short: workflowExecutionConfigShort,
			Long:  workflowExecutionConfigLong, PFlagProvider: workflowexecutionconfig.DefaultFetchConfig},
	}

	cmdcore.AddCommands(getCmd, getResourcesFuncs)

	return getCmd
}
