package delete

import (
	"context"

	"github.com/flyteorg/flytectl/cmd/config"
	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/taskresourceattribute"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

const (
	taskResourceAttributesShort = "Deletes matchable resources of task attributes"
	taskResourceAttributesLong  = `
Deletes task  resource attributes for given project,domain combination or additionally with workflow name.

Deletes task resource attribute for project and domain
Here the command delete task resource attributes for  project flytectldemo and development domain.
::

 flytectl delete task-resource-attribute -p flytectldemo -d development 


Deletes task resource attribute using config file which was used for creating it.
Here the command deletes task resource attributes from the config file tra.yaml
defaults/limits are optional in the file as they are unread during the delete command but can be kept as the same file can be used for get, update or delete 
eg:  content of tra.yaml which will use the project domain and workflow name for deleting the resource

::

 flytectl delete task-resource-attribute --attrFile tra.yaml


.. code-block:: yaml

    domain: development
    project: flytectldemo
    defaults:
      cpu: "1"
      memory: "150Mi"
    limits:
      cpu: "2"
      memory: "450Mi"

Deletes task resource attribute for a workflow
Here the command deletes task resource attributes for a workflow core.control_flow.run_merge_sort.merge_sort

::

 flytectl delete task-resource-attribute -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

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
		admin.MatchableResource_TASK_RESOURCE); err != nil {
		return err
	}

	return nil
}
