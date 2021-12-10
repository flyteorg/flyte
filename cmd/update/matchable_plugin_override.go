package update

import (
	"context"
	"fmt"

	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	pluginoverride "github.com/flyteorg/flytectl/cmd/config/subcommand/plugin_override"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
)

const (
	pluginOverrideShort = "Update matchable resources of plugin overrides"
	pluginOverrideLong  = `
Updates plugin overrides for given project and domain combination or additionally with workflow name.

Updating to the plugin override is only available from a generated file. See the get section for generating this file.
This will completely overwrite any existing plugins overrides on custom project, domain and workflow combination.
It is preferable to do get and generate a plugin override file if there is an existing override already set and then update it to have new values.
Refer to get plugin-override section on how to generate this file
It takes input for plugin overrides from the config file po.yaml,
e.g., content of po.yaml:

.. code-block:: yaml

    domain: development
    project: flytectldemo
    overrides:
       - task_type: python_task # Task type for which to apply plugin implementation overrides
         plugin_id:             # Plugin id(s) to be used in place of the default for the task type.
           - plugin_override1
           - plugin_override2
         missing_plugin_behavior: 1 # Behavior when no specified plugin_id has an associated handler. 0 : FAIL , 1: DEFAULT

::

 flytectl update plugin-override --attrFile po.yaml

Updates plugin override for project, domain and workflow combination. This will take precedence over any other
plugin overrides defined at project domain level.
For workflow 'core.control_flow.run_merge_sort.merge_sort' in flytectldemo project, development domain, it is:

.. code-block:: yaml

    domain: development
    project: flytectldemo
    workflow: core.control_flow.run_merge_sort.merge_sort
    overrides:
       - task_type: python_task # Task type for which to apply plugin implementation overrides
         plugin_id:             # Plugin id(s) to be used in place of the default for the task type.
           - plugin_override1
           - plugin_override2
         missing_plugin_behavior: 1 # Behavior when no specified plugin_id has an associated handler. 0 : FAIL , 1: DEFAULT

::

 flytectl update plugin-override --attrFile po.yaml

Usage

`
)

func updatePluginOverridesFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	updateConfig := pluginoverride.DefaultUpdateConfig
	if len(updateConfig.AttrFile) == 0 {
		return fmt.Errorf("attrFile is mandatory while calling update for plugin override")
	}

	pluginOverrideFileConfig := pluginoverride.FileConfig{}
	if err := sconfig.ReadConfigFromFile(&pluginOverrideFileConfig, updateConfig.AttrFile); err != nil {
		return err
	}

	// Get project domain workflow name from the read file.
	project := pluginOverrideFileConfig.Project
	domain := pluginOverrideFileConfig.Domain
	workflowName := pluginOverrideFileConfig.Workflow

	// Updates the admin matchable attribute from pluginOverrideFileConfig
	if err := DecorateAndUpdateMatchableAttr(ctx, project, domain, workflowName, cmdCtx.AdminUpdaterExt(),
		pluginOverrideFileConfig, updateConfig.DryRun); err != nil {
		return err
	}
	return nil
}
