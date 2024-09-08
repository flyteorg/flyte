package batchscheduler

import (
	"context"
	schedulerConfig "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/batchscheduler/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/batchscheduler/scheduler/yunikorn"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SchedulerPlugin interface {
	// Mutate is responsible for mutating the object to be scheduled.
	// It will add the necessary annotations, labels, etc. to the object.
	Mutate(ctx context.Context, object client.Object) error
}

type NoopSchedulerPlugin struct{}

func NewNoopSchedulerPlugin() *NoopSchedulerPlugin {
	return &NoopSchedulerPlugin{}
}

func (p *NoopSchedulerPlugin) Mutate(ctx context.Context, object client.Object) error {
	return nil
}

func NewSchedulerPlugin(t reflect.Type, cfg *schedulerConfig.Config) SchedulerPlugin {
	switch cfg.GetScheduler() {
	case yunikorn.Yunikorn:
		return yunikorn.NewPlugin(t, cfg.GetParameters())
	default:
		return NewNoopSchedulerPlugin()
	}
}
