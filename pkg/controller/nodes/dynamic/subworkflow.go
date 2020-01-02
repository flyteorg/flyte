package dynamic

import (
	"context"

	"github.com/lyft/flytepropeller/pkg/controller/executors"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytestdlib/storage"
)

// Defines a sub-contextual workflow that is built in-memory to represent a dynamic job execution plan.
type contextualWorkflow struct {
	v1alpha1.ExecutableWorkflow

	extraTasks     map[v1alpha1.TaskID]*v1alpha1.TaskSpec
	extraWorkflows map[v1alpha1.WorkflowID]*v1alpha1.WorkflowSpec
	status         *ContextualWorkflowStatus
}

func newContextualWorkflow(baseWorkflow v1alpha1.ExecutableWorkflow,
	subwf v1alpha1.ExecutableSubWorkflow,
	status v1alpha1.ExecutableNodeStatus,
	tasks map[v1alpha1.TaskID]*v1alpha1.TaskSpec,
	workflows map[v1alpha1.WorkflowID]*v1alpha1.WorkflowSpec,
	refConstructor storage.ReferenceConstructor) v1alpha1.ExecutableWorkflow {

	return &contextualWorkflow{
		ExecutableWorkflow: executors.NewSubContextualWorkflow(baseWorkflow, subwf, status),
		extraTasks:         tasks,
		extraWorkflows:     workflows,
		status:             newContextualWorkflowStatus(baseWorkflow.GetExecutionStatus(), status, refConstructor),
	}
}

func (w contextualWorkflow) GetExecutionStatus() v1alpha1.ExecutableWorkflowStatus {
	return w.status
}

func (w contextualWorkflow) GetTask(id v1alpha1.TaskID) (v1alpha1.ExecutableTask, error) {
	if task, found := w.extraTasks[id]; found {
		return task, nil
	}

	return w.ExecutableWorkflow.GetTask(id)
}

func (w contextualWorkflow) FindSubWorkflow(id v1alpha1.WorkflowID) v1alpha1.ExecutableSubWorkflow {
	if wf, found := w.extraWorkflows[id]; found {
		return wf
	}

	return w.ExecutableWorkflow.FindSubWorkflow(id)
}

// A contextual workflow status to override some of the implementations.
type ContextualWorkflowStatus struct {
	v1alpha1.ExecutableWorkflowStatus
	baseStatus           v1alpha1.ExecutableNodeStatus
	referenceConstructor storage.ReferenceConstructor
}

func (w ContextualWorkflowStatus) GetDataDir() v1alpha1.DataReference {
	return w.baseStatus.GetDataDir()
}

// Overrides default node data dir to work around the contractual assumption between Propeller and Futures to write all
// sub-node inputs into current node data directory.
// E.g.
//   if current node data dir is /wf_exec/node-1/data/
//   and the task ran and yielded 2 nodes, the structure will look like this:
//   /wf_exec/node-1/data/
//                   |_ inputs.pb
//                   |_ futures.pb
//                   |_ sub-node1/inputs.pb
//                   |_ sub-node2/inputs.pb
// TODO: This is just a stop-gap until we transition the DynamicJobSpec to be a full-fledged workflow spec.
// TODO: this will allow us to have proper data bindings between nodes then we can stop making assumptions about data refs.
func (w ContextualWorkflowStatus) ConstructNodeDataDir(ctx context.Context, name v1alpha1.NodeID) (storage.DataReference, error) {
	return w.referenceConstructor.ConstructReference(ctx, w.GetDataDir(), name)
}

func newContextualWorkflowStatus(baseWfStatus v1alpha1.ExecutableWorkflowStatus,
	baseStatus v1alpha1.ExecutableNodeStatus, constructor storage.ReferenceConstructor) *ContextualWorkflowStatus {

	return &ContextualWorkflowStatus{
		ExecutableWorkflowStatus: baseWfStatus,
		baseStatus:               baseStatus,
		referenceConstructor:     constructor,
	}
}
