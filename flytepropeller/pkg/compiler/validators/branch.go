package validators

import (
	flyte "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	c "github.com/flyteorg/flytepropeller/pkg/compiler/common"
	"github.com/flyteorg/flytepropeller/pkg/compiler/errors"
	"k8s.io/apimachinery/pkg/util/sets"
)

func validateBranchInterface(w c.WorkflowBuilder, node c.NodeBuilder, errs errors.CompileErrors) (iface *flyte.TypedInterface, ok bool) {
	if node.GetInterface() != nil {
		return node.GetInterface(), true
	}

	if branch := node.GetBranchNode(); branch == nil {
		errs.Collect(errors.NewValueRequiredErr(node.GetId(), "Branch"))
		return nil, false
	}

	if ifBlock := node.GetBranchNode().IfElse; ifBlock == nil {
		errs.Collect(errors.NewValueRequiredErr(node.GetId(), "Branch.IfElse"))
		return nil, false
	}

	if ifCase := node.GetBranchNode().IfElse.Case; ifCase == nil {
		errs.Collect(errors.NewValueRequiredErr(node.GetId(), "Branch.IfElse.Case"))
		return nil, false
	}

	if thenNode := node.GetBranchNode().IfElse.Case.ThenNode; thenNode == nil {
		errs.Collect(errors.NewValueRequiredErr(node.GetId(), "Branch.IfElse.Case.ThenNode"))
		return nil, false
	}

	var outputs map[string]*flyte.Variable
	finalOutputParameterNames := sets.NewString()

	validateIfaceMatch := func(nodeId string, iface2 *flyte.TypedInterface, errsScope errors.CompileErrors) (match bool) {
		outputs2, outputs2Set := buildVariablesIndex(iface2.Outputs)
		// Validate that parameters that exist in both interfaces have compatible types.
		finalOutputParameterNames = finalOutputParameterNames.Intersection(outputs2Set)
		for paramName := range finalOutputParameterNames {
			if validateVarType(nodeId, paramName, outputs[paramName], outputs2[paramName].Type, errs.NewScope()) {
				validateVarType(nodeId, paramName, outputs2[paramName], outputs[paramName].Type, errs.NewScope())
			}
		}

		return !errsScope.HasErrors()
	}

	cases := make([]*flyte.Node, 0, len(node.GetBranchNode().IfElse.Other)+1)
	caseBlock := node.GetBranchNode().IfElse.Case
	cases = append(cases, caseBlock.ThenNode)

	otherCases := node.GetBranchNode().IfElse.Other
	for _, otherCase := range otherCases {
		if otherCase.ThenNode == nil {
			errs.Collect(errors.NewValueRequiredErr(node.GetId(), "IfElse.Case.ThenNode"))
			continue
		}

		cases = append(cases, otherCase.ThenNode)
	}

	if elseNode := node.GetBranchNode().IfElse.GetElseNode(); elseNode != nil {
		cases = append(cases, elseNode)
	}

	for _, block := range cases {
		n := w.GetOrCreateNodeBuilder(block)
		n.SetID(branchNodeIDFormatter(node.GetId(), n.GetId()))
		iface2, ok := ValidateUnderlyingInterface(w, n, errs.NewScope())
		if !ok {
			continue
		}

		// Clear out the Inputs. We do not care if the inputs of each of the underlying nodes
		// match. We will pull the inputs needed for the underlying branch node at runtime.
		iface2 = &flyte.TypedInterface{
			Inputs:  &flyte.VariableMap{Variables: map[string]*flyte.Variable{}},
			Outputs: iface2.Outputs,
		}

		if iface == nil {
			iface = iface2
			outputs, finalOutputParameterNames = buildVariablesIndex(iface.Outputs)
		} else {
			validateIfaceMatch(n.GetId(), iface2, errs.NewScope())
		}
	}

	// Discover inputs from bindings... these should include all the variables used in the conditional statements.
	// When we come to validate the conditions themselves, we will look up these variables and fail if a variable is used
	// in a condition but doesn't have a node input binding.
	inputVarsFromBindings, _ := ValidateBindings(w, node, node.GetInputs(), &flyte.VariableMap{Variables: map[string]*flyte.Variable{}},
		false, c.EdgeDirectionUpstream, errs.NewScope())

	if !errs.HasErrors() && iface != nil {
		iface = &flyte.TypedInterface{
			Inputs:  inputVarsFromBindings,
			Outputs: filterVariables(iface.Outputs, finalOutputParameterNames),
		}
	} else {
		iface = nil
	}

	return iface, !errs.HasErrors()
}
