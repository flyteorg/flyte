package subworkflow

import (
	"context"
	"fmt"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/executors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytestdlib/storage"
)

// TODO Add unit tests for subworkflow handler

// Subworkflow handler handles inline subworkflows
type subworkflowHandler struct {
	nodeExecutor executors.Node
}

func (s *subworkflowHandler) DoInlineSubWorkflow(ctx context.Context, nCtx handler.NodeExecutionContext, w v1alpha1.ExecutableWorkflow,
	parentNodeStatus v1alpha1.ExecutableNodeStatus, startNode v1alpha1.ExecutableNode) (handler.Transition, error) {

	// TODO we need to handle failing and success nodes
	state, err := s.nodeExecutor.RecursiveNodeHandler(ctx, w, startNode)
	if err != nil {
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoUndefined), err
	}

	if state.HasFailed() {
		if w.GetOnFailureNode() != nil {
			// TODO ssingh: this is supposed to be failing
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errors.SubWorkflowExecutionFailed, state.Err.Error(), nil)), err
		}

		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errors.SubWorkflowExecutionFailed, state.Err.Error(), nil)), err
	}

	if state.IsComplete() {
		// If the WF interface has outputs, validate that the outputs file was written.
		var oInfo *handler.OutputInfo
		if outputBindings := w.GetOutputBindings(); len(outputBindings) > 0 {
			endNodeStatus := w.GetNodeExecutionStatus(ctx, v1alpha1.EndNodeID)
			store := nCtx.DataStore()
			if endNodeStatus == nil {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errors.SubWorkflowExecutionFailed, "No end node found in subworkflow.", nil)), err
			}

			sourcePath := v1alpha1.GetOutputsFile(endNodeStatus.GetOutputDir())
			if metadata, err := store.Head(ctx, sourcePath); err == nil {
				if !metadata.Exists() {
					errMsg := fmt.Sprintf("Subworkflow is expected to produce outputs but no outputs file was written to %v.", sourcePath)
					return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errors.SubWorkflowExecutionFailed, errMsg, nil)), nil
				}
			} else {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoUndefined), nil
			}

			// TODO optimization, we could just point the outputInfo to the path of the subworkflows output
			destinationPath := v1alpha1.GetOutputsFile(parentNodeStatus.GetOutputDir())
			if err := store.CopyRaw(ctx, sourcePath, destinationPath, storage.Options{}); err != nil {
				errMsg := fmt.Sprintf("Failed to copy subworkflow outputs from [%v] to [%v]", sourcePath, destinationPath)
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errors.SubWorkflowExecutionFailed, errMsg, nil)), nil
			}
			oInfo = &handler.OutputInfo{OutputURI: destinationPath}
		}

		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(&handler.ExecutionInfo{
			OutputInfo: oInfo,
		})), nil
	}

	if state.PartiallyComplete() {
		if err := nCtx.EnqueueOwnerFunc()(); err != nil {
			return handler.UnknownTransition, err
		}
	}

	return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(nil)), nil
}

func (s *subworkflowHandler) DoInFailureHandling(ctx context.Context, nCtx handler.NodeExecutionContext, w v1alpha1.ExecutableWorkflow) (handler.Transition, error) {
	if w.GetOnFailureNode() != nil {
		state, err := s.nodeExecutor.RecursiveNodeHandler(ctx, w, w.GetOnFailureNode())
		if err != nil {
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoUndefined), err
		}
		if state.HasFailed() {
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errors.SubWorkflowExecutionFailed, state.Err.Error(), nil)), nil
		}

		if state.IsComplete() {
			if err := nCtx.EnqueueOwnerFunc()(); err != nil {
				return handler.UnknownTransition, err
			}
		}

		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errors.SubWorkflowExecutionFailed, "failure node handling completed", nil)), nil
	}

	return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailureErr(nil, nil)), nil
}

func (s *subworkflowHandler) StartSubWorkflow(ctx context.Context, nCtx handler.NodeExecutionContext) (handler.Transition, error) {
	node := nCtx.Node()
	subID := *node.GetWorkflowNode().GetSubWorkflowRef()
	subWorkflow := nCtx.Workflow().FindSubWorkflow(subID)
	if subWorkflow == nil {
		errMsg := fmt.Sprintf("No subWorkflow [%s], workflow.", subID)
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errors.SubWorkflowExecutionFailed, errMsg, nil)), nil
	}

	w := nCtx.Workflow()
	status := w.GetNodeExecutionStatus(ctx, node.GetID())
	contextualSubWorkflow := executors.NewSubContextualWorkflow(w, subWorkflow, status)
	startNode := contextualSubWorkflow.StartNode()
	if startNode == nil {
		errMsg := "No start node found in subworkflow."
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errors.SubWorkflowExecutionFailed, errMsg, nil)), nil
	}

	// Before starting the subworkflow, lets set the inputs for the Workflow. The inputs for a SubWorkflow are essentially
	// Copy of the inputs to the Node
	nodeInputs, err := nCtx.InputReader().Get(ctx)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to read input. Error [%s]", err)
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errors.RuntimeExecutionError, errMsg, nil)), nil
	}

	startStatus, err := s.nodeExecutor.SetInputsForStartNode(ctx, contextualSubWorkflow, nodeInputs)
	if err != nil {
		// TODO we are considering an error when setting inputs are retryable
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoUndefined), err
	}

	if startStatus.HasFailed() {
		errorCode, _ := errors.GetErrorCode(startStatus.Err)
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errorCode, startStatus.Err.Error(), nil)), nil
	}

	// assert startStatus.IsComplete() == true
	return s.DoInlineSubWorkflow(ctx, nCtx, contextualSubWorkflow, status, startNode)
}

func (s *subworkflowHandler) CheckSubWorkflowStatus(ctx context.Context, nCtx handler.NodeExecutionContext, w v1alpha1.ExecutableWorkflow, status v1alpha1.ExecutableNodeStatus) (handler.Transition, error) {
	// Handle subworkflow
	subID := *nCtx.Node().GetWorkflowNode().GetSubWorkflowRef()
	subWorkflow := w.FindSubWorkflow(subID)
	if subWorkflow == nil {
		errMsg := fmt.Sprintf("No subWorkflow [%s], workflow.", subID)
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errors.SubWorkflowExecutionFailed, errMsg, nil)), nil
	}

	contextualSubWorkflow := executors.NewSubContextualWorkflow(w, subWorkflow, status)
	startNode := w.StartNode()
	if startNode == nil {
		errMsg := "No start node found in subworkflow"
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errors.SubWorkflowExecutionFailed, errMsg, nil)), nil
	}

	parentNodeStatus := w.GetNodeExecutionStatus(ctx, nCtx.NodeID())
	return s.DoInlineSubWorkflow(ctx, nCtx, contextualSubWorkflow, parentNodeStatus, startNode)
}

func (s *subworkflowHandler) HandleSubWorkflowFailingNode(ctx context.Context, nCtx handler.NodeExecutionContext, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) (handler.Transition, error) {
	status := w.GetNodeExecutionStatus(ctx, node.GetID())
	subID := *node.GetWorkflowNode().GetSubWorkflowRef()
	subWorkflow := w.FindSubWorkflow(subID)
	if subWorkflow == nil {
		errMsg := fmt.Sprintf("No subWorkflow [%s], workflow.", subID)
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errors.SubWorkflowExecutionFailed, errMsg, nil)), nil
	}
	contextualSubWorkflow := executors.NewSubContextualWorkflow(w, subWorkflow, status)
	return s.DoInFailureHandling(ctx, nCtx, contextualSubWorkflow)
}

func (s *subworkflowHandler) HandleAbort(ctx context.Context, nCtx handler.NodeExecutionContext, w v1alpha1.ExecutableWorkflow, workflowID v1alpha1.WorkflowID) error {
	subWorkflow := w.FindSubWorkflow(workflowID)
	if subWorkflow == nil {
		return fmt.Errorf("no sub workflow [%s] found in node [%s]", workflowID, nCtx.NodeID())
	}

	nodeStatus := w.GetNodeExecutionStatus(ctx, nCtx.NodeID())
	contextualSubWorkflow := executors.NewSubContextualWorkflow(w, subWorkflow, nodeStatus)

	startNode := w.StartNode()
	if startNode == nil {
		return fmt.Errorf("no sub workflow [%s] found in node [%s]", workflowID, nCtx.NodeID())
	}

	return s.nodeExecutor.AbortHandler(ctx, contextualSubWorkflow, startNode, "")
}

func newSubworkflowHandler(nodeExecutor executors.Node) subworkflowHandler {
	return subworkflowHandler{
		nodeExecutor: nodeExecutor,
	}
}
