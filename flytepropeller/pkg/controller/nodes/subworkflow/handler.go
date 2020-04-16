package subworkflow

import (
	"context"

	"github.com/lyft/flytestdlib/promutils"

	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils/labeled"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/executors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
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

func (w *workflowNodeHandler) Setup(_ context.Context, _ handler.SetupContext) error {
	return nil
}

func (w *workflowNodeHandler) Handle(ctx context.Context, nCtx handler.NodeExecutionContext) (handler.Transition, error) {

	logger.Debug(ctx, "Starting workflow Node")
	invalidWFNodeError := func() (handler.Transition, error) {
		errMsg := "workflow wfNode does not have a subworkflow or child workflow reference"
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(
			errors.BadSpecificationError, errMsg, nil)), nil
	}

	updateNodeStateFn := func(transition handler.Transition, err error) (handler.Transition, error) {
		if err != nil {
			return transition, err
		}
		workflowNodeState := handler.WorkflowNodeState{Phase: v1alpha1.WorkflowNodePhaseExecuting}
		err = nCtx.NodeStateWriter().PutWorkflowNodeState(workflowNodeState)
		if err != nil {
			logger.Errorf(ctx, "Failed to store WorkflowNodeState, err :%s", err.Error())
			return handler.UnknownTransition, err
		}

		return transition, err
	}

	wfNode := nCtx.Node().GetWorkflowNode()
	workflowPhase := nCtx.NodeStateReader().GetWorkflowNodeState().Phase
	if workflowPhase == v1alpha1.WorkflowNodePhaseUndefined {
		if wfNode == nil {
			errMsg := "Invoked workflow handler, for a non workflow Node."
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(errors.RuntimeExecutionError, errMsg, nil)), nil
		}

		if wfNode.GetSubWorkflowRef() != nil {
			return updateNodeStateFn(w.subWfHandler.StartSubWorkflow(ctx, nCtx))
		} else if wfNode.GetLaunchPlanRefID() != nil {
			return updateNodeStateFn(w.lpHandler.StartLaunchPlan(ctx, nCtx))
		}

		return invalidWFNodeError()
	}

	if wfNode.GetSubWorkflowRef() != nil {
		return w.subWfHandler.CheckSubWorkflowStatus(ctx, nCtx)
	} else if wfNode.GetLaunchPlanRefID() != nil {
		return w.lpHandler.CheckLaunchPlanStatus(ctx, nCtx)
	}

	return invalidWFNodeError()
}

func (w *workflowNodeHandler) Abort(ctx context.Context, nCtx handler.NodeExecutionContext, reason string) error {
	wfNode := nCtx.Node().GetWorkflowNode()
	if wfNode.GetSubWorkflowRef() != nil {
		return w.subWfHandler.HandleAbort(ctx, nCtx, reason)
	}

	if wfNode.GetLaunchPlanRefID() != nil {
		return w.lpHandler.HandleAbort(ctx, nCtx, reason)
	}
	return nil
}

func (w *workflowNodeHandler) Finalize(ctx context.Context, _ handler.NodeExecutionContext) error {
	logger.Warnf(ctx, "Subworkflow finalize invoked. Nothing to be done")
	return nil
}

func New(executor executors.Node, workflowLauncher launchplan.Executor, scope promutils.Scope) handler.Node {
	workflowScope := scope.NewSubScope("workflow")
	return &workflowNodeHandler{
		subWfHandler: newSubworkflowHandler(executor),
		lpHandler: launchPlanHandler{
			launchPlan: workflowLauncher,
		},
		metrics: newMetrics(workflowScope),
	}
}
