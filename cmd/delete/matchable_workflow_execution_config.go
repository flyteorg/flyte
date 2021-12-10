package delete

import (
	"context"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/workflowexecutionconfig"

	"github.com/flyteorg/flytectl/cmd/config"
	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

const (
	workflowExecutionConfigShort = "Delete matchable resources of workflow execution config"
	workflowExecutionConfigLong  = `
Deletes workflow execution config for the given project and domain combination or additionally with workflow name.

For project flytectldemo and development domain, it is:
::

 flytectl delete workflow-execution-config -p flytectldemo -d development 


Deletes workflow execution config using config file which was used to create it.
Here, the config file is written to wec.yaml.
Max_parallelism is optional in the file as it is unread during the delete command but can be kept since the same file can be used for get, update or delete commands.
e.g., content of wec.yaml:

::

 flytectl delete workflow-execution-config --attrFile wec.yaml


.. code-block:: yaml
	
    domain: development
    project: flytectldemo
    max_parallelism: 5

Deletes workflow execution config for a workflow.
For the workflow 'core.control_flow.run_merge_sort.merge_sort', it is:

::

 flytectl delete workflow-execution-config -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

Usage
`
)

func deleteWorkflowExecutionConfig(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	var pwdGetter sconfig.ProjectDomainWorkflowGetter
	pwdGetter = sconfig.PDWGetterCommandLine{Config: config.GetConfig(), Args: args}
	delConfig := workflowexecutionconfig.DefaultDelConfig

	// Get the project domain workflowName from the config file or commandline params
	if len(delConfig.AttrFile) > 0 {
		// Initialize FileConfig which will be used if delConfig.AttrFile is non empty
		// And Reads from the workflow execution config file
		pwdGetter = &workflowexecutionconfig.FileConfig{}
		if err := sconfig.ReadConfigFromFile(pwdGetter, delConfig.AttrFile); err != nil {
			return err
		}
	}
	// Use the pwdGetter to initialize the project domain and workflow
	project := pwdGetter.GetProject()
	domain := pwdGetter.GetDomain()
	workflowName := pwdGetter.GetWorkflow()

	// Deletes the matchable attributes using the WorkflowExecutionConfigFileConfig
	if err := deleteMatchableAttr(ctx, project, domain, workflowName, cmdCtx.AdminDeleterExt(),
		admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG, delConfig.DryRun); err != nil {
		return err
	}

	return nil
}
