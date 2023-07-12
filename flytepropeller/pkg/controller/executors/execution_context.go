package executors

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
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
	GetExecutionConfig() v1alpha1.ExecutionConfig
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
}

type NodeExecutor interface {
	AddNodeFuture(func(chan<- NodeExecutionResult))
	Wait() (NodeStatus, error)
}

type ExecutionContext interface {
	ImmutableExecutionContext
	TaskDetailsGetter
	SubWorkflowGetter
	ParentInfoGetter
	ControlFlow
	NodeExecutor
}

type execContext struct {
	ControlFlow
	ImmutableExecutionContext
	TaskDetailsGetter
	SubWorkflowGetter
	parentInfo ImmutableParentInfo
	NodeExecutor
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
	v uint32
}

func (c *controlFlow) CurrentParallelism() uint32 {
	return c.v
}

func (c *controlFlow) IncrementParallelism() uint32 {
	c.v = c.v + 1
	return c.v
}

type NodeExecutionResult struct {
	Err        error
	NodeStatus NodeStatus
}

type nodeExecutor struct {
	nodeFutures []<-chan NodeExecutionResult
}

func (n *nodeExecutor) AddNodeFuture(f func(chan<- NodeExecutionResult)) {
	nodeFuture := make(chan NodeExecutionResult, 1)
	go f(nodeFuture)
	n.nodeFutures = append(n.nodeFutures, nodeFuture)
}

func (n *nodeExecutor) Wait() (NodeStatus, error) {
	if len(n.nodeFutures) == 0 {
		return NodeStatusComplete, nil
	}

	// If any downstream node is failed, fail, all
	// Else if all are success then success
	// Else if any one is running then Downstream is still running
	allCompleted := true
	partialNodeCompletion := false
	onFailurePolicy := v1alpha1.WorkflowOnFailurePolicy(core.WorkflowMetadata_FAIL_IMMEDIATELY)
	//onFailurePolicy := execContext.GetOnFailurePolicy() // TODO @hamersaw - need access to this
	stateOnComplete := NodeStatusComplete
	for _, nodeFuture := range n.nodeFutures {
		nodeExecutionResult := <-nodeFuture
		state := nodeExecutionResult.NodeStatus
		err := nodeExecutionResult.Err
		if err != nil { // TODO @hamersaw - do we want to fail right away? or wait until all nodes are done?
			return NodeStatusUndefined, err
		}

		if state.HasFailed() || state.HasTimedOut() {
			// TODO @hamersaw - Debug?
			//logger.Debugf(ctx, "Some downstream node has failed. Failed: [%v]. TimedOut: [%v]. Error: [%s]", state.HasFailed(), state.HasTimedOut(), state.Err)
			if onFailurePolicy == v1alpha1.WorkflowOnFailurePolicy(core.WorkflowMetadata_FAIL_AFTER_EXECUTABLE_NODES_COMPLETE) {
				// If the failure policy allows other nodes to continue running, do not exit the loop,
				// Keep track of the last failed state in the loop since it'll be the one to return.
				// TODO: If multiple nodes fail (which this mode allows), consolidate/summarize failure states in one.
				stateOnComplete = state
			} else {
				return state, nil
			}
		} else if !state.IsComplete() {
			// A Failed/Timedout node is implicitly considered "complete" this means none of the downstream nodes from
			// that node will ever be allowed to run.
			// This else block, therefore, deals with all other states. IsComplete will return true if and only if this
			// node as well as all of its downstream nodes have finished executing with success statuses. Otherwise we
			// mark this node's state as not completed to ensure we will visit it again later.
			allCompleted = false
		}

		if state.PartiallyComplete() {
			// This implies that one of the downstream nodes has just succeeded and workflow is ready for propagation
			// We do not propagate in current cycle to make it possible to store the state between transitions
			partialNodeCompletion = true
		}
	}

	if allCompleted {
		// TODO @hamersaw - Debug?
		//logger.Debugf(ctx, "All downstream nodes completed")
		return stateOnComplete, nil
	}

	if partialNodeCompletion {
		return NodeStatusSuccess, nil
	}

	return NodeStatusPending, nil
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
		NodeExecutor:              &nodeExecutor{},
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
		v: 0,
	}
}
