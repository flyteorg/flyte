package get

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytectl/cmd/config"
	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/clusterresourceattribute"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
)

const (
	clusterResourceAttributesShort = "Gets matchable resources of cluster resource attributes."
	clusterResourceAttributesLong  = `
Retrieve cluster resource attributes for the given project and domain.
For project flytesnacks and development domain:
::

 flytectl get cluster-resource-attribute -p flytesnacks -d development

Example: output from the command:

.. code-block:: json

 {"project":"flytesnacks","domain":"development","attributes":{"buzz":"lightyear","foo":"bar"}}

Retrieve cluster resource attributes for the given project, domain, and workflow.
For project flytesnacks, development domain, and workflow 'core.control_flow.merge_sort.merge_sort':
::

 flytectl get cluster-resource-attribute -p flytesnacks -d development core.control_flow.merge_sort.merge_sort

Example: output from the command:

.. code-block:: json

 {"project":"flytesnacks","domain":"development","workflow":"core.control_flow.merge_sort.merge_sort","attributes":{"buzz":"lightyear","foo":"bar"}}

Write the cluster resource attributes to a file. If there are no cluster resource attributes, the command throws an error.
The config file is written to cra.yaml file.
Example: content of cra.yaml:

::

 flytectl get task-resource-attribute --attrFile cra.yaml


.. code-block:: yaml

    domain: development
    project: flytesnacks
    attributes:
      foo: "bar"
      buzz: "lightyear"

Usage
`
)

func getClusterResourceAttributes(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	var project string
	var domain string
	var workflowName string

	// Get the project domain workflow name parameters from the command line. Project and domain are mandatory for this command
	project = config.GetConfig().Project
	domain = config.GetConfig().Domain
	if len(args) == 1 {
		workflowName = args[0]
	}
	// Construct a shadow config for ClusterResourceAttribute. The shadow config is not using ProjectDomainAttribute/Workflowattribute directly inorder to simplify the inputs.
	clusterResourceAttrFileConfig := clusterresourceattribute.AttrFileConfig{Project: project, Domain: domain, Workflow: workflowName}
	// Get the attribute file name from the command line config
	fileName := clusterresourceattribute.DefaultFetchConfig.AttrFile

	// Updates the taskResourceAttrFileConfig with the fetched matchable attribute
	if err := FetchAndUnDecorateMatchableAttr(ctx, project, domain, workflowName, cmdCtx.AdminFetcherExt(),
		&clusterResourceAttrFileConfig, admin.MatchableResource_CLUSTER_RESOURCE); err != nil {
		return err
	}

	// Write the config to the file which can be used for update
	if err := sconfig.DumpTaskResourceAttr(clusterResourceAttrFileConfig, fileName); err != nil {
		return err
	}
	return nil
}
