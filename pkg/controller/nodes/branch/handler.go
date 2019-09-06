package branch

import (
	"context"

	"github.com/lyft/flyteidl/clients/go/events"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/executors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"
)

type branchHandler struct {
	nodeExecutor executors.Node
	recorder     events.NodeEventRecorder
}

func (b *branchHandler) recurseDownstream(ctx context.Context, w v1alpha1.ExecutableWorkflow, nodeStatus v1alpha1.ExecutableNodeStatus, branchTakenNode v1alpha1.ExecutableNode) (handler.Status, error) {
	downstreamStatus, err := b.nodeExecutor.RecursiveNodeHandler(ctx, w, branchTakenNode)
	if err != nil {
		return handler.StatusUndefined, err
	}

	if downstreamStatus.IsComplete() {
		// For branch node we set the output node to be the same as the child nodes output
		childNodeStatus := w.GetNodeExecutionStatus(branchTakenNode.GetID())
		nodeStatus.SetDataDir(childNodeStatus.GetDataDir())
		return handler.StatusSuccess, nil
	}

	if downstreamStatus.HasFailed() {
		return handler.StatusFailed(downstreamStatus.Err), nil
	}

	return handler.StatusRunning, nil
}

func (b *branchHandler) StartNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, nodeInputs *handler.Data) (handler.Status, error) {
	logger.Debugf(ctx, "Starting Branch Node")
	branch := node.GetBranchNode()
	if branch == nil {
		return handler.StatusFailed(errors.Errorf(errors.IllegalStateError, w.GetID(), node.GetID(), "Invoked branch handler, for a non branch node.")), nil
	}
	nodeStatus := w.GetNodeExecutionStatus(node.GetID())
	branchStatus := nodeStatus.GetOrCreateBranchStatus()
	finalNode, err := DecideBranch(ctx, w, node.GetID(), branch, nodeInputs)
	if err != nil {
		branchStatus.SetBranchNodeError()
		logger.Debugf(ctx, "Branch evaluation failed. Error [%s]", err)
		return handler.StatusFailed(err), nil
	}
	branchStatus.SetBranchNodeSuccess(*finalNode)
	var ok bool
	childNode, ok := w.GetNode(*finalNode)
	if !ok {
		logger.Debugf(ctx, "Branch downstream finalized node not found. FinalizedNode [%s]", *finalNode)
		return handler.StatusFailed(errors.Errorf(errors.DownstreamNodeNotFoundError, w.GetID(), node.GetID(), "Downstream node [%v] not found", *finalNode)), nil
	}
	i := node.GetID()
	childNodeStatus := w.GetNodeExecutionStatus(childNode.GetID())
	childNodeStatus.SetParentNodeID(&i)

	logger.Debugf(ctx, "Recursing down branch node")
	return b.recurseDownstream(ctx, w, nodeStatus, childNode)
}

func (b *branchHandler) CheckNodeStatus(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, nodeStatus v1alpha1.ExecutableNodeStatus) (handler.Status, error) {
	branch := node.GetBranchNode()
	if branch == nil {
		return handler.StatusFailed(errors.Errorf(errors.IllegalStateError, w.GetID(), node.GetID(), "Invoked branch handler, for a non branch node.")), nil
	}
	// If the branch was already evaluated i.e, Node is in Running status
	branchStatus := nodeStatus.GetOrCreateBranchStatus()
	userError := branch.GetElseFail()
	finalNodeID := branchStatus.GetFinalizedNode()
	if finalNodeID == nil {
		if userError != nil {
			// We should never reach here, but for safety and completeness
			return handler.StatusFailed(errors.Errorf(errors.UserProvidedError, w.GetID(), node.GetID(), userError.Message)), nil
		}
		return handler.StatusRunning, errors.Errorf(errors.IllegalStateError, w.GetID(), node.GetID(), "No node finalized through previous branch evaluation.")
	}
	var ok bool
	branchTakenNode, ok := w.GetNode(*finalNodeID)
	if !ok {
		return handler.StatusFailed(errors.Errorf(errors.DownstreamNodeNotFoundError, w.GetID(), node.GetID(), "Downstream node [%v] not found", *finalNodeID)), nil
	}
	// Recurse downstream
	return b.recurseDownstream(ctx, w, nodeStatus, branchTakenNode)
}

func (b *branchHandler) HandleFailingNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) (handler.Status, error) {
	return handler.StatusFailed(errors.Errorf(errors.IllegalStateError, w.GetID(), node.GetID(), "A Branch node cannot enter a failing state")), nil
}

func (b *branchHandler) Initialize(ctx context.Context) error {
	return nil
}

func (b *branchHandler) AbortNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) error {
	branch := node.GetBranchNode()
	if branch == nil {
		return errors.Errorf(errors.IllegalStateError, w.GetID(), node.GetID(), "Invoked branch handler, for a non branch node.")
	}
	// If the branch was already evaluated i.e, Node is in Running status
	nodeStatus := w.GetNodeExecutionStatus(node.GetID())
	branchStatus := nodeStatus.GetOrCreateBranchStatus()
	userError := branch.GetElseFail()
	finalNodeID := branchStatus.GetFinalizedNode()
	if finalNodeID == nil {
		if userError != nil {
			// We should never reach here, but for safety and completeness
			return errors.Errorf(errors.UserProvidedError, w.GetID(), node.GetID(), userError.Message)
		}
		return errors.Errorf(errors.IllegalStateError, w.GetID(), node.GetID(), "No node finalized through previous branch evaluation.")
	}
	var ok bool
	branchTakenNode, ok := w.GetNode(*finalNodeID)

	if !ok {
		return errors.Errorf(errors.DownstreamNodeNotFoundError, w.GetID(), node.GetID(), "Downstream node [%v] not found", *finalNodeID)
	}
	// Recurse downstream
	return b.nodeExecutor.AbortHandler(ctx, w, branchTakenNode)
}

func New(executor executors.Node, eventSink events.EventSink, scope promutils.Scope) handler.IFace {
	branchScope := scope.NewSubScope("branch")
	return &branchHandler{
		nodeExecutor: executor,
		recorder:     events.NewNodeEventRecorder(eventSink, branchScope),
	}
}
