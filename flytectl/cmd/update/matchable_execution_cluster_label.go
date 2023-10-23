package update

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/executionclusterlabel"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
)

const (
	executionClusterLabelShort = "Update matchable resources of execution cluster label"
	executionClusterLabelLong  = `
Update execution cluster label for the given project and domain combination or additionally with workflow name.

Updating to the execution cluster label is only available from a generated file. See the get section to generate this file.
It takes input for execution cluster label from the config file ecl.yaml
Example: content of ecl.yaml:

.. code-block:: yaml

    domain: development
    project: flytesnacks
    value: foo

::

 flytectl update execution-cluster-label --attrFile ecl.yaml

Update execution cluster label for project, domain, and workflow combination. This will take precedence over any other
execution cluster label defined at project domain level.
For workflow 'core.control_flow.merge_sort.merge_sort' in flytesnacks project, development domain, it is:

.. code-block:: yaml

    domain: development
    project: flytesnacks
    workflow: core.control_flow.merge_sort.merge_sort
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

	if err := DecorateAndUpdateMatchableAttr(ctx, cmdCtx, project, domain, workflowName,
		admin.MatchableResource_EXECUTION_CLUSTER_LABEL, executionClusterLabelFileConfig,
		updateConfig.DryRun, updateConfig.Force); err != nil {
		return err
	}
	return nil
}
