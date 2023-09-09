package subworkflow

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/common"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/recovery"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type launchPlanHandler struct {
	launchPlan     launchplan.Executor
	recoveryClient recovery.Client
	eventConfig    *config.EventConfig
}

func getParentNodeExecutionID(nCtx interfaces.NodeExecutionContext) (*core.NodeExecutionIdentifier, error) {
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

func (l *launchPlanHandler) StartLaunchPlan(ctx context.Context, nCtx interfaces.NodeExecutionContext) (handler.Transition, error) {
	nodeInputs, err := nCtx.InputReader().Get(ctx)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to read input. Error [%s]", err)
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.RuntimeExecutionError, errMsg, nil)), nil
	}

	parentNodeExecutionID, err := getParentNodeExecutionID(nCtx)
	if err != nil {
		return handler.UnknownTransition, err
	}

	childID, err := GetChildWorkflowExecutionIDForExecution(
		parentNodeExecutionID,
		nCtx,
	)

	if err != nil {
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.RuntimeExecutionError, "failed to create unique ID", nil)), nil
	}

	launchCtx := launchplan.LaunchContext{
		ParentNodeExecution:  parentNodeExecutionID,
		MaxParallelism:       nCtx.ExecutionContext().GetExecutionConfig().MaxParallelism,
		SecurityContext:      nCtx.ExecutionContext().GetSecurityContext(),
		RawOutputDataConfig:  nCtx.ExecutionContext().GetRawOutputDataConfig().RawOutputDataConfig,
		Labels:               nCtx.ExecutionContext().GetLabels(),
		Annotations:          nCtx.ExecutionContext().GetAnnotations(),
		Interruptible:        nCtx.ExecutionContext().GetExecutionConfig().Interruptible,
		OverwriteCache:       nCtx.ExecutionContext().GetExecutionConfig().OverwriteCache,
		EnvironmentVariables: nCtx.ExecutionContext().GetExecutionConfig().EnvironmentVariables,
	}

	if nCtx.ExecutionContext().GetExecutionConfig().RecoveryExecution.WorkflowExecutionIdentifier != nil {
		fullyQualifiedNodeID := nCtx.NodeExecutionMetadata().GetNodeExecutionID().NodeId
		if nCtx.ExecutionContext().GetEventVersion() != v1alpha1.EventVersion0 {
			// compute fully qualified node id (prefixed with parent id and retry attempt) to ensure uniqueness
			var err error
			fullyQualifiedNodeID, err = common.GenerateUniqueID(nCtx.ExecutionContext().GetParentInfo(), nCtx.NodeExecutionMetadata().GetNodeExecutionID().NodeId)
			if err != nil {
				return handler.UnknownTransition, err
			}
		}

		recovered, err := l.recoveryClient.RecoverNodeExecution(ctx, nCtx.ExecutionContext().GetExecutionConfig().RecoveryExecution.WorkflowExecutionIdentifier, fullyQualifiedNodeID)
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
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_USER, errors.RuntimeExecutionError, err.Error(), nil)), nil
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

func GetChildWorkflowExecutionIDForExecution(parentNodeExecID *core.NodeExecutionIdentifier, nCtx interfaces.NodeExecutionContext) (*core.WorkflowExecutionIdentifier, error) {
	// Handle launch plan
	if nCtx.ExecutionContext().GetDefinitionVersion() == v1alpha1.WorkflowDefinitionVersion0 {
		return GetChildWorkflowExecutionID(
			parentNodeExecID,
			nCtx.CurrentAttempt(),
		)
	}

	return GetChildWorkflowExecutionIDV2(
		parentNodeExecID,
		nCtx.CurrentAttempt(),
	)
}

func (l *launchPlanHandler) CheckLaunchPlanStatus(ctx context.Context, nCtx interfaces.NodeExecutionContext) (handler.Transition, error) {
	parentNodeExecutionID, err := getParentNodeExecutionID(nCtx)
	if err != nil {
		return handler.UnknownTransition, err
	}

	// Handle launch plan
	childID, err := GetChildWorkflowExecutionIDForExecution(
		parentNodeExecutionID,
		nCtx,
	)

	if err != nil {
		// THIS SHOULD NEVER HAPPEN
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.RuntimeExecutionError, "failed to create unique ID", nil)), nil
	}

	wfStatusClosure, outputs, err := l.launchPlan.GetStatus(ctx, childID)
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
		if outputs != nil {
			outputFile := v1alpha1.GetOutputsFile(nCtx.NodeStatus().GetOutputDir())
			if err := nCtx.DataStore().WriteProtobuf(ctx, outputFile, storage.Options{}, outputs); err != nil {
				logger.Debugf(ctx, "failed to write data to Storage, err: %v", err.Error())
				return handler.UnknownTransition, errors.Wrapf(errors.CausedByError, nCtx.NodeID(), err, "failed to copy outputs for child workflow")
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

func (l *launchPlanHandler) HandleAbort(ctx context.Context, nCtx interfaces.NodeExecutionContext, reason string) error {
	parentNodeExecutionID, err := getParentNodeExecutionID(nCtx)
	if err != nil {
		return err
	}
	childID, err := GetChildWorkflowExecutionIDForExecution(
		parentNodeExecutionID,
		nCtx,
	)
	if err != nil {
		// THIS SHOULD NEVER HAPPEN
		return err
	}
	return l.launchPlan.Kill(ctx, childID, fmt.Sprintf("cascading abort as parent execution id [%s] aborted, reason [%s]", nCtx.ExecutionContext().GetName(), reason))
}
