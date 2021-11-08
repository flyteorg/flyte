package get

import (
	"context"

	"github.com/flyteorg/flytectl/cmd/config"
	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/taskresourceattribute"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

const (
	taskResourceAttributesShort = "Gets matchable resources of task attributes"
	taskResourceAttributesLong  = `
Retrieve task resource attributes for the given project and domain.
Here, the command gets the task resource attributes for the project flytectldemo and development domain.
::

 flytectl get task-resource-attribute -p flytectldemo -d development 

e.g. : output from the command

.. code-block:: json

 {"project":"flytectldemo","domain":"development","workflow":"","defaults":{"cpu":"1","memory":"150Mi"},"limits":{"cpu":"2","memory":"450Mi"}}

Retrieve task resource attributes for the given project, domain, and workflow.
Here, the command gets task resource attributes for the project flytectldemo, development domain, and workflow 'core.control_flow.run_merge_sort.merge_sort'.
::

 flytectl get task-resource-attribute -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

e.g. : output from the command

.. code-block:: json

 {"project":"flytectldemo","domain":"development","workflow":"core.control_flow.run_merge_sort.merge_sort","defaults":{"cpu":"1","memory":"150Mi"},"limits":{"cpu":"2","memory":"450Mi"}}


Write the task resource attributes to a file. If there are no task resource attributes, a file would be populated with the basic data.
Here the command gets task resource attributes and writes the config file to tra.yaml
e.g. :  content of tra.yaml

::

 flytectl get task-resource-attribute --attrFile tra.yaml


.. code-block:: yaml

    domain: development
    project: flytectldemo
    defaults:
      cpu: "1"
      memory: "150Mi"
    limits:
      cpu: "2"
      memory: "450Mi"

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
	taskResourceAttrFileConfig := taskresourceattribute.TaskResourceAttrFileConfig{Project: project, Domain: domain, Workflow: workflowName}
	// Get the attribute file name from the command line config
	fileName := taskresourceattribute.DefaultFetchConfig.AttrFile

	// Updates the taskResourceAttrFileConfig with the fetched matchable attribute
	if err := FetchAndUnDecorateMatchableAttr(ctx, project, domain, workflowName, cmdCtx.AdminFetcherExt(),
		&taskResourceAttrFileConfig, admin.MatchableResource_TASK_RESOURCE); err != nil {
		return err
	}

	// Write the config to the file which can be used for update
	if err := sconfig.DumpTaskResourceAttr(taskResourceAttrFileConfig, fileName); err != nil {
		return err
	}
	return nil
}
