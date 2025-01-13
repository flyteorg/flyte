package executors

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

type FailureNodeLookup struct {
	NodeLookup
	FailureNode       v1alpha1.ExecutableNode
	FailureNodeStatus v1alpha1.ExecutableNodeStatus
	OriginalError     *core.ExecutionError
}

func (f FailureNodeLookup) GetNode(nodeID v1alpha1.NodeID) (v1alpha1.ExecutableNode, bool) {
	if nodeID == v1alpha1.StartNodeID {
		return f.NodeLookup.GetNode(nodeID)
	}
	return f.FailureNode, true
}

func (f FailureNodeLookup) GetNodeExecutionStatus(ctx context.Context, id v1alpha1.NodeID) v1alpha1.ExecutableNodeStatus {
	if id == v1alpha1.StartNodeID {
		return f.NodeLookup.GetNodeExecutionStatus(ctx, id)
	}
	return f.FailureNodeStatus
}

func (f FailureNodeLookup) ToNode(id v1alpha1.NodeID) ([]v1alpha1.NodeID, error) {
	// The upstream node of the failure node is always the start node
	return []v1alpha1.NodeID{v1alpha1.StartNodeID}, nil
}

func (f FailureNodeLookup) FromNode(id v1alpha1.NodeID) ([]v1alpha1.NodeID, error) {
	return nil, nil
}

func (f FailureNodeLookup) GetOriginalError() *core.ExecutionError {
	return f.OriginalError
}

func NewFailureNodeLookup(nodeLookup NodeLookup, failureNode v1alpha1.ExecutableNode, failureNodeStatus v1alpha1.ExecutableNodeStatus, originalError *core.ExecutionError) NodeLookup {
	return FailureNodeLookup{
		NodeLookup:        nodeLookup,
		FailureNode:       failureNode,
		FailureNodeStatus: failureNodeStatus,
		OriginalError:     originalError,
	}
}
