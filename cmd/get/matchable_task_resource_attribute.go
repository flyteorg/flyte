package get

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

const (
	taskResourceAttributesShort = "Gets matchable resources of task attributes"
	taskResourceAttributesLong  = `
Retrieves task  resource attributes for given project,domain combination or additionally with workflow name.

Retrieves task resource attribute for project and domain
Here the command get task resource attributes for  project flytectldemo and development domain.
::

 flytectl get task-resource-attribute -p flytectldemo -d development 

eg : O/P

.. code-block:: json

 {"Project":"flytectldemo","Domain":"development","Workflow":"","defaults":{"cpu":"1","memory":"150Mi"},"limits":{"cpu":"2","memory":"450Mi"}}

Writing the task resource attribute to a file. If there are no task resource attributes a file would be written with basic data populated.
Here the command gets task resource attributes and writes the config file to tra.yaml
eg:  content of tra.yaml

::

 flytectl get task-resource-attribute --attrFile tra.yaml


.. code-block:: yaml

	domain: development
	project: flytectldemo
	defaults:
	  cpu: "1"
	  memory: 150Mi
	limits:
	  cpu: "2"
	  memory: 450Mi

Usage
`
)

func getTaskResourceAttributes(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	var project string
	var domain string
	var workflowName string

	// Get the project domain workflow name parameters from the command line. Project and domain are mandatory for this command
	project = config.GetConfig().Project
	domain = config.GetConfig().Domain
	if len(args) == 1 {
		workflowName = args[0]
	}
	// Construct a shadow config for TaskResourceAttribute. The shadow config is not using ProjectDomainAttribute/Workflowattribute directly inorder to simplify the inputs.
	taskResourceAttrFileConfig := subcommand.TaskResourceAttrFileConfig{Project: project, Domain: domain, Workflow: workflowName}
	// Get the attribute file name from the command line config
	fileName := subcommand.DefaultTaskResourceFetchConfig.AttrFile

	if len(workflowName) > 0 {
		// Fetch the workflow attribute from the admin
		workflowAttr, err := cmdCtx.AdminFetcherExt().FetchWorkflowAttributes(ctx,
			project, domain, workflowName, admin.MatchableResource_TASK_RESOURCE)
		if err != nil {
			return err
		}
		if workflowAttr.GetAttributes() == nil || workflowAttr.GetAttributes().GetMatchingAttributes() == nil {
			return fmt.Errorf("attribute doesn't exist")
		}
		// Update the shadow config with the fetched taskResourceAttribute which can then be written to a file which can then be called for an update.
		taskResourceAttrFileConfig.TaskResourceAttributes = workflowAttr.GetAttributes().GetMatchingAttributes().GetTaskResourceAttributes()
	} else {
		// Fetch the project domain attribute from the admin
		projectDomainAttr, err := cmdCtx.AdminFetcherExt().FetchProjectDomainAttributes(ctx,
			project, domain, admin.MatchableResource_TASK_RESOURCE)
		if err != nil {
			return err
		}
		if projectDomainAttr.GetAttributes() == nil || projectDomainAttr.GetAttributes().GetMatchingAttributes() == nil {
			return fmt.Errorf("attribute doesn't exist")
		}
		// Update the shadow config with the fetched taskResourceAttribute which can then be written to a file which can then be called for an update.
		taskResourceAttrFileConfig.TaskResourceAttributes = projectDomainAttr.GetAttributes().GetMatchingAttributes().GetTaskResourceAttributes()
	}
	// Write the config to the file which can be used for update
	taskResourceAttrFileConfig.DumpTaskResourceAttr(ctx, fileName)
	return nil
}
