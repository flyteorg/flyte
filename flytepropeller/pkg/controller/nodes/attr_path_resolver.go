package nodes

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/errors"
	"google.golang.org/protobuf/types/known/structpb"
)

// resolveAttrPathInPromise resolves the literal with attribute path
// If the promise is chained with attributes (e.g. promise.a["b"][0]), then we need to resolve the promise
func resolveAttrPathInPromise(ctx context.Context, nodeID string, literal *core.Literal, bindAttrPath []*core.PromiseAtrribute) (*core.Literal, error) {
	var currVal *core.Literal = literal
	var tmpVal *core.Literal
	var err error
	var exist bool
	count := 0

	for _, attr := range bindAttrPath {
		switch currVal.GetValue().(type) {
		case *core.Literal_Map:
			tmpVal, exist = currVal.GetMap().GetLiterals()[attr.GetStringValue()]
			if exist == false {
				return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "key [%v] does not exist in literal %v", attr.GetStringValue(), currVal.GetMap().GetLiterals())
			}
			currVal = tmpVal
			count += 1
		case *core.Literal_Collection:
			if int(attr.GetIntValue()) >= len(currVal.GetCollection().GetLiterals()) {
				return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "index [%v] is out of range of %v", attr.GetIntValue(), currVal.GetCollection().GetLiterals())
			}
			currVal = currVal.GetCollection().GetLiterals()[attr.GetIntValue()]
			count += 1
		// scalar is always the leaf, so we can break here
		case *core.Literal_Scalar:
			break
		}
	}

	// resolve dataclass
	if currVal.GetScalar() != nil && currVal.GetScalar().GetGeneric() != nil {
		st := currVal.GetScalar().GetGeneric()
		// start from index "count"
		currVal, err = resolveAttrPathInPbStruct(ctx, nodeID, st, bindAttrPath[count:])
		if err != nil {
			return nil, err
		}
	}

	return currVal, nil
}

func resolveAttrPathInPbStruct(ctx context.Context, nodeID string, st *structpb.Struct, bindAttrPath []*core.PromiseAtrribute) (*core.Literal, error) {

	var currVal interface{}
	var tmpVal interface{}
	var exist bool

	currVal = st.AsMap()

	// Turn the current value to a map so it can be resolved more easily
	for _, attr := range bindAttrPath {
		switch currVal.(type) {
		// map
		case map[string]interface{}:
			tmpVal, exist = currVal.(map[string]interface{})[attr.GetStringValue()]
			if exist == false {
				return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "key [%v] does not exist in literal %v", attr.GetStringValue(), currVal)
			}
			currVal = tmpVal
		// list
		case []interface{}:
			if int(attr.GetIntValue()) >= len(currVal.([]interface{})) {
				return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "index [%v] is out of range of %v", attr.GetIntValue(), currVal)
			}
			currVal = currVal.([]interface{})[attr.GetIntValue()]
		}
	}

	literal := &core.Literal{}

	// Different types of current value will be wrapped in different ways
	switch currVal.(type) {
	// We don't need to wrap map[string]interface{} to LiteralMap, because DictTransformer in flytekit can also handle pb struct
	case map[string]interface{}:
		new_st, err := structpb.NewStruct(currVal.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		literal.Value = &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Generic{
					Generic: new_st,
				},
			},
		}
	case []interface{}:
		literals := []*core.Literal{}
		for _, v := range currVal.([]interface{}) {
			literalItem := &core.Literal{}
			scalar, err := resolveInterfaceToLiteralScalar(ctx, nodeID, v)
			if err != nil {
				return nil, err
			}
			literalItem.Value = scalar
			literals = append(literals, literalItem)
		}

		literal.Value = &core.Literal_Collection{
			Collection: &core.LiteralCollection{
				Literals: literals,
			},
		}
	case interface{}:
		scalar, err := resolveInterfaceToLiteralScalar(ctx, nodeID, currVal)
		if err != nil {
			return nil, err
		}
		literal.Value = scalar
	}

	return literal, nil
}

func resolveInterfaceToLiteralScalar(ctx context.Context, nodeID string, obj interface{}) (*core.Literal_Scalar, error) {
	value := &core.Primitive{}

	switch obj.(type) {
	case string:
		value.Value = &core.Primitive_StringValue{StringValue: obj.(string)}
	case int:
		value.Value = &core.Primitive_Integer{Integer: int64(obj.(int))}
	case float64:
		value.Value = &core.Primitive_FloatValue{FloatValue: obj.(float64)}
	case bool:
		value.Value = &core.Primitive_Boolean{Boolean: obj.(bool)}
	default:
		return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "Failed to resolve interface to literal scalar")
	}

	return &core.Literal_Scalar{
		Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{
				Primitive: value,
			},
		},
	}, nil
}
