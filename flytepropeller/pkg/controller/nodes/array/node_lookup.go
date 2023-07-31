package array

import (
	"context"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
)

type arrayNodeLookup struct {
	executors.NodeLookup
	subNodeID     v1alpha1.NodeID
	subNodeSpec   *v1alpha1.NodeSpec
	subNodeStatus *v1alpha1.NodeStatus
}

func (a *arrayNodeLookup) GetNode(nodeID v1alpha1.NodeID) (v1alpha1.ExecutableNode, bool) {
	if nodeID == a.subNodeID {
		return a.subNodeSpec, true
	}

	return a.NodeLookup.GetNode(nodeID)
}

func (a *arrayNodeLookup) GetNodeExecutionStatus(ctx context.Context, id v1alpha1.NodeID) v1alpha1.ExecutableNodeStatus {
	if id == a.subNodeID {
		return a.subNodeStatus
	}

	return a.NodeLookup.GetNodeExecutionStatus(ctx, id)
}

func newArrayNodeLookup(nodeLookup executors.NodeLookup, subNodeID v1alpha1.NodeID, subNodeSpec *v1alpha1.NodeSpec, subNodeStatus *v1alpha1.NodeStatus) arrayNodeLookup {
	return arrayNodeLookup{
		NodeLookup:    nodeLookup,
		subNodeID:     subNodeID,
		subNodeSpec:   subNodeSpec,
		subNodeStatus: subNodeStatus,
	}
}
