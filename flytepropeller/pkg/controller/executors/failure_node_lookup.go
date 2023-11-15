package executors

import (
	"context"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

type FailureNodeLookup struct {
	NodeSpec   *v1alpha1.NodeSpec
	NodeStatus v1alpha1.ExecutableNodeStatus
}

func (f FailureNodeLookup) GetNode(nodeID v1alpha1.NodeID) (v1alpha1.ExecutableNode, bool) {
	return f.NodeSpec, true
}

func (f FailureNodeLookup) GetNodeExecutionStatus(ctx context.Context, id v1alpha1.NodeID) v1alpha1.ExecutableNodeStatus {
	return f.NodeStatus
}

func (f FailureNodeLookup) ToNode(id v1alpha1.NodeID) ([]v1alpha1.NodeID, error) {
	return []v1alpha1.NodeID{v1alpha1.StartNodeID}, nil
}

func (f FailureNodeLookup) FromNode(id v1alpha1.NodeID) ([]v1alpha1.NodeID, error) {
	return nil, nil
}

func NewFailureNodeLookup(nodeSpec *v1alpha1.NodeSpec, nodeStatus v1alpha1.ExecutableNodeStatus) NodeLookup {
	return FailureNodeLookup{
		NodeSpec:   nodeSpec,
		NodeStatus: nodeStatus,
	}
}
