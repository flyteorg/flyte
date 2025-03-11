package update

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/flyteorg/flyte/flytectl/cmd/config"
	"github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/clusterresourceattribute"
	"github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/execution"
	"github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/executionclusterlabel"
	"github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/executionqueueattribute"
	"github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/launchplan"
	pluginoverride "github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/plugin_override"
	"github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/project"
	"github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/taskresourceattribute"
	"github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/workflowexecutionconfig"
	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Long descriptions are whitespace sensitive when generating docs using sphinx.
const (
	updateUse     = "update"
	updateShort   = `Update Flyte resources e.g., project.`
	updatecmdLong = `
Provides subcommands to update Flyte resources, such as tasks, workflows, launch plans, executions, and projects.
Update Flyte resource; e.g., to activate a project:
::

 flytectl update project -p flytesnacks --activate
`
)

// CreateUpdateCommand will return update command
var updateResourcesFuncs = map[string]cmdCore.CommandEntry{
	"launchplan": {CmdFunc: updateLPFunc, Aliases: []string{}, ProjectDomainNotRequired: false, PFlagProvider: launchplan.UConfig,
		Short: updateLPShort, Long: updateLPLong},
	"launchplan-meta": {CmdFunc: getUpdateLPMetaFunc(namedEntityConfig), Aliases: []string{}, ProjectDomainNotRequired: false, PFlagProvider: namedEntityConfig,
		Short: updateLPMetaShort, Long: updateLPMetaLong},
	"project": {CmdFunc: updateProjectsFunc, Aliases: []string{}, ProjectDomainNotRequired: true, PFlagProvider: project.DefaultProjectConfig,
		Short: projectShort, Long: projectLong},
	"execution": {CmdFunc: updateExecutionFunc, Aliases: []string{}, ProjectDomainNotRequired: false, PFlagProvider: execution.UConfig,
		Short: updateExecutionShort, Long: updateExecutionLong},
	"task-meta": {CmdFunc: getUpdateTaskFunc(namedEntityConfig), Aliases: []string{}, ProjectDomainNotRequired: false, PFlagProvider: namedEntityConfig,
		Short: updateTaskShort, Long: updateTaskLong},
	"workflow-meta": {CmdFunc: getUpdateWorkflowFunc(namedEntityConfig), Aliases: []string{}, ProjectDomainNotRequired: false, PFlagProvider: namedEntityConfig,
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

var configStructMap = map[string]interface{}{
	"launchplan": launchplan.UConfig,
    "launchplan-meta": namedEntityConfig,
    "project": project.DefaultProjectConfig,
	"execution": execution.UConfig,
	"task-meta": namedEntityConfig,
	"workflow-meta": namedEntityConfig,
	"task-resource-attribute": taskresourceattribute.DefaultUpdateConfig,  // TODO
	"cluster-resource-attribute": clusterresourceattribute.DefaultUpdateConfig, // TODO
	"execution-queue-attribute": executionqueueattribute.DefaultUpdateConfig, // TODO
	"execution-cluster-label": executionclusterlabel.DefaultUpdateConfig,  // TODO
	"plugin-override": pluginoverride.DefaultUpdateConfig,  // TODO
	"workflow-execution-config": workflowexecutionconfig.DefaultFileConfig,
}

var subCmdMap = make(map[string]*cobra.Command)

func CreateUpdateCommand() *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   updateUse,
		Short: updateShort,
		Long:  updatecmdLong,
		RunE:  runUpdateFiles,
	}
	updateCmd.Flags().AddFlagSet(config.DefaultUpdateConfig.GetPFlagSet(""))

	cmdCore.AddCommands(updateCmd, updateResourcesFuncs)

	// Create a map of subcommands
	for _, subCmd := range updateCmd.Commands() {
		subCmdMap[subCmd.Name()] = subCmd
	}
	return updateCmd
}

func runUpdateFiles(cmd *cobra.Command, args []string) error {
	var updateConfig = config.DefaultUpdateConfig
	if len(updateConfig.AttrFile) == 0 {
		return fmt.Errorf("file is mandatory while calling update")
	}

	type generic map[string]interface{}

	for _, file := range updateConfig.AttrFile {
		yamlFile, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		decoder := yaml.NewDecoder(bytes.NewReader(yamlFile))

		for {
			obj := generic{}
			if err := decoder.Decode(&obj); err != nil {
				if err.Error() == "EOF" {
					break
				}
				return err
			}

			kind, ok := obj["kind"].(string)
			if !ok {
				return errors.New("kind not set or isn't string type")
			}

			// Find the corresponding subcommand
			// subCmdEntry, found := updateResourcesFuncs[kind]
            cmdStruct, found := configStructMap[kind]
			if !found {
				return fmt.Errorf("subcommand not found for kind %s", kind)
			}
			subCmd, found := subCmdMap[kind]
			if !found {
				return fmt.Errorf("subcommand not found for kind %s", kind)
			}

			// Convert obj to JSON and then to the appropriate struct
			data, err := json.Marshal(obj)
			if err != nil {
				return err
			}

			if err := json.Unmarshal(data, cmdStruct); err != nil {
				return err
			}

			// Execute the subcommand
			if err := subCmd.RunE(subCmd, args); err != nil {
				return err
			}
		}
	}

	return nil
}
