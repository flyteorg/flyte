package yunikorn

import (
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/batchscheduler"
	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"reflect"
)

const (
	Yunikorn            = "yunikorn"
	AppID               = "yunikorn.apache.org/app-id"
	TaskGroupNameKey    = "yunikorn.apache.org/task-group-name"
	TaskGroupsKey       = "yunikorn.apache.org/task-groups"
	TaskGroupParameters = "yunikorn.apache.org/schedulingPolicyParameters"
)

func NewPlugin(t reflect.Type, parameters string) batchscheduler.SchedulerPlugin {
	switch t {
	case reflect.TypeOf(rayv1.RayJob{}):
		return NewRayHandler(parameters)
	default:
		return batchscheduler.NewNoopSchedulerPlugin()
	}
}
