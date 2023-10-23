package get

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytectl/cmd/config"
	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/executionqueueattribute"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
)

const (
	executionQueueAttributesShort = "Gets matchable resources of execution queue attributes."
	executionQueueAttributesLong  = `
Retrieve the execution queue attribute for the given project and domain.
For project flytesnacks and development domain:
::

 flytectl get execution-queue-attribute -p flytesnacks -d development

Example: output from the command:

.. code-block:: json

 {"project":"flytesnacks","domain":"development","tags":["foo", "bar"]}

Retrieve the execution queue attribute for the given project, domain, and workflow.
For project flytesnacks, development domain, and workflow 'core.control_flow.merge_sort.merge_sort':
::

 flytectl get execution-queue-attribute -p flytesnacks -d development core.control_flow.merge_sort.merge_sort

Example: output from the command:

.. code-block:: json

 {"project":"flytesnacks","domain":"development","workflow":"core.control_flow.merge_sort.merge_sort","tags":["foo", "bar"]}

Write the execution queue attribute to a file. If there are no execution queue attributes, the command throws an error.
The config file is written to era.yaml file.
Example: content of era.yaml:

::

 flytectl get execution-queue-attribute --attrFile era.yaml


.. code-block:: yaml

    domain: development
    project: flytesnacks
    tags:
      - foo
      - bar
      - buzz
      - lightyear

Usage
`
)

func getExecutionQueueAttributes(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	var project string
	var domain string
	var workflowName string

	// Get the project domain workflow name parameters from the command line. Project and domain are mandatory for this command
	project = config.GetConfig().Project
	domain = config.GetConfig().Domain
	if len(args) == 1 {
		workflowName = args[0]
	}
	// Construct a shadow config for ExecutionQueueAttribute. The shadow config is not using ProjectDomainAttribute/Workflowattribute directly inorder to simplify the inputs.
	executionQueueAttrFileConfig := executionqueueattribute.AttrFileConfig{Project: project, Domain: domain, Workflow: workflowName}
	// Get the attribute file name from the command line config
	fileName := executionqueueattribute.DefaultFetchConfig.AttrFile

	// Updates the taskResourceAttrFileConfig with the fetched matchable attribute
	if err := FetchAndUnDecorateMatchableAttr(ctx, project, domain, workflowName, cmdCtx.AdminFetcherExt(),
		&executionQueueAttrFileConfig, admin.MatchableResource_EXECUTION_QUEUE); err != nil {
		return err
	}

	// Write the config to the file which can be used for update
	if err := sconfig.DumpTaskResourceAttr(executionQueueAttrFileConfig, fileName); err != nil {
		return err
	}
	return nil
}
