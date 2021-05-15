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
	executionClusterLabelShort = "Gets matchable resources of execution cluster label"
	executionClusterLabelLong  = `
Retrieves execution cluster label for given project and domain combination or additionally with workflow name.

Retrieves execution cluster label for project and domain
Here the command get execution cluster label for project flytectldemo and development domain.
::

 flytectl get execution-cluster-label -p flytectldemo -d development 

eg : output from the command

.. code-block:: json

 {"project":"flytectldemo","domain":"development","value":"foo"}

Retrieves execution cluster label for project and domain and workflow
Here the command get execution cluster label for  project flytectldemo, development domain and workflow core.control_flow.run_merge_sort.merge_sort
::

 flytectl get execution-cluster-label -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

eg : output from the command

.. code-block:: json

 {"project":"flytectldemo","domain":"development","workflow":"core.control_flow.run_merge_sort.merge_sort","value":"foo"}

Writing the execution cluster label to a file. If there are no execution cluster label, command would return an error.
Here the command gets execution cluster label and writes the config file to ecl.yaml
eg:  content of ecl.yaml

::

 flytectl get execution-cluster-label --attrFile ecl.yaml


.. code-block:: yaml

    domain: development
    project: flytectldemo
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
