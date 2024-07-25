package batchscheduler

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
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
	config *Config,
	metadata *metav1.ObjectMeta,
	workerGroupsSpec []*plugins.WorkerGroupSpec,
	pod *v1.PodSpec,
	primaryContainerIdx int,
) error {
	return nil
}
func (d *DefaultPlugin) ProcessHead(metadata *metav1.ObjectMeta, head *v1.PodSpec)                {}
func (d *DefaultPlugin) ProcessWorker(metadata *metav1.ObjectMeta, worker *v1.PodSpec, index int) {}
func (d *DefaultPlugin) AfterProcess(metadata *metav1.ObjectMeta)                                 {}
