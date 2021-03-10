package executors

import (
	"fmt"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

// An interface that captures the Directed Acyclic Graph structure in which the nodes are connected.
// If NodeLookup and DAGStructure are used together a traversal can be implemented.
type DAGStructure interface {
	// Lookup for upstream edges, find all node ids from which this node can be reached.
	ToNode(id v1alpha1.NodeID) ([]v1alpha1.NodeID, error)
	// Lookup for downstream edges, find all node ids that can be reached from the given node id.
	FromNode(id v1alpha1.NodeID) ([]v1alpha1.NodeID, error)
}

type DAGStructureWithStartNode interface {
	DAGStructure
	// The Starting node for the DAG
	StartNode() v1alpha1.ExecutableNode
}

type leafNodeDAGStructure struct {
	parentNodes []v1alpha1.NodeID
	currentNode v1alpha1.NodeID
}

func (l leafNodeDAGStructure) ToNode(id v1alpha1.NodeID) ([]v1alpha1.NodeID, error) {
	if id == l.currentNode {
		return l.parentNodes, nil
	}
	return nil, fmt.Errorf("unknown Node ID [%s]", id)
}

func (l leafNodeDAGStructure) FromNode(id v1alpha1.NodeID) ([]v1alpha1.NodeID, error) {
	return []v1alpha1.NodeID{}, nil
}

// Returns a new DAGStructure for a leafNode. i.e., there are only incoming edges and no outgoing edges.
// Also there is no StartNode for this Structure
func NewLeafNodeDAGStructure(leafNode v1alpha1.NodeID, parentNodes ...v1alpha1.NodeID) DAGStructure {
	return leafNodeDAGStructure{currentNode: leafNode, parentNodes: parentNodes}
}
