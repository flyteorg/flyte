// This package contains validators for all elements of the workflow spec (node, task, branch, interface, bindings... etc.)
package validators

import (
	flyte "github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	c "github.com/lyft/flytepropeller/pkg/compiler/common"
	"github.com/lyft/flytepropeller/pkg/compiler/errors"
)

// Computes output parameters after applying all aliases -if any-.
func validateEffectiveOutputParameters(n c.NodeBuilder, errs errors.CompileErrors) (
	params *flyte.VariableMap, ok bool) {
	aliases := make(map[string]string, len(n.GetOutputAliases()))
	for _, alias := range n.GetOutputAliases() {
		if _, found := aliases[alias.Var]; found {
			errs.Collect(errors.NewDuplicateAliasErr(n.GetId(), alias.Alias))
		} else {
			aliases[alias.Var] = alias.Alias
		}
	}

	if n.GetInterface() != nil {
		params = &flyte.VariableMap{
			Variables: make(map[string]*flyte.Variable, len(n.GetInterface().GetOutputs().Variables)),
		}

		for paramName, param := range n.GetInterface().GetOutputs().Variables {
			if alias, found := aliases[paramName]; found {
				if newParam, paramOk := withVariableName(param); paramOk {
					params.Variables[alias] = newParam
				} else {
					errs.Collect(errors.NewParameterNotBoundErr(n.GetId(), alias))
				}

				delete(aliases, paramName)
			} else {
				params.Variables[paramName] = param
			}
		}

		// If there are still more aliases at this point, they point to non-existent variables.
		for _, alias := range aliases {
			errs.Collect(errors.NewParameterNotBoundErr(n.GetId(), alias))
		}
	}

	return params, !errs.HasErrors()
}

func validateBranchNode(w c.WorkflowBuilder, n c.NodeBuilder, errs errors.CompileErrors) bool {
	cases := make([]*flyte.IfBlock, 0, len(n.GetBranchNode().IfElse.Other)+1)
	cases = append(cases, n.GetBranchNode().IfElse.Case)
	cases = append(cases, n.GetBranchNode().IfElse.Other...)
	for _, block := range cases {
		// Validate condition
		ValidateBooleanExpression(n, block.Condition, errs.NewScope())

		if block.GetThenNode() == nil {
			errs.Collect(errors.NewBranchNodeNotSpecified(n.GetId()))
		} else {
			wrapperNode := w.NewNodeBuilder(block.GetThenNode())
			if ValidateNode(w, wrapperNode, errs.NewScope()) {
				// Add to the global nodes to be able to reference it later
				w.AddNode(wrapperNode, errs.NewScope())
				w.AddExecutionEdge(n.GetId(), block.GetThenNode().Id)
			}
		}
	}

	return !errs.HasErrors()
}

func validateNodeID(w c.WorkflowBuilder, nodeID string, errs errors.CompileErrors) (node c.NodeBuilder, ok bool) {
	if nodeID == "" {
		n, _ := w.GetNode(c.StartNodeID)
		return n, !errs.HasErrors()
	} else if node, ok = w.GetNode(nodeID); !ok {
		errs.Collect(errors.NewNodeReferenceNotFoundErr(nodeID, nodeID))
	}

	return node, !errs.HasErrors()
}

func ValidateNode(w c.WorkflowBuilder, n c.NodeBuilder, errs errors.CompileErrors) (ok bool) {
	if n.GetId() == "" {
		errs.Collect(errors.NewValueRequiredErr("<node>", "Id"))
	}

	if _, ifaceOk := ValidateUnderlyingInterface(w, n, errs.NewScope()); ifaceOk {
		// Validate node output aliases
		validateEffectiveOutputParameters(n, errs.NewScope())
	}

	// Validate branch node conditions and inner nodes.
	if n.GetBranchNode() != nil {
		validateBranchNode(w, n, errs.NewScope())
	} else if workflowN := n.GetWorkflowNode(); workflowN != nil && workflowN.GetSubWorkflowRef() != nil {
		workflowID := *workflowN.GetSubWorkflowRef()
		if wf, wfOk := w.GetSubWorkflow(workflowID); wfOk {
			// This might lead to redundant errors if the same subWorkflow is invoked from multiple nodes in the main
			// workflow.
			if subWorkflow, workflowOk := w.ValidateWorkflow(wf, errs.NewScope()); workflowOk {
				n.SetSubWorkflow(subWorkflow)
				w.UpdateSubWorkflow(workflowID, subWorkflow.GetCoreWorkflow())
			}
		} else {
			errs.Collect(errors.NewWorkflowReferenceNotFoundErr(n.GetId(), workflowN.GetSubWorkflowRef().String()))
		}
	} else if taskN := n.GetTaskNode(); taskN != nil && taskN.GetReferenceId() != nil {
		if task, found := w.GetTask(*taskN.GetReferenceId()); found {
			n.SetTask(task)
		} else if taskN.GetReferenceId() == nil {
			errs.Collect(errors.NewValueRequiredErr(n.GetId(), "TaskNode.ReferenceId"))
		} else {
			errs.Collect(errors.NewTaskReferenceNotFoundErr(n.GetId(), taskN.GetReferenceId().String()))
		}
	}

	return !errs.HasErrors()
}
