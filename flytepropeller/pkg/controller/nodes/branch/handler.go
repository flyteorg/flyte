package branch

import (
	"context"
	"fmt"

	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/executors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
)

type metrics struct {
	scope promutils.Scope
}

type branchHandler struct {
	nodeExecutor executors.Node
	m            metrics
}

func (b *branchHandler) FinalizeRequired() bool {
	return false
}

func (b *branchHandler) Setup(ctx context.Context, setupContext handler.SetupContext) error {
	logger.Debugf(ctx, "BranchNode::Setup: nothing to do")
	return nil
}

func (b *branchHandler) Handle(ctx context.Context, nCtx handler.NodeExecutionContext) (handler.Transition, error) {
	logger.Debug(ctx, "Starting Branch Node")
	branchNode := nCtx.Node().GetBranchNode()
	if branchNode == nil {
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(string(errors.IllegalStateError), "Invoked branch handler, for a non branch node.", nil)), nil
	}

	if nCtx.NodeStateReader().GetBranchNode().FinalizedNodeID == nil {
		nodeInputs, err := nCtx.InputReader().Get(ctx)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to read input. Error [%s]", err)
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(string(errors.RuntimeExecutionError), errMsg, nil)), nil
		}
		w := nCtx.Workflow()
		finalNodeID, err := DecideBranch(ctx, w, nCtx.NodeID(), branchNode, nodeInputs)
		if err != nil {
			errMsg := fmt.Sprintf("Branch evaluation failed. Error [%s]", err)
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(string(errors.IllegalStateError), errMsg, nil)), nil
		}

		branchNodeState := handler.BranchNodeState{FinalizedNodeID: finalNodeID, Phase: v1alpha1.BranchNodeSuccess}
		err = nCtx.NodeStateWriter().PutBranchNode(branchNodeState)
		if err != nil {
			logger.Errorf(ctx, "Failed to store BranchNode state, err :%s", err.Error())
			return handler.UnknownTransition, err
		}

		var ok bool
		finalNode, ok := w.GetNode(*finalNodeID)
		if !ok {
			errMsg := fmt.Sprintf("Branch downstream finalized node not found. FinalizedNode [%s]", *finalNodeID)
			logger.Debugf(ctx, errMsg)
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(string(errors.DownstreamNodeNotFoundError), errMsg, nil)), nil
		}
		i := nCtx.NodeID()
		childNodeStatus := w.GetNodeExecutionStatus(finalNode.GetID())
		childNodeStatus.SetParentNodeID(&i)

		logger.Debugf(ctx, "Recursing down branchNodestatus node")
		nodeStatus := w.GetNodeExecutionStatus(nCtx.NodeID())
		return b.recurseDownstream(ctx, nCtx, nodeStatus, finalNode)
	}

	// If the branchNodestatus was already evaluated i.e, Node is in Running status
	branchStatus := nCtx.NodeStateReader().GetBranchNode()
	userError := branchNode.GetElseFail()
	finalNodeID := branchStatus.FinalizedNodeID
	if finalNodeID == nil {
		if userError != nil {
			// We should never reach here, but for safety and completeness
			errMsg := fmt.Sprintf("Branch node userError [%s]", userError)
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(string(errors.UserProvidedError), errMsg, nil)), nil
		}
		errMsg := "no node finalized through previous branchNodestatus evaluation"
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(string(errors.IllegalStateError), errMsg, nil)), nil
	}

	w := nCtx.Workflow()
	branchTakenNode, ok := w.GetNode(*finalNodeID)
	if !ok {
		errMsg := fmt.Sprintf("Downstream node [%v] not found", *finalNodeID)
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(string(errors.DownstreamNodeNotFoundError), errMsg, nil)), nil
	}

	// Recurse downstream
	nodeStatus := w.GetNodeExecutionStatus(nCtx.NodeID())
	return b.recurseDownstream(ctx, nCtx, nodeStatus, branchTakenNode)
}

func (b *branchHandler) recurseDownstream(ctx context.Context, nCtx handler.NodeExecutionContext, nodeStatus v1alpha1.ExecutableNodeStatus, branchTakenNode v1alpha1.ExecutableNode) (handler.Transition, error) {
	w := nCtx.Workflow()
	downstreamStatus, err := b.nodeExecutor.RecursiveNodeHandler(ctx, w, branchTakenNode)
	if err != nil {
		return handler.UnknownTransition, err
	}

	if downstreamStatus.IsComplete() {
		// For branch node we set the output node to be the same as the child nodes output
		childNodeStatus := w.GetNodeExecutionStatus(branchTakenNode.GetID())
		nodeStatus.SetDataDir(childNodeStatus.GetDataDir())
		phase := handler.PhaseInfoSuccess(&handler.ExecutionInfo{
			OutputInfo: &handler.OutputInfo{OutputURI: v1alpha1.GetOutputsFile(childNodeStatus.GetDataDir())},
		})
		return handler.DoTransition(handler.TransitionTypeEphemeral, phase), nil
	}

	if downstreamStatus.HasFailed() {
		errMsg := downstreamStatus.Err.Error()
		code, nodeError := errors.GetErrorCode(downstreamStatus.Err)
		if !nodeError {
			code = errors.UnknownError
		}
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(string(code), errMsg, nil)), nil
	}

	phase := handler.PhaseInfoRunning(nil)
	return handler.DoTransition(handler.TransitionTypeEphemeral, phase), nil
}

func (b *branchHandler) Abort(ctx context.Context, nCtx handler.NodeExecutionContext, reason string) error {

	branch := nCtx.Node().GetBranchNode()
	w := nCtx.Workflow()
	if branch == nil {
		return errors.Errorf(errors.IllegalStateError, w.GetID(), nCtx.NodeID(), "Invoked branch handler, for a non branch node.")
	}

	// If the branch was already evaluated i.e, Node is in Running status
	branchNodeState := nCtx.NodeStateReader().GetBranchNode()
	if branchNodeState.Phase == v1alpha1.BranchNodeNotYetEvaluated {
		logger.Errorf(ctx, "No node finalized through previous branch evaluation.")
		return nil
	} else if branchNodeState.Phase == v1alpha1.BranchNodeError {
		// We should never reach here, but for safety and completeness
		errMsg := "branch evaluation failed"
		if branch.GetElseFail() != nil {
			errMsg = branch.GetElseFail().Message
		}
		return errors.Errorf(errors.UserProvidedError, nCtx.NodeID(), errMsg)
	}

	finalNodeID := branchNodeState.FinalizedNodeID
	branchTakenNode, ok := w.GetNode(*finalNodeID)
	if !ok {
		logger.Errorf(ctx, "Downstream node [%v] not found", *finalNodeID)
		return nil
	}

	// Recurse downstream
	return b.nodeExecutor.AbortHandler(ctx, w, branchTakenNode, reason)
}

func (b *branchHandler) Finalize(ctx context.Context, executionContext handler.NodeExecutionContext) error {
	logger.Debugf(ctx, "BranchNode::Finalizer: nothing to do")
	return nil
}

func New(executor executors.Node, scope promutils.Scope) handler.Node {
	return &branchHandler{
		nodeExecutor: executor,
		m:            metrics{scope: scope},
	}
}
