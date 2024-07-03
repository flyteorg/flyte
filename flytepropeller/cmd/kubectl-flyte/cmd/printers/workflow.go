package printers

import (
	"context"
	"fmt"
	"time"

	gotree "github.com/DiSiqueira/GoTree"
	"github.com/fatih/color"

	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/visualize"
)

func ColorizeWorkflowPhase(p v1alpha1.WorkflowPhase) string {
	switch p {
	case v1alpha1.WorkflowPhaseReady:
		return p.String()
	case v1alpha1.WorkflowPhaseRunning:
		return color.YellowString("%s", p.String())
	case v1alpha1.WorkflowPhaseSuccess:
		return color.HiGreenString("%s", p.String())
	case v1alpha1.WorkflowPhaseFailed:
		return color.HiRedString("%s", p.String())
	}
	return color.CyanString("%s", p.String())
}

func CalculateWorkflowRuntime(s v1alpha1.ExecutionTimeInfo) string {
	if s.GetStartedAt() != nil {
		if s.GetStoppedAt() != nil {
			return s.GetStoppedAt().Sub(s.GetStartedAt().Time).String()
		}
		return time.Since(s.GetStartedAt().Time).String()
	}
	return "na"
}

type ContextualWorkflow struct {
	v1alpha1.MetaExtended
	v1alpha1.ExecutableSubWorkflow
	v1alpha1.NodeStatusGetter
}

func (w *ContextualWorkflow) GetExecutionConfig() v1alpha1.ExecutionConfig {
	// ExecutionConfig isn't rendered in the printed workflow.
	return v1alpha1.ExecutionConfig{}
}

func (w *ContextualWorkflow) GetConsoleURL() string { return "" }

type WorkflowPrinter struct {
}

func (p WorkflowPrinter) Print(ctx context.Context, tree gotree.Tree, w v1alpha1.ExecutableWorkflow) error {
	sortedNodes, err := visualize.TopologicalSort(w)
	if err != nil {
		return err
	}
	newTree := gotree.New(fmt.Sprintf("%s/%s [ExecId: %s] (%s %s %s)",
		w.GetNamespace(), boldString.Sprint(w.GetName()), w.GetExecutionID(), CalculateWorkflowRuntime(w.GetExecutionStatus()),
		ColorizeWorkflowPhase(w.GetExecutionStatus().GetPhase()), w.GetExecutionStatus().GetMessage()))
	if tree != nil {
		tree.AddTree(newTree)
	}
	np := NodePrinter{}
	return np.PrintList(ctx, newTree, w, sortedNodes)
}

func (p WorkflowPrinter) PrintSubWorkflow(ctx context.Context, tree gotree.Tree, w v1alpha1.ExecutableWorkflow, swf v1alpha1.ExecutableSubWorkflow, ns v1alpha1.ExecutableNodeStatus) error {
	sortedNodes, err := visualize.TopologicalSort(swf)
	if err != nil {
		return err
	}
	newTree := gotree.New(fmt.Sprintf("SubWorkflow [%s] (%s %s %s)",
		swf.GetID(), CalculateWorkflowRuntime(ns),
		ColorizeNodePhase(ns.GetPhase()), ns.GetMessage()))
	if tree != nil {
		tree.AddTree(newTree)
	}
	np := NodePrinter{}

	return np.PrintList(ctx, newTree, &ContextualWorkflow{MetaExtended: w, ExecutableSubWorkflow: swf, NodeStatusGetter: ns}, sortedNodes)
}

func (p WorkflowPrinter) PrintShort(tree gotree.Tree, w v1alpha1.ExecutableWorkflow) error {
	if tree == nil {
		return fmt.Errorf("bad state in printer")
	}
	tree.Add(fmt.Sprintf("%s/%s [ExecId: %s] (%s %s) - Time SinceCreation(%s)",
		w.GetNamespace(), boldString.Sprint(w.GetName()), w.GetExecutionID(), CalculateWorkflowRuntime(w.GetExecutionStatus()),
		ColorizeWorkflowPhase(w.GetExecutionStatus().GetPhase()), time.Since(w.GetCreationTimestamp().Time)))
	return nil
}
