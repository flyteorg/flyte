package executors

import (
	"context"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

// NodeLookup provides a structure that enables looking up all nodes within the current execution hierarchy/context.
// NOTE: execution hierarchy may change the nodes available, this is because when a SubWorkflow is being executed, only
// the nodes within the subworkflow are visible
type NodeLookup interface {
	GetNode(nodeID v1alpha1.NodeID) (v1alpha1.ExecutableNode, bool)
	GetNodeExecutionStatus(ctx context.Context, id v1alpha1.NodeID) v1alpha1.ExecutableNodeStatus
}

// Implements a contextual NodeLookup that can be composed of a disparate NodeGetter and a NodeStatusGetter
type contextualNodeLookup struct {
	v1alpha1.NodeGetter
	v1alpha1.NodeStatusGetter
}

// Returns a Contextual NodeLookup using the given NodeGetter and a separate NodeStatusGetter.
// Very useful in Subworkflows where the Subworkflow is the reservoir of the nodes, but the status for these nodes
// maybe stored int he Top-level workflow node itself.
func NewNodeLookup(n v1alpha1.NodeGetter, s v1alpha1.NodeStatusGetter) NodeLookup {
	return contextualNodeLookup{
		NodeGetter:       n,
		NodeStatusGetter: s,
	}
}

// Implements a nodeLookup using Maps, very useful in Testing
type staticNodeLookup struct {
	nodes  map[v1alpha1.NodeID]v1alpha1.ExecutableNode
	status map[v1alpha1.NodeID]v1alpha1.ExecutableNodeStatus
}

func (s staticNodeLookup) GetNode(nodeID v1alpha1.NodeID) (v1alpha1.ExecutableNode, bool) {
	n, ok := s.nodes[nodeID]
	return n, ok
}

func (s staticNodeLookup) GetNodeExecutionStatus(_ context.Context, id v1alpha1.NodeID) v1alpha1.ExecutableNodeStatus {
	return s.status[id]
}

// Returns a new NodeLookup useful in Testing. Not recommended to be used in production
func NewTestNodeLookup(nodes map[v1alpha1.NodeID]v1alpha1.ExecutableNode, status map[v1alpha1.NodeID]v1alpha1.ExecutableNodeStatus) NodeLookup {
	return staticNodeLookup{
		nodes:  nodes,
		status: status,
	}
}
