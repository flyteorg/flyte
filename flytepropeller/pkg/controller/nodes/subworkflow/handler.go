package subworkflow

import (
	"context"

	"github.com/lyft/flyteidl/clients/go/events"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/executors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"
)

type workflowNodeHandler struct {
	recorder     events.WorkflowEventRecorder
	lpHandler    launchPlanHandler
	subWfHandler subworkflowHandler
}

func (w *workflowNodeHandler) Initialize(ctx context.Context) error {
	return nil
}

func (w *workflowNodeHandler) StartNode(ctx context.Context, wf v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, nodeInputs *handler.Data) (handler.Status, error) {
	if node.GetWorkflowNode().GetSubWorkflowRef() != nil {
		return w.subWfHandler.StartSubWorkflow(ctx, wf, node, nodeInputs)
	}

	if node.GetWorkflowNode().GetLaunchPlanRefID() != nil {
		return w.lpHandler.StartLaunchPlan(ctx, wf, node, nodeInputs)
	}

	return handler.StatusFailed(errors.Errorf(errors.BadSpecificationError, node.GetID(), "SubWorkflow is incorrectly specified.")), nil
}

func (w *workflowNodeHandler) CheckNodeStatus(ctx context.Context, wf v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, status v1alpha1.ExecutableNodeStatus) (handler.Status, error) {
	if node.GetWorkflowNode().GetSubWorkflowRef() != nil {
		return w.subWfHandler.CheckSubWorkflowStatus(ctx, wf, node, status)
	}

	if node.GetWorkflowNode().GetLaunchPlanRefID() != nil {
		return w.lpHandler.CheckLaunchPlanStatus(ctx, wf, node, status)
	}

	return handler.StatusFailed(errors.Errorf(errors.BadSpecificationError, node.GetID(), "workflow node does not have a subworkflow or child workflow reference")), nil
}

func (w *workflowNodeHandler) HandleFailingNode(ctx context.Context, wf v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) (handler.Status, error) {
	if node.GetWorkflowNode() != nil && node.GetWorkflowNode().GetSubWorkflowRef() != nil {
		return w.subWfHandler.HandleSubWorkflowFailingNode(ctx, wf, node)
	}
	return handler.StatusFailed(nil), nil
}

func (w *workflowNodeHandler) AbortNode(ctx context.Context, wf v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) error {
	if node.GetWorkflowNode().GetSubWorkflowRef() != nil {
		return w.subWfHandler.HandleAbort(ctx, wf, node)
	}

	if node.GetWorkflowNode().GetLaunchPlanRefID() != nil {
		return w.lpHandler.HandleAbort(ctx, wf, node)
	}
	return nil
}

func New(executor executors.Node, eventSink events.EventSink, workflowLauncher launchplan.Executor, enQWorkflow v1alpha1.EnqueueWorkflow, store *storage.DataStore, scope promutils.Scope) handler.IFace {
	subworkflowScope := scope.NewSubScope("workflow")
	return &workflowNodeHandler{
		subWfHandler: newSubworkflowHandler(executor, enQWorkflow, store),
		lpHandler: launchPlanHandler{
			store:      store,
			launchPlan: workflowLauncher,
		},
		recorder: events.NewWorkflowEventRecorder(eventSink, subworkflowScope),
	}
}
