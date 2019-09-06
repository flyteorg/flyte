package subworkflow

import (
	"context"
	"fmt"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/storage"
)

type launchPlanHandler struct {
	launchPlan launchplan.Executor
	store      *storage.DataStore
}

func (l *launchPlanHandler) StartLaunchPlan(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, nodeInputs *handler.Data) (handler.Status, error) {
	nodeStatus := w.GetNodeExecutionStatus(node.GetID())
	childID, err := GetChildWorkflowExecutionID(
		w.GetExecutionID().WorkflowExecutionIdentifier,
		node.GetID(),
		nodeStatus.GetAttempts(),
	)

	if err != nil {
		return handler.StatusFailed(errors.Wrapf(errors.RuntimeExecutionError, node.GetID(), err, "failed to create unique ID")), nil
	}

	launchCtx := launchplan.LaunchContext{
		// TODO we need to add principal and nestinglevel as annotations or labels?
		Principal:    "unknown",
		NestingLevel: 0,
		ParentNodeExecution: &core.NodeExecutionIdentifier{
			NodeId:      node.GetID(),
			ExecutionId: w.GetExecutionID().WorkflowExecutionIdentifier,
		},
	}
	err = l.launchPlan.Launch(ctx, launchCtx, childID, node.GetWorkflowNode().GetLaunchPlanRefID().Identifier, nodeInputs)
	if err != nil {
		if launchplan.IsAlreadyExists(err) {
			logger.Info(ctx, "Execution already exists [%s].", childID.Name)
		} else if launchplan.IsUserError(err) {
			return handler.StatusFailed(err), nil
		} else {
			return handler.StatusUndefined, err
		}
	} else {
		logger.Infof(ctx, "Launched launchplan with ID [%s]", childID.Name)
	}

	nodeStatus.GetOrCreateWorkflowStatus().SetWorkflowExecutionName(childID.Name)
	return handler.StatusRunning, nil
}

func (l *launchPlanHandler) CheckLaunchPlanStatus(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, status v1alpha1.ExecutableNodeStatus) (handler.Status, error) {
	// Handle launch plan
	nodeStatus := w.GetNodeExecutionStatus(node.GetID())
	childID, err := GetChildWorkflowExecutionID(
		w.GetExecutionID().WorkflowExecutionIdentifier,
		node.GetID(),
		nodeStatus.GetAttempts(),
	)

	if err != nil {
		// THIS SHOULD NEVER HAPPEN
		return handler.StatusFailed(errors.Wrapf(errors.RuntimeExecutionError, node.GetID(), err, "failed to create unique ID")), nil
	}

	wfStatusClosure, err := l.launchPlan.GetStatus(ctx, childID)
	if err != nil {
		if launchplan.IsNotFound(err) { //NotFound
			return handler.StatusFailed(err), nil
		}

		return handler.StatusUndefined, err
	}

	if wfStatusClosure == nil {
		logger.Info(ctx, "Retrieved Launch Plan status is nil. This might indicate pressure on the admin cache."+
			" Consider tweaking its size to allow for more concurrent executions to be cached.")
		return handler.StatusRunning, nil
	}

	var wErr error
	switch wfStatusClosure.GetPhase() {
	case core.WorkflowExecution_ABORTED:
		wErr = fmt.Errorf("launchplan execution aborted")
		return handler.StatusFailed(errors.Wrapf(errors.RemoteChildWorkflowExecutionFailed, node.GetID(), wErr, "launchplan [%s] failed", childID.Name)), nil
	case core.WorkflowExecution_FAILED:
		wErr = fmt.Errorf("launchplan execution failed without explicit error")
		if wfStatusClosure.GetError() != nil {
			wErr = fmt.Errorf(" errorCode[%s]: %s", wfStatusClosure.GetError().Code, wfStatusClosure.GetError().Message)
		}
		return handler.StatusFailed(errors.Wrapf(errors.RemoteChildWorkflowExecutionFailed, node.GetID(), wErr, "launchplan [%s] failed", childID.Name)), nil
	case core.WorkflowExecution_SUCCEEDED:
		if wfStatusClosure.GetOutputs() != nil {
			outputFile := v1alpha1.GetOutputsFile(nodeStatus.GetDataDir())
			childOutput := &core.LiteralMap{}
			uri := wfStatusClosure.GetOutputs().GetUri()
			if uri != "" {
				// Copy remote data to local S3 path
				if err := l.store.ReadProtobuf(ctx, storage.DataReference(uri), childOutput); err != nil {
					if storage.IsNotFound(err) {
						return handler.StatusFailed(errors.Wrapf(errors.RemoteChildWorkflowExecutionFailed, node.GetID(), err, "remote output for launchplan execution was not found, uri [%s]", uri)), nil
					}
					return handler.StatusUndefined, errors.Wrapf(errors.RuntimeExecutionError, node.GetID(), err, "failed to read outputs from child workflow @ [%s]", uri)
				}

			} else if wfStatusClosure.GetOutputs().GetValues() != nil {
				// Store data to S3Path
				childOutput = wfStatusClosure.GetOutputs().GetValues()
			}
			if err := l.store.WriteProtobuf(ctx, outputFile, storage.Options{}, childOutput); err != nil {
				logger.Debugf(ctx, "failed to write data to Storage, err: %v", err.Error())
				return handler.StatusUndefined, errors.Wrapf(errors.CausedByError, node.GetID(), err, "failed to copy outputs for child workflow")
			}
		}
		return handler.StatusSuccess, nil
	}
	return handler.StatusRunning, nil
}

func (l *launchPlanHandler) HandleAbort(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) error {
	nodeStatus := w.GetNodeExecutionStatus(node.GetID())
	childID, err := GetChildWorkflowExecutionID(
		w.GetExecutionID().WorkflowExecutionIdentifier,
		node.GetID(),
		nodeStatus.GetAttempts(),
	)
	if err != nil {
		// THIS SHOULD NEVER HAPPEN
		return err
	}
	return l.launchPlan.Kill(ctx, childID, fmt.Sprintf("parent execution id [%s] aborted", w.GetName()))
}
