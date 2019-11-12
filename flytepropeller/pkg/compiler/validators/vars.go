package validators

import (
	flyte "github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	c "github.com/lyft/flytepropeller/pkg/compiler/common"
	"github.com/lyft/flytepropeller/pkg/compiler/errors"
	"k8s.io/apimachinery/pkg/util/sets"
)

func validateOutputVar(n c.NodeBuilder, paramName string, errs errors.CompileErrors) (
	param *flyte.Variable, ok bool) {
	if outputs, effectiveOk := validateEffectiveOutputParameters(n, errs.NewScope()); effectiveOk {
		var paramFound bool
		if param, paramFound = findVariableByName(outputs, paramName); !paramFound {
			errs.Collect(errors.NewVariableNameNotFoundErr(n.GetId(), n.GetId(), paramName))
		}
	}

	return param, !errs.HasErrors()
}

func validateInputVar(n c.NodeBuilder, paramName string, errs errors.CompileErrors) (param *flyte.Variable, ok bool) {
	if n.GetInterface() == nil {
		return nil, false
	}

	if param, ok = findVariableByName(n.GetInterface().GetInputs(), paramName); !ok {
		errs.Collect(errors.NewVariableNameNotFoundErr(n.GetId(), n.GetId(), paramName))
	}

	return
}

func validateVarType(nodeID c.NodeID, paramName string, param *flyte.Variable,
	expectedType *flyte.LiteralType, errs errors.CompileErrors) (ok bool) {
	if param.GetType().String() != expectedType.String() {
		errs.Collect(errors.NewMismatchingTypesErr(nodeID, paramName, param.GetType().String(), expectedType.String()))
	}

	return !errs.HasErrors()
}

func validateVarsSetMatch(nodeID string, params1, params2 map[string]*flyte.Variable,
	params1Set, params2Set sets.String, errs errors.CompileErrors) {
	// Validate that parameters that exist in both interfaces have compatible types.
	inBoth := params1Set.Intersection(params2Set)
	for paramName := range inBoth {
		if validateVarType(nodeID, paramName, params1[paramName], params2[paramName].Type, errs.NewScope()) {
			validateVarType(nodeID, paramName, params2[paramName], params1[paramName].Type, errs.NewScope())
		}
	}

	// All remaining params on either sides indicate errors
	inLeftSide := params1Set.Intersection(params2Set)
	for range inLeftSide {
		errs.Collect(errors.NewMismatchingInterfacesErr(nodeID, nodeID))
	}

	inRightSide := params2Set.Intersection(params1Set)
	for range inRightSide {
		errs.Collect(errors.NewMismatchingInterfacesErr(nodeID, nodeID))
	}
}

// Validate parameters have their required attributes set
func validateVariables(nodeID c.NodeID, params *flyte.VariableMap, errs errors.CompileErrors) {
	for paramName, param := range params.Variables {
		if len(paramName) == 0 {
			errs.Collect(errors.NewValueRequiredErr(nodeID, "paramName"))
		}

		if param.Type == nil {
			errs.Collect(errors.NewValueRequiredErr(nodeID, "param.Type"))
		}
	}
}
