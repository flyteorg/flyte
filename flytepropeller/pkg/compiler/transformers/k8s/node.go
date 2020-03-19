package k8s

import (
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/compiler/common"
	"github.com/lyft/flytepropeller/pkg/compiler/errors"
	"github.com/lyft/flytepropeller/pkg/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Gets the compiled subgraph if this node contains an inline-declared coreWorkflow. Otherwise nil.
func buildNodeSpec(n *core.Node, tasks []*core.CompiledTask, errs errors.CompileErrors) (*v1alpha1.NodeSpec, bool) {
	if n == nil {
		errs.Collect(errors.NewValueRequiredErr("root", "node"))
		return nil, !errs.HasErrors()
	}

	if n.GetId() != common.StartNodeID && n.GetId() != common.EndNodeID &&
		n.GetTarget() == nil {

		errs.Collect(errors.NewValueRequiredErr(n.GetId(), "target"))
		return nil, !errs.HasErrors()
	}

	var task *core.TaskTemplate
	if n.GetTaskNode() != nil {
		taskID := n.GetTaskNode().GetReferenceId().String()
		// TODO: Use task index for quick lookup
		for _, t := range tasks {
			if t.Template.Id.String() == taskID {
				task = t.Template
				break
			}
		}

		if task == nil {
			errs.Collect(errors.NewTaskReferenceNotFoundErr(n.GetId(), taskID))
			return nil, !errs.HasErrors()
		}
	}

	res, err := utils.ToK8sResourceRequirements(getResources(task))
	if err != nil {
		errs.Collect(errors.NewWorkflowBuildError(err))
		return nil, false
	}

	timeout, err := computeDeadline(n)
	// TODO: Active deadline accounts for the retries and queueing delays. using active deadline = execution deadline.
	var activeDeadline *v1.Duration
	if timeout != nil {
		activeDeadline = &v1.Duration{Duration: 2 * timeout.Duration}
	}
	if err != nil {
		errs.Collect(errors.NewSyntaxError(n.GetId(), "node:metadata:timeout", nil))
		return nil, !errs.HasErrors()
	}

	var interruptible *bool
	if n.GetMetadata() != nil && n.GetMetadata().GetInterruptibleValue() != nil {
		interruptVal := n.GetMetadata().GetInterruptible()
		interruptible = &interruptVal
	}

	nodeSpec := &v1alpha1.NodeSpec{
		ID:                n.GetId(),
		RetryStrategy:     computeRetryStrategy(n, task),
		ExecutionDeadline: timeout,
		Resources:         res,
		OutputAliases:     toAliasValueArray(n.GetOutputAliases()),
		InputBindings:     toBindingValueArray(n.GetInputs()),
		ActiveDeadline:    activeDeadline,
		Interruptibe:      interruptible,
	}

	switch v := n.GetTarget().(type) {
	case *core.Node_TaskNode:
		nodeSpec.Kind = v1alpha1.NodeKindTask
		nodeSpec.TaskRef = refStr(n.GetTaskNode().GetReferenceId().String())
	case *core.Node_WorkflowNode:
		if n.GetWorkflowNode().Reference == nil {
			errs.Collect(errors.NewValueRequiredErr(n.GetId(), "WorkflowNode.Reference"))
			return nil, !errs.HasErrors()
		}

		switch n.GetWorkflowNode().Reference.(type) {
		case *core.WorkflowNode_LaunchplanRef:
			nodeSpec.Kind = v1alpha1.NodeKindWorkflow
			nodeSpec.WorkflowNode = &v1alpha1.WorkflowNodeSpec{
				LaunchPlanRefID: &v1alpha1.LaunchPlanRefID{Identifier: n.GetWorkflowNode().GetLaunchplanRef()},
			}
		case *core.WorkflowNode_SubWorkflowRef:
			nodeSpec.Kind = v1alpha1.NodeKindWorkflow
			if v.WorkflowNode.GetSubWorkflowRef() != nil {
				nodeSpec.WorkflowNode = &v1alpha1.WorkflowNodeSpec{
					SubWorkflowReference: refStr(v.WorkflowNode.GetSubWorkflowRef().String()),
				}
			} else if v.WorkflowNode.GetLaunchplanRef() != nil {
				nodeSpec.WorkflowNode = &v1alpha1.WorkflowNodeSpec{
					LaunchPlanRefID: &v1alpha1.LaunchPlanRefID{Identifier: n.GetWorkflowNode().GetLaunchplanRef()},
				}
			} else {
				errs.Collect(errors.NewValueRequiredErr(n.GetId(), "WorkflowNode.WorkflowTemplate"))
				return nil, !errs.HasErrors()
			}
		}
	case *core.Node_BranchNode:
		nodeSpec.Kind = v1alpha1.NodeKindBranch
		nodeSpec.BranchNode = buildBranchNodeSpec(n.GetBranchNode(), errs.NewScope())
	default:
		if n.GetId() == v1alpha1.StartNodeID {
			nodeSpec.Kind = v1alpha1.NodeKindStart
		} else if n.GetId() == v1alpha1.EndNodeID {
			nodeSpec.Kind = v1alpha1.NodeKindEnd
		}
	}

	return nodeSpec, !errs.HasErrors()
}

func buildIfBlockSpec(block *core.IfBlock, _ errors.CompileErrors) *v1alpha1.IfBlock {
	return &v1alpha1.IfBlock{
		Condition: v1alpha1.BooleanExpression{BooleanExpression: block.Condition},
		ThenNode:  refStr(block.ThenNode.Id),
	}
}

func buildBranchNodeSpec(branch *core.BranchNode, errs errors.CompileErrors) *v1alpha1.BranchNodeSpec {
	if branch == nil {
		return nil
	}

	res := &v1alpha1.BranchNodeSpec{
		If: *buildIfBlockSpec(branch.IfElse.Case, errs.NewScope()),
	}

	switch branch.IfElse.GetDefault().(type) {
	case *core.IfElseBlock_ElseNode:
		res.Else = refStr(branch.IfElse.GetElseNode().Id)
	case *core.IfElseBlock_Error:
		res.ElseFail = &v1alpha1.Error{Error: branch.IfElse.GetError()}
	}

	other := make([]*v1alpha1.IfBlock, 0, len(branch.IfElse.Other))
	for _, block := range branch.IfElse.Other {
		other = append(other, buildIfBlockSpec(block, errs.NewScope()))
	}

	res.ElseIf = other

	return res
}

func buildNodes(nodes []*core.Node, tasks []*core.CompiledTask, errs errors.CompileErrors) (map[common.NodeID]*v1alpha1.NodeSpec, bool) {
	res := make(map[common.NodeID]*v1alpha1.NodeSpec, len(nodes))
	for _, nodeBuidler := range nodes {
		n, ok := buildNodeSpec(nodeBuidler, tasks, errs.NewScope())
		if !ok {
			return nil, ok
		}

		if _, exists := res[n.ID]; exists {
			errs.Collect(errors.NewValueCollisionError(nodeBuidler.GetId(), "Id", n.ID))
		}

		res[n.ID] = n
	}

	return res, !errs.HasErrors()
}

func buildTasks(tasks []*core.CompiledTask, errs errors.CompileErrors) map[common.TaskIDKey]*v1alpha1.TaskSpec {
	res := make(map[common.TaskIDKey]*v1alpha1.TaskSpec, len(tasks))
	for _, flyteTask := range tasks {
		if flyteTask == nil {
			errs.Collect(errors.NewValueRequiredErr("root", "coreTask"))
		} else {
			taskID := flyteTask.Template.Id.String()
			if _, exists := res[taskID]; exists {
				errs.Collect(errors.NewValueCollisionError(taskID, "Id", taskID))
			}

			res[taskID] = &v1alpha1.TaskSpec{TaskTemplate: flyteTask.Template}
		}
	}

	return res
}
