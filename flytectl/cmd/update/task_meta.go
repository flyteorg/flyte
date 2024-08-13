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
	updateTaskShort = "Update task metadata"
	updateTaskLong  = `
Update the description on the task:
::

 flytectl update task-meta  -d development -p flytesnacks core.control_flow.merge_sort.merge --description "Merge sort example"

Archiving task named entity is not supported and would throw an error:
::

 flytectl update task-meta -d development -p flytesnacks core.control_flow.merge_sort.merge --archive

Activating task named entity would be a noop since archiving is not possible:
::

 flytectl update task-meta -d development -p flytesnacks core.control_flow.merge_sort.merge --activate

Usage
`
)

func getUpdateTaskFunc(namedEntityConfig *NamedEntityConfig) func(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	return func(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
		project := config.GetConfig().Project
		domain := config.GetConfig().Domain
		if len(args) != 1 {
			return fmt.Errorf(clierrors.ErrTaskNotPassed)
		}

		name := args[0]
		err := namedEntityConfig.UpdateNamedEntity(ctx, name, project, domain, core.ResourceType_TASK, cmdCtx)
		if err != nil {
			fmt.Printf(clierrors.ErrFailedTaskUpdate, name, err)
			return err
		}

		fmt.Printf("updated metadata successfully on %v", name)
		return nil
	}
}
