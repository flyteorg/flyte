package update

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/workflowexecutionconfig"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

	sconfig "github.com/flyteorg/flyte/flytectl/cmd/config/subcommand"
	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
)

const (
	workflowExecutionConfigShort = "Updates matchable resources of workflow execution config"
	workflowExecutionConfigLong  = `
Updates the workflow execution config for the given project and domain combination or additionally with workflow name.

Updating the workflow execution config is only available from a generated file. See the get section for generating this file.
This will completely overwrite any existing custom project and domain and workflow combination execution config.
It is preferable to do get and generate a config file if there is an existing execution config already set and then update it to have new values.
Refer to get workflow-execution-config section on how to generate this file.
It takes input for workflow execution config from the config file wec.yaml,
Example: content of wec.yaml:

.. code-block:: yaml

    domain: development
    project: flytesnacks
    max_parallelism: 5
    security_context:
      run_as:
        k8s_service_account: demo

::

 flytectl update workflow-execution-config --attrFile wec.yaml

Update workflow execution config for project, domain, and workflow combination. This will take precedence over any other
execution config defined at project domain level.
For workflow 'core.control_flow.merge_sort.merge_sort' in flytesnacks project, development domain, it is:

.. code-block:: yaml

    domain: development
    project: flytesnacks
    workflow: core.control_flow.merge_sort.merge_sort
    max_parallelism: 5
    security_context:
      run_as:
        k8s_service_account: mergesortsa

::

 flytectl update workflow-execution-config --attrFile wec.yaml

Usage

`
)

func updateWorkflowExecutionConfigFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	updateConfig := workflowexecutionconfig.DefaultUpdateConfig
	if len(updateConfig.AttrFile) == 0 {
		return fmt.Errorf("attrFile is mandatory while calling update for workflow execution config")
	}

	workflowExecutionConfigFileConfig := workflowexecutionconfig.FileConfig{}
	if err := sconfig.ReadConfigFromFile(&workflowExecutionConfigFileConfig, updateConfig.AttrFile); err != nil {
		return err
	}

	// Get project domain workflow name from the read file.
	project := workflowExecutionConfigFileConfig.Project
	domain := workflowExecutionConfigFileConfig.Domain
	workflowName := workflowExecutionConfigFileConfig.Workflow

	if err := DecorateAndUpdateMatchableAttr(ctx, cmdCtx, project, domain, workflowName,
		admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG, workflowExecutionConfigFileConfig,
		updateConfig.DryRun, updateConfig.Force); err != nil {
		return err
	}
	return nil
}
