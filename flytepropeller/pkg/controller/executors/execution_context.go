package executors

import (
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

// go:generate mockery -case=underscore

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
	GetExecutionConfig() v1alpha1.ExecutionConfig
	GetConsoleURL() string
}

type ParentInfoGetter interface {
	GetParentInfo() ImmutableParentInfo
}

type ImmutableParentInfo interface {
	GetUniqueID() v1alpha1.NodeID
	CurrentAttempt() uint32
}

type ControlFlow interface {
	CurrentParallelism() uint32
	IncrementParallelism() uint32
	CurrentNodeExecutionCount() uint32
	IncrementNodeExecutionCount() uint32
	CurrentTaskExecutionCount() uint32
	IncrementTaskExecutionCount() uint32
}

type ExecutionContext interface {
	ImmutableExecutionContext
	TaskDetailsGetter
	SubWorkflowGetter
	ParentInfoGetter
	ControlFlow
}

type execContext struct {
	ControlFlow
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

type controlFlow struct {
	// We could use atomic.Uint32, but this is not required for current Propeller. As every round is run in a single
	// thread and using atomic will introduce memory barriers
	parallelism        uint32
	nodeExecutionCount uint32
	taskExecutionCount uint32
}

func (c *controlFlow) CurrentParallelism() uint32 {
	return c.parallelism
}

func (c *controlFlow) IncrementParallelism() uint32 {
	c.parallelism = c.parallelism + 1
	return c.parallelism
}

func (c *controlFlow) CurrentNodeExecutionCount() uint32 {
	return c.nodeExecutionCount
}

func (c *controlFlow) IncrementNodeExecutionCount() uint32 {
	c.nodeExecutionCount++
	return c.nodeExecutionCount
}

func (c *controlFlow) CurrentTaskExecutionCount() uint32 {
	return c.taskExecutionCount
}

func (c *controlFlow) IncrementTaskExecutionCount() uint32 {
	c.taskExecutionCount++
	return c.taskExecutionCount
}

func NewExecutionContextWithTasksGetter(prevExecContext ExecutionContext, taskGetter TaskDetailsGetter) ExecutionContext {
	return NewExecutionContext(prevExecContext, taskGetter, prevExecContext, prevExecContext.GetParentInfo(), prevExecContext)
}

func NewExecutionContextWithWorkflowGetter(prevExecContext ExecutionContext, getter SubWorkflowGetter) ExecutionContext {
	return NewExecutionContext(prevExecContext, prevExecContext, getter, prevExecContext.GetParentInfo(), prevExecContext)
}

func NewExecutionContextWithParentInfo(prevExecContext ExecutionContext, parentInfo ImmutableParentInfo) ExecutionContext {
	return NewExecutionContext(prevExecContext, prevExecContext, prevExecContext, parentInfo, prevExecContext)
}

func NewExecutionContext(immExecContext ImmutableExecutionContext, tasksGetter TaskDetailsGetter, workflowGetter SubWorkflowGetter, parentInfo ImmutableParentInfo, flow ControlFlow) ExecutionContext {
	return execContext{
		ImmutableExecutionContext: immExecContext,
		TaskDetailsGetter:         tasksGetter,
		SubWorkflowGetter:         workflowGetter,
		parentInfo:                parentInfo,
		ControlFlow:               flow,
	}
}

func NewParentInfo(uniqueID string, currentAttempts uint32) ImmutableParentInfo {
	return &parentExecutionInfo{
		currentAttempts: currentAttempts,
		uniqueID:        uniqueID,
	}
}

func InitializeControlFlow() ControlFlow {
	return &controlFlow{
		parallelism:        0,
		nodeExecutionCount: 0,
		taskExecutionCount: 0,
	}
}
