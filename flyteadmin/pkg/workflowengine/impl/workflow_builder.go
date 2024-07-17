package impl

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/flyteorg/flyte/flyteadmin/pkg/workflowengine/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/transformers/k8s"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
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
	wfClosure *core.CompiledWorkflowClosure, inputs *core.InputData, executionID *core.WorkflowExecutionIdentifier,
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
