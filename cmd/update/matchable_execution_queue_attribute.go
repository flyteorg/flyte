package update

import (
	"context"
	"fmt"

	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/executionqueueattribute"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
)

const (
	executionQueueAttributesShort = "Update matchable resources of execution queue attributes"
	executionQueueAttributesLong  = `
Update execution queue attributes for the given project and domain combination or additionally with workflow name.

Updating the execution queue attribute is only available from a generated file. See the get section for generating this file.
This will completely overwrite any existing custom project, domain, and workflow combination attributes.
It is preferable to do get and generate an attribute file if there is an existing attribute that is already set and then update it to have new values.
Refer to get execution-queue-attribute section on how to generate this file
It takes input for execution queue attributes from the config file era.yaml,
Example: content of era.yaml:

.. code-block:: yaml

    domain: development
    project: flytectldemo
    tags:
      - foo
      - bar
      - buzz
      - lightyear

::

 flytectl update execution-queue-attribute --attrFile era.yaml

Update execution queue attribute for project, domain, and workflow combination. This will take precedence over any other
execution queue attribute defined at project domain level.
For workflow 'core.control_flow.run_merge_sort.merge_sort' in flytectldemo project, development domain, it is:

.. code-block:: yaml

    domain: development
    project: flytectldemo
    workflow: core.control_flow.run_merge_sort.merge_sort
    tags:
      - foo
      - bar
      - buzz
      - lightyear

::

 flytectl update execution-queue-attribute --attrFile era.yaml

Usage

`
)

func updateExecutionQueueAttributesFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	updateConfig := executionqueueattribute.DefaultUpdateConfig
	if len(updateConfig.AttrFile) == 0 {
		return fmt.Errorf("attrFile is mandatory while calling update for execution queue attribute")
	}

	executionQueueAttrFileConfig := executionqueueattribute.AttrFileConfig{}
	if err := sconfig.ReadConfigFromFile(&executionQueueAttrFileConfig, updateConfig.AttrFile); err != nil {
		return err
	}

	// Get project domain workflow name from the read file.
	project := executionQueueAttrFileConfig.Project
	domain := executionQueueAttrFileConfig.Domain
	workflowName := executionQueueAttrFileConfig.Workflow

	// Updates the admin matchable attribute from executionQueueAttrFileConfig
	if err := DecorateAndUpdateMatchableAttr(ctx, project, domain, workflowName, cmdCtx.AdminUpdaterExt(),
		executionQueueAttrFileConfig, updateConfig.DryRun); err != nil {
		return err
	}
	return nil
}
