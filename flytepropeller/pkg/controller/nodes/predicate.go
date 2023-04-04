package nodes

import (
	"context"
	"time"

	"github.com/flyteorg/flytestdlib/logger"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/errors"
)

// Special enum to indicate if the node under consideration is ready to be executed or should be skipped
type PredicatePhase int

const (
	// Indicates node is not yet ready to be executed
	PredicatePhaseNotReady PredicatePhase = iota
	// Indicates node is ready to be executed - execution should proceed
	PredicatePhaseReady
	// Indicates that the node execution should be skipped as one of its parents was skipped or the branch was not taken
	PredicatePhaseSkip
	// Indicates failure during Predicate check
	PredicatePhaseUndefined
)

func (p PredicatePhase) String() string {
	switch p {
	case PredicatePhaseNotReady:
		return "NotReady"
	case PredicatePhaseReady:
		return "Ready"
	case PredicatePhaseSkip:
		return "Skip"
	}
	return "undefined"
}

// CanExecute method informs the callee if the given node can begin execution. This is dependent on
// primarily that all nodes upstream to the given node are successful and the results are available.
func CanExecute(ctx context.Context, dag executors.DAGStructure, nl executors.NodeLookup, node v1alpha1.BaseNode) (
	PredicatePhase, error) {

	nodeID := node.GetID()
	if nodeID == v1alpha1.StartNodeID {
		logger.Debugf(ctx, "Start Node id is assumed to be ready.")
		return PredicatePhaseReady, nil
	}

	upstreamNodes, err := dag.ToNode(nodeID)
	if err != nil {
		return PredicatePhaseUndefined, errors.Errorf(errors.BadSpecificationError, nodeID, "Unable to find upstream nodes for Node")
	}

	skipped := false
	for _, upstreamNodeID := range upstreamNodes {
		upstreamNode, ok := nl.GetNode(upstreamNodeID)
		if !ok {
			return PredicatePhaseUndefined, errors.Errorf(errors.BadSpecificationError, nodeID, "Upstream node [%v] of node [%v] not defined", upstreamNodeID, nodeID)
		}

		upstreamNodeStatus := nl.GetNodeExecutionStatus(ctx, upstreamNodeID)

		if upstreamNodeStatus.IsDirty() {
			return PredicatePhaseNotReady, nil
		}

		// BranchNodes are special. When the upstreamNode is a branch node it could represent one of two cases
		// Case 1: the upstreamNode is the parent branch node (the one that houses the condition) and the child node
		//         is one of the branches for the condition.
		//         i.e For an example Branch (if condition then take-branch1 else take-branch2)
		//         condition is the branch node, take-branch1/2 are the 2 child nodes and this condition can be triggered
		//         for either one of them
		// Case 2: the upstreamNode is some branch node and the current node is simply a node that consumes the output
		//         of the branch
		// Thus,
		//   In case1, we can proceed to execute one of the branches as soon as the branch evaluation completes, but
		//   branch node itself will not be successful (as it contains the branches)
		//   In case2, we will continue the evaluation of the CanExecute code block and can only proceed if the entire
		//   branch node is complete
		if upstreamNode.GetBranchNode() != nil && upstreamNodeStatus.GetBranchStatus() != nil {
			if upstreamNodeStatus.GetBranchStatus().GetPhase() != v1alpha1.BranchNodeSuccess {
				return PredicatePhaseNotReady, nil
			}
			// Branch node is success, so we are free to go ahead and execute the child nodes of the branch node, but
			// not any of the dependent nodes
			nodeStatus := nl.GetNodeExecutionStatus(ctx, nodeID)
			if nodeStatus.GetParentNodeID() != nil && *nodeStatus.GetParentNodeID() == upstreamNodeID {
				continue
			}
		}

		if upstreamNodeStatus.GetPhase() == v1alpha1.NodePhaseSkipped ||
			upstreamNodeStatus.GetPhase() == v1alpha1.NodePhaseFailed ||
			upstreamNodeStatus.GetPhase() == v1alpha1.NodePhaseTimedOut {
			skipped = true
		} else if !(upstreamNodeStatus.GetPhase() == v1alpha1.NodePhaseSucceeded ||
			upstreamNodeStatus.GetPhase() == v1alpha1.NodePhaseRecovered) {
			return PredicatePhaseNotReady, nil
		}
	}

	if skipped {
		return PredicatePhaseSkip, nil
	}

	return PredicatePhaseReady, nil
}

func GetParentNodeMaxEndTime(ctx context.Context, dag executors.DAGStructure, nl executors.NodeLookup, node v1alpha1.BaseNode) (t v1.Time, err error) {
	zeroTime := v1.NewTime(time.Time{})
	nodeID := node.GetID()
	if nodeID == v1alpha1.StartNodeID {
		logger.Debugf(ctx, "Start Node id is assumed to be ready.")
		return zeroTime, nil
	}

	upstreamNodes, err := dag.ToNode(nodeID)
	if err != nil {
		return zeroTime, errors.Errorf(errors.BadSpecificationError, nodeID, "Unable to find upstream nodes for Node")
	}

	var latest v1.Time
	for _, upstreamNodeID := range upstreamNodes {
		upstreamNodeStatus := nl.GetNodeExecutionStatus(ctx, upstreamNodeID)
		upstreamNode, ok := nl.GetNode(upstreamNodeID)
		if !ok {
			return zeroTime, errors.Errorf(errors.BadSpecificationError, nodeID, "Upstream node [%v] of node [%v] not defined", upstreamNodeID, nodeID)
		}

		// Special handling for branch node. The status for branch does not get updated to success, as it is the parent node
		if upstreamNode.GetBranchNode() != nil && upstreamNodeStatus.GetBranchStatus() != nil {
			if upstreamNodeStatus.GetBranchStatus().GetPhase() != v1alpha1.BranchNodeSuccess {
				return zeroTime, nil
			}
			branchUpdatedAt := upstreamNodeStatus.GetStartedAt()
			if upstreamNodeStatus.GetLastUpdatedAt() != nil {
				branchUpdatedAt = upstreamNodeStatus.GetLastUpdatedAt()
			}
			if branchUpdatedAt != nil && branchUpdatedAt.Unix() > latest.Unix() {
				latest = *branchUpdatedAt
			}
		} else if stoppedAt := upstreamNodeStatus.GetStoppedAt(); stoppedAt != nil && stoppedAt.Unix() > latest.Unix() {
			latest = *upstreamNodeStatus.GetStoppedAt()
		}
	}

	return latest, nil
}
