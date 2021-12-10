package delete

import (
	"context"

	"github.com/flyteorg/flytectl/cmd/config"
	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/clusterresourceattribute"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

const (
	clusterResourceAttributesShort = "Delete matchable resources of cluster attributes"
	clusterResourceAttributesLong  = `
Deletes cluster resource attributes for the given project and domain combination or additionally with workflow name.

For project flytectldemo and development domain, it is:
::

 flytectl delete cluster-resource-attribute -p flytectldemo -d development 


Deletes cluster resource attribute using config file which was used to create it.
Here, the config file is written to cra.yaml.
Attributes are optional in the file as they are unread during the delete command but can be kept since the same file can be used for get, update or delete commands.
e.g., content of cra.yaml:

::

 flytectl delete cluster-resource-attribute --attrFile cra.yaml


.. code-block:: yaml
	
    domain: development
    project: flytectldemo
    attributes:
      foo: "bar"
      buzz: "lightyear"

Deletes cluster resource attribute for a workflow.
For the workflow 'core.control_flow.run_merge_sort.merge_sort', it is:

::

 flytectl delete cluster-resource-attribute -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

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
