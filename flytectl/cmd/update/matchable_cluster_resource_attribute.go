package update

import (
	"context"
	"fmt"

	sconfig "github.com/flyteorg/flyte/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/clusterresourceattribute"
	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

const (
	clusterResourceAttributesShort = "Update matchable resources of cluster attributes"
	clusterResourceAttributesLong  = `
Update cluster resource attributes for given project and domain combination or additionally with workflow name.

Updating to the cluster resource attribute is only available from a generated file. See the get section to generate this file.
It takes input for cluster resource attributes from the config file cra.yaml,
Example: content of cra.yaml:

.. code-block:: yaml

    domain: development
    project: flytesnacks
    attributes:
      foo: "bar"
      buzz: "lightyear"

::

 flytectl update cluster-resource-attribute --attrFile cra.yaml

Update cluster resource attribute for project and domain and workflow combination. This will take precedence over any other
resource attribute defined at project domain level.
This will completely overwrite any existing custom project, domain and workflow combination attributes.
It is preferable to do get and generate an attribute file if there is an existing attribute that is already set and then update it to have new values.
Refer to get cluster-resource-attribute section on how to generate this file.
For workflow 'core.control_flow.merge_sort.merge_sort' in flytesnacks project, development domain, it is:

.. code-block:: yaml

    domain: development
    project: flytesnacks
    workflow: core.control_flow.merge_sort.merge_sort
    attributes:
      foo: "bar"
      buzz: "lightyear"

::

 flytectl update cluster-resource-attribute --attrFile cra.yaml

Usage

`
)

func updateClusterResourceAttributesFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	updateConfig := clusterresourceattribute.DefaultUpdateConfig
	if len(updateConfig.AttrFile) == 0 {
		return fmt.Errorf("attrFile is mandatory while calling update for cluster resource attribute")
	}

	clustrResourceAttrFileConfig := clusterresourceattribute.AttrFileConfig{}
	if err := sconfig.ReadConfigFromFile(&clustrResourceAttrFileConfig, updateConfig.AttrFile); err != nil {
		return err
	}

	// Get project domain workflow name from the read file.
	project := clustrResourceAttrFileConfig.Project
	domain := clustrResourceAttrFileConfig.Domain
	workflowName := clustrResourceAttrFileConfig.Workflow

	if err := DecorateAndUpdateMatchableAttr(ctx, cmdCtx, project, domain, workflowName,
		admin.MatchableResource_CLUSTER_RESOURCE, clustrResourceAttrFileConfig,
		updateConfig.DryRun, updateConfig.Force); err != nil {
		return err
	}
	return nil
}
