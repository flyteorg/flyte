package kubernetes

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
	schedulerConfig "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/ray/batchscheduler/config"
)

var (
	DefaultScheduler = "default"
)

type Plugin struct{}

func NewDefaultPlugin() *Plugin {
	return &Plugin{}
}

func (d *Plugin) GetSchedulerName() string { return DefaultScheduler }

func (d *Plugin) ParseJob(
	config *schedulerConfig.Config,
	metadata *metav1.ObjectMeta,
	workerGroupsSpec []*plugins.WorkerGroupSpec,
	pod *v1.PodSpec,
	primaryContainerIdx int,
) error {
	return nil
}
func (d *Plugin) ProcessHead(metadata *metav1.ObjectMeta, head *v1.PodSpec, index int)     {}
func (d *Plugin) ProcessWorker(metadata *metav1.ObjectMeta, worker *v1.PodSpec, index int) {}
func (d *Plugin) AfterProcess(metadata *metav1.ObjectMeta)                                 {}
