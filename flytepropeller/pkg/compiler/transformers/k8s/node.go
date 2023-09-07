package k8s

import (
	"strings"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/compiler/common"
	"github.com/flyteorg/flytepropeller/pkg/compiler/errors"
	"github.com/go-test/deep"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Gets the compiled subgraph if this node contains an inline-declared coreWorkflow. Otherwise nil.
func buildNodeSpec(n *core.Node, tasks []*core.CompiledTask, errs errors.CompileErrors) ([]*v1alpha1.NodeSpec, bool) {
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
	var resources *core.Resources
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

		if n.GetTaskNode().Overrides != nil && n.GetTaskNode().Overrides.Resources != nil {
			resources = n.GetTaskNode().Overrides.Resources
		}
	}

	res, err := flytek8s.ToK8sResourceRequirements(resources)
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
	var name string
	if n.GetMetadata() != nil {
		if n.GetMetadata().GetInterruptibleValue() != nil {
			interruptVal := n.GetMetadata().GetInterruptible()
			interruptible = &interruptVal
		}
		name = n.GetMetadata().Name
	}

	nodeSpec := &v1alpha1.NodeSpec{
		ID:                n.GetId(),
		Name:              name,
		RetryStrategy:     computeRetryStrategy(n, task),
		ExecutionDeadline: timeout,
		Resources:         res,
		OutputAliases:     toAliasValueArray(n.GetOutputAliases()),
		InputBindings:     toBindingValueArray(n.GetInputs()),
		ActiveDeadline:    activeDeadline,
		Interruptible:     interruptible,
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
		b, ns := buildBranchNodeSpec(n.GetBranchNode(), tasks, errs.NewScope())
		nodeSpec.BranchNode = b
		// The reason why we create a separate list and then append the passed list to it is to maintain the actualNode
		// as the first element in the list. That way list[0] will always be the first node
		actualNode := []*v1alpha1.NodeSpec{nodeSpec}
		return append(actualNode, ns...), !errs.HasErrors()
	case *core.Node_GateNode:
		nodeSpec.Kind = v1alpha1.NodeKindGate
		gateNode := n.GetGateNode()
		switch gateNode.Condition.(type) {
		case *core.GateNode_Approve:
			nodeSpec.GateNode = &v1alpha1.GateNodeSpec{
				Kind: v1alpha1.ConditionKindApprove,
				Approve: &v1alpha1.ApproveCondition{
					ApproveCondition: gateNode.GetApprove(),
				},
			}
		case *core.GateNode_Signal:
			nodeSpec.GateNode = &v1alpha1.GateNodeSpec{
				Kind: v1alpha1.ConditionKindSignal,
				Signal: &v1alpha1.SignalCondition{
					SignalCondition: gateNode.GetSignal(),
				},
			}
		case *core.GateNode_Sleep:
			nodeSpec.GateNode = &v1alpha1.GateNodeSpec{
				Kind: v1alpha1.ConditionKindSleep,
				Sleep: &v1alpha1.SleepCondition{
					SleepCondition: gateNode.GetSleep(),
				},
			}
		}
	case *core.Node_ArrayNode:
		arrayNode := n.GetArrayNode()

		// build subNodeSpecs
		subNodeSpecs, ok := buildNodeSpec(arrayNode.Node, tasks, errs)
		if !ok {
			return nil, ok
		}

		// build ArrayNode
		nodeSpec.Kind = v1alpha1.NodeKindArray
		nodeSpec.ArrayNode = &v1alpha1.ArrayNodeSpec{
			SubNodeSpec: subNodeSpecs[0],
			Parallelism: arrayNode.Parallelism,
		}

		switch successCriteria := arrayNode.SuccessCriteria.(type) {
		case *core.ArrayNode_MinSuccesses:
			nodeSpec.ArrayNode.MinSuccesses = &successCriteria.MinSuccesses
		case *core.ArrayNode_MinSuccessRatio:
			nodeSpec.ArrayNode.MinSuccessRatio = &successCriteria.MinSuccessRatio
		}
	default:
		if n.GetId() == v1alpha1.StartNodeID {
			nodeSpec.Kind = v1alpha1.NodeKindStart
		} else if n.GetId() == v1alpha1.EndNodeID {
			nodeSpec.Kind = v1alpha1.NodeKindEnd
		}
	}

	return []*v1alpha1.NodeSpec{nodeSpec}, !errs.HasErrors()
}

func buildIfBlockSpec(block *core.IfBlock, tasks []*core.CompiledTask, errs errors.CompileErrors) (*v1alpha1.IfBlock, []*v1alpha1.NodeSpec) {
	nodeSpecs, ok := buildNodeSpec(block.ThenNode, tasks, errs)
	if !ok {
		return nil, []*v1alpha1.NodeSpec{}
	}
	return &v1alpha1.IfBlock{
		Condition: v1alpha1.BooleanExpression{BooleanExpression: block.Condition},
		ThenNode:  refStr(block.ThenNode.Id),
	}, nodeSpecs
}

func buildBranchNodeSpec(branch *core.BranchNode, tasks []*core.CompiledTask, errs errors.CompileErrors) (*v1alpha1.BranchNodeSpec, []*v1alpha1.NodeSpec) {
	if branch == nil {
		return nil, []*v1alpha1.NodeSpec{}
	}

	var childNodes []*v1alpha1.NodeSpec

	branchNode, nodeSpecs := buildIfBlockSpec(branch.IfElse.Case, tasks, errs.NewScope())
	res := &v1alpha1.BranchNodeSpec{
		If: *branchNode,
	}
	childNodes = append(childNodes, nodeSpecs...)

	switch branch.IfElse.GetDefault().(type) {
	case *core.IfElseBlock_ElseNode:
		ns, ok := buildNodeSpec(branch.IfElse.GetElseNode(), tasks, errs)
		if !ok {
			return nil, []*v1alpha1.NodeSpec{}
		}
		childNodes = append(childNodes, ns...)
		res.Else = refStr(branch.IfElse.GetElseNode().Id)
	case *core.IfElseBlock_Error:
		res.ElseFail = &v1alpha1.Error{Error: branch.IfElse.GetError()}
	}

	other := make([]*v1alpha1.IfBlock, 0, len(branch.IfElse.Other))
	for _, block := range branch.IfElse.Other {
		b, ns := buildIfBlockSpec(block, tasks, errs.NewScope())
		other = append(other, b)
		childNodes = append(childNodes, ns...)
	}

	res.ElseIf = other

	return res, childNodes
}

func buildNodes(nodes []*core.Node, tasks []*core.CompiledTask, errs errors.CompileErrors) (map[common.NodeID]*v1alpha1.NodeSpec, bool) {
	res := make(map[common.NodeID]*v1alpha1.NodeSpec, len(nodes))
	for _, nodeBuilder := range nodes {
		nodeSpecs, ok := buildNodeSpec(nodeBuilder, tasks, errs.NewScope())
		if !ok {
			return nil, ok
		}

		for _, nref := range nodeSpecs {
			n := nref
			if existingNode, exists := res[n.ID]; exists {
				if diff := deep.Equal(existingNode, n); diff != nil {
					errs.Collect(errors.NewValueCollisionError(nodeBuilder.GetId(), strings.Join(diff, "\r\n"), n.ID))
				}
			}

			res[n.ID] = n
		}
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
