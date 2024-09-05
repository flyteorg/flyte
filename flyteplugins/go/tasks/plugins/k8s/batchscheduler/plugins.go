package batchscheduler

import (
	schedulerConfig "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/batchscheduler/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/batchscheduler/scheduler/kubernetes"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/batchscheduler/scheduler/yunikorn"
)

type SchedulerPlugin interface {
	Process(app interface{}) error
}

func NewSchedulerPlugin(cfg *schedulerConfig.Config) SchedulerPlugin {
	switch cfg.GetScheduler() {
	case yunikorn.Yunikorn:
		return yunikorn.NewPlugin(cfg.GetParameters())
	default:
		return kubernetes.NewPlugin()
	}
}
