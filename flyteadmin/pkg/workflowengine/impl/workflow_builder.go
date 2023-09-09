package impl

import (
	"github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"

	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/compiler/transformers/k8s"
)

type builderMetrics struct {
	Scope                  promutils.Scope
	WorkflowBuildSuccesses prometheus.Counter
	WorkflowBuildFailures  prometheus.Counter
	InvalidExecutionID     prometheus.Counter
}

type flyteWorkflowBuilder struct {
	metrics builderMetrics
}

func (b *flyteWorkflowBuilder) Build(
	wfClosure *core.CompiledWorkflowClosure, inputs *core.LiteralMap, executionID *core.WorkflowExecutionIdentifier,
	namespace string) (*v1alpha1.FlyteWorkflow, error) {
	flyteWorkflow, err := k8s.BuildFlyteWorkflow(wfClosure, inputs, executionID, namespace)
	if err != nil {
		b.metrics.WorkflowBuildFailures.Inc()
		return nil, err
	}
	b.metrics.WorkflowBuildSuccesses.Inc()
	return flyteWorkflow, nil
}

func newBuilderMetrics(scope promutils.Scope) builderMetrics {
	return builderMetrics{
		Scope: scope,
		WorkflowBuildSuccesses: scope.MustNewCounter("build_successes",
			"count of workflows built by propeller without error"),
		WorkflowBuildFailures: scope.MustNewCounter("build_failures",
			"count of workflows built by propeller with errors"),
	}
}

func NewFlyteWorkflowBuilder(scope promutils.Scope) interfaces.FlyteWorkflowBuilder {
	return &flyteWorkflowBuilder{
		metrics: newBuilderMetrics(scope),
	}
}
