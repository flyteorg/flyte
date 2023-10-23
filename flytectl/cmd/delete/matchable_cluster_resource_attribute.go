package delete

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytectl/cmd/config"
	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/clusterresourceattribute"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
)

const (
	clusterResourceAttributesShort = "Deletes matchable resources of cluster attributes."
	clusterResourceAttributesLong  = `
Delete cluster resource attributes for the given project and domain, in combination with the workflow name.

For project flytesnacks and development domain, run:
::

 flytectl delete cluster-resource-attribute -p flytesnacks -d development


To delete cluster resource attribute using the config file that was used to create it, run:

::

 flytectl delete cluster-resource-attribute --attrFile cra.yaml

For example, here's the config file cra.yaml:

.. code-block:: yaml

    domain: development
    project: flytesnacks
    attributes:
      foo: "bar"
      buzz: "lightyear"

Attributes are optional in the file, which are unread during the 'delete' command but can be retained since the same file can be used for 'get', 'update' and 'delete' commands.

To delete cluster resource attribute for the workflow 'core.control_flow.merge_sort.merge_sort', run:

::

 flytectl delete cluster-resource-attribute -p flytesnacks -d development core.control_flow.merge_sort.merge_sort

Usage
`
)

func deleteClusterResourceAttributes(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	var pwdGetter sconfig.ProjectDomainWorkflowGetter
	pwdGetter = sconfig.PDWGetterCommandLine{Config: config.GetConfig(), Args: args}
	delConfig := clusterresourceattribute.DefaultDelConfig

	// Get the project domain workflowName from the config file or commandline params
	if len(delConfig.AttrFile) > 0 {
		// Initialize TaskResourceAttrFileConfig which will be used if delConfig.AttrFile is non empty
		// And Reads from the attribute file
		pwdGetter = &clusterresourceattribute.AttrFileConfig{}
		if err := sconfig.ReadConfigFromFile(pwdGetter, delConfig.AttrFile); err != nil {
			return err
		}
	}
	// Use the pwdGetter to initialize the project domain and workflow
	project := pwdGetter.GetProject()
	domain := pwdGetter.GetDomain()
	workflowName := pwdGetter.GetWorkflow()

	// Deletes the matchable attributes using the taskResourceAttrFileConfig
	if err := deleteMatchableAttr(ctx, project, domain, workflowName, cmdCtx.AdminDeleterExt(),
		admin.MatchableResource_CLUSTER_RESOURCE, delConfig.DryRun); err != nil {
		return err
	}

	return nil
}
