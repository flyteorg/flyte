package get

import (
	"context"

	"github.com/flyteorg/flytectl/cmd/config"
	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/executionqueueattribute"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

const (
	executionQueueAttributesShort = "Gets matchable resources of execution queue attributes"
	executionQueueAttributesLong  = `
Retrieves execution queue attributes for given project and domain combination or additionally with workflow name.

Retrieves execution queue attribute for project and domain
Here the command get execution queue attributes for  project flytectldemo and development domain.
::

 flytectl get execution-queue-attribute -p flytectldemo -d development 

eg : output from the command

.. code-block:: json

 {"project":"flytectldemo","domain":"development","tags":["foo", "bar"]}

Retrieves execution queue attribute for project and domain and workflow
Here the command get execution queue attributes for  project flytectldemo ,development domain and workflow core.control_flow.run_merge_sort.merge_sort
::

 flytectl get execution-queue-attribute -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

eg : output from the command

.. code-block:: json

 {"project":"flytectldemo","domain":"development","workflow":"core.control_flow.run_merge_sort.merge_sort","tags":["foo", "bar"]}

Writing the execution queue attribute to a file. If there are no execution queue attributes, command would return an error.
Here the command gets execution queue attributes and writes the config file to era.yaml
eg:  content of era.yaml

::

 flytectl get execution-queue-attribute --attrFile era.yaml


.. code-block:: yaml

    domain: development
    project: flytectldemo
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
