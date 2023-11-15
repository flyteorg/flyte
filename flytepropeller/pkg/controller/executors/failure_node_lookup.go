package executors

import (
	"context"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

type FailureNodeLookup struct {
	NodeSpec        *v1alpha1.NodeSpec
	NodeStatus      v1alpha1.ExecutableNodeStatus
	StartNode       v1alpha1.ExecutableNode
	StartNodeStatus v1alpha1.ExecutableNodeStatus
}

func (f FailureNodeLookup) GetNode(nodeID v1alpha1.NodeID) (v1alpha1.ExecutableNode, bool) {
	if nodeID == v1alpha1.StartNodeID {
		return f.StartNode, true
	}
	return f.NodeSpec, true
}

func (f FailureNodeLookup) GetNodeExecutionStatus(ctx context.Context, id v1alpha1.NodeID) v1alpha1.ExecutableNodeStatus {
	if id == v1alpha1.StartNodeID {
		return f.StartNodeStatus
	}
	return f.NodeStatus
}

func (f FailureNodeLookup) ToNode(id v1alpha1.NodeID) ([]v1alpha1.NodeID, error) {
	return []v1alpha1.NodeID{v1alpha1.StartNodeID}, nil
}

func (f FailureNodeLookup) FromNode(id v1alpha1.NodeID) ([]v1alpha1.NodeID, error) {
	return nil, nil
}

func NewFailureNodeLookup(nodeSpec *v1alpha1.NodeSpec, startNode v1alpha1.ExecutableNode, nodeStatusGetter v1alpha1.NodeStatusGetter) NodeLookup {
	startNodeStatus := nodeStatusGetter.GetNodeExecutionStatus(context.TODO(), v1alpha1.StartNodeID)
	errNodeStatus := nodeStatusGetter.GetNodeExecutionStatus(context.TODO(), nodeSpec.GetID())
	return FailureNodeLookup{
		NodeSpec:        nodeSpec,
		NodeStatus:      errNodeStatus,
		StartNode:       startNode,
		StartNodeStatus: startNodeStatus,
	}
}
