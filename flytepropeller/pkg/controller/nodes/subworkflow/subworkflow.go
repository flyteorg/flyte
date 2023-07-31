package subworkflow

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

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"
)

// Subworkflow handler handles inline subWorkflows
type subworkflowHandler struct {
	nodeExecutor interfaces.Node
	eventConfig  *config.EventConfig
}

// Helper method that extracts the SubWorkflow from the ExecutionContext
func GetSubWorkflow(ctx context.Context, nCtx interfaces.NodeExecutionContext) (v1alpha1.ExecutableSubWorkflow, error) {
	node := nCtx.Node()
	subID := *node.GetWorkflowNode().GetSubWorkflowRef()
	subWorkflow := nCtx.ExecutionContext().FindSubWorkflow(subID)
	if subWorkflow == nil {
		return nil, fmt.Errorf("failed to find sub workflow with ID [%s]", subID)
	}
	return subWorkflow, nil
}

// Performs an additional step of passing in and setting the inputs, before handling the execution of a SubWorkflow.
func (s *subworkflowHandler) startAndHandleSubWorkflow(ctx context.Context, nCtx interfaces.NodeExecutionContext, subWorkflow v1alpha1.ExecutableSubWorkflow, nl executors.NodeLookup) (handler.Transition, error) {
	// Before starting the subworkflow, lets set the inputs for the Workflow. The inputs for a SubWorkflow are essentially
	// Copy of the inputs to the Node
	nodeInputs, err := nCtx.InputReader().Get(ctx)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to read input. Error [%s]", err)
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.RuntimeExecutionError, errMsg, nil)), nil
	}

	execContext, err := s.getExecutionContextForDownstream(nCtx)
	if err != nil {
		return handler.UnknownTransition, err
	}
	startStatus, err := s.nodeExecutor.SetInputsForStartNode(ctx, execContext, subWorkflow, nl, nodeInputs)
	if err != nil {
		// NOTE: We are implicitly considering an error when setting the inputs as a system error and hence automatically retryable!
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoUndefined), err
	}

	if startStatus.HasFailed() {
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailureErr(startStatus.Err, nil)), nil
	}

	return s.handleSubWorkflow(ctx, nCtx, subWorkflow, nl)
}

// Calls the recursive node executor to handle the SubWorkflow and translates the results after the success
func (s *subworkflowHandler) handleSubWorkflow(ctx context.Context, nCtx interfaces.NodeExecutionContext, subworkflow v1alpha1.ExecutableSubWorkflow, nl executors.NodeLookup) (handler.Transition, error) {
	// The current node would end up becoming the parent for the sub workflow nodes.
	// This is done to track the lineage. For level zero, the CreateParentInfo will return nil
	newParentInfo, err := common.CreateParentInfo(nCtx.ExecutionContext().GetParentInfo(), nCtx.NodeID(), nCtx.CurrentAttempt())
	if err != nil {
		return handler.UnknownTransition, err
	}
	execContext := executors.NewExecutionContextWithParentInfo(nCtx.ExecutionContext(), newParentInfo)
	state, err := s.nodeExecutor.RecursiveNodeHandler(ctx, execContext, subworkflow, nl, subworkflow.StartNode())
	if err != nil {
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoUndefined), err
	}

	if state.HasFailed() {
		workflowNodeState := handler.WorkflowNodeState{
			Phase: v1alpha1.WorkflowNodePhaseFailing,
			Error: state.Err,
		}

		err = nCtx.NodeStateWriter().PutWorkflowNodeState(workflowNodeState)
		if err != nil {
			logger.Warnf(ctx, "failed to store failing subworkflow state with err: [%v]", err)
			return handler.UnknownTransition, err
		}

		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(nil)), nil
	}

	if state.IsComplete() {
		// If the WF interface has outputs, validate that the outputs file was written.
		var oInfo *handler.OutputInfo
		if outputBindings := subworkflow.GetOutputBindings(); len(outputBindings) > 0 {
			endNodeStatus := nl.GetNodeExecutionStatus(ctx, v1alpha1.EndNodeID)
			store := nCtx.DataStore()
			if endNodeStatus == nil {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.SubWorkflowExecutionFailed, "No end node found in subworkflow.", nil)), err
			}

			sourcePath := v1alpha1.GetOutputsFile(endNodeStatus.GetOutputDir())
			if metadata, err := store.Head(ctx, sourcePath); err == nil {
				if !metadata.Exists() {
					errMsg := fmt.Sprintf("Subworkflow is expected to produce outputs but no outputs file was written to %v.", sourcePath)
					return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.SubWorkflowExecutionFailed, errMsg, nil)), nil
				}
			} else {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoUndefined), nil
			}

			// TODO optimization, we could just point the outputInfo to the path of the subworkflows output
			destinationPath := v1alpha1.GetOutputsFile(nCtx.NodeStatus().GetOutputDir())

			if err := store.CopyRaw(ctx, sourcePath, destinationPath, storage.Options{}); err != nil {
				errMsg := fmt.Sprintf("Failed to copy subworkflow outputs from [%v] to [%v]", sourcePath, destinationPath)
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.SubWorkflowExecutionFailed, errMsg, nil)), nil
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

func (s *subworkflowHandler) getExecutionContextForDownstream(nCtx interfaces.NodeExecutionContext) (executors.ExecutionContext, error) {
	newParentInfo, err := common.CreateParentInfo(nCtx.ExecutionContext().GetParentInfo(), nCtx.NodeID(), nCtx.CurrentAttempt())
	if err != nil {
		return nil, err
	}
	return executors.NewExecutionContextWithParentInfo(nCtx.ExecutionContext(), newParentInfo), nil
}

func (s *subworkflowHandler) HandleFailureNodeOfSubWorkflow(ctx context.Context, nCtx interfaces.NodeExecutionContext, subworkflow v1alpha1.ExecutableSubWorkflow, nl executors.NodeLookup) (handler.Transition, error) {
	originalError := nCtx.NodeStateReader().GetWorkflowNodeState().Error
	if subworkflow.GetOnFailureNode() != nil {
		execContext, err := s.getExecutionContextForDownstream(nCtx)
		if err != nil {
			return handler.UnknownTransition, err
		}
		state, err := s.nodeExecutor.RecursiveNodeHandler(ctx, execContext, subworkflow, nl, subworkflow.GetOnFailureNode())
		if err != nil {
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoUndefined), err
		}

		if state.NodePhase == interfaces.NodePhaseRunning {
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(nil)), nil
		}

		if state.HasFailed() {
			// If handling failure node resulted in failure, its failure will mask the original failure for the workflow
			// TODO: Consider returning both errors.
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailureErr(state.Err, nil)), nil
		}

		if state.PartiallyComplete() {
			if err := nCtx.EnqueueOwnerFunc()(); err != nil {
				return handler.UnknownTransition, err
			}

			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailingErr(originalError, nil)), nil
		}

		// When handling the failure node succeeds, the final status will still be failure
		// we use the original error as the failure reason.
		if state.IsComplete() {
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailureErr(
				originalError, nil)), nil
		}
	}

	return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailureErr(
		originalError, nil)), nil
}

func (s *subworkflowHandler) HandleFailingSubWorkflow(ctx context.Context, nCtx interfaces.NodeExecutionContext) (handler.Transition, error) {
	subWorkflow, err := GetSubWorkflow(ctx, nCtx)
	if err != nil {
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.SubWorkflowExecutionFailed, err.Error(), nil)), nil
	}

	if err := s.HandleAbort(ctx, nCtx, "subworkflow failed"); err != nil {
		logger.Warnf(ctx, "failed to abort failing subworkflow with err: [%v]", err)
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoUndefined), err
	}

	if subWorkflow.GetOnFailureNode() == nil {
		logger.Infof(ctx, "Subworkflow has no failure nodes, failing immediately.")
		state := nCtx.NodeStateReader().GetWorkflowNodeState()
		if state.Error != nil {
			return handler.DoTransition(handler.TransitionTypeEphemeral,
				handler.PhaseInfoFailureErr(state.Error, nil)), nil
		}

		return handler.DoTransition(handler.TransitionTypeEphemeral,
			handler.PhaseInfoFailure(core.ExecutionError_UNKNOWN, "SubworkflowNodeFailing", "", nil)), nil
	}

	status := nCtx.NodeStatus()
	nodeLookup := executors.NewNodeLookup(subWorkflow, status, subWorkflow)
	return s.HandleFailureNodeOfSubWorkflow(ctx, nCtx, subWorkflow, nodeLookup)
}

func (s *subworkflowHandler) StartSubWorkflow(ctx context.Context, nCtx interfaces.NodeExecutionContext) (handler.Transition, error) {
	subWorkflow, err := GetSubWorkflow(ctx, nCtx)
	if err != nil {
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.SubWorkflowExecutionFailed, err.Error(), nil)), nil
	}

	status := nCtx.NodeStatus()
	nodeLookup := executors.NewNodeLookup(subWorkflow, status, subWorkflow)

	// assert startStatus.IsComplete() == true
	return s.startAndHandleSubWorkflow(ctx, nCtx, subWorkflow, nodeLookup)
}

func (s *subworkflowHandler) CheckSubWorkflowStatus(ctx context.Context, nCtx interfaces.NodeExecutionContext) (handler.Transition, error) {
	subWorkflow, err := GetSubWorkflow(ctx, nCtx)
	if err != nil {
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.SubWorkflowExecutionFailed, err.Error(), nil)), nil
	}

	status := nCtx.NodeStatus()
	nodeLookup := executors.NewNodeLookup(subWorkflow, status, subWorkflow)
	return s.handleSubWorkflow(ctx, nCtx, subWorkflow, nodeLookup)
}

func (s *subworkflowHandler) HandleAbort(ctx context.Context, nCtx interfaces.NodeExecutionContext, reason string) error {
	subWorkflow, err := GetSubWorkflow(ctx, nCtx)
	if err != nil {
		return err
	}
	status := nCtx.NodeStatus()
	nodeLookup := executors.NewNodeLookup(subWorkflow, status, subWorkflow)
	execContext, err := s.getExecutionContextForDownstream(nCtx)
	if err != nil {
		return err
	}
	return s.nodeExecutor.AbortHandler(ctx, execContext, subWorkflow, nodeLookup, subWorkflow.StartNode(), reason)
}

func newSubworkflowHandler(nodeExecutor interfaces.Node, eventConfig *config.EventConfig) subworkflowHandler {
	return subworkflowHandler{
		nodeExecutor: nodeExecutor,
		eventConfig:  eventConfig,
	}
}
