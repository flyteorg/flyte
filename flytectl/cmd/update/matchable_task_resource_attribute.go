package update

import (
	"context"
	"fmt"

	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/taskresourceattribute"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
)

const (
	taskResourceAttributesShort = "Update matchable resources of task attributes"
	taskResourceAttributesLong  = `
Updates the task resource attributes for the given project and domain combination or additionally with workflow name.

Updating the task resource attribute is only available from a generated file. See the get section for generating this file.
This will completely overwrite any existing custom project, domain, and workflow combination attributes.
It is preferable to do get and generate an attribute file if there is an existing attribute already set and then update it to have new values.
Refer to get task-resource-attribute section on how to generate this file.
It takes input for task resource attributes from the config file tra.yaml,
Example: content of tra.yaml:

.. code-block:: yaml

    domain: development
    project: flytesnacks
    defaults:
      cpu: "1"
      memory: "150Mi"
    limits:
      cpu: "2"
      memory: "450Mi"

::

 flytectl update task-resource-attribute --attrFile tra.yaml

Update task resource attribute for project, domain, and workflow combination. This will take precedence over any other
resource attribute defined at project domain level.
For workflow 'core.control_flow.merge_sort.merge_sort' in flytesnacks project, development domain, it is:

.. code-block:: yaml

    domain: development
    project: flytesnacks
    workflow: core.control_flow.merge_sort.merge_sort
    defaults:
      cpu: "1"
      memory: "150Mi"
    limits:
      cpu: "2"
      memory: "450Mi"

::

 flytectl update task-resource-attribute --attrFile tra.yaml

Usage

`
)

func updateTaskResourceAttributesFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	updateConfig := taskresourceattribute.DefaultUpdateConfig
	if len(updateConfig.AttrFile) == 0 {
		return fmt.Errorf("attrFile is mandatory while calling update for task resource attribute")
	}

	taskResourceAttrFileConfig := taskresourceattribute.TaskResourceAttrFileConfig{}
	if err := sconfig.ReadConfigFromFile(&taskResourceAttrFileConfig, updateConfig.AttrFile); err != nil {
		return err
	}

	// Get project domain workflow name from the read file.
	project := taskResourceAttrFileConfig.Project
	domain := taskResourceAttrFileConfig.Domain
	workflowName := taskResourceAttrFileConfig.Workflow

	// Updates the admin matchable attribute from taskResourceAttrFileConfig
	if err := DecorateAndUpdateMatchableAttr(ctx, project, domain, workflowName, cmdCtx.AdminUpdaterExt(),
		taskResourceAttrFileConfig, updateConfig.DryRun); err != nil {
		return err
	}
	return nil
}
