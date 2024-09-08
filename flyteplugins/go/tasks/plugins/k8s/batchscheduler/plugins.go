package batchscheduler

import (
	"context"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/batchscheduler/scheduler"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/batchscheduler/scheduler/yunikorn"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SchedulerManager interface {
	// Mutate is responsible for mutating the object to be scheduled.
	// It will add the necessary annotations, labels, etc. to the object.
	Mutate(ctx context.Context, object client.Object) error
}

func NewSchedulerManager(cfg *Config) SchedulerManager {
	switch cfg.GetScheduler() {
	case yunikorn.Yunikorn:
		return yunikorn.NewSchedulerManager(cfg)
	default:
		return scheduler.NewNoopSchedulerManager()
	}
}
