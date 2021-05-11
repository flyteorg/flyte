package delete

import (
	"context"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"
)

const (
	taskResourceAttributesShort = "Deletes matchable resources of task attributes"
	taskResourceAttributesLong  = `
Deletes task  resource attributes for given project,domain combination or additionally with workflow name.

Deletes task resource attribute for project and domain
Here the command delete task resource attributes for  project flytectldemo and development domain.
::

 flytectl delete task-resource-attribute -p flytectldemo -d development 


Deleting task resource attribute using config file which was used for creating it.
Here the command deletes task resource attributes from the config file tra.yaml
eg:  content of tra.yaml which will use the project domain and workflow name for deleting the resource

::

 flytectl delete task-resource-attribute --attrFile tra.yaml


.. code-block:: yaml

	domain: development
	project: flytectldemo
	defaults:
	  cpu: "1"
	  memory: 150Mi
	limits:
	  cpu: "2"
	  memory: 450Mi

Deleting task resource attribute for a workflow
Here the command deletes task resource attributes for a workflow

::

 flytectl delete task-resource-attribute -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

Usage
`
)

func deleteTaskResourceAttributes(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	delConfig := subcommand.DefaultTaskResourceDelConfig
	var project string
	var domain string
	var workflowName string

	if len(delConfig.AttrFile) > 0 {
		// Read the config from the file
		taskResourceAttrFileConfig := subcommand.TaskResourceAttrFileConfig{}
		if err := taskResourceAttrFileConfig.ReadConfigFromFile(delConfig.AttrFile); err != nil {
			return err
		}
		// Get project domain workflow name from the read file.
		project = taskResourceAttrFileConfig.Project
		domain = taskResourceAttrFileConfig.Domain
		workflowName = taskResourceAttrFileConfig.Workflow
	} else {
		// Get all the parameters for deletion from the command line
		project = config.GetConfig().Project
		domain = config.GetConfig().Domain
		if len(args) == 1 {
			workflowName = args[0]
		}
	}

	if len(workflowName) > 0 {
		// Delete the workflow attribute from the admin. If the attribute doesn't exist , admin deesn't return an error and same behavior is followed here
		err := cmdCtx.AdminDeleterExt().DeleteWorkflowAttributes(ctx, project, domain, workflowName, admin.MatchableResource_TASK_RESOURCE)
		if err != nil {
			return err
		}
		logger.Debugf(ctx, "Deleted task resource attributes from %v project and domain %v and workflow %v", project, domain, workflowName)
	} else {
		// Delete the project domain attribute from the admin. If the attribute doesn't exist , admin deesn't return an error and same behavior is followed here
		err := cmdCtx.AdminDeleterExt().DeleteProjectDomainAttributes(ctx, project, domain, admin.MatchableResource_TASK_RESOURCE)
		if err != nil {
			return err
		}
		logger.Debugf(ctx, "Deleted task resource attributes from %v project and domain %v", project, domain)
	}
	return nil
}
