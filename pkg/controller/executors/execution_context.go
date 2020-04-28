package executors

import (
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

type TaskDetailsGetter interface {
	GetTask(id v1alpha1.TaskID) (v1alpha1.ExecutableTask, error)
}

type SubWorkflowGetter interface {
	FindSubWorkflow(subID v1alpha1.WorkflowID) v1alpha1.ExecutableSubWorkflow
}

type ImmutableExecutionContext interface {
	v1alpha1.Meta
	GetID() v1alpha1.WorkflowID
}

type ExecutionContext interface {
	ImmutableExecutionContext
	TaskDetailsGetter
	SubWorkflowGetter
}

type execContext struct {
	ImmutableExecutionContext
	TaskDetailsGetter
	SubWorkflowGetter
}

func NewExecutionContextWithTasksGetter(prevExecContext ExecutionContext, taskGetter TaskDetailsGetter) ExecutionContext {
	return NewExecutionContext(prevExecContext, taskGetter, prevExecContext)
}

func NewExecutionContextWithWorkflowGetter(prevExecContext ExecutionContext, getter SubWorkflowGetter) ExecutionContext {
	return NewExecutionContext(prevExecContext, prevExecContext, getter)
}

func NewExecutionContext(immExecContext ImmutableExecutionContext, tasksGetter TaskDetailsGetter, workflowGetter SubWorkflowGetter) ExecutionContext {
	return execContext{
		ImmutableExecutionContext: immExecContext,
		TaskDetailsGetter:         tasksGetter,
		SubWorkflowGetter:         workflowGetter,
	}
}
