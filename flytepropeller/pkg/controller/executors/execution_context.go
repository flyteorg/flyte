package executors

import (
	"fmt"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

type TaskDetailsGetter interface {
	GetTask(id v1alpha1.TaskID) (v1alpha1.ExecutableTask, error)
}

// Retrieves the Task details from a static inmemory HashMap
type staticTaskDetailsGetter struct {
	// As this is an additional taskmap created, we can use a name to identify the taskmap. Every error message will have this name
	name  string
	tasks map[v1alpha1.TaskID]v1alpha1.ExecutableTask
}

func (t *staticTaskDetailsGetter) GetTask(id v1alpha1.TaskID) (v1alpha1.ExecutableTask, error) {
	if task, ok := t.tasks[id]; ok {
		return task, nil
	}
	return nil, fmt.Errorf("unable to find task with id [%s] in task set [%s]", id, t.name)
}

// As this is an additional taskmap created, we can use a name to identify the taskmap. Every error message will have this name
func NewStaticTaskDetailsGetter(name string, tasks map[v1alpha1.TaskID]v1alpha1.ExecutableTask) TaskDetailsGetter {
	return &staticTaskDetailsGetter{name: name, tasks: tasks}
}

type SubWorkflowGetter interface {
	FindSubWorkflow(subID v1alpha1.WorkflowID) v1alpha1.ExecutableSubWorkflow
}

// Retrieves the Task details from a static inmemory HashMap
type staticSubWorkflowGetter struct {
	// As this is an additional subWorkflow Map created, we can use a name to identify the SubWorkflow Set. Every error message will have this name
	name         string
	subWorkflows map[v1alpha1.WorkflowID]v1alpha1.ExecutableSubWorkflow
}

func (t *staticSubWorkflowGetter) FindSubWorkflow(subID v1alpha1.WorkflowID) v1alpha1.ExecutableSubWorkflow {
	if swf, ok := t.subWorkflows[subID]; ok {
		return swf
	}
	return nil
}

// As this is an additional taskmap created, we can use a name to identify the taskmap. Every error message will have this name
func NewStaticSubWorkflowsGetter(name string, subworkflows map[v1alpha1.WorkflowID]v1alpha1.ExecutableSubWorkflow) SubWorkflowGetter {
	return &staticSubWorkflowGetter{name: name, subWorkflows: subworkflows}
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
