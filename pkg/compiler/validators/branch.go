package validators

import (
	flyte "github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	c "github.com/lyft/flytepropeller/pkg/compiler/common"
	"github.com/lyft/flytepropeller/pkg/compiler/errors"
	"k8s.io/apimachinery/pkg/util/sets"
)

func validateBranchInterface(w c.WorkflowBuilder, node c.NodeBuilder, errs errors.CompileErrors) (iface *flyte.TypedInterface, ok bool) {
	if branch := node.GetBranchNode(); branch == nil {
		errs.Collect(errors.NewValueRequiredErr(node.GetId(), "Branch"))
		return
	}

	if ifBlock := node.GetBranchNode().IfElse; ifBlock == nil {
		errs.Collect(errors.NewValueRequiredErr(node.GetId(), "Branch.IfElse"))
		return
	}

	if ifCase := node.GetBranchNode().IfElse.Case; ifCase == nil {
		errs.Collect(errors.NewValueRequiredErr(node.GetId(), "Branch.IfElse.Case"))
		return
	}

	if thenNode := node.GetBranchNode().IfElse.Case.ThenNode; thenNode == nil {
		errs.Collect(errors.NewValueRequiredErr(node.GetId(), "Branch.IfElse.Case.ThenNode"))
		return
	}

	finalInputParameterNames := sets.NewString()
	finalOutputParameterNames := sets.NewString()

	var outputs map[string]*flyte.Variable
	outputsSet := sets.NewString()

	validateIfaceMatch := func(nodeId string, iface2 *flyte.TypedInterface, errsScope errors.CompileErrors) (match bool) {
		outputs2, outputs2Set := buildVariablesIndex(iface2.Outputs)
		validateVarsSetMatch(nodeId, outputs, outputs2, outputsSet, outputs2Set, errsScope.NewScope())
		finalOutputParameterNames = finalOutputParameterNames.Intersection(outputs2Set)

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
		n := w.NewNodeBuilder(block)
		if iface == nil {
			// if this is the first node to validate, just assume all other nodes will match the interface
			if iface, ok = ValidateUnderlyingInterface(w, n, errs.NewScope()); ok {
				// Clear out the Inputs. We do not care if the inputs of each of the underlying nodes
				// match. We will pull the inputs needed for the underlying branch node at runtime.
				iface = &flyte.TypedInterface{
					Inputs:  &flyte.VariableMap{Variables: map[string]*flyte.Variable{}},
					Outputs: iface.Outputs,
				}

				outputs, outputsSet = buildVariablesIndex(iface.Outputs)
				finalOutputParameterNames = finalOutputParameterNames.Union(outputsSet)
			}
		} else {
			if iface2, ok2 := ValidateUnderlyingInterface(w, n, errs.NewScope()); ok2 {
				iface2 = &flyte.TypedInterface{
					Inputs:  &flyte.VariableMap{Variables: map[string]*flyte.Variable{}},
					Outputs: iface2.Outputs,
				}

				validateIfaceMatch(n.GetId(), iface2, errs.NewScope())
			}
		}
	}

	if !errs.HasErrors() && iface != nil {
		iface = &flyte.TypedInterface{
			Inputs:  filterVariables(iface.Inputs, finalInputParameterNames),
			Outputs: filterVariables(iface.Outputs, finalOutputParameterNames),
		}
	} else {
		iface = nil
	}

	return iface, !errs.HasErrors()
}
