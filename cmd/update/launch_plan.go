package update

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytectl/clierrors"
	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/launchplan"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
)

const (
	updateLPShort = "Updates launch plan status"
	updateLPLong  = `
Activates a launch plan which activates the scheduled job associated with it:
::

 flytectl update launchplan -p flytectldemo -d development  core.advanced.run_merge_sort.merge_sort --version v1 --activate

Archives a launch plan which deschedules any scheduled job associated with it:
::

 flytectl update launchplan -p flytectldemo -d development  core.advanced.run_merge_sort.merge_sort --version v1 --archive


Usage
`
)

func updateLPFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	project := config.GetConfig().Project
	domain := config.GetConfig().Domain
	if len(args) != 1 {
		return fmt.Errorf(clierrors.ErrLPNotPassed)
	}
	name := args[0]
	version := launchplan.UConfig.Version
	if len(version) == 0 {
		return fmt.Errorf(clierrors.ErrLPVersionNotPassed)
	}
	activateLP := launchplan.UConfig.Activate
	archiveLP := launchplan.UConfig.Archive
	if activateLP == archiveLP && archiveLP {
		return fmt.Errorf(clierrors.ErrInvalidStateUpdate)
	}

	var lpState admin.LaunchPlanState
	if activateLP {
		lpState = admin.LaunchPlanState_ACTIVE
	} else if archiveLP {
		lpState = admin.LaunchPlanState_INACTIVE
	}

	if launchplan.UConfig.DryRun {
		logger.Debugf(ctx, "skipping CreateExecution request (DryRun)")
	} else {
		_, err := cmdCtx.AdminClient().UpdateLaunchPlan(ctx, &admin.LaunchPlanUpdateRequest{
			Id: &core.Identifier{
				Project: project,
				Domain:  domain,
				Name:    name,
				Version: version,
			},
			State: lpState,
		})
		if err != nil {
			fmt.Printf(clierrors.ErrFailedLPUpdate, name, err)
			return err
		}
	}
	fmt.Printf("updated launchplan successfully on %v", name)

	return nil
}
