package nodes

import (
	"context"
	"time"

	"github.com/lyft/flytestdlib/logger"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/executors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
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

func CanExecute(ctx context.Context, dag executors.DAGStructure, nl executors.NodeLookup, node v1alpha1.BaseNode) (PredicatePhase, error) {
	nodeID := node.GetID()
	if nodeID == v1alpha1.StartNodeID {
		logger.Debugf(ctx, "Start Node id is assumed to be ready.")
		return PredicatePhaseReady, nil
	}

	nodeStatus := nl.GetNodeExecutionStatus(ctx, nodeID)
	parentNodeID := nodeStatus.GetParentNodeID()
	upstreamNodes, err := dag.ToNode(nodeID)
	if err != nil {
		return PredicatePhaseUndefined, errors.Errorf(errors.BadSpecificationError, nodeID, "Unable to find upstream nodes for Node")
	}

	skipped := false
	for _, upstreamNodeID := range upstreamNodes {
		upstreamNodeStatus := nl.GetNodeExecutionStatus(ctx, upstreamNodeID)

		if upstreamNodeStatus.IsDirty() {
			return PredicatePhaseNotReady, nil
		}

		if parentNodeID != nil && *parentNodeID == upstreamNodeID {
			upstreamNode, ok := nl.GetNode(upstreamNodeID)
			if !ok {
				return PredicatePhaseUndefined, errors.Errorf(errors.BadSpecificationError, nodeID, "Upstream node [%v] of node [%v] not defined", upstreamNodeID, nodeID)
			}

			// This only happens if current node is the child node of a branch node
			if upstreamNode.GetBranchNode() == nil || upstreamNodeStatus.GetBranchStatus().GetPhase() != v1alpha1.BranchNodeSuccess {
				logger.Debugf(ctx, "Branch sub node is expected to have parent branch node in succeeded state")
				return PredicatePhaseUndefined, errors.Errorf(errors.IllegalStateError, nodeID, "Upstream node [%v] is set as parent, but is not a branch node of [%v] or in illegal state.", upstreamNodeID, nodeID)
			}

			continue
		}

		if upstreamNodeStatus.GetPhase() == v1alpha1.NodePhaseSkipped ||
			upstreamNodeStatus.GetPhase() == v1alpha1.NodePhaseFailed ||
			upstreamNodeStatus.GetPhase() == v1alpha1.NodePhaseTimedOut {
			skipped = true
		} else if upstreamNodeStatus.GetPhase() != v1alpha1.NodePhaseSucceeded {
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

	nodeStatus := nl.GetNodeExecutionStatus(ctx, node.GetID())
	parentNodeID := nodeStatus.GetParentNodeID()
	upstreamNodes, err := dag.ToNode(nodeID)
	if err != nil {
		return zeroTime, errors.Errorf(errors.BadSpecificationError, nodeID, "Unable to find upstream nodes for Node")
	}

	var latest v1.Time
	for _, upstreamNodeID := range upstreamNodes {
		upstreamNodeStatus := nl.GetNodeExecutionStatus(ctx, upstreamNodeID)
		if parentNodeID != nil && *parentNodeID == upstreamNodeID {
			upstreamNode, ok := nl.GetNode(upstreamNodeID)
			if !ok {
				return zeroTime, errors.Errorf(errors.BadSpecificationError, nodeID, "Upstream node [%v] of node [%v] not defined", upstreamNodeID, nodeID)
			}

			// This only happens if current node is the child node of a branch node
			if upstreamNode.GetBranchNode() == nil || upstreamNodeStatus.GetBranchStatus().GetPhase() != v1alpha1.BranchNodeSuccess {
				logger.Debugf(ctx, "Branch sub node is expected to have parent branch node in succeeded state")
				return zeroTime, errors.Errorf(errors.IllegalStateError, nodeID, "Upstream node [%v] is set as parent, but is not a branch node of [%v] or in illegal state.", upstreamNodeID, nodeID)
			}

			continue
		}

		if stoppedAt := upstreamNodeStatus.GetStoppedAt(); stoppedAt != nil && stoppedAt.Unix() > latest.Unix() {
			latest = *upstreamNodeStatus.GetStoppedAt()
		}
	}

	return latest, nil
}
