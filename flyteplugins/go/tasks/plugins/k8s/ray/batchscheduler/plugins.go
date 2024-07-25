package batchscheduler

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
)

type SchedulerPlugin interface {
	GetSchedulerName() string
	ParseJob(config *Config, metadata *metav1.ObjectMeta, workerGroupsSpec []*plugins.WorkerGroupSpec, pod *v1.PodSpec, primaryContainerIdx int) error
	ProcessHead(metadata *metav1.ObjectMeta, head *v1.PodSpec)
	ProcessWorker(metadata *metav1.ObjectMeta, worker *v1.PodSpec, index int)
	AfterProcess(metadata *metav1.ObjectMeta)
}

func NewSchedulerPlugin(config *Config) SchedulerPlugin {
	switch config.GetScheduler() {
	case Yunikorn:
		return NewYunikornPlugin()
	default:
		return NewDefaultPlugin()
	}
}
