package k8s

import (
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/common"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/errors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/validators"
)

func validateInputs(nodeID common.NodeID, iface *core.TypedInterface, inputs core.InputData, errs errors.CompileErrors) (ok bool) {
	if iface == nil {
		errs.Collect(errors.NewValueRequiredErr(nodeID, "interface"))
		return false
	}

	if iface.Inputs == nil {
		errs.Collect(errors.NewValueRequiredErr(nodeID, "interface.InputsRef"))
		return false
	}

	varMap := make(map[string]*core.Variable, len(iface.Inputs.Variables))
	requiredInputsSet := sets.String{}
	for name, v := range iface.Inputs.Variables {
		varMap[name] = v
		requiredInputsSet.Insert(name)
	}

	boundInputsSet := sets.String{}
	for inputVar, inputVal := range inputs.GetInputs().GetLiterals() {
		v, exists := varMap[inputVar]
		if !exists {
			errs.Collect(errors.NewVariableNameNotFoundErr(nodeID, "", inputVar))
			continue
		}

		inputType := validators.LiteralTypeForLiteral(inputVal)
		if !validators.AreTypesCastable(inputType, v.Type) {
			errs.Collect(errors.NewMismatchingTypesErr(nodeID, inputVar, v.Type.String(), inputType.String()))
			continue
		}

		boundInputsSet.Insert(inputVar)
	}

	if diff := requiredInputsSet.Difference(boundInputsSet); len(diff) > 0 {
		for param := range diff {
			errs.Collect(errors.NewParameterNotBoundErr(nodeID, param))
		}
	}

	return !errs.HasErrors()
}
