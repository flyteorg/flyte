package nodes

import (
	"context"

	"github.com/lyft/flytestdlib/promutils"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
)

type setupContext struct {
	enq   func(string)
	scope promutils.Scope
}

func (s *setupContext) EnqueueOwner() func(string) {
	return s.enq
}

func (s *setupContext) OwnerKind() string {
	return v1alpha1.FlyteWorkflowKind
}

func (s *setupContext) MetricsScope() promutils.Scope {
	return s.scope
}

func (c *nodeExecutor) newSetupContext(_ context.Context) handler.SetupContext {
	return &setupContext{
		enq:   c.enqueueWorkflow,
		scope: c.metrics.Scope,
	}
}
