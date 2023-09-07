package util

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/pbhash"
	"google.golang.org/grpc/codes"
)

// Returns a unique digest for functionally equivalent launch plans
func GetLaunchPlanDigest(ctx context.Context, launchPlan *admin.LaunchPlan) ([]byte, error) {
	launchPlanDigest, err := pbhash.ComputeHash(ctx, launchPlan)
	if err != nil {
		logger.Warningf(ctx, "failed to hash launch plan [%+v] to digest with err %v",
			launchPlan.Id, err)
		return nil, errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to hash launch plan [%+v] to digest with err %v", launchPlan.Id, err)
	}

	return launchPlanDigest, nil
}

// Returns a unique digest for functionally equivalent compiled tasks
func GetTaskDigest(ctx context.Context, task *core.CompiledTask) ([]byte, error) {
	taskDigest, err := pbhash.ComputeHash(ctx, task)
	if err != nil {
		logger.Warningf(ctx, "failed to hash task [%+v] to digest with err %v",
			task.Template.Id, err)
		return nil, errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to hash task [%+v] to digest with err %v", task.Template.Id, err)
	}

	return taskDigest, nil
}

// Returns a unique digest for functionally equivalent compiled workflows
func GetWorkflowDigest(ctx context.Context, workflowClosure *core.CompiledWorkflowClosure) ([]byte, error) {
	workflowDigest, err := pbhash.ComputeHash(ctx, workflowClosure)
	if err != nil {
		logger.Warningf(ctx, "failed to hash workflow [%+v] to digest with err %v",
			workflowClosure.Primary.Template.Id, err)
		return nil, errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to hash workflow [%+v] to digest with err %v", workflowClosure.Primary.Template.Id, err)
	}

	return workflowDigest, nil
}
