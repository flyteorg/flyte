package validators

import (
	"reflect"

	"github.com/flyteorg/flytepropeller/pkg/compiler/typing"

	flyte "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	c "github.com/flyteorg/flytepropeller/pkg/compiler/common"
	"github.com/flyteorg/flytepropeller/pkg/compiler/errors"
	"k8s.io/apimachinery/pkg/util/sets"
)

func validateBinding(w c.WorkflowBuilder, nodeID c.NodeID, nodeParam string, binding *flyte.BindingData,
	expectedType *flyte.LiteralType, errs errors.CompileErrors, validateParamTypes bool) (
	resolvedType *flyte.LiteralType, upstreamNodes []c.NodeID, ok bool) {

	// Non-scalar bindings will fail to introspect the type through a union type so we resolve them beforehand
	switch binding.GetValue().(type) {
	case *flyte.BindingData_Scalar:
		// Goes through union-aware AreTypesCastable
		break
	case *flyte.BindingData_Promise:
		// Goes through union-aware AreTypesCastable
		break
	default:
		if expectedType.GetUnionType() != nil {
			var matchingType *flyte.LiteralType
			var resolvedType *flyte.LiteralType
			var nodeIds []c.NodeID
			var ok bool

			for _, t := range expectedType.GetUnionType().GetVariants() {
				resolvedType1, nodeIds1, ok1 := validateBinding(w, nodeID, nodeParam, binding, t, errors.NewCompileErrors(), validateParamTypes)
				if ok1 {
					if ok {
						errs.Collect(errors.NewAmbiguousBindingUnionValue(nodeID, nodeParam, expectedType.String(), binding.String(), matchingType.String(), t.String()))
						return nil, nil, !errs.HasErrors()
					}

					matchingType = t
					resolvedType, nodeIds, ok = resolvedType1, nodeIds1, ok1
				}
			}

			if ok {
				return resolvedType, nodeIds, ok
			}

			errs.Collect(errors.NewIncompatibleBindingUnionValue(nodeID, nodeParam, expectedType.String(), binding.String()))
			return nil, nil, !errs.HasErrors()
		}
	}

	switch val := binding.GetValue().(type) {
	case *flyte.BindingData_Collection:
		if val.Collection == nil {
			errs.Collect(errors.NewParameterNotBoundErr(nodeID, nodeParam))
			return nil, nil, !errs.HasErrors()
		}

		if expectedType.GetCollectionType() != nil {
			allNodeIds := make([]c.NodeID, 0, len(val.Collection.GetBindings()))
			var subType *flyte.LiteralType
			for _, v := range val.Collection.GetBindings() {
				if resolvedType, nodeIds, ok := validateBinding(w, nodeID, nodeParam, v, expectedType.GetCollectionType(), errs.NewScope(), validateParamTypes); ok {
					allNodeIds = append(allNodeIds, nodeIds...)
					subType = resolvedType
				}
			}

			return &flyte.LiteralType{
				Type: &flyte.LiteralType_CollectionType{
					CollectionType: subType,
				},
			}, allNodeIds, !errs.HasErrors()
		}

		errs.Collect(errors.NewMismatchingBindingsErr(nodeID, nodeParam, expectedType.String(), val.Collection.String()))
	case *flyte.BindingData_Map:
		if val.Map == nil {
			errs.Collect(errors.NewParameterNotBoundErr(nodeID, nodeParam))
			return nil, nil, !errs.HasErrors()
		}

		if expectedType.GetMapValueType() != nil {
			allNodeIds := make([]c.NodeID, 0, len(val.Map.GetBindings()))
			var subType *flyte.LiteralType
			for _, v := range val.Map.GetBindings() {
				if resolvedType, nodeIds, ok := validateBinding(w, nodeID, nodeParam, v, expectedType.GetMapValueType(), errs.NewScope(), validateParamTypes); ok {
					allNodeIds = append(allNodeIds, nodeIds...)
					subType = resolvedType
				}
			}

			return &flyte.LiteralType{
				Type: &flyte.LiteralType_MapValueType{
					MapValueType: subType,
				},
			}, allNodeIds, !errs.HasErrors()
		}

		errs.Collect(errors.NewMismatchingBindingsErr(nodeID, nodeParam, expectedType.String(), val.Map.String()))
	case *flyte.BindingData_Promise:
		if val.Promise == nil {
			errs.Collect(errors.NewParameterNotBoundErr(nodeID, nodeParam))
			return nil, nil, !errs.HasErrors()
		}

		if upNode, found := validateNodeID(w, val.Promise.NodeId, errs.NewScope()); found {
			v, err := typing.ParseVarName(val.Promise.GetVar())
			if err != nil {
				errs.Collect(errors.NewSyntaxError(nodeID, val.Promise.GetVar(), err))
				return nil, nil, !errs.HasErrors()
			}

			if param, paramFound := validateOutputVar(upNode, v.Name, errs.NewScope()); paramFound {
				sourceType := param.Type
				// If the variable has an index. We expect param to be a collection.
				if v.Index != nil {
					if cType := param.GetType().GetCollectionType(); cType == nil {
						errs.Collect(errors.NewMismatchingTypesErr(nodeID, val.Promise.Var, param.Type.String(), expectedType.String()))
					} else {
						sourceType = cType
					}
				}

				if !validateParamTypes || AreTypesCastable(sourceType, expectedType) {
					val.Promise.NodeId = upNode.GetId()
					return param.GetType(), []c.NodeID{val.Promise.NodeId}, true
				}

				errs.Collect(errors.NewMismatchingTypesErr(nodeID, val.Promise.Var, sourceType.String(), expectedType.String()))
			}
		}

		errs.Collect(errors.NewParameterNotBoundErr(nodeID, nodeParam))
	case *flyte.BindingData_Scalar:
		if val.Scalar == nil {
			errs.Collect(errors.NewParameterNotBoundErr(nodeID, nodeParam))
			return nil, nil, !errs.HasErrors()
		}

		literalType := literalTypeForScalar(val.Scalar)
		if literalType == nil {
			errs.Collect(errors.NewUnrecognizedValueErr(nodeID, reflect.TypeOf(val.Scalar.GetValue()).String()))
		} else if validateParamTypes && !AreTypesCastable(literalType, expectedType) {
			errs.Collect(errors.NewMismatchingTypesErr(nodeID, nodeParam, literalType.String(), expectedType.String()))
		}

		if expectedType.GetEnumType() != nil {
			v := val.Scalar.GetPrimitive().GetStringValue()
			// Let us assert that the bound value is a correct enum Value
			found := false
			for _, ev := range expectedType.GetEnumType().Values {
				if ev == v {
					found = true
					break
				}
			}
			if !found {
				errs.Collect(errors.NewIllegalEnumValueError(nodeID, nodeParam, v, expectedType.GetEnumType().Values))
			}
		}

		return literalType, []c.NodeID{}, !errs.HasErrors()
	default:
		bindingType := ""
		if val != nil {
			bindingType = reflect.TypeOf(val).String()
		}

		errs.Collect(errors.NewMismatchingBindingsErr(nodeID, nodeParam, expectedType.String(), bindingType))
	}

	return nil, nil, !errs.HasErrors()
}

func ValidateBindings(w c.WorkflowBuilder, node c.Node, bindings []*flyte.Binding, params *flyte.VariableMap,
	validateParamTypes bool, edgeDirection c.EdgeDirection, errs errors.CompileErrors) (resolved *flyte.VariableMap, ok bool) {

	resolved = &flyte.VariableMap{
		Variables: make(map[string]*flyte.Variable, len(bindings)),
	}

	providedBindings := sets.NewString()
	for _, binding := range bindings {
		if param, ok := findVariableByName(params, binding.GetVar()); !ok && validateParamTypes {
			errs.Collect(errors.NewVariableNameNotFoundErr(node.GetId(), node.GetId(), binding.GetVar()))
		} else if binding.GetBinding() == nil {
			errs.Collect(errors.NewValueRequiredErr(node.GetId(), "Binding"))
		} else if providedBindings.Has(binding.GetVar()) {
			errs.Collect(errors.NewParameterBoundMoreThanOnceErr(node.GetId(), binding.GetVar()))
		} else {
			if !validateParamTypes && param == nil {
				param = &flyte.Variable{
					Type: &flyte.LiteralType{
						Type: &flyte.LiteralType_Simple{},
					},
				}
			}

			providedBindings.Insert(binding.GetVar())
			if resolvedType, upstreamNodes, bindingOk := validateBinding(w, node.GetId(), binding.GetVar(), binding.GetBinding(),
				param.Type, errs.NewScope(), validateParamTypes); bindingOk {
				for _, upNode := range upstreamNodes {
					// Add implicit Edges
					switch edgeDirection {
					case c.EdgeDirectionBidirectional:
						w.AddExecutionEdge(upNode, node.GetId())
					case c.EdgeDirectionDownstream:
						w.AddDownstreamEdge(upNode, node.GetId())
					case c.EdgeDirectionUpstream:
						w.AddUpstreamEdge(upNode, node.GetId())
					}
				}

				resolved.Variables[binding.GetVar()] = &flyte.Variable{
					Type: resolvedType,
				}
			}
		}
	}

	// If we missed binding some params, add errors
	if params != nil {
		for paramName, Variable := range params.Variables {
			if !providedBindings.Has(paramName) && !IsOptionalType(*Variable) {
				errs.Collect(errors.NewParameterNotBoundErr(node.GetId(), paramName))
			}
		}
	}

	return resolved, !errs.HasErrors()
}

// IsOptionalType Return true if there is a None type in Union Type
func IsOptionalType(variable flyte.Variable) bool {
	if variable.Type.GetUnionType() == nil {
		return false
	}
	for _, variant := range variable.Type.GetUnionType().Variants {
		if flyte.SimpleType_NONE == variant.GetSimple() {
			return true
		}
	}
	return false
}
