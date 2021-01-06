package compiler

import (
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/compiler/common"
	"github.com/lyft/flytepropeller/pkg/compiler/errors"
)

type TaskIdentifier = common.Identifier
type LaunchPlanRefIdentifier = common.Identifier

// Represents the set of required resources for a given Workflow's execution. All of the resources should be loaded before
// hand and passed to the compiler.
type WorkflowExecutionRequirements struct {
	taskIds       []TaskIdentifier
	launchPlanIds []LaunchPlanRefIdentifier
}

// Gets a slice of required Task ids to load.
func (g WorkflowExecutionRequirements) GetRequiredTaskIds() []TaskIdentifier {
	return g.taskIds
}

// Gets a slice of required Workflow ids to load.
func (g WorkflowExecutionRequirements) GetRequiredLaunchPlanIds() []LaunchPlanRefIdentifier {
	return g.launchPlanIds
}

// Computes requirements for a given Workflow.
func GetRequirements(fg *core.WorkflowTemplate, subWfs []*core.WorkflowTemplate) (reqs WorkflowExecutionRequirements, err error) {
	errs := errors.NewCompileErrors()
	compiledSubWfs := toCompiledWorkflows(subWfs...)

	index, ok := common.NewWorkflowIndex(compiledSubWfs, errs)

	if ok {
		return getRequirements(fg, index, true, errs), nil
	}

	return WorkflowExecutionRequirements{}, errs
}

func getRequirements(fg *core.WorkflowTemplate, subWfs common.WorkflowIndex, followSubworkflows bool,
	errs errors.CompileErrors) (reqs WorkflowExecutionRequirements) {

	taskIds := common.NewIdentifierSet()
	launchPlanIds := common.NewIdentifierSet()
	updateWorkflowRequirements(fg, subWfs, taskIds, launchPlanIds, followSubworkflows, errs)

	reqs.taskIds = taskIds.List()
	reqs.launchPlanIds = launchPlanIds.List()

	return
}

// Augments taskIds and launchPlanIds with referenced tasks/workflows within coreWorkflow nodes
func updateWorkflowRequirements(workflow *core.WorkflowTemplate, subWfs common.WorkflowIndex,
	taskIds, workflowIds common.IdentifierSet, followSubworkflows bool, errs errors.CompileErrors) {

	for _, node := range workflow.Nodes {
		updateNodeRequirements(node, subWfs, taskIds, workflowIds, followSubworkflows, errs)
	}
}

func updateNodeRequirements(node *flyteNode, subWfs common.WorkflowIndex, taskIds, workflowIds common.IdentifierSet,
	followSubworkflows bool, errs errors.CompileErrors) {

	if taskN := node.GetTaskNode(); taskN != nil && taskN.GetReferenceId() != nil {
		taskIds.Insert(*taskN.GetReferenceId())
	} else if workflowNode := node.GetWorkflowNode(); workflowNode != nil {
		if workflowNode.GetLaunchplanRef() != nil {
			workflowIds.Insert(*workflowNode.GetLaunchplanRef())
		} else if workflowNode.GetSubWorkflowRef() != nil && followSubworkflows {
			if subWf, found := subWfs[workflowNode.GetSubWorkflowRef().String()]; !found {
				errs.Collect(errors.NewWorkflowReferenceNotFoundErr(node.Id, workflowNode.GetSubWorkflowRef().String()))
			} else {
				updateWorkflowRequirements(subWf.Template, subWfs, taskIds, workflowIds, followSubworkflows, errs)
			}
		}
	} else if branchN := node.GetBranchNode(); branchN != nil {
		updateNodeRequirements(branchN.IfElse.Case.ThenNode, subWfs, taskIds, workflowIds, followSubworkflows, errs)
		for _, otherCase := range branchN.IfElse.Other {
			updateNodeRequirements(otherCase.ThenNode, subWfs, taskIds, workflowIds, followSubworkflows, errs)
		}

		if elseNode := branchN.IfElse.GetElseNode(); elseNode != nil {
			updateNodeRequirements(elseNode, subWfs, taskIds, workflowIds, followSubworkflows, errs)
		}
	}
}
