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
	clusterResourceAttributesShort = "Deletes matchable resources of cluster attributes"
	clusterResourceAttributesLong  = `
Deletes cluster resource attributes for given project and domain combination or additionally with workflow name.

Deletes cluster resource attribute for project and domain
Here the command delete cluster resource attributes for  project flytectldemo and development domain.
::

 flytectl delete cluster-resource-attribute -p flytectldemo -d development 


Deletes cluster resource attribute using config file which was used for creating it.
Here the command deletes cluster resource attributes from the config file cra.yaml
Attributes are optional in the file as they are unread during the delete command but can be kept as the same file can be used for get, update or delete 
eg:  content of cra.yaml which will use the project domain and workflow name for deleting the resource

::

 flytectl delete cluster-resource-attribute --attrFile cra.yaml


.. code-block:: yaml
	
    domain: development
    project: flytectldemo
    attributes:
      foo: "bar"
      buzz: "lightyear"

Deletes cluster resource attribute for a workflow
Here the command deletes cluster resource attributes for a workflow core.control_flow.run_merge_sort.merge_sort

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
