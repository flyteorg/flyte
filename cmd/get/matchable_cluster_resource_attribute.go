package get

import (
	"context"

	"github.com/flyteorg/flytectl/cmd/config"
	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/clusterresourceattribute"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

const (
	clusterResourceAttributesShort = "Gets matchable resources of cluster resource attributes"
	clusterResourceAttributesLong  = `
Retrieves cluster resource attributes for given project and domain combination or additionally with workflow name.

Retrieves cluster resource attribute for project and domain
Here the command get cluster resource attributes for  project flytectldemo and development domain.
::

 flytectl get cluster-resource-attribute -p flytectldemo -d development 

eg : output from the command

.. code-block:: json

 {"project":"flytectldemo","domain":"development","attributes":{"buzz":"lightyear","foo":"bar"}}

Retrieves cluster resource attribute for project and domain and workflow
Here the command get cluster resource attributes for  project flytectldemo, development domain and workflow core.control_flow.run_merge_sort.merge_sort
::

 flytectl get cluster-resource-attribute -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

eg : output from the command

.. code-block:: json

 {"project":"flytectldemo","domain":"development","workflow":"core.control_flow.run_merge_sort.merge_sort","attributes":{"buzz":"lightyear","foo":"bar"}}

Writing the cluster resource attribute to a file. If there are no cluster resource attributes , command would return an error.
Here the command gets task resource attributes and writes the config file to cra.yaml
eg:  content of cra.yaml

::

 flytectl get task-resource-attribute --attrFile cra.yaml


.. code-block:: yaml

    domain: development
    project: flytectldemo
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
