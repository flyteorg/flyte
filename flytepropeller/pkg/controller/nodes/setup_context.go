package nodes

import (
	"context"

	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/interfaces"
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

func (c *recursiveNodeExecutor) newSetupContext(_ context.Context) interfaces.SetupContext {
	return &setupContext{
		enq:   c.enqueueWorkflow,
		scope: c.metrics.Scope,
	}
}
