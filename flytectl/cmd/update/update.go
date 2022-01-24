package update

import (
	"github.com/flyteorg/flytectl/cmd/config/subcommand/clusterresourceattribute"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/execution"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/executionclusterlabel"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/executionqueueattribute"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/launchplan"
	pluginoverride "github.com/flyteorg/flytectl/cmd/config/subcommand/plugin_override"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/project"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/taskresourceattribute"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/workflowexecutionconfig"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"

	"github.com/spf13/cobra"
)

// Long descriptions are whitespace sensitive when generating docs using sphinx.
const (
	updateUse     = "update"
	updateShort   = `Update Flyte resources e.g., project.`
	updatecmdLong = `
Currently, this command only provides subcommands to update project.
Takes input project that needs to be archived or unarchived. Name of the project to be updated is a mandatory field.
To update a project:
::

 flytectl update project -p flytesnacks --activateProject
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
		"launchplan": {CmdFunc: updateLPFunc, Aliases: []string{}, ProjectDomainNotRequired: false, PFlagProvider: launchplan.UConfig,
			Short: updateLPShort, Long: updateLPLong},
		"launchplan-meta": {CmdFunc: updateLPMetaFunc, Aliases: []string{}, ProjectDomainNotRequired: false, PFlagProvider: namedEntityConfig,
			Short: updateLPMetaShort, Long: updateLPMetaLong},
		"project": {CmdFunc: updateProjectsFunc, Aliases: []string{}, ProjectDomainNotRequired: true, PFlagProvider: project.DefaultProjectConfig,
			Short: projectShort, Long: projectLong},
		"execution": {CmdFunc: updateExecutionFunc, Aliases: []string{}, ProjectDomainNotRequired: false, PFlagProvider: execution.UConfig,
			Short: updateExecutionShort, Long: updateExecutionLong},
		"task-meta": {CmdFunc: updateTaskFunc, Aliases: []string{}, ProjectDomainNotRequired: false, PFlagProvider: namedEntityConfig,
			Short: updateTaskShort, Long: updateTaskLong},
		"workflow-meta": {CmdFunc: updateWorkflowFunc, Aliases: []string{}, ProjectDomainNotRequired: false, PFlagProvider: namedEntityConfig,
			Short: updateWorkflowShort, Long: updateWorkflowLong},
		"task-resource-attribute": {CmdFunc: updateTaskResourceAttributesFunc, Aliases: []string{}, PFlagProvider: taskresourceattribute.DefaultUpdateConfig,
			Short: taskResourceAttributesShort, Long: taskResourceAttributesLong, ProjectDomainNotRequired: true},
		"cluster-resource-attribute": {CmdFunc: updateClusterResourceAttributesFunc, Aliases: []string{}, PFlagProvider: clusterresourceattribute.DefaultUpdateConfig,
			Short: clusterResourceAttributesShort, Long: clusterResourceAttributesLong, ProjectDomainNotRequired: true},
		"execution-queue-attribute": {CmdFunc: updateExecutionQueueAttributesFunc, Aliases: []string{}, PFlagProvider: executionqueueattribute.DefaultUpdateConfig,
			Short: executionQueueAttributesShort, Long: executionQueueAttributesLong, ProjectDomainNotRequired: true},
		"execution-cluster-label": {CmdFunc: updateExecutionClusterLabelFunc, Aliases: []string{}, PFlagProvider: executionclusterlabel.DefaultUpdateConfig,
			Short: executionClusterLabelShort, Long: executionClusterLabelLong, ProjectDomainNotRequired: true},
		"plugin-override": {CmdFunc: updatePluginOverridesFunc, Aliases: []string{}, PFlagProvider: pluginoverride.DefaultUpdateConfig,
			Short: pluginOverrideShort, Long: pluginOverrideLong, ProjectDomainNotRequired: true},
		"workflow-execution-config": {CmdFunc: updateWorkflowExecutionConfigFunc, Aliases: []string{}, PFlagProvider: workflowexecutionconfig.DefaultUpdateConfig,
			Short: workflowExecutionConfigShort, Long: workflowExecutionConfigLong, ProjectDomainNotRequired: true},
	}
	cmdCore.AddCommands(updateCmd, updateResourcesFuncs)
	return updateCmd
}
