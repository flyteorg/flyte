package update

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytectl/cmd/config/subcommand"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytestdlib/logger"
)

const (
	taskResourceAttributesShort = "Updates matchable resources of task attributes"
	taskResourceAttributesLong  = `
Updates task  resource attributes for given project and domain combination or additionally with workflow name.

Updating the task resource attribute is only available from a generated file. See the get section for generating this file.
Here the command updates takes the input for task resource attributes from the config file tra.yaml
eg:  content of tra.yaml

.. code-block:: yaml

	domain: development
	project: flytectldemo
	defaults:
	  cpu: "1"
	  memory: 150Mi
	limits:
	  cpu: "2"
	  memory: 450Mi

::

 flytectl update task-resource-attribute -attrFile tra.yaml

Updating task resource attribute for project and domain and workflow combination. This will take precedence over any other
resource attribute defined at project domain level.
Update the resource attributes for workflow core.control_flow.run_merge_sort.merge_sort in flytectldemo , development domain
.. code-block:: yaml

	domain: development
	project: flytectldemo
	workflow: core.control_flow.run_merge_sort.merge_sort
	defaults:
	  cpu: "1"
	  memory: 150Mi
	limits:
	  cpu: "2"
	  memory: 450Mi

::

 flytectl update task-resource-attribute -attrFile tra.yaml

Usage

`
)

func updateTaskResourceAttributesFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	updateConfig := subcommand.DefaultTaskResourceUpdateConfig
	if len(updateConfig.AttrFile) == 0 {
		return fmt.Errorf("attrFile is mandatory while calling update for task resource attribute")
	}

	taskResourceAttrFileConfig := subcommand.TaskResourceAttrFileConfig{}
	if err := taskResourceAttrFileConfig.ReadConfigFromFile(updateConfig.AttrFile); err != nil {
		return err
	}

	// Get project domain workflow name from the read file.
	project := taskResourceAttrFileConfig.Project
	domain := taskResourceAttrFileConfig.Domain
	workflowName := taskResourceAttrFileConfig.Workflow

	// decorate the taskresource Attributes with MatchingAttributes
	matchingAttr := taskResourceAttrFileConfig.MatchableAttributeDecorator()

	if len(workflowName) > 0 {
		// Update the workflow attribute using the admin.
		err := cmdCtx.AdminUpdaterExt().UpdateWorkflowAttributes(ctx, project, domain, workflowName, matchingAttr)
		if err != nil {
			return err
		}
		logger.Debugf(ctx, "Updated task resource attributes from %v project and domain %v and workflow %v", project, domain, workflowName)
	} else {
		// Update the project domain attribute using the admin.
		err := cmdCtx.AdminUpdaterExt().UpdateProjectDomainAttributes(ctx, project, domain, matchingAttr)
		if err != nil {
			return err
		}
		logger.Debugf(ctx, "Updated task resource attributes from %v project and domain %v", project, domain)
	}
	return nil
}
