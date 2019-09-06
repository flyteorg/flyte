package subworkflow

import (
	"context"
	"fmt"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/executors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/storage"
)

//TODO Add unit tests for subworkflow handler

// Subworkflow handler handles inline subworkflows
type subworkflowHandler struct {
	nodeExecutor    executors.Node
	enqueueWorkflow v1alpha1.EnqueueWorkflow
	store           *storage.DataStore
}

func (s *subworkflowHandler) DoInlineSubWorkflow(ctx context.Context, w v1alpha1.ExecutableWorkflow,
	parentNodeStatus v1alpha1.ExecutableNodeStatus, startNode v1alpha1.ExecutableNode) (handler.Status, error) {

	//TODO we need to handle failing and success nodes
	state, err := s.nodeExecutor.RecursiveNodeHandler(ctx, w, startNode)
	if err != nil {
		return handler.StatusUndefined, err
	}

	if state.HasFailed() {
		if w.GetOnFailureNode() != nil {
			return handler.StatusFailing(state.Err), nil
		}
		return handler.StatusFailed(state.Err), nil
	}

	if state.IsComplete() {
		nodeID := ""
		if parentNodeStatus.GetParentNodeID() != nil {
			nodeID = *parentNodeStatus.GetParentNodeID()
		}

		// If the WF interface has outputs, validate that the outputs file was written.
		if outputBindings := w.GetOutputBindings(); len(outputBindings) > 0 {
			endNodeStatus := w.GetNodeExecutionStatus(v1alpha1.EndNodeID)
			if endNodeStatus == nil {
				return handler.StatusFailed(errors.Errorf(errors.SubWorkflowExecutionFailed, nodeID,
					"No end node found in subworkflow.")), nil
			}

			sourcePath := v1alpha1.GetOutputsFile(endNodeStatus.GetDataDir())
			if metadata, err := s.store.Head(ctx, sourcePath); err == nil {
				if !metadata.Exists() {
					return handler.StatusFailed(errors.Errorf(errors.SubWorkflowExecutionFailed, nodeID,
						"Subworkflow is expected to produce outputs but no outputs file was written to %v.",
						sourcePath)), nil
				}
			} else {
				return handler.StatusUndefined, err
			}

			destinationPath := v1alpha1.GetOutputsFile(parentNodeStatus.GetDataDir())
			if err := s.store.CopyRaw(ctx, sourcePath, destinationPath, storage.Options{}); err != nil {
				return handler.StatusFailed(errors.Wrapf(errors.OutputsNotFoundError, nodeID,
					err, "Failed to copy subworkflow outputs from [%v] to [%v]",
					sourcePath, destinationPath)), nil
			}
		}

		return handler.StatusSuccess, nil
	}

	if state.PartiallyComplete() {
		// Re-enqueue the workflow
		s.enqueueWorkflow(w.GetK8sWorkflowID().String())
	}

	return handler.StatusRunning, nil
}

func (s *subworkflowHandler) DoInFailureHandling(ctx context.Context, w v1alpha1.ExecutableWorkflow) (handler.Status, error) {
	if w.GetOnFailureNode() != nil {
		state, err := s.nodeExecutor.RecursiveNodeHandler(ctx, w, w.GetOnFailureNode())
		if err != nil {
			return handler.StatusUndefined, err
		}
		if state.HasFailed() {
			return handler.StatusFailed(state.Err), nil
		}
		if state.IsComplete() {
			// Re-enqueue the workflow
			s.enqueueWorkflow(w.GetK8sWorkflowID().String())
			return handler.StatusFailed(nil), nil
		}
		return handler.StatusFailing(nil), nil
	}
	return handler.StatusFailed(nil), nil
}

func (s *subworkflowHandler) StartSubWorkflow(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, nodeInputs *handler.Data) (handler.Status, error) {
	subID := *node.GetWorkflowNode().GetSubWorkflowRef()
	subWorkflow := w.FindSubWorkflow(subID)
	if subWorkflow == nil {
		return handler.StatusFailed(errors.Errorf(errors.SubWorkflowExecutionFailed, node.GetID(), "No subWorkflow [%s], workflow.", subID)), nil
	}

	status := w.GetNodeExecutionStatus(node.GetID())
	contextualSubWorkflow := executors.NewSubContextualWorkflow(w, subWorkflow, status)
	startNode := contextualSubWorkflow.StartNode()
	if startNode == nil {
		return handler.StatusFailed(errors.Errorf(errors.SubWorkflowExecutionFailed, "", "No start node found in subworkflow.")), nil
	}

	// Before starting the subworkflow, lets set the inputs for the Workflow. The inputs for a SubWorkflow are essentially
	// Copy of the inputs to the Node
	nodeStatus := contextualSubWorkflow.GetNodeExecutionStatus(startNode.GetID())
	if len(nodeStatus.GetDataDir()) == 0 {
		dataDir, err := contextualSubWorkflow.GetExecutionStatus().ConstructNodeDataDir(ctx, s.store, startNode.GetID())
		if err != nil {
			logger.Errorf(ctx, "Failed to create metadata store key. Error [%v]", err)
			return handler.StatusUndefined, errors.Wrapf(errors.CausedByError, startNode.GetID(), err, "Failed to create metadata store key.")
		}

		nodeStatus.SetDataDir(dataDir)
		startStatus, err := s.nodeExecutor.SetInputsForStartNode(ctx, contextualSubWorkflow, nodeInputs)
		if err != nil {
			// TODO we are considering an error when setting inputs are retryable
			return handler.StatusUndefined, err
		}

		if startStatus.HasFailed() {
			return handler.StatusFailed(startStatus.Err), nil
		}
	}

	return s.DoInlineSubWorkflow(ctx, contextualSubWorkflow, status, startNode)
}

func (s *subworkflowHandler) CheckSubWorkflowStatus(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, status v1alpha1.ExecutableNodeStatus) (handler.Status, error) {
	// Handle subworkflow
	subID := *node.GetWorkflowNode().GetSubWorkflowRef()
	subWorkflow := w.FindSubWorkflow(subID)
	if subWorkflow == nil {
		return handler.StatusFailed(errors.Errorf(errors.SubWorkflowExecutionFailed, node.GetID(), "No subWorkflow [%s], workflow.", subID)), nil
	}

	contextualSubWorkflow := executors.NewSubContextualWorkflow(w, subWorkflow, status)
	startNode := w.StartNode()
	if startNode == nil {
		return handler.StatusFailed(errors.Errorf(errors.SubWorkflowExecutionFailed, node.GetID(), "No start node found in subworkflow")), nil
	}

	parentNodeStatus := w.GetNodeExecutionStatus(node.GetID())
	return s.DoInlineSubWorkflow(ctx, contextualSubWorkflow, parentNodeStatus, startNode)
}

func (s *subworkflowHandler) HandleSubWorkflowFailingNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) (handler.Status, error) {
	status := w.GetNodeExecutionStatus(node.GetID())
	subID := *node.GetWorkflowNode().GetSubWorkflowRef()
	subWorkflow := w.FindSubWorkflow(subID)
	if subWorkflow == nil {
		return handler.StatusFailed(errors.Errorf(errors.SubWorkflowExecutionFailed, node.GetID(), "No subWorkflow [%s], workflow.", subID)), nil
	}
	contextualSubWorkflow := executors.NewSubContextualWorkflow(w, subWorkflow, status)
	return s.DoInFailureHandling(ctx, contextualSubWorkflow)
}

func (s *subworkflowHandler) HandleAbort(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) error {
	subID := *node.GetWorkflowNode().GetSubWorkflowRef()
	subWorkflow := w.FindSubWorkflow(subID)
	if subWorkflow == nil {
		return fmt.Errorf("no sub workflow [%s] found in node [%s]", subID, node.GetID())
	}

	nodeStatus := w.GetNodeExecutionStatus(node.GetID())
	contextualSubWorkflow := executors.NewSubContextualWorkflow(w, subWorkflow, nodeStatus)

	startNode := w.StartNode()
	if startNode == nil {
		return fmt.Errorf("no sub workflow [%s] found in node [%s]", subID, node.GetID())
	}

	return s.nodeExecutor.AbortHandler(ctx, contextualSubWorkflow, startNode)
}

func newSubworkflowHandler(nodeExecutor executors.Node, enqueueWorkflow v1alpha1.EnqueueWorkflow, store *storage.DataStore) subworkflowHandler {
	return subworkflowHandler{
		nodeExecutor:    nodeExecutor,
		enqueueWorkflow: enqueueWorkflow,
		store:           store,
	}
}
