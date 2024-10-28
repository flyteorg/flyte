package validators

import (
	flyte "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	c "github.com/flyteorg/flyte/flytepropeller/pkg/compiler/common"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/errors"
	"github.com/flyteorg/flyte/flytestdlib/utils"
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

func validateInputVar(n c.NodeBuilder, paramName string, requireParamType bool, errs errors.CompileErrors) (param *flyte.Variable, ok bool) {
	if n.GetInterface() == nil {
		return nil, false
	}

	if param, ok = findVariableByName(n.GetInterface().GetInputs(), paramName); ok {
		return
	}

	if !requireParamType {
		if containsBindingByVariableName(n.GetInputs(), paramName) {
			return
		}
	}

	errs.Collect(errors.NewVariableNameNotFoundErr(n.GetId(), n.GetId(), paramName))
	return
}

func validateVarType(nodeID c.NodeID, paramName string, param *flyte.Variable,
	expectedType *flyte.LiteralType, errs errors.CompileErrors) (ok bool) {
	if param.GetType().String() != expectedType.String() {
		errs.Collect(errors.NewMismatchingTypesErr(nodeID, paramName, utils.LiteralTypeToStr(param.GetType()), utils.LiteralTypeToStr(expectedType)))
	}

	return !errs.HasErrors()
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
