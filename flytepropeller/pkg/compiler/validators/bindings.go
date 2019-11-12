package validators

import (
	"reflect"

	"github.com/lyft/flytepropeller/pkg/compiler/typing"

	flyte "github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	c "github.com/lyft/flytepropeller/pkg/compiler/common"
	"github.com/lyft/flytepropeller/pkg/compiler/errors"
	"k8s.io/apimachinery/pkg/util/sets"
)

func validateBinding(w c.WorkflowBuilder, nodeID c.NodeID, nodeParam string, binding *flyte.BindingData, expectedType *flyte.LiteralType, errs errors.CompileErrors) (
	[]c.NodeID, bool) {

	switch binding.GetValue().(type) {
	case *flyte.BindingData_Collection:
		if expectedType.GetCollectionType() != nil {
			allNodeIds := make([]c.NodeID, 0, len(binding.GetMap().GetBindings()))
			for _, v := range binding.GetCollection().GetBindings() {
				if nodeIds, ok := validateBinding(w, nodeID, nodeParam, v, expectedType.GetCollectionType(), errs.NewScope()); ok {
					allNodeIds = append(allNodeIds, nodeIds...)
				}
			}
			return allNodeIds, !errs.HasErrors()
		}
		errs.Collect(errors.NewMismatchingBindingsErr(nodeID, nodeParam, expectedType.String(), binding.GetCollection().String()))
	case *flyte.BindingData_Map:
		if expectedType.GetMapValueType() != nil {
			allNodeIds := make([]c.NodeID, 0, len(binding.GetMap().GetBindings()))
			for _, v := range binding.GetMap().GetBindings() {
				if nodeIds, ok := validateBinding(w, nodeID, nodeParam, v, expectedType.GetMapValueType(), errs.NewScope()); ok {
					allNodeIds = append(allNodeIds, nodeIds...)
				}
			}
			return allNodeIds, !errs.HasErrors()
		}

		errs.Collect(errors.NewMismatchingBindingsErr(nodeID, nodeParam, expectedType.String(), binding.GetMap().String()))
	case *flyte.BindingData_Promise:
		if upNode, found := validateNodeID(w, binding.GetPromise().NodeId, errs.NewScope()); found {
			v, err := typing.ParseVarName(binding.GetPromise().GetVar())
			if err != nil {
				errs.Collect(errors.NewSyntaxError(nodeID, binding.GetPromise().GetVar(), err))
				return nil, !errs.HasErrors()
			}

			if param, paramFound := validateOutputVar(upNode, v.Name, errs.NewScope()); paramFound {
				sourceType := param.Type
				// If the variable has an index. We expect param to be a collection.
				if v.Index != nil {
					if cType := param.GetType().GetCollectionType(); cType == nil {
						errs.Collect(errors.NewMismatchingTypesErr(nodeID, binding.GetPromise().Var, param.Type.String(), expectedType.String()))
					} else {
						sourceType = cType
					}
				}

				if AreTypesCastable(sourceType, expectedType) {
					binding.GetPromise().NodeId = upNode.GetId()
					return []c.NodeID{binding.GetPromise().NodeId}, true
				}

				errs.Collect(errors.NewMismatchingTypesErr(nodeID, binding.GetPromise().Var, sourceType.String(), expectedType.String()))
			}
		}
	case *flyte.BindingData_Scalar:
		literalType := literalTypeForScalar(binding.GetScalar())
		if literalType == nil {
			errs.Collect(errors.NewUnrecognizedValueErr(nodeID, reflect.TypeOf(binding.GetScalar().GetValue()).String()))
		} else if !AreTypesCastable(literalType, expectedType) {
			errs.Collect(errors.NewMismatchingTypesErr(nodeID, nodeParam, literalType.String(), expectedType.String()))
		}

		return []c.NodeID{}, !errs.HasErrors()
	default:
		errs.Collect(errors.NewUnrecognizedValueErr(nodeID, reflect.TypeOf(binding.GetValue()).String()))
	}

	return nil, !errs.HasErrors()
}

func ValidateBindings(w c.WorkflowBuilder, node c.Node, bindings []*flyte.Binding, params *flyte.VariableMap,
	errs errors.CompileErrors) (ok bool) {

	providedBindings := sets.NewString()
	for _, binding := range bindings {
		if param, ok := findVariableByName(params, binding.GetVar()); !ok {
			errs.Collect(errors.NewVariableNameNotFoundErr(node.GetId(), node.GetId(), binding.GetVar()))
		} else if binding.GetBinding() == nil {
			errs.Collect(errors.NewValueRequiredErr(node.GetId(), "Binding"))
		} else if providedBindings.Has(binding.GetVar()) {
			errs.Collect(errors.NewParameterBoundMoreThanOnceErr(node.GetId(), binding.GetVar()))
		} else {
			providedBindings.Insert(binding.GetVar())
			if upstreamNodes, bindingOk := validateBinding(w, node.GetId(), binding.GetVar(), binding.GetBinding(), param.Type, errs.NewScope()); bindingOk {
				for _, upNode := range upstreamNodes {
					// Add implicit Edges
					w.AddExecutionEdge(upNode, node.GetId())
				}
			}
		}
	}

	// If we missed binding some params, add errors
	for paramName := range params.Variables {
		if !providedBindings.Has(paramName) {
			errs.Collect(errors.NewParameterNotBoundErr(node.GetId(), paramName))
		}
	}

	return !errs.HasErrors()
}
