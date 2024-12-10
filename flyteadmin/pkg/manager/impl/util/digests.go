package util

import (
	"context"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/pbhash"
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
	// using a CompiledWorkflowClosure where the LaunchPlans have been stripped to mitigate forced
	// re-registration of all workflows that contain launchplans on the addition of subworkflow
	// and launchplan caching support. this is not the proper fix, the correct approach is to
	// include a compiler version in workflow version to ensure CompiledWorkflowClosures will
	// seamlessly re-register upon compiler updates.
	strippedWorkflowClosure := proto.Clone(workflowClosure).(*core.CompiledWorkflowClosure)
	strippedWorkflowClosure.LaunchPlans = nil

	workflowDigest, err := pbhash.ComputeHash(ctx, strippedWorkflowClosure)
	if err != nil {
		logger.Warningf(ctx, "failed to hash workflow [%+v] to digest with err %v",
			workflowClosure.Primary.Template.Id, err)
		return nil, errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to hash workflow [%+v] to digest with err %v", workflowClosure.Primary.Template.Id, err)
	}

	return workflowDigest, nil
}

// Returns a unique string digest for equivalent project configuration document
// To avoid digest inconsistencies, the version field is removed and the configuration document is cleaned up before hashing
func GetConfigurationDocumentStringDigest(ctx context.Context, configDoc *admin.ConfigurationDocument) (string, error) {
	configDoc.Version = ""
	digest, err := pbhash.ComputeHashString(ctx, configDoc)
	if err != nil {
		logger.Warningf(ctx, "failed to hash configuration [%+v] to string digest with err %v",
			configDoc, err)
		return "", errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to hash configuration [%+v] to string digest with err %v", configDoc, err)
	}

	return digest, nil
}
