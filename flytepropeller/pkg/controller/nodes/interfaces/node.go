package interfaces

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
)

//go:generate mockery -all -case=underscore

// p of the node
type NodePhase int

const (
	// Indicates that the node is not yet ready to be executed and is pending any previous nodes completion
	NodePhasePending NodePhase = iota
	// Indicates that the node was queued and will start running soon
	NodePhaseQueued
	// Indicates that the payload associated with this node is being executed and is not yet done
	NodePhaseRunning
	// Indicates that the nodes payload has been successfully completed, but any downstream nodes from this node may not yet have completed
	// We could make Success = running, but this enables more granular control
	NodePhaseSuccess
	// Complete indicates successful completion of a node. For singular nodes (nodes that have only one execution) success = complete, but, the executor
	// will always signal completion
	NodePhaseComplete
	// Node failed in execution, either this node or anything in the downstream chain
	NodePhaseFailed
	// Internal error observed. This state should always be accompanied with an `error`. if not the behavior is undefined
	NodePhaseUndefined
	// Finalize node failing due to timeout
	NodePhaseTimingOut
	// Node failed because execution timed out
	NodePhaseTimedOut
	// Node recovered from a prior execution.
	NodePhaseRecovered
)

func (p NodePhase) String() string {
	switch p {
	case NodePhaseRunning:
		return "Running"
	case NodePhaseQueued:
		return "Queued"
	case NodePhasePending:
		return "Pending"
	case NodePhaseFailed:
		return "Failed"
	case NodePhaseSuccess:
		return "Success"
	case NodePhaseComplete:
		return "Complete"
	case NodePhaseUndefined:
		return "Undefined"
	case NodePhaseTimedOut:
		return "NodePhaseTimedOut"
	case NodePhaseRecovered:
		return "NodePhaseRecovered"
	}
	return fmt.Sprintf("Unknown - %d", p)
}

// Core Node Executor that is used to execute a node. This is a recursive node executor and understands node dependencies
type Node interface {
	// This method is used specifically to set inputs for start node. This is because start node does not retrieve inputs
	// from predecessors, but the inputs are inputs to the workflow or inputs to the parent container (workflow) node.
	SetInputsForStartNode(ctx context.Context, execContext executors.ExecutionContext, dag executors.DAGStructureWithStartNode,
		nl executors.NodeLookup, inputs *core.LiteralMap) (NodeStatus, error)

	// This is the main entrypoint to execute a node. It recursively depth-first goes through all ready nodes and starts their execution
	// This returns either
	// - 1. It finds a blocking node (not ready, or running)
	// - 2. A node fails and hence the workflow will fail
	// - 3. The final/end node has completed and the workflow should be stopped
	RecursiveNodeHandler(ctx context.Context, execContext executors.ExecutionContext, dag executors.DAGStructure,
		nl executors.NodeLookup, currentNode v1alpha1.ExecutableNode) (NodeStatus, error)

	// This aborts the given node. If the given node is complete then it recursively finds the running nodes and aborts them
	AbortHandler(ctx context.Context, execContext executors.ExecutionContext, dag executors.DAGStructure,
		nl executors.NodeLookup, currentNode v1alpha1.ExecutableNode, reason string) error

	FinalizeHandler(ctx context.Context, execContext executors.ExecutionContext, dag executors.DAGStructure,
		nl executors.NodeLookup, currentNode v1alpha1.ExecutableNode) error

	// This method should be used to initialize Node executor
	Initialize(ctx context.Context) error

	// GetNodeExecutionContextBuilder returns the current NodeExecutionContextBuilder
	GetNodeExecutionContextBuilder() NodeExecutionContextBuilder

	// WithNodeExecutionContextBuilder returns a new Node with the given NodeExecutionContextBuilder
	WithNodeExecutionContextBuilder(NodeExecutionContextBuilder) Node
}

// NodeExecutionContextBuilder defines how a NodeExecutionContext is built
type NodeExecutionContextBuilder interface {
	BuildNodeExecutionContext(ctx context.Context, executionContext executors.ExecutionContext,
		nl executors.NodeLookup, currentNodeID v1alpha1.NodeID) (NodeExecutionContext, error)
}

// Helper struct to allow passing of status between functions
type NodeStatus struct {
	NodePhase NodePhase
	Err       *core.ExecutionError
}

func (n *NodeStatus) IsComplete() bool {
	return n.NodePhase == NodePhaseComplete
}

func (n *NodeStatus) HasFailed() bool {
	return n.NodePhase == NodePhaseFailed
}

func (n *NodeStatus) HasTimedOut() bool {
	return n.NodePhase == NodePhaseTimedOut
}

func (n *NodeStatus) PartiallyComplete() bool {
	return n.NodePhase == NodePhaseSuccess
}

var NodeStatusPending = NodeStatus{NodePhase: NodePhasePending}
var NodeStatusQueued = NodeStatus{NodePhase: NodePhaseQueued}
var NodeStatusRunning = NodeStatus{NodePhase: NodePhaseRunning}
var NodeStatusSuccess = NodeStatus{NodePhase: NodePhaseSuccess}
var NodeStatusComplete = NodeStatus{NodePhase: NodePhaseComplete}
var NodeStatusUndefined = NodeStatus{NodePhase: NodePhaseUndefined}
var NodeStatusTimedOut = NodeStatus{NodePhase: NodePhaseTimedOut}
var NodeStatusRecovered = NodeStatus{NodePhase: NodePhaseRecovered}

func NodeStatusFailed(err *core.ExecutionError) NodeStatus {
	return NodeStatus{NodePhase: NodePhaseFailed, Err: err}
}
