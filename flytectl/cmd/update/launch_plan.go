package update

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/flyteorg/flyte/flytectl/clierrors"
	"github.com/flyteorg/flyte/flytectl/cmd/config"
	"github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/launchplan"
	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	cmdUtil "github.com/flyteorg/flyte/flytectl/pkg/commandutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

const (
	updateLPShort = "Updates launch plan status"
	updateLPLong  = `
Activates a ` + "`launch plan <https://docs.flyte.org/en/latest/user_guide/productionizing/schedules.html#activating-a-schedule>`__" + ` which activates the scheduled job associated with it:
::

 flytectl update launchplan -p flytesnacks -d development core.control_flow.merge_sort.merge_sort --version v1 --activate

Deactivates a ` + "`launch plan <https://docs.flyte.org/en/latest/user_guide/productionizing/schedules.html#deactivating-a-schedule>`__" + ` which deschedules any scheduled job associated with it:
::

 flytectl update launchplan -p flytesnacks -d development core.control_flow.merge_sort.merge_sort --version v1 --deactivate

Archives a launch plan version, marking it as old/unused. Archived launch plans can be filtered out
of list queries using ne(state,2) and any active schedule is disabled. Archived launch plans can
still be used to launch executions:
::

 flytectl update launchplan -p flytesnacks -d development core.control_flow.merge_sort.merge_sort --version v1 --archive

Usage
`
)

func updateLPFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	project := config.GetConfig().Project
	domain := config.GetConfig().Domain
	if len(args) != 1 {
		return errors.New(clierrors.ErrLPNotPassed)
	}
	name := args[0]
	version := launchplan.UConfig.Version
	if len(version) == 0 {
		return errors.New(clierrors.ErrLPVersionNotPassed)
	}

	activate := launchplan.UConfig.Activate
	archive := launchplan.UConfig.Archive
	deactivate := launchplan.UConfig.Deactivate

	setCount := 0
	if activate {
		setCount++
	}
	if deactivate {
		setCount++
	}
	if archive {
		setCount++
	}
	if setCount > 1 {
		return errors.New(clierrors.ErrInvalidBothStateUpdate)
	}

	var newState admin.LaunchPlanState
	if activate {
		newState = admin.LaunchPlanState_ACTIVE
	} else if deactivate {
		newState = admin.LaunchPlanState_INACTIVE
	} else if archive {
		newState = admin.LaunchPlanState_ARCHIVED
	}

	id := &core.Identifier{
		Project:      project,
		Domain:       domain,
		Name:         name,
		Version:      version,
		ResourceType: core.ResourceType_LAUNCH_PLAN,
	}

	launchPlan, err := cmdCtx.AdminClient().GetLaunchPlan(ctx, &admin.ObjectGetRequest{Id: id})
	if err != nil {
		return fmt.Errorf("update launch plan %s: could not fetch launch plan: %w", name, err)
	}
	oldState := launchPlan.GetClosure().GetState()

	type LaunchPlan struct {
		State admin.LaunchPlanState `json:"state"`
	}
	patch, err := DiffAsYaml(diffPathBefore, diffPathAfter, LaunchPlan{oldState}, LaunchPlan{newState})
	if err != nil {
		panic(err)
	}

	if patch == "" {
		fmt.Printf("No changes detected. Skipping the update.\n")
		return nil
	}

	fmt.Printf("The following changes are to be applied.\n%s\n", patch)

	if launchplan.UConfig.DryRun {
		fmt.Printf("skipping LaunchPlanUpdate request (DryRun)")
		return nil
	}

	if !launchplan.UConfig.Force && !cmdUtil.AskForConfirmation("Continue?", os.Stdin) {
		return fmt.Errorf("update aborted by user")
	}

	_, err = cmdCtx.AdminClient().UpdateLaunchPlan(ctx, &admin.LaunchPlanUpdateRequest{
		Id:    id,
		State: newState,
	})
	if err != nil {
		return fmt.Errorf(clierrors.ErrFailedLPUpdate, name, err)
	}

	fmt.Printf("updated launch plan successfully on %s", name)

	return nil
}
