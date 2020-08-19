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
	GetOnFailurePolicy() v1alpha1.WorkflowOnFailurePolicy
}

type ParentInfoGetter interface {
	GetParentInfo() ImmutableParentInfo
}

type ImmutableParentInfo interface {
	GetUniqueID() v1alpha1.NodeID
	CurrentAttempt() uint32
}

type ExecutionContext interface {
	ImmutableExecutionContext
	TaskDetailsGetter
	SubWorkflowGetter
	ParentInfoGetter
}

type execContext struct {
	ImmutableExecutionContext
	TaskDetailsGetter
	SubWorkflowGetter
	parentInfo ImmutableParentInfo
}

func (e execContext) GetParentInfo() ImmutableParentInfo {
	return e.parentInfo
}

type parentExecutionInfo struct {
	uniqueID        v1alpha1.NodeID
	currentAttempts uint32
}

func (p *parentExecutionInfo) GetUniqueID() v1alpha1.NodeID {
	return p.uniqueID
}

func (p *parentExecutionInfo) CurrentAttempt() uint32 {
	return p.currentAttempts
}

func NewExecutionContextWithTasksGetter(prevExecContext ExecutionContext, taskGetter TaskDetailsGetter) ExecutionContext {
	return NewExecutionContext(prevExecContext, taskGetter, prevExecContext, prevExecContext.GetParentInfo())
}

func NewExecutionContextWithWorkflowGetter(prevExecContext ExecutionContext, getter SubWorkflowGetter) ExecutionContext {
	return NewExecutionContext(prevExecContext, prevExecContext, getter, prevExecContext.GetParentInfo())
}

func NewExecutionContextWithParentInfo(prevExecContext ExecutionContext, parentInfo ImmutableParentInfo) ExecutionContext {
	return NewExecutionContext(prevExecContext, prevExecContext, prevExecContext, parentInfo)
}

func NewExecutionContext(immExecContext ImmutableExecutionContext, tasksGetter TaskDetailsGetter, workflowGetter SubWorkflowGetter, parentInfo ImmutableParentInfo) ExecutionContext {
	return execContext{
		ImmutableExecutionContext: immExecContext,
		TaskDetailsGetter:         tasksGetter,
		SubWorkflowGetter:         workflowGetter,
		parentInfo:                parentInfo,
	}
}

func NewParentInfo(uniqueID string, currentAttempts uint32) ImmutableParentInfo {
	return &parentExecutionInfo{
		currentAttempts: currentAttempts,
		uniqueID:        uniqueID,
	}
}
