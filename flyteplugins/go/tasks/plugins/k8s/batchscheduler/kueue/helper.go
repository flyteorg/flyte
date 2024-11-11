package kueue

import (
	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/batchscheduler/utils"
)

const (
	QueueName         = "kueue.x-k8s.io/queue-name"
	PriorityClassName = "kueue.x-k8s.io/priority-class"
)

func UpdateKueueLabels(labels map[string]string, app *rayv1.RayJob) {
	utils.UpdateLabels(labels, &app.ObjectMeta)
}
