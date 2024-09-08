package scheduler

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NoopSchedulerManager struct{}

func NewNoopSchedulerManager() *NoopSchedulerManager {
	return &NoopSchedulerManager{}
}

func (p *NoopSchedulerManager) Mutate(ctx context.Context, object client.Object) error {
	return nil
}
