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

	var inputs map[string]*flyte.Variable
	var outputs map[string]*flyte.Variable
	inputsSet := sets.NewString()
	outputsSet := sets.NewString()

	validateIfaceMatch := func(nodeId string, iface2 *flyte.TypedInterface, errsScope errors.CompileErrors) (match bool) {
		inputs2, inputs2Set := buildVariablesIndex(iface2.Inputs)
		validateVarsSetMatch(nodeId, inputs, inputs2, inputsSet, inputs2Set, errsScope.NewScope())
		finalInputParameterNames = finalInputParameterNames.Intersection(inputs2Set)

		outputs2, outputs2Set := buildVariablesIndex(iface2.Outputs)
		validateVarsSetMatch(nodeId, outputs, outputs2, outputsSet, outputs2Set, errsScope.NewScope())
		finalOutputParameterNames = finalOutputParameterNames.Intersection(outputs2Set)

		return !errsScope.HasErrors()
	}

	cases := make([]*flyte.IfBlock, 0, len(node.GetBranchNode().IfElse.Other)+1)
	caseBlock := node.GetBranchNode().IfElse.Case
	cases = append(cases, caseBlock)

	if otherCases := node.GetBranchNode().IfElse.Other; otherCases != nil {
		cases = append(cases, otherCases...)
	}

	for _, block := range cases {
		if block.ThenNode == nil {
			errs.Collect(errors.NewValueRequiredErr(node.GetId(), "IfElse.Case.ThenNode"))
			continue
		}

		n := w.NewNodeBuilder(block.ThenNode)
		if iface == nil {
			// if this is the first node to validate, just assume all other nodes will match the interface
			if iface, ok = ValidateUnderlyingInterface(w, n, errs.NewScope()); ok {
				inputs, inputsSet = buildVariablesIndex(iface.Inputs)
				finalInputParameterNames = finalInputParameterNames.Union(inputsSet)

				outputs, outputsSet = buildVariablesIndex(iface.Outputs)
				finalOutputParameterNames = finalOutputParameterNames.Union(outputsSet)
			}
		} else {
			if iface2, ok2 := ValidateUnderlyingInterface(w, n, errs.NewScope()); ok2 {
				validateIfaceMatch(n.GetId(), iface2, errs.NewScope())
			}
		}
	}

	if !errs.HasErrors() {
		iface = &flyte.TypedInterface{
			Inputs:  filterVariables(iface.Inputs, finalInputParameterNames),
			Outputs: filterVariables(iface.Outputs, finalOutputParameterNames),
		}
	} else {
		iface = nil
	}

	return iface, !errs.HasErrors()
}
