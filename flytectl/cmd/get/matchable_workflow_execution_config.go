package get

import (
	"context"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/workflowexecutionconfig"

	"github.com/flyteorg/flytectl/cmd/config"
	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

const (
	workflowExecutionConfigShort = "Gets matchable resources of workflow execution config"
	workflowExecutionConfigLong  = `
Retrieves workflow execution config for given project and domain combination or additionally with workflow name.

Retrieves workflow execution config for project and domain
Here the command get workflow execution config for project flytectldemo and development domain.

::

 flytectl get workflow-execution-config -p flytectldemo -d development 

eg : output from the command

.. code-block:: json

 {
	"project": "flytectldemo",
	"domain": "development",
	"max_parallelism": 5
 }

Retrieves workflow execution config for project and domain and workflow
Here the command get workflow execution config for project flytectldemo ,development domain and workflow core.control_flow.run_merge_sort.merge_sort

::

 flytectl get workflow-execution-config -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

eg : output from the command

.. code-block:: json

 {
	"project": "flytectldemo",
	"domain": "development",
    "workflow": "core.control_flow.run_merge_sort.merge_sort"
	"max_parallelism": 5
 }

Writing the workflow execution config to a file. If there are no workflow execution config, command would return an error.
Here the command gets workflow execution config and writes the config file to wec.yaml
eg:  content of wec.yaml

::

 flytectl get workflow-execution-config -p flytectldemo -d development --attrFile wec.yaml


.. code-block:: yaml

    domain: development
    project: flytectldemo
    max_parallelism: 5

Usage
`
)

func getWorkflowExecutionConfigFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	var project string
	var domain string
	var workflowName string

	// Get the project domain workflow name parameters from the command line. Project and domain are mandatory for this command
	project = config.GetConfig().Project
	domain = config.GetConfig().Domain
	if len(args) == 1 {
		workflowName = args[0]
	}
	// Construct a shadow config for WorkflowExecutionConfig. The shadow config is not using ProjectDomainAttribute/Workflowattribute directly inorder to simplify the inputs.
	workflowExecutionConfigFileConfig := workflowexecutionconfig.FileConfig{Project: project, Domain: domain, Workflow: workflowName}
	// Get the workflow execution config from the command line config
	fileName := workflowexecutionconfig.DefaultFetchConfig.AttrFile

	// Updates the workflowExecutionConfigFileConfig with the fetched matchable attribute
	if err := FetchAndUnDecorateMatchableAttr(ctx, project, domain, workflowName, cmdCtx.AdminFetcherExt(),
		&workflowExecutionConfigFileConfig, admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG); err != nil {
		return err
	}

	// Write the config to the file which can be used for update
	if err := sconfig.DumpTaskResourceAttr(workflowExecutionConfigFileConfig, fileName); err != nil {
		return err
	}
	return nil
}
