package subworkflow

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/recovery"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

type workflowNodeHandler struct {
	lpHandler    launchPlanHandler
	subWfHandler subworkflowHandler
	metrics      metrics
}

type metrics struct {
	CacheError labeled.Counter
}

func newMetrics(scope promutils.Scope) metrics {
	return metrics{
		CacheError: labeled.NewCounter("cache_err", "workflow handler failed to store or load from data store.", scope),
	}
}

func (w *workflowNodeHandler) FinalizeRequired() bool {
	return false
}

func (w *workflowNodeHandler) Setup(_ context.Context, _ interfaces.SetupContext) error {
	return nil
}

func (w *workflowNodeHandler) Handle(ctx context.Context, nCtx interfaces.NodeExecutionContext) (handler.Transition, error) {

	logger.Debug(ctx, "Starting workflow Node")
	invalidWFNodeError := func() (handler.Transition, error) {
		errMsg := "workflow wfNode does not have a subworkflow or child workflow reference"
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM,
			errors.BadSpecificationError, errMsg, nil)), nil
	}

	updateNodeStateFn := func(transition handler.Transition, workflowNodeState handler.WorkflowNodeState, err error) (handler.Transition, error) {
		if err != nil {
			return transition, err
		}

		err = nCtx.NodeStateWriter().PutWorkflowNodeState(workflowNodeState)
		if err != nil {
			logger.Errorf(ctx, "Failed to store WorkflowNodeState, err :%s", err.Error())
			return handler.UnknownTransition, err
		}

		return transition, err
	}

	wfNode := nCtx.Node().GetWorkflowNode()
	wfNodeState := nCtx.NodeStateReader().GetWorkflowNodeState()
	workflowPhase := wfNodeState.Phase
	switch workflowPhase {
	case v1alpha1.WorkflowNodePhaseUndefined:
		if wfNode == nil {
			errMsg := "Invoked workflow handler, for a non workflow Node."
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.RuntimeExecutionError, errMsg, nil)), nil
		}

		if wfNode.GetSubWorkflowRef() != nil {
			trns, err := w.subWfHandler.StartSubWorkflow(ctx, nCtx)
			return updateNodeStateFn(trns, handler.WorkflowNodeState{Phase: v1alpha1.WorkflowNodePhaseExecuting}, err)
		} else if wfNode.GetLaunchPlanRefID() != nil {
			trns, err := w.lpHandler.StartLaunchPlan(ctx, nCtx)
			return updateNodeStateFn(trns, handler.WorkflowNodeState{Phase: v1alpha1.WorkflowNodePhaseExecuting}, err)
		}

		return invalidWFNodeError()
	case v1alpha1.WorkflowNodePhaseExecuting:
		if wfNode.GetSubWorkflowRef() != nil {
			return w.subWfHandler.CheckSubWorkflowStatus(ctx, nCtx)
		} else if wfNode.GetLaunchPlanRefID() != nil {
			return w.lpHandler.CheckLaunchPlanStatus(ctx, nCtx)
		}
	case v1alpha1.WorkflowNodePhaseFailing:
		if wfNode == nil {
			errMsg := "Invoked workflow handler, for a non workflow Node."
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.RuntimeExecutionError, errMsg, nil)), nil
		}

		if wfNode.GetSubWorkflowRef() != nil {
			originalError := nCtx.NodeStateReader().GetWorkflowNodeState().Error
			trns, err := w.subWfHandler.HandleFailingSubWorkflow(ctx, nCtx)
			return updateNodeStateFn(trns, handler.WorkflowNodeState{Phase: workflowPhase, Error: originalError}, err)
		} else if wfNode.GetLaunchPlanRefID() != nil {
			// There is no failure node for launch plans, terminate immediately.
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailureErr(wfNodeState.Error, nil)), nil
		}

		return invalidWFNodeError()
	}

	return invalidWFNodeError()
}

func (w *workflowNodeHandler) Abort(ctx context.Context, nCtx interfaces.NodeExecutionContext, reason string) error {
	wfNode := nCtx.Node().GetWorkflowNode()
	if wfNode.GetSubWorkflowRef() != nil {
		return w.subWfHandler.HandleAbort(ctx, nCtx, reason)
	}

	if wfNode.GetLaunchPlanRefID() != nil {
		return w.lpHandler.HandleAbort(ctx, nCtx, reason)
	}
	return nil
}

func (w *workflowNodeHandler) Finalize(ctx context.Context, _ interfaces.NodeExecutionContext) error {
	logger.Warnf(ctx, "Subworkflow finalize invoked. Nothing to be done")
	return nil
}

func New(executor interfaces.Node, workflowLauncher launchplan.Executor, recoveryClient recovery.Client, eventConfig *config.EventConfig, scope promutils.Scope) interfaces.NodeHandler {
	workflowScope := scope.NewSubScope("workflow")
	return &workflowNodeHandler{
		subWfHandler: newSubworkflowHandler(executor, eventConfig),
		lpHandler: launchPlanHandler{
			launchPlan:     workflowLauncher,
			recoveryClient: recoveryClient,
			eventConfig:    eventConfig,
		},
		metrics: newMetrics(workflowScope),
	}
}
