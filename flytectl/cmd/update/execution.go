package update

import (
	"context"
	"fmt"
	"os"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytectl/clierrors"
	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/execution"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	cmdUtil "github.com/flyteorg/flytectl/pkg/commandutils"
)

const (
	updateExecutionShort = "Updates the execution status"
	updateExecutionLong  = `
Activate an execution; and it shows up in the CLI and UI:
::

 flytectl update execution -p flytesnacks -d development  oeh94k9r2r --activate

Archive an execution; and it is hidden from the CLI and UI:
::

 flytectl update execution -p flytesnacks -d development  oeh94k9r2r --archive


Usage
`
)

func updateExecutionFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	project := config.GetConfig().Project
	domain := config.GetConfig().Domain
	if len(args) != 1 {
		return fmt.Errorf(clierrors.ErrExecutionNotPassed)
	}
	executionName := args[0]
	activate := execution.UConfig.Activate
	archive := execution.UConfig.Archive
	if activate && archive {
		return fmt.Errorf(clierrors.ErrInvalidStateUpdate)
	}

	var newState admin.ExecutionState
	if activate {
		newState = admin.ExecutionState_EXECUTION_ACTIVE
	} else if archive {
		newState = admin.ExecutionState_EXECUTION_ARCHIVED
	}

	exec, err := cmdCtx.AdminFetcherExt().FetchExecution(ctx, executionName, project, domain)
	if err != nil {
		return fmt.Errorf("update execution: could not fetch execution %s: %w", executionName, err)
	}
	oldState := exec.GetClosure().GetStateChangeDetails().GetState()

	type Execution struct {
		State admin.ExecutionState `json:"state"`
	}
	patch, err := DiffAsYaml(diffPathBefore, diffPathAfter, Execution{oldState}, Execution{newState})
	if err != nil {
		panic(err)
	}

	if patch == "" {
		fmt.Printf("No changes detected. Skipping the update.\n")
		return nil
	}

	fmt.Printf("The following changes are to be applied.\n%s\n", patch)

	if execution.UConfig.DryRun {
		fmt.Printf("skipping UpdateExecution request (DryRun)\n")
		return nil
	}

	if !execution.UConfig.Force && !cmdUtil.AskForConfirmation("Continue?", os.Stdin) {
		return fmt.Errorf("update aborted by user")
	}

	_, err = cmdCtx.AdminClient().UpdateExecution(ctx, &admin.ExecutionUpdateRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: project,
			Domain:  domain,
			Name:    executionName,
		},
		State: newState,
	})
	if err != nil {
		fmt.Printf(clierrors.ErrFailedExecutionUpdate, executionName, err)
		return err
	}

	fmt.Printf("updated execution %s successfully to state %s\n", executionName, newState)
	return nil
}
