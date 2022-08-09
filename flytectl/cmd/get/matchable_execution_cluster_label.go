package get

import (
	"context"

	"github.com/flyteorg/flytectl/cmd/config"
	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/executionclusterlabel"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

const (
	executionClusterLabelShort = "Gets matchable resources of execution cluster label."
	executionClusterLabelLong  = `
Retrieve the execution cluster label for a given project and domain, in combination with the workflow name.

For project flytesnacks and development domain, run:
::

 flytectl get execution-cluster-label -p flytesnacks -d development

The output would look like:

.. code-block:: json

 {"project":"flytesnacks","domain":"development","value":"foo"}

Retrieve the execution cluster label for the given project, domain, and workflow.
For project flytesnacks, development domain, and workflow 'core.control_flow.merge_sort.merge_sort':
::

 flytectl get execution-cluster-label -p flytesnacks -d development core.control_flow.merge_sort.merge_sort

Example: output from the command:

.. code-block:: json

 {"project":"flytesnacks","domain":"development","workflow":"core.control_flow.merge_sort.merge_sort","value":"foo"}

Write the execution cluster label to a file. If there is no execution cluster label, the command throws an error.
The config file is written to ecl.yaml file.
Example: content of ecl.yaml:

::

 flytectl get execution-cluster-label --attrFile ecl.yaml


.. code-block:: yaml

    domain: development
    project: flytesnacks
    value: foo

Usage
`
)

func getExecutionClusterLabel(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	var project string
	var domain string
	var workflowName string

	// Get the project domain workflow name parameters from the command line. Project and domain are mandatory for this command
	project = config.GetConfig().Project
	domain = config.GetConfig().Domain
	if len(args) == 1 {
		workflowName = args[0]
	}
	// Construct a shadow config for ExecutionClusterLabel. The shadow config is not using ProjectDomainAttribute/Workflowattribute directly inorder to simplify the inputs.
	executionClusterLabelFileConfig := executionclusterlabel.FileConfig{Project: project, Domain: domain, Workflow: workflowName}
	// Get the attribute file name from the command line config
	fileName := executionclusterlabel.DefaultFetchConfig.AttrFile

	// Updates the taskResourceAttrFileConfig with the fetched matchable attribute
	if err := FetchAndUnDecorateMatchableAttr(ctx, project, domain, workflowName, cmdCtx.AdminFetcherExt(),
		&executionClusterLabelFileConfig, admin.MatchableResource_EXECUTION_CLUSTER_LABEL); err != nil {
		return err
	}

	// Write the config to the file which can be used for update
	if err := sconfig.DumpTaskResourceAttr(executionClusterLabelFileConfig, fileName); err != nil {
		return err
	}
	return nil
}
