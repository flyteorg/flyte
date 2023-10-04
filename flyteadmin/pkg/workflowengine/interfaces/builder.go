package interfaces

import (
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

//go:generate mockery -name FlyteWorkflowBuilder -output=../mocks -case=underscore

// FlyteWorkflowBuilder produces a v1alpha1.FlyteWorkflow definition from a compiled workflow closure and execution inputs
type FlyteWorkflowBuilder interface {
	Build(
		wfClosure *core.CompiledWorkflowClosure, inputs *core.LiteralMap, executionID *core.WorkflowExecutionIdentifier,
		namespace string) (*v1alpha1.FlyteWorkflow, error)
}
