package delete

import (
	"context"

	"github.com/flyteorg/flytectl/cmd/config"
	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	pluginoverride "github.com/flyteorg/flytectl/cmd/config/subcommand/plugin_override"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

const (
	pluginOverrideShort = "Deletes matchable resources of plugin overrides."
	pluginOverrideLong  = `
Delete plugin override for the given project and domain, in combination with the workflow name.

For project flytesnacks and development domain, run:
::

 flytectl delete plugin-override -p flytesnacks -d development


To delete plugin override using the config file which was used to create it, run:
::

 flytectl delete plugin-override --attrFile po.yaml

For example, here's the config file po.yaml:

.. code-block:: yaml

    domain: development
    project: flytesnacks
    overrides:
       - task_type: python_task # Task type for which to apply plugin implementation overrides
         plugin_id:             # Plugin id(s) to be used in place of the default for the task type.
           - plugin_override1
           - plugin_override2
         missing_plugin_behavior: 1 # Behavior when no specified plugin_id has an associated handler. 0 : FAIL , 1: DEFAULT

Overrides are optional in the file as they are unread during the delete command but can be retained since the same file can be used for get, update and delete commands.

To delete plugin override for the workflow 'core.control_flow.merge_sort.merge_sort', run the following command:

::

 flytectl delete plugin-override -p flytesnacks -d development core.control_flow.merge_sort.merge_sort

Usage
`
)

func deletePluginOverride(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	var pwdGetter sconfig.ProjectDomainWorkflowGetter
	pwdGetter = sconfig.PDWGetterCommandLine{Config: config.GetConfig(), Args: args}
	delConfig := pluginoverride.DefaultDelConfig

	// Get the project domain workflowName from the config file or commandline params
	if len(delConfig.AttrFile) > 0 {
		// Initialize AttrFileConfig which will be used if delConfig.AttrFile is non empty
		// And Reads from the attribute file
		pwdGetter = &pluginoverride.FileConfig{}
		if err := sconfig.ReadConfigFromFile(pwdGetter, delConfig.AttrFile); err != nil {
			return err
		}
	}
	// Use the pwdGetter to initialize the project domain and workflow
	project := pwdGetter.GetProject()
	domain := pwdGetter.GetDomain()
	workflowName := pwdGetter.GetWorkflow()

	// Deletes the matchable attributes using the AttrFileConfig
	if err := deleteMatchableAttr(ctx, project, domain, workflowName, cmdCtx.AdminDeleterExt(),
		admin.MatchableResource_PLUGIN_OVERRIDE, delConfig.DryRun); err != nil {
		return err
	}

	return nil
}
