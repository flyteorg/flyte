package printers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	gotree "github.com/DiSiqueira/GoTree"
	"github.com/fatih/color"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/encoding"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task"
)

var boldString = color.New(color.Bold)

func ColorizeNodePhase(p v1alpha1.NodePhase) string {
	switch p {
	case v1alpha1.NodePhaseNotYetStarted:
		return p.String()
	case v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseDynamicRunning:
		return color.YellowString("%s", p.String())
	case v1alpha1.NodePhaseSucceeded:
		return color.HiGreenString("%s", p.String())
	case v1alpha1.NodePhaseTimedOut:
		return color.HiRedString("%s", p.String())
	case v1alpha1.NodePhaseFailed:
		return color.HiRedString("%s", p.String())
	}
	return color.CyanString("%s", p.String())
}

func CalculateRuntime(s v1alpha1.ExecutableNodeStatus) string {
	if s.GetStartedAt() != nil {
		if s.GetStoppedAt() != nil {
			return s.GetStoppedAt().Sub(s.GetStartedAt().Time).String()
		}
		return time.Since(s.GetStartedAt().Time).String()
	}
	return "na"
}

type NodePrinter struct {
	NodeStatusPrinter
}

func (p NodeStatusPrinter) BaseNodeInfo(node v1alpha1.BaseNode, nodeStatus v1alpha1.ExecutableNodeStatus) []string {
	return []string{
		fmt.Sprintf("%s (%s)", boldString.Sprint(node.GetID()), node.GetKind().String()),
		CalculateRuntime(nodeStatus),
		ColorizeNodePhase(nodeStatus.GetPhase()),
		nodeStatus.GetMessage(),
	}
}

func (p NodeStatusPrinter) NodeInfo(wName string, node v1alpha1.BaseNode, nodeStatus v1alpha1.ExecutableNodeStatus) []string {
	resourceName, err := encoding.FixedLengthUniqueIDForParts(task.IDMaxLength, []string{wName, node.GetID(), strconv.Itoa(int(nodeStatus.GetAttempts()))})
	if err != nil {
		resourceName = "na"
	}
	return append(
		p.BaseNodeInfo(node, nodeStatus),
		fmt.Sprintf("resource=%s", resourceName),
	)
}

func (p NodePrinter) BranchNodeInfo(node v1alpha1.ExecutableNode, nodeStatus v1alpha1.ExecutableNodeStatus) []string {
	info := p.BaseNodeInfo(node, nodeStatus)
	branchStatus := nodeStatus.GetBranchStatus()
	info = append(info, branchStatus.GetPhase().String())
	if branchStatus.GetFinalizedNode() != nil {
		info = append(info, *branchStatus.GetFinalizedNode())
	}
	return info

}

func (p NodePrinter) traverseNode(ctx context.Context, tree gotree.Tree, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, nodeStatus v1alpha1.ExecutableNodeStatus) error {
	switch node.GetKind() {
	case v1alpha1.NodeKindBranch:
		subTree := tree.Add(strings.Join(p.BranchNodeInfo(node, nodeStatus), " | "))
		f := func(nodeID *v1alpha1.NodeID) error {
			if nodeID != nil {
				ifNode, ok := w.GetNode(*nodeID)
				if !ok {
					return fmt.Errorf("failed to find branch node %s", *nodeID)
				}
				if err := p.traverseNode(ctx, subTree, w, ifNode, nodeStatus.GetNodeExecutionStatus(ctx, *nodeID)); err != nil {
					return err
				}
			}
			return nil
		}
		if err := f(node.GetBranchNode().GetIf().GetThenNode()); err != nil {
			return err
		}
		if len(node.GetBranchNode().GetElseIf()) > 0 {
			for _, n := range node.GetBranchNode().GetElseIf() {
				if err := f(n.GetThenNode()); err != nil {
					return err
				}
			}
		}
		if err := f(node.GetBranchNode().GetElse()); err != nil {
			return err
		}
	case v1alpha1.NodeKindWorkflow:
		if node.GetWorkflowNode().GetSubWorkflowRef() != nil {
			s := w.FindSubWorkflow(*node.GetWorkflowNode().GetSubWorkflowRef())
			wp := WorkflowPrinter{}
			return wp.PrintSubWorkflow(ctx, tree, w, s, nodeStatus)
		}
	case v1alpha1.NodeKindTask:
		sub := tree.Add(strings.Join(p.NodeInfo(w.GetName(), node, nodeStatus), " | "))
		if err := p.PrintRecursive(sub, w.GetName(), nodeStatus); err != nil {
			return err
		}
	default:
		_ = tree.Add(strings.Join(p.NodeInfo(w.GetName(), node, nodeStatus), " | "))
	}
	return nil
}

func (p NodePrinter) PrintList(ctx context.Context, tree gotree.Tree, w v1alpha1.ExecutableWorkflow, nodes []v1alpha1.ExecutableNode) error {
	for _, n := range nodes {
		s := w.GetNodeExecutionStatus(ctx, n.GetID())
		if err := p.traverseNode(ctx, tree, w, n, s); err != nil {
			return err
		}
	}
	return nil
}

type NodeStatusPrinter struct {
}

func (p NodeStatusPrinter) PrintRecursive(tree gotree.Tree, wfName string, s v1alpha1.ExecutableNodeStatus) error {
	orderedKeys := sets.String{}
	allStatuses := map[v1alpha1.NodeID]v1alpha1.ExecutableNodeStatus{}
	s.VisitNodeStatuses(func(node v1alpha1.NodeID, status v1alpha1.ExecutableNodeStatus) {
		orderedKeys.Insert(node)
		allStatuses[node] = status
	})

	for _, id := range orderedKeys.List() {
		ns := allStatuses[id]
		sub := tree.Add(strings.Join(p.NodeInfo(wfName, &v1alpha1.NodeSpec{ID: id}, ns), " | "))
		if err := p.PrintRecursive(sub, wfName, ns); err != nil {
			return err
		}
	}

	return nil
}
