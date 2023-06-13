package launchplan

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytestdlib/errors"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
)

type failFastWorkflowLauncher struct {
	Executor
	Reader
}

func (failFastWorkflowLauncher) Launch(ctx context.Context, launchCtx LaunchContext, executionID *core.WorkflowExecutionIdentifier, launchPlanRef *core.Identifier, inputs *core.LiteralMap) error {
	logger.Infof(ctx, "Fail: Launch Workflow requested with ExecID [%s], LaunchPlan [%s]", executionID.Name, fmt.Sprintf("%s:%s:%s", launchPlanRef.Project, launchPlanRef.Domain, launchPlanRef.Name))
	return errors.Wrapf(RemoteErrorUser, fmt.Errorf("badly configured system"), "please enable admin workflow launch to use launchplans")
}

func (failFastWorkflowLauncher) GetStatus(ctx context.Context, executionID *core.WorkflowExecutionIdentifier) (*admin.ExecutionClosure, *core.LiteralMap, error) {
	logger.Infof(ctx, "NOOP: Workflow Status ExecID [%s]", executionID.Name)
	return nil, nil, errors.Wrapf(RemoteErrorUser, fmt.Errorf("badly configured system"), "please enable admin workflow launch to use launchplans")
}

func (failFastWorkflowLauncher) Kill(ctx context.Context, executionID *core.WorkflowExecutionIdentifier, reason string) error {
	return nil
}

func (failFastWorkflowLauncher) GetLaunchPlan(ctx context.Context, launchPlanRef *core.Identifier) (*admin.LaunchPlan, error) {
	return nil, nil
}

// Initializes Executor.
func (failFastWorkflowLauncher) Initialize(ctx context.Context) error {
	return nil
}

func NewFailFastLaunchPlanExecutor() FlyteAdmin {
	logger.Infof(context.TODO(), "created failFast workflow launcher, will not launch subworkflows.")
	return &failFastWorkflowLauncher{}
}
