package subworkflow

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytepropeller/pkg/controller/config"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/recovery"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/common"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
)

type launchPlanHandler struct {
	launchPlan     launchplan.Executor
	recoveryClient recovery.Client
	eventConfig    *config.EventConfig
}

func getParentNodeExecutionID(nCtx handler.NodeExecutionContext) (*core.NodeExecutionIdentifier, error) {
	nodeExecID := &core.NodeExecutionIdentifier{
		ExecutionId: nCtx.NodeExecutionMetadata().GetNodeExecutionID().ExecutionId,
	}
	if nCtx.ExecutionContext().GetEventVersion() != v1alpha1.EventVersion0 {
		var err error
		currentNodeUniqueID, err := common.GenerateUniqueID(nCtx.ExecutionContext().GetParentInfo(), nCtx.NodeExecutionMetadata().GetNodeExecutionID().NodeId)
		if err != nil {
			return nil, err
		}
		nodeExecID.NodeId = currentNodeUniqueID
	} else {
		nodeExecID.NodeId = nCtx.NodeExecutionMetadata().GetNodeExecutionID().NodeId
	}
	return nodeExecID, nil
}

func (l *launchPlanHandler) StartLaunchPlan(ctx context.Context, nCtx handler.NodeExecutionContext) (handler.Transition, error) {
	nodeInputs, err := nCtx.InputReader().Get(ctx)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to read input. Error [%s]", err)
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.RuntimeExecutionError, errMsg, nil)), nil
	}

	parentNodeExecutionID, err := getParentNodeExecutionID(nCtx)
	if err != nil {
		return handler.UnknownTransition, err
	}
	childID, err := GetChildWorkflowExecutionID(
		parentNodeExecutionID,
		nCtx.CurrentAttempt(),
	)
	if err != nil {
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.RuntimeExecutionError, "failed to create unique ID", nil)), nil
	}

	launchCtx := launchplan.LaunchContext{
		ParentNodeExecution: parentNodeExecutionID,
		MaxParallelism:      nCtx.ExecutionContext().GetExecutionConfig().MaxParallelism,
		SecurityContext:     nCtx.ExecutionContext().GetSecurityContext(),
		RawOutputDataConfig: nCtx.ExecutionContext().GetRawOutputDataConfig().RawOutputDataConfig,
		Labels:              nCtx.ExecutionContext().GetLabels(),
		Annotations:         nCtx.ExecutionContext().GetAnnotations(),
	}

	if nCtx.ExecutionContext().GetExecutionConfig().RecoveryExecution.WorkflowExecutionIdentifier != nil {
		recovered, err := l.recoveryClient.RecoverNodeExecution(ctx, nCtx.ExecutionContext().GetExecutionConfig().RecoveryExecution.WorkflowExecutionIdentifier, nCtx.NodeExecutionMetadata().GetNodeExecutionID())
		if err != nil {
			st, ok := status.FromError(err)
			if !ok || st.Code() != codes.NotFound {
				logger.Warnf(ctx, "Failed to recover workflow node [%+v] with err [%+v]", nCtx.NodeExecutionMetadata().GetNodeExecutionID(), err)
			}
		}
		if recovered != nil && recovered.Closure != nil && recovered.Closure.Phase == core.NodeExecution_SUCCEEDED {
			if recovered.Closure.GetWorkflowNodeMetadata() != nil {
				launchCtx.RecoveryExecution = recovered.Closure.GetWorkflowNodeMetadata().ExecutionId
			} else {
				logger.Debugf(ctx, "Attempted to recovered workflow node execution [%+v] but was missing workflow node metadata", recovered.Id)
			}
		}
	}
	err = l.launchPlan.Launch(ctx, launchCtx, childID, nCtx.Node().GetWorkflowNode().GetLaunchPlanRefID().Identifier, nodeInputs)
	if err != nil {
		if launchplan.IsAlreadyExists(err) {
			logger.Infof(ctx, "Execution already exists [%s].", childID.Name)
		} else if launchplan.IsUserError(err) {
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_USER, errors.RuntimeExecutionError, err.Error(), &handler.ExecutionInfo{
				WorkflowNodeInfo: &handler.WorkflowNodeInfo{LaunchedWorkflowID: childID},
			})), nil
		} else {
			return handler.UnknownTransition, err
		}
	} else {
		eCtx := nCtx.ExecutionContext()
		logger.Infof(ctx, "Launched launchplan with ID [%s], Parallelism is now set to [%d]", childID.Name, eCtx.IncrementParallelism())
	}

	return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(&handler.ExecutionInfo{
		WorkflowNodeInfo: &handler.WorkflowNodeInfo{LaunchedWorkflowID: childID},
	})), nil
}

func (l *launchPlanHandler) CheckLaunchPlanStatus(ctx context.Context, nCtx handler.NodeExecutionContext) (handler.Transition, error) {
	parentNodeExecutionID, err := getParentNodeExecutionID(nCtx)
	if err != nil {
		return handler.UnknownTransition, err
	}
	// Handle launch plan
	childID, err := GetChildWorkflowExecutionID(
		parentNodeExecutionID,
		nCtx.CurrentAttempt(),
	)

	if err != nil {
		// THIS SHOULD NEVER HAPPEN
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.RuntimeExecutionError, "failed to create unique ID", nil)), nil
	}

	wfStatusClosure, err := l.launchPlan.GetStatus(ctx, childID)
	if err != nil {
		if launchplan.IsNotFound(err) { // NotFound
			errorCode, _ := errors.GetErrorCode(err)
			err = errors.Wrapf(errorCode, nCtx.NodeID(), err, "launch-plan not found")
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errorCode, err.Error(), &handler.ExecutionInfo{
				WorkflowNodeInfo: &handler.WorkflowNodeInfo{LaunchedWorkflowID: childID},
			})), nil
		}

		return handler.UnknownTransition, err
	}

	if wfStatusClosure == nil {
		logger.Info(ctx, "Retrieved Launch Plan status is nil. This might indicate pressure on the admin cache."+
			" Consider tweaking its size to allow for more concurrent executions to be cached."+
			" Assuming LP is running, parallelism [%d].", nCtx.ExecutionContext().IncrementParallelism())
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(&handler.ExecutionInfo{
			WorkflowNodeInfo: &handler.WorkflowNodeInfo{LaunchedWorkflowID: childID},
		})), nil
	}

	var wErr error
	switch wfStatusClosure.GetPhase() {
	case core.WorkflowExecution_ABORTED:
		wErr = fmt.Errorf("launchplan execution aborted")
		err = errors.Wrapf(errors.RemoteChildWorkflowExecutionFailed, nCtx.NodeID(), wErr, "launchplan [%s] aborted", childID.Name)
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_USER, errors.RemoteChildWorkflowExecutionFailed, err.Error(), &handler.ExecutionInfo{
			WorkflowNodeInfo: &handler.WorkflowNodeInfo{LaunchedWorkflowID: childID},
		})), nil
	case core.WorkflowExecution_FAILED:
		execErr := &core.ExecutionError{Code: "LaunchPlanExecutionFailed", Message: "Unknown Error"}
		if wfStatusClosure.GetError() != nil {
			execErr = wfStatusClosure.GetError()
		}
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailureErr(execErr, &handler.ExecutionInfo{
			WorkflowNodeInfo: &handler.WorkflowNodeInfo{LaunchedWorkflowID: childID},
		})), nil
	case core.WorkflowExecution_SUCCEEDED:
		// TODO do we need to massage the output to match the alias or is the alias resolution done at the downstream consumer
		// nCtx.Node().GetOutputAlias()
		var oInfo *handler.OutputInfo
		if wfStatusClosure.GetOutputs() != nil {
			outputFile := v1alpha1.GetOutputsFile(nCtx.NodeStatus().GetOutputDir())
			if wfStatusClosure.GetOutputs().GetUri() != "" {
				uri := wfStatusClosure.GetOutputs().GetUri()
				store := nCtx.DataStore()
				err := store.CopyRaw(ctx, storage.DataReference(uri), outputFile, storage.Options{})
				if err != nil {
					logger.Warnf(ctx, "remote output for launchplan execution was not found, uri [%s], err %s", uri, err.Error())
					return handler.UnknownTransition, errors.Wrapf(errors.RuntimeExecutionError, nCtx.NodeID(), err, "remote output for launchplan execution was not found, uri [%s]", uri)
				}
			} else {
				childOutput := wfStatusClosure.GetOutputs().GetValues()
				if err := nCtx.DataStore().WriteProtobuf(ctx, outputFile, storage.Options{}, childOutput); err != nil {
					logger.Debugf(ctx, "failed to write data to Storage, err: %v", err.Error())
					return handler.UnknownTransition, errors.Wrapf(errors.CausedByError, nCtx.NodeID(), err, "failed to copy outputs for child workflow")
				}
			}
			oInfo = &handler.OutputInfo{OutputURI: outputFile}
		}
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(&handler.ExecutionInfo{
			WorkflowNodeInfo: &handler.WorkflowNodeInfo{LaunchedWorkflowID: childID},
			OutputInfo:       oInfo,
		})), nil
	}
	logger.Infof(ctx, "LaunchPlan running, parallelism is now set to [%d]", nCtx.ExecutionContext().IncrementParallelism())
	return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(nil)), nil
}

func (l *launchPlanHandler) HandleAbort(ctx context.Context, nCtx handler.NodeExecutionContext, reason string) error {
	parentNodeExecutionID, err := getParentNodeExecutionID(nCtx)
	if err != nil {
		return err
	}
	childID, err := GetChildWorkflowExecutionID(
		parentNodeExecutionID,
		nCtx.CurrentAttempt(),
	)
	if err != nil {
		// THIS SHOULD NEVER HAPPEN
		return err
	}
	return l.launchPlan.Kill(ctx, childID, fmt.Sprintf("cascading abort as parent execution id [%s] aborted, reason [%s]", nCtx.ExecutionContext().GetName(), reason))
}
