package yunikorn

import (
	"context"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/batchscheduler"
	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Yunikorn            = "yunikorn"
	AppID               = "yunikorn.apache.org/app-id"
	TaskGroupNameKey    = "yunikorn.apache.org/task-group-name"
	TaskGroupsKey       = "yunikorn.apache.org/task-groups"
	TaskGroupParameters = "yunikorn.apache.org/schedulingPolicyParameters"
)

type YunikornSchedulerManager struct {
	parameters string
}

func (y *YunikornSchedulerManager) Mutate(ctx context.Context, object client.Object) error {
	switch object.(type) {
	case *rayv1.RayJob:
		return MutateRayJob(y.parameters, object.(*rayv1.RayJob))
	default:

	}
	return nil
}

func NewSchedulerManager(cfg *batchscheduler.Config) batchscheduler.SchedulerManager {
	return &YunikornSchedulerManager{parameters: cfg.Parameters}
}
