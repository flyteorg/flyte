package update

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/workflowexecutionconfig"

	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
)

const (
	workflowExecutionConfigShort = "Updates matchable resources of workflow execution config"
	workflowExecutionConfigLong  = `
Updates workflow execution config for given project and domain combination or additionally with workflow name.

Updating the workflow execution config is only available from a generated file. See the get section for generating this file.
Also this will completely overwrite any existing custom project and domain and workflow combination execution config.
Would be preferable to do get and generate an config file if there is an existing execution config already set and then update it to have new values
Refer to get workflow-execution-config section on how to generate this file
Here the command updates takes the input for workflow execution config from the config file wec.yaml
eg:  content of wec.yaml

.. code-block:: yaml

    domain: development
    project: flytectldemo
    max_parallelism: 5

::

 flytectl update workflow-execution-config --attrFile wec.yaml

Updating workflow execution config for project and domain and workflow combination. This will take precedence over any other
execution config defined at project domain level.
Update the workflow execution config for workflow core.control_flow.run_merge_sort.merge_sort in flytectldemo, development domain

.. code-block:: yaml

    domain: development
    project: flytectldemo
    workflow: core.control_flow.run_merge_sort.merge_sort
    max_parallelism: 5

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

	// Updates the admin matchable attribute from workflowExecutionConfigFileConfig
	if err := DecorateAndUpdateMatchableAttr(ctx, project, domain, workflowName, cmdCtx.AdminUpdaterExt(),
		workflowExecutionConfigFileConfig); err != nil {
		return err
	}
	return nil
}
