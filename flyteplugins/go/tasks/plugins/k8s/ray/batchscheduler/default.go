package batchscheduler

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	DefaultScheduler = "default"
)

type DefaultPlugin struct{}

func NewDefaultPlugin() *DefaultPlugin {
	return &DefaultPlugin{}
}

func (d *DefaultPlugin) GetSchedulerName() string { return DefaultScheduler }

func (d *DefaultPlugin) ParseJob(
	config *BatchSchedulerConfig,
	metadata *metav1.ObjectMeta,
	workerGroupsSpec []*plugins.WorkerGroupSpec,
	pod *v1.PodSpec,
	primaryContainerIdx int,
) error {
	return nil
}
func (d *DefaultPlugin) ProcessHead(metadata *metav1.ObjectMeta, head *v1.PodSpec) {}
func (d *DefaultPlugin) ProcessWorker(metadata *metav1.ObjectMeta, worker *v1.PodSpec, index int) {}
func (d *DefaultPlugin) AfterProcess(metadata *metav1.ObjectMeta) {}
