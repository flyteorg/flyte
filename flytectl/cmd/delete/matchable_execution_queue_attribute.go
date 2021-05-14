package delete

import (
	"context"

	"github.com/flyteorg/flytectl/cmd/config"
	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/executionqueueattribute"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

const (
	executionQueueAttributesShort = "Deletes matchable resources of execution queue attributes"
	executionQueueAttributesLong  = `
Deletes execution queue attributes for given project and domain combination or additionally with workflow name.

Deletes execution queue attribute for project and domain
Here the command delete execution queue attributes for project flytectldemo and development domain.
::

 flytectl delete execution-queue-attribute -p flytectldemo -d development 


Deletes execution queue attribute using config file which was used for creating it.
Here the command deletes execution queue attributes from the config file era.yaml
Tags are optional in the file as they are unread during the delete command but can be kept as the same file can be used for get, update or delete 
eg:  content of era.yaml which will use the project domain and workflow name for deleting the resource

::

 flytectl delete execution-queue-attribute --attrFile era.yaml


.. code-block:: yaml

    domain: development
    project: flytectldemo
    tags:
      - foo
      - bar
      - buzz
      - lightyear

Deletes execution queue attribute for a workflow
Here the command deletes the execution queue attributes for a workflow

::

 flytectl delete execution-queue-attribute -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

Usage
`
)

func deleteExecutionQueueAttributes(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	var pwdGetter sconfig.ProjectDomainWorkflowGetter
	pwdGetter = sconfig.PDWGetterCommandLine{Config: config.GetConfig(), Args: args}
	delConfig := executionqueueattribute.DefaultDelConfig

	// Get the project domain workflowName from the config file or commandline params
	if len(delConfig.AttrFile) > 0 {
		// Initialize AttrFileConfig which will be used if delConfig.AttrFile is non empty
		// And Reads from the attribute file
		pwdGetter = &executionqueueattribute.AttrFileConfig{}
		if err := sconfig.ReadConfigFromFile(pwdGetter, delConfig.AttrFile); err != nil {
			return err
		}
	}
	// Use the pwdGetter to initialize the project domain and workflow
	project := pwdGetter.GetProject()
	domain := pwdGetter.GetDomain()
	workflowName := pwdGetter.GetWorkflow()

	// Deletes the matchable attributes using the AttrFileConfig
	if err := deleteMatchableAttr(ctx, project, domain, workflowName, cmdCtx.AdminDeleterExt(),
		admin.MatchableResource_EXECUTION_QUEUE); err != nil {
		return err
	}

	return nil
}
