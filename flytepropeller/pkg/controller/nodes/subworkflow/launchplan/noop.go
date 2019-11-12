package launchplan

import (
	"context"
	"fmt"

	"github.com/lyft/flytestdlib/errors"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/logger"
)

type failFastWorkflowLauncher struct {
}

func (failFastWorkflowLauncher) Launch(ctx context.Context, launchCtx LaunchContext, executionID *core.WorkflowExecutionIdentifier, launchPlanRef *core.Identifier, inputs *core.LiteralMap) error {
	logger.Infof(ctx, "Fail: Launch Workflow requested with ExecID [%s], LaunchPlan [%s]", executionID.Name, fmt.Sprintf("%s:%s:%s", launchPlanRef.Project, launchPlanRef.Domain, launchPlanRef.Name))
	return errors.Wrapf(RemoteErrorUser, fmt.Errorf("badly configured system"), "please enable admin workflow launch to use launchplans")
}

func (failFastWorkflowLauncher) GetStatus(ctx context.Context, executionID *core.WorkflowExecutionIdentifier) (*admin.ExecutionClosure, error) {
	logger.Infof(ctx, "NOOP: Workflow Status ExecID [%s]", executionID.Name)
	return nil, errors.Wrapf(RemoteErrorUser, fmt.Errorf("badly configured system"), "please enable admin workflow launch to use launchplans")
}

func (failFastWorkflowLauncher) Kill(ctx context.Context, executionID *core.WorkflowExecutionIdentifier, reason string) error {
	return nil
}

// Initializes Executor.
func (failFastWorkflowLauncher) Initialize(ctx context.Context) error {
	return nil
}

func NewFailFastLaunchPlanExecutor() Executor {
	logger.Infof(context.TODO(), "created failFast workflow launcher, will not launch subworkflows.")
	return &failFastWorkflowLauncher{}
}
