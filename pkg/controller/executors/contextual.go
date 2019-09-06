package executors

import (
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

type ContextualWorkflow struct {
	v1alpha1.WorkflowMetaExtended
	v1alpha1.ExecutableSubWorkflow
	v1alpha1.NodeStatusGetter
}

func NewBaseContextualWorkflow(baseWorkflow v1alpha1.ExecutableWorkflow) v1alpha1.ExecutableWorkflow {
	return &ContextualWorkflow{
		ExecutableSubWorkflow: baseWorkflow,
		WorkflowMetaExtended:  baseWorkflow,
		NodeStatusGetter:      baseWorkflow.GetExecutionStatus(),
	}
}

// Creates a contextual workflow using the provided interface implementations.
func NewSubContextualWorkflow(baseWorkflow v1alpha1.ExecutableWorkflow, subWF v1alpha1.ExecutableSubWorkflow,
	nodeStatus v1alpha1.ExecutableNodeStatus) v1alpha1.ExecutableWorkflow {

	return &ContextualWorkflow{
		ExecutableSubWorkflow: subWF,
		WorkflowMetaExtended:  baseWorkflow,
		NodeStatusGetter:      nodeStatus,
	}
}
