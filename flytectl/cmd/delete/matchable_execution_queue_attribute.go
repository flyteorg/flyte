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
	executionQueueAttributesShort = "Deletes matchable resources of execution queue attributes."
	executionQueueAttributesLong  = `
Delete execution queue attributes for the given project and domain, in combination with the workflow name.

For project flytectldemo and development domain, run:
::

 flytectl delete execution-queue-attribute -p flytectldemo -d development

Delete execution queue attribute using the config file which was used to create it.

::

 flytectl delete execution-queue-attribute --attrFile era.yaml

For example, here's the config file era.yaml:

.. code-block:: yaml

    domain: development
    project: flytectldemo
    tags:
      - foo
      - bar
      - buzz
      - lightyear

Value is optional in the file as it is unread during the delete command but it can be retained since the same file can be used for get, update and delete commands.

To delete the execution queue attribute for the workflow 'core.control_flow.run_merge_sort.merge_sort', run the following command:

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
		admin.MatchableResource_EXECUTION_QUEUE, delConfig.DryRun); err != nil {
		return err
	}

	return nil
}
