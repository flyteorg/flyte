package update

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyte/flytectl/clierrors"
	"github.com/flyteorg/flyte/flytectl/cmd/config"
	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

const (
	updateWorkflowShort = "Update workflow metadata"
	updateWorkflowLong  = `
Update the description on the workflow:
::

 flytectl update workflow-meta -p flytesnacks -d development core.control_flow.merge_sort.merge_sort --description "Mergesort workflow example"

Archiving workflow named entity would cause this to disappear from flyteconsole UI:
::

 flytectl update workflow-meta -p flytesnacks -d development core.control_flow.merge_sort.merge_sort --archive

Activate workflow named entity:
::

 flytectl update workflow-meta -p flytesnacks -d development core.control_flow.merge_sort.merge_sort --activate

Usage
`
)

func getUpdateWorkflowFunc(namedEntityConfig *NamedEntityConfig) func(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	return func(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
		project := config.GetConfig().Project
		domain := config.GetConfig().Domain
		if len(args) != 1 {
			return fmt.Errorf(clierrors.ErrWorkflowNotPassed)
		}
		name := args[0]
		err := namedEntityConfig.UpdateNamedEntity(ctx, name, project, domain, core.ResourceType_WORKFLOW, cmdCtx)
		if err != nil {
			fmt.Printf(clierrors.ErrFailedWorkflowUpdate, name, err)
			return err
		}
		fmt.Printf("updated metadata successfully on %v", name)
		return nil
	}
}
