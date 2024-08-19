package batchscheduler

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
	schedulerConfig "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/batchscheduler/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/batchscheduler/scheduler/kubernetes"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/batchscheduler/scheduler/yunikorn"
)

type SchedulerPlugin interface {
	GetSchedulerName() string
	ParseJob(config *schedulerConfig.Config, metadata *metav1.ObjectMeta, workerGroupsSpec []*plugins.WorkerGroupSpec, pod *v1.PodSpec, primaryContainerIdx int) error
	ProcessHead(metadata *metav1.ObjectMeta, head *v1.PodSpec, index int)
	ProcessWorker(metadata *metav1.ObjectMeta, worker *v1.PodSpec, index int)
	AfterProcess(metadata *metav1.ObjectMeta)
}

func NewSchedulerPlugin(config *schedulerConfig.Config) SchedulerPlugin {
	switch config.GetScheduler() {
	case yunikorn.Yunikorn:
		return yunikorn.NewYunikornPlugin()
	default:
		return kubernetes.NewDefaultPlugin()
	}
}
