package subworkflow

import (
	"context"
	"fmt"

	"github.com/lyft/flytestdlib/storage"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/executors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
)

// Subworkflow handler handles inline subWorkflows
type subworkflowHandler struct {
	nodeExecutor executors.Node
}

// Helper method that extracts the SubWorkflow from the ExecutionContext
func GetSubWorkflow(ctx context.Context, nCtx handler.NodeExecutionContext) (v1alpha1.ExecutableSubWorkflow, error) {
	node := nCtx.Node()
	subID := *node.GetWorkflowNode().GetSubWorkflowRef()
	subWorkflow := nCtx.ExecutionContext().FindSubWorkflow(subID)
	if subWorkflow == nil {
		return nil, fmt.Errorf("failed to find sub workflow with ID [%s]", subID)
	}
	return subWorkflow, nil
}

// Performs an additional step of passing in and setting the inputs, before handling the execution of a SubWorkflow.
func (s *subworkflowHandler) startAndHandleSubWorkflow(ctx context.Context, nCtx handler.NodeExecutionContext, subWorkflow v1alpha1.ExecutableSubWorkflow, nl executors.NodeLookup) (handler.Transition, error) {
	// Before starting the subworkflow, lets set the inputs for the Workflow. The inputs for a SubWorkflow are essentially
	// Copy of the inputs to the Node
	nodeInputs, err := nCtx.InputReader().Get(ctx)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to read input. Error [%s]", err)
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errors.RuntimeExecutionError, errMsg, nil)), nil
	}

	startStatus, err := s.nodeExecutor.SetInputsForStartNode(ctx, nCtx.ExecutionContext(), subWorkflow, nl, nodeInputs)
	if err != nil {
		// NOTE: We are implicitly considering an error when setting the inputs as a system error and hence automatically retryable!
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoUndefined), err
	}

	if startStatus.HasFailed() {
		errorCode, _ := errors.GetErrorCode(startStatus.Err)
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errorCode, startStatus.Err.Error(), nil)), nil
	}
	return s.handleSubWorkflow(ctx, nCtx, subWorkflow, nl)
}

// Calls the recursive node executor to handle the SubWorkflow and translates the results after the success
func (s *subworkflowHandler) handleSubWorkflow(ctx context.Context, nCtx handler.NodeExecutionContext, subworkflow v1alpha1.ExecutableSubWorkflow, nl executors.NodeLookup) (handler.Transition, error) {

	state, err := s.nodeExecutor.RecursiveNodeHandler(ctx, nCtx.ExecutionContext(), subworkflow, nl, subworkflow.StartNode())
	if err != nil {
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoUndefined), err
	}

	if state.HasFailed() {
		if subworkflow.GetOnFailureNode() != nil {
			// TODO Handle Failure node for subworkflows. We need to add new state to the executor so that, we can continue returning Running, but in the next round start executing DoInFailureHandling - NOTE1
			// https://github.com/lyft/flyte/issues/265
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errors.SubWorkflowExecutionFailed, state.Err.Error(), nil)), err
		}

		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errors.SubWorkflowExecutionFailed, state.Err.Error(), nil)), err
	}

	if state.IsComplete() {
		// If the WF interface has outputs, validate that the outputs file was written.
		var oInfo *handler.OutputInfo
		if outputBindings := subworkflow.GetOutputBindings(); len(outputBindings) > 0 {
			endNodeStatus := nl.GetNodeExecutionStatus(ctx, v1alpha1.EndNodeID)
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
			destinationPath := v1alpha1.GetOutputsFile(nCtx.NodeStatus().GetOutputDir())
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

// TODO related to NOTE1, this is not used currently, but should be used. For this we will need to clean up the state machine in the main handle function
// https://github.com/lyft/flyte/issues/265
func (s *subworkflowHandler) HandleFailureNodeOfSubWorkflow(ctx context.Context, nCtx handler.NodeExecutionContext, subworkflow v1alpha1.ExecutableSubWorkflow, nl executors.NodeLookup) (handler.Transition, error) {
	if subworkflow.GetOnFailureNode() != nil {
		state, err := s.nodeExecutor.RecursiveNodeHandler(ctx, nCtx.ExecutionContext(), subworkflow, nl, subworkflow.GetOnFailureNode())
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
	subWorkflow, err := GetSubWorkflow(ctx, nCtx)
	if err != nil {
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errors.SubWorkflowExecutionFailed, err.Error(), nil)), nil
	}

	status := nCtx.NodeStatus()
	nodeLookup := executors.NewNodeLookup(subWorkflow, status)

	// assert startStatus.IsComplete() == true
	return s.startAndHandleSubWorkflow(ctx, nCtx, subWorkflow, nodeLookup)
}

func (s *subworkflowHandler) CheckSubWorkflowStatus(ctx context.Context, nCtx handler.NodeExecutionContext) (handler.Transition, error) {
	subWorkflow, err := GetSubWorkflow(ctx, nCtx)
	if err != nil {
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errors.SubWorkflowExecutionFailed, err.Error(), nil)), nil
	}

	status := nCtx.NodeStatus()
	nodeLookup := executors.NewNodeLookup(subWorkflow, status)
	return s.startAndHandleSubWorkflow(ctx, nCtx, subWorkflow, nodeLookup)
}

func (s *subworkflowHandler) HandleAbort(ctx context.Context, nCtx handler.NodeExecutionContext, reason string) error {
	subWorkflow, err := GetSubWorkflow(ctx, nCtx)
	if err != nil {
		return err
	}
	status := nCtx.NodeStatus()
	nodeLookup := executors.NewNodeLookup(subWorkflow, status)
	return s.nodeExecutor.AbortHandler(ctx, nCtx.ExecutionContext(), subWorkflow, nodeLookup, subWorkflow.StartNode(), reason)
}

func newSubworkflowHandler(nodeExecutor executors.Node) subworkflowHandler {
	return subworkflowHandler{
		nodeExecutor: nodeExecutor,
	}
}
