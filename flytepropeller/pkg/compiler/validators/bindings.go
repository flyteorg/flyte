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
	expectedType *flyte.LiteralType, errs errors.CompileErrors) (
	resolvedType *flyte.LiteralType, upstreamNodes []c.NodeID, ok bool) {

	switch binding.GetValue().(type) {
	case *flyte.BindingData_Collection:
		if expectedType.GetCollectionType() != nil {
			allNodeIds := make([]c.NodeID, 0, len(binding.GetMap().GetBindings()))
			var subType *flyte.LiteralType
			for _, v := range binding.GetCollection().GetBindings() {
				if resolvedType, nodeIds, ok := validateBinding(w, nodeID, nodeParam, v, expectedType.GetCollectionType(), errs.NewScope()); ok {
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

		errs.Collect(errors.NewMismatchingBindingsErr(nodeID, nodeParam, expectedType.String(), binding.GetCollection().String()))
	case *flyte.BindingData_Map:
		if expectedType.GetMapValueType() != nil {
			allNodeIds := make([]c.NodeID, 0, len(binding.GetMap().GetBindings()))
			var subType *flyte.LiteralType
			for _, v := range binding.GetMap().GetBindings() {
				if resolvedType, nodeIds, ok := validateBinding(w, nodeID, nodeParam, v, expectedType.GetMapValueType(), errs.NewScope()); ok {
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

		errs.Collect(errors.NewMismatchingBindingsErr(nodeID, nodeParam, expectedType.String(), binding.GetMap().String()))
	case *flyte.BindingData_Promise:
		if upNode, found := validateNodeID(w, binding.GetPromise().NodeId, errs.NewScope()); found {
			v, err := typing.ParseVarName(binding.GetPromise().GetVar())
			if err != nil {
				errs.Collect(errors.NewSyntaxError(nodeID, binding.GetPromise().GetVar(), err))
				return nil, nil, !errs.HasErrors()
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
					return param.GetType(), []c.NodeID{binding.GetPromise().NodeId}, true
				}

				errs.Collect(errors.NewMismatchingTypesErr(nodeID, binding.GetPromise().Var, sourceType.String(), expectedType.String()))
			}
		}

		errs.Collect(errors.NewParameterNotBoundErr(nodeID, nodeParam))
	case *flyte.BindingData_Scalar:
		literalType := literalTypeForScalar(binding.GetScalar())
		if literalType == nil {
			errs.Collect(errors.NewUnrecognizedValueErr(nodeID, reflect.TypeOf(binding.GetScalar().GetValue()).String()))
		} else if !AreTypesCastable(literalType, expectedType) {
			errs.Collect(errors.NewMismatchingTypesErr(nodeID, nodeParam, literalType.String(), expectedType.String()))
		}

		if expectedType.GetEnumType() != nil {
			v := binding.GetScalar().GetPrimitive().GetStringValue()
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
		errs.Collect(errors.NewUnrecognizedValueErr(nodeID, reflect.TypeOf(binding.GetValue()).String()))
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
				param.Type, errs.NewScope()); bindingOk {
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
		for paramName := range params.Variables {
			if !providedBindings.Has(paramName) {
				errs.Collect(errors.NewParameterNotBoundErr(node.GetId(), paramName))
			}
		}
	}

	return resolved, !errs.HasErrors()
}
