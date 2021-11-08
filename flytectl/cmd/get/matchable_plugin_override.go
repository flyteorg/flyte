package get

import (
	"context"

	"github.com/flyteorg/flytectl/cmd/config"
	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	pluginoverride "github.com/flyteorg/flytectl/cmd/config/subcommand/plugin_override"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

const (
	pluginOverrideShort = "Gets matchable resources of plugin override"
	pluginOverrideLong  = `
Retrieve the plugin overrides for the given project and domain.
Here, the command gets the plugin overrides for the project flytectldemo and development domain.

::

 flytectl get plugin-override -p flytectldemo -d development 

e.g. : output from the command

.. code-block:: json

 {
	"project": "flytectldemo",
	"domain": "development",
	"overrides": [{
		"task_type": "python_task",
		"plugin_id": ["pluginoverride1", "pluginoverride2"],
        "missing_plugin_behavior": 0 
	}]
 }

Retrieves the plugin overrides for project, domain and workflow
Here the command gets the plugin overrides for project flytectldemo, development domain and workflow core.control_flow.run_merge_sort.merge_sort

::

 flytectl get plugin-override -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

e.g. : output from the command

.. code-block:: json

 {
	"project": "flytectldemo",
	"domain": "development",
    "workflow": "core.control_flow.run_merge_sort.merge_sort"
	"overrides": [{
		"task_type": "python_task",
		"plugin_id": ["pluginoverride1", "pluginoverride2"],
        "missing_plugin_behavior": 0
	}]
 }

Writing the plugin overrides to a file. If there are no plugin overrides, command would return an error.
Here the command gets plugin overrides and writes the config file to po.yaml
eg:  content of po.yaml

::

 flytectl get plugin-override --attrFile po.yaml


.. code-block:: yaml

    domain: development
    project: flytectldemo
    overrides:
       - task_type: python_task # Task type for which to apply plugin implementation overrides
         plugin_id:             # Plugin id(s) to be used in place of the default for the task type.
           - plugin_override1
           - plugin_override2
         missing_plugin_behavior: 1 # Behavior when no specified plugin_id has an associated handler. 0 : FAIL , 1: DEFAULT

Usage
`
)

func getPluginOverridesFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	var project string
	var domain string
	var workflowName string

	// Get the project domain workflow name parameters from the command line. Project and domain are mandatory for this command
	project = config.GetConfig().Project
	domain = config.GetConfig().Domain
	if len(args) == 1 {
		workflowName = args[0]
	}
	// Construct a shadow config for PluginOverrides. The shadow config is not using ProjectDomainAttribute/Workflowattribute directly inorder to simplify the inputs.
	pluginOverrideFileConfig := pluginoverride.FileConfig{Project: project, Domain: domain, Workflow: workflowName}
	// Get the plugin overrides from the command line config
	fileName := pluginoverride.DefaultFetchConfig.AttrFile

	// Updates the pluginOverrideFileConfig with the fetched matchable attribute
	if err := FetchAndUnDecorateMatchableAttr(ctx, project, domain, workflowName, cmdCtx.AdminFetcherExt(),
		&pluginOverrideFileConfig, admin.MatchableResource_PLUGIN_OVERRIDE); err != nil {
		return err
	}

	// Write the config to the file which can be used for update
	if err := sconfig.DumpTaskResourceAttr(pluginOverrideFileConfig, fileName); err != nil {
		return err
	}
	return nil
}
