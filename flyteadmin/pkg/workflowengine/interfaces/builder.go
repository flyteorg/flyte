package interfaces

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

//go:generate mockery -name FlyteWorkflowBuilder -output=../mocks -case=underscore

// FlyteWorkflowBuilder produces a v1alpha1.FlyteWorkflow definition from a compiled workflow closure and execution inputs
type FlyteWorkflowBuilder interface {
	Build(
		wfClosure *core.CompiledWorkflowClosure, inputs *core.LiteralMap, executionID *core.WorkflowExecutionIdentifier,
		namespace string) (*v1alpha1.FlyteWorkflow, error)
}
