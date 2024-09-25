package k8s

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/common"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/errors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/validators"
)

func validateInputs(nodeID common.NodeID, iface *core.TypedInterface, inputs core.LiteralMap, errs errors.CompileErrors) (ok bool) {
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
	for inputVar, inputVal := range inputs.Literals {
		v, exists := varMap[inputVar]
		if !exists {
			errs.Collect(errors.NewVariableNameNotFoundErr(nodeID, "", inputVar))
			continue
		}

		var inputType *core.LiteralType
		switch inputVal.GetValue().(type) {
		case *core.Literal_OffloadedMetadata:
			inputType = inputVal.GetOffloadedMetadata().GetInferredType()
		default:
			inputType = validators.LiteralTypeForLiteral(inputVal)
		}

		err := validators.ValidateLiteralType(inputType)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to validate literal type for [%s] with err: %s", inputVar, err)
			errs.Collect(errors.NewIDLTypeNotFoundErr(nodeID, errMsg))
			continue
		}

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
