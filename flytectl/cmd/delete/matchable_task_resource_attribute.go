package delete

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytectl/cmd/config"
	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/taskresourceattribute"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
)

const (
	taskResourceAttributesShort = "Deletes matchable resources of task attributes."
	taskResourceAttributesLong  = `
Delete task resource attributes for the given project and domain, in combination with the workflow name.

For project flytesnacks and development domain, run:
::

 flytectl delete task-resource-attribute -p flytesnacks -d development

To delete task resource attribute using the config file which was used to create it, run:

::

 flytectl delete task-resource-attribute --attrFile tra.yaml

For example, here's the config file tra.yaml:

.. code-block:: yaml

    domain: development
    project: flytesnacks
    defaults:
      cpu: "1"
      memory: "150Mi"
    limits:
      cpu: "2"
      memory: "450Mi"

The defaults/limits are optional in the file as they are unread during the delete command, but can be retained since the same file can be used for 'get', 'update' and 'delete' commands.

To delete task resource attribute for the workflow 'core.control_flow.merge_sort.merge_sort', run the following command:

::

 flytectl delete task-resource-attribute -p flytesnacks -d development core.control_flow.merge_sort.merge_sort

Usage
`
)

func deleteTaskResourceAttributes(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	var pwdGetter sconfig.ProjectDomainWorkflowGetter
	pwdGetter = sconfig.PDWGetterCommandLine{Config: config.GetConfig(), Args: args}
	delConfig := taskresourceattribute.DefaultDelConfig

	// Get the project domain workflowName from the config file or commandline params
	if len(delConfig.AttrFile) > 0 {
		// Initialize TaskResourceAttrFileConfig which will be used if delConfig.AttrFile is non empty
		// And Reads from the attribute file
		pwdGetter = &taskresourceattribute.TaskResourceAttrFileConfig{}
		if err := sconfig.ReadConfigFromFile(pwdGetter, delConfig.AttrFile); err != nil {
			return err
		}
	}
	// Use the pwdGetter to initialize the project domain and workflow
	project := pwdGetter.GetProject()
	domain := pwdGetter.GetDomain()
	workflowName := pwdGetter.GetWorkflow()

	// Deletes the matchable attributes using the taskResourceAttrFileConfig
	if err := deleteMatchableAttr(ctx, project, domain, workflowName, cmdCtx.AdminDeleterExt(),
		admin.MatchableResource_TASK_RESOURCE, delConfig.DryRun); err != nil {
		return err
	}

	return nil
}
