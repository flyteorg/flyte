package update

import (
	"context"
	"fmt"

	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/executionclusterlabel"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
)

const (
	executionClusterLabelShort = "Updates matchable resources of execution cluster label"
	executionClusterLabelLong  = `
Updates execution cluster label for given project and domain combination or additionally with workflow name.

Updating to the execution cluster label is only available from a generated file. See the get section for generating this file.
Here the command updates takes the input for execution cluster label from the config file ecl.yaml
eg:  content of ecl.yaml

.. code-block:: yaml

    domain: development
    project: flytectldemo
    value: foo

::

 flytectl update execution-cluster-label --attrFile ecl.yaml

Updating execution cluster label for project and domain and workflow combination. This will take precedence over any other
execution cluster label defined at project domain level.
Update the execution cluster label for workflow core.control_flow.run_merge_sort.merge_sort in flytectldemo, development domain

.. code-block:: yaml

    domain: development
    project: flytectldemo
    workflow: core.control_flow.run_merge_sort.merge_sort
    value: foo

::

 flytectl update execution-cluster-label --attrFile ecl.yaml

Usage

`
)

func updateExecutionClusterLabelFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	updateConfig := executionclusterlabel.DefaultUpdateConfig
	if len(updateConfig.AttrFile) == 0 {
		return fmt.Errorf("attrFile is mandatory while calling update for execution cluster label")
	}

	executionClusterLabelFileConfig := executionclusterlabel.FileConfig{}
	if err := sconfig.ReadConfigFromFile(&executionClusterLabelFileConfig, updateConfig.AttrFile); err != nil {
		return err
	}

	// Get project domain workflow name from the read file.
	project := executionClusterLabelFileConfig.Project
	domain := executionClusterLabelFileConfig.Domain
	workflowName := executionClusterLabelFileConfig.Workflow

	// Updates the admin matchable attribute from executionClusterLabelFileConfig
	if err := DecorateAndUpdateMatchableAttr(ctx, project, domain, workflowName, cmdCtx.AdminUpdaterExt(),
		executionClusterLabelFileConfig); err != nil {
		return err
	}
	return nil
}
