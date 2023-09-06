package branch

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/common"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/interfaces"

	stdErrors "github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
)

type metrics struct {
	scope promutils.Scope
}

type branchHandler struct {
	nodeExecutor interfaces.Node
	m            metrics
	eventConfig  *config.EventConfig
}

func (b *branchHandler) FinalizeRequired() bool {
	return false
}

func (b *branchHandler) Setup(ctx context.Context, _ interfaces.SetupContext) error {
	logger.Debugf(ctx, "BranchNode::Setup: nothing to do")
	return nil
}

func (b *branchHandler) HandleBranchNode(ctx context.Context, branchNode v1alpha1.ExecutableBranchNode, nCtx interfaces.NodeExecutionContext, nl executors.NodeLookup) (handler.Transition, error) {
	if nCtx.NodeStateReader().GetBranchNodeState().FinalizedNodeID == nil {
		nodeInputs, err := nCtx.InputReader().Get(ctx)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to read input. Error [%s]", err)
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.RuntimeExecutionError, errMsg, nil)), nil
		}
		finalNodeID, err := DecideBranch(ctx, nl, nCtx.NodeID(), branchNode, nodeInputs)
		if err != nil {
			ec, ok := stdErrors.GetErrorCode(err)
			if ok {
				if ec == ErrorCodeMalformedBranch || ec == ErrorCodeUserProvidedError {
					return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_USER, ec, err.Error(), nil)), nil
				}
			}
			errMsg := fmt.Sprintf("Branch evaluation failed. Error [%s]", err)
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.IllegalStateError, errMsg, nil)), nil
		}

		branchNodeState := handler.BranchNodeState{FinalizedNodeID: finalNodeID, Phase: v1alpha1.BranchNodeSuccess}
		err = nCtx.NodeStateWriter().PutBranchNode(branchNodeState)
		if err != nil {
			logger.Errorf(ctx, "Failed to store BranchNode state, err :%s", err.Error())
			return handler.UnknownTransition, err
		}

		var ok bool
		finalNode, ok := nl.GetNode(*finalNodeID)
		if !ok {
			errMsg := fmt.Sprintf("Branch downstream finalized node not found. FinalizedNode [%s]", *finalNodeID)
			logger.Debugf(ctx, errMsg)
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.DownstreamNodeNotFoundError, errMsg, nil)), nil
		}
		i := nCtx.NodeID()
		childNodeStatus := nl.GetNodeExecutionStatus(ctx, finalNode.GetID())
		childNodeStatus.SetParentNodeID(&i)

		logger.Debugf(ctx, "Recursively executing branchNode's chosen path")
		nodeStatus := nl.GetNodeExecutionStatus(ctx, nCtx.NodeID())
		return b.recurseDownstream(ctx, nCtx, nodeStatus, finalNode)
	}

	// If the branchNodestatus was already evaluated i.e, Node is in Running status
	branchStatus := nCtx.NodeStateReader().GetBranchNodeState()
	userError := branchNode.GetElseFail()
	finalNodeID := branchStatus.FinalizedNodeID
	if finalNodeID == nil {
		if userError != nil {
			// We should never reach here, but for safety and completeness
			errMsg := fmt.Sprintf("Branch node userError [%s]", userError)
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_USER, errors.UserProvidedError, errMsg, nil)), nil
		}
		errMsg := "no node finalized through previous branchNodestatus evaluation"
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.IllegalStateError, errMsg, nil)), nil
	}

	branchTakenNode, ok := nl.GetNode(*finalNodeID)
	if !ok {
		errMsg := fmt.Sprintf("Downstream node [%v] not found", *finalNodeID)
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.DownstreamNodeNotFoundError, errMsg, nil)), nil
	}

	// Recurse downstream
	nodeStatus := nl.GetNodeExecutionStatus(ctx, nCtx.NodeID())
	return b.recurseDownstream(ctx, nCtx, nodeStatus, branchTakenNode)
}

func (b *branchHandler) Handle(ctx context.Context, nCtx interfaces.NodeExecutionContext) (handler.Transition, error) {
	logger.Debug(ctx, "Starting Branch Node")
	branchNode := nCtx.Node().GetBranchNode()
	if branchNode == nil {
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.IllegalStateError, "Invoked branch handler, for a non branch node.", nil)), nil
	}

	nl := nCtx.ContextualNodeLookup()

	return b.HandleBranchNode(ctx, branchNode, nCtx, nl)
}

func (b *branchHandler) getExecutionContextForDownstream(nCtx interfaces.NodeExecutionContext) (executors.ExecutionContext, error) {
	newParentInfo, err := common.CreateParentInfo(nCtx.ExecutionContext().GetParentInfo(), nCtx.NodeID(), nCtx.CurrentAttempt())
	if err != nil {
		return nil, err
	}
	return executors.NewExecutionContextWithParentInfo(nCtx.ExecutionContext(), newParentInfo), nil
}

func (b *branchHandler) recurseDownstream(ctx context.Context, nCtx interfaces.NodeExecutionContext, nodeStatus v1alpha1.ExecutableNodeStatus, branchTakenNode v1alpha1.ExecutableNode) (handler.Transition, error) {
	// TODO we should replace the call to RecursiveNodeHandler with a call to SingleNode Handler. The inputs are also already known ahead of time
	// There is no DAGStructure for the branch nodes, the branch taken node is the leaf node. The node itself may be arbitrarily complex, but in that case the node should reference a subworkflow etc
	// The parent of the BranchTaken Node is the actual Branch Node and all the data is just forwarded from the Branch to the executed node.
	nl := nCtx.ContextualNodeLookup()
	if nl == nil {
		return handler.DoTransition(handler.TransitionTypeBarrier, handler.PhaseInfo{}),
			errors.Errorf(errors.IllegalStateError, nCtx.NodeID(), "nodeLookup must be supplied.")
	}

	childNodeStatus := nl.GetNodeExecutionStatus(ctx, branchTakenNode.GetID())
	childNodeStatus.SetDataDir(nodeStatus.GetDataDir())
	childNodeStatus.SetOutputDir(nodeStatus.GetOutputDir())
	upstreamNodeIds, err := nCtx.ContextualNodeLookup().ToNode(branchTakenNode.GetID())
	if err != nil {
		return handler.UnknownTransition, err
	}
	dag := executors.NewLeafNodeDAGStructure(branchTakenNode.GetID(), append(upstreamNodeIds, nCtx.NodeID())...)
	execContext, err := b.getExecutionContextForDownstream(nCtx)
	if err != nil {
		return handler.UnknownTransition, err
	}
	downstreamStatus, err := b.nodeExecutor.RecursiveNodeHandler(ctx, execContext, dag, nCtx.ContextualNodeLookup(), branchTakenNode)
	if err != nil {
		return handler.UnknownTransition, err
	}

	if downstreamStatus.IsComplete() {
		// For branch node we set the output node to be the same as the child nodes output
		phase := handler.PhaseInfoSuccess(&handler.ExecutionInfo{
			OutputInfo: &handler.OutputInfo{OutputURI: v1alpha1.GetOutputsFile(childNodeStatus.GetOutputDir())},
		})

		return handler.DoTransition(handler.TransitionTypeEphemeral, phase), nil
	}

	if downstreamStatus.HasFailed() {
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailureErr(downstreamStatus.Err, nil)), nil
	}

	phase := handler.PhaseInfoRunning(nil)
	return handler.DoTransition(handler.TransitionTypeEphemeral, phase), nil
}

func (b *branchHandler) Abort(ctx context.Context, nCtx interfaces.NodeExecutionContext, reason string) error {

	branch := nCtx.Node().GetBranchNode()
	if branch == nil {
		return errors.Errorf(errors.IllegalStateError, nCtx.NodeID(), "Invoked branch handler, for a non branch node.")
	}

	// If the branch was already evaluated i.e, Node is in Running status
	branchNodeState := nCtx.NodeStateReader().GetBranchNodeState()
	if branchNodeState.Phase == v1alpha1.BranchNodeNotYetEvaluated {
		logger.Errorf(ctx, "No node finalized through previous branch evaluation.")
		return nil
	} else if branchNodeState.Phase == v1alpha1.BranchNodeError {
		// We should never reach here, but for safety and completeness
		errMsg := "branch evaluation failed"
		if branch.GetElseFail() != nil {
			errMsg = branch.GetElseFail().Message
		}
		logger.Errorf(ctx, errMsg)
		return nil
	}

	finalNodeID := branchNodeState.FinalizedNodeID
	branchTakenNode, ok := nCtx.ContextualNodeLookup().GetNode(*finalNodeID)
	if !ok {
		logger.Errorf(ctx, "Downstream node [%v] not found", *finalNodeID)
		return nil
	}

	// Recurse downstream
	// TODO we should replace the call to RecursiveNodeHandler with a call to SingleNode Handler. The inputs are also already known ahead of time
	// There is no DAGStructure for the branch nodes, the branch taken node is the leaf node. The node itself may be arbitrarily complex, but in that case the node should reference a subworkflow etc
	// The parent of the BranchTaken Node is the actual Branch Node and all the data is just forwarded from the Branch to the executed node.
	upstreamNodeIds, err := nCtx.ContextualNodeLookup().ToNode(branchTakenNode.GetID())
	if err != nil {
		return err
	}
	dag := executors.NewLeafNodeDAGStructure(branchTakenNode.GetID(), append(upstreamNodeIds, nCtx.NodeID())...)
	execContext, err := b.getExecutionContextForDownstream(nCtx)
	if err != nil {
		return err
	}
	return b.nodeExecutor.AbortHandler(ctx, execContext, dag, nCtx.ContextualNodeLookup(), branchTakenNode, reason)
}

func (b *branchHandler) Finalize(ctx context.Context, nCtx interfaces.NodeExecutionContext) error {
	branch := nCtx.Node().GetBranchNode()
	if branch == nil {
		return errors.Errorf(errors.IllegalStateError, nCtx.NodeID(), "Invoked branch handler, for a non branch node.")
	}

	// If the branch was already evaluated i.e, Node is in Running status
	branchNodeState := nCtx.NodeStateReader().GetBranchNodeState()
	if branchNodeState.Phase == v1alpha1.BranchNodeNotYetEvaluated {
		logger.Errorf(ctx, "No node finalized through previous branch evaluation.")
		return nil
	} else if branchNodeState.Phase == v1alpha1.BranchNodeError {
		// We should never reach here, but for safety and completeness
		errMsg := "branch evaluation failed"
		if branch.GetElseFail() != nil {
			errMsg = branch.GetElseFail().Message
		}
		logger.Errorf(ctx, "failed to evaluate branch - user error: %s", errMsg)
		return nil
	}

	finalNodeID := branchNodeState.FinalizedNodeID
	branchTakenNode, ok := nCtx.ContextualNodeLookup().GetNode(*finalNodeID)
	if !ok {
		logger.Errorf(ctx, "Downstream node [%v] not found", *finalNodeID)
		return nil
	}

	// Recurse downstream
	// TODO we should replace the call to RecursiveNodeHandler with a call to SingleNode Handler. The inputs are also already known ahead of time
	// There is no DAGStructure for the branch nodes, the branch taken node is the leaf node. The node itself may be arbitrarily complex, but in that case the node should reference a subworkflow etc
	// The parent of the BranchTaken Node is the actual Branch Node and all the data is just forwarded from the Branch to the executed node.
	upstreamNodeIds, err := nCtx.ContextualNodeLookup().ToNode(branchTakenNode.GetID())
	if err != nil {
		return err
	}
	dag := executors.NewLeafNodeDAGStructure(branchTakenNode.GetID(), append(upstreamNodeIds, nCtx.NodeID())...)
	execContext, err := b.getExecutionContextForDownstream(nCtx)
	if err != nil {
		return err
	}
	return b.nodeExecutor.FinalizeHandler(ctx, execContext, dag, nCtx.ContextualNodeLookup(), branchTakenNode)
}

func New(executor interfaces.Node, eventConfig *config.EventConfig, scope promutils.Scope) interfaces.NodeHandler {
	return &branchHandler{
		nodeExecutor: executor,
		m:            metrics{scope: scope},
		eventConfig:  eventConfig,
	}
}
