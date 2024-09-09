package nodes

import (
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/errors"
)

// resolveAttrPathInPromise resolves the literal with attribute path
// If the promise is chained with attributes (e.g. promise.a["b"][0]), then we need to resolve the promise
func resolveAttrPathInPromise(nodeID string, literal *core.Literal, bindAttrPath []*core.PromiseAttribute) (*core.Literal, error) {
	var currVal *core.Literal = literal
	var tmpVal *core.Literal
	var exist bool
	count := 0

	for _, attr := range bindAttrPath {
		switch currVal.GetValue().(type) {
		case *core.Literal_Map:
			tmpVal, exist = currVal.GetMap().GetLiterals()[attr.GetStringValue()]
			if !exist {
				return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "key [%v] does not exist in literal %v", attr.GetStringValue(), currVal.GetMap().GetLiterals())
			}
			currVal = tmpVal
			count++
		case *core.Literal_Collection:
			if int(attr.GetIntValue()) >= len(currVal.GetCollection().GetLiterals()) {
				return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "index [%v] is out of range of %v", attr.GetIntValue(), currVal.GetCollection().GetLiterals())
			}
			currVal = currVal.GetCollection().GetLiterals()[attr.GetIntValue()]
			count++
		// scalar is always the leaf, so we can break here
		case *core.Literal_Scalar:
			break
		}
	}

	// resolve dataclass
	if scalar := currVal.GetScalar(); scalar != nil {
		// start from index "count"
		var err error

		if jsonIDL := scalar.GetJson(); jsonIDL != nil {
			serializationFormat := jsonIDL.GetSerializationFormat()
			currVal, err = resolveAttrPathInJSON(nodeID, jsonIDL.GetValue(), bindAttrPath[count:], serializationFormat)
		} else if generic := scalar.GetGeneric(); generic != nil {
			currVal, err = resolveAttrPathInPbStruct(nodeID, generic, bindAttrPath[count:])
		}
		if err != nil {
			return nil, err
		}

	}

	return currVal, nil
}

// resolveAttrPathInPbStruct resolves the protobuf struct (e.g. dataclass) with attribute path
func resolveAttrPathInPbStruct(nodeID string, st *structpb.Struct, bindAttrPath []*core.PromiseAttribute) (*core.Literal, error) {

	var currVal interface{}
	var tmpVal interface{}
	var exist bool

	currVal = st.AsMap()

	for _, attr := range bindAttrPath {
		switch resolvedVal := currVal.(type) {
		// map
		case map[string]interface{}:
			tmpVal, exist = resolvedVal[attr.GetStringValue()]
			if !exist {
				return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "key [%v] does not exist in literal %v", attr.GetStringValue(), currVal)
			}
			currVal = tmpVal
		// list
		case []interface{}:
			if int(attr.GetIntValue()) >= len(resolvedVal) {
				return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "index [%v] is out of range of %v", attr.GetIntValue(), currVal)
			}
			currVal = resolvedVal[attr.GetIntValue()]
		}
	}

	// After resolve, convert the interface to literal
	literal, err := convertStructToLiteral(nodeID, currVal)

	return literal, err
}

// resolveAttrPathInJSON resolves the msgpack bytes (e.g. dataclass) with attribute path
func resolveAttrPathInJSON(nodeID string, msgpackBytes []byte, bindAttrPath []*core.PromiseAttribute,
	serializationFormat string) (*core.Literal,
	error) {

	var currVal interface{}
	var tmpVal interface{}
	var exist bool

	if serializationFormat == "msgpack" {
		err := msgpack.Unmarshal(msgpackBytes, &currVal)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID,
			"Unsupported format '%v' found for literal value.\n"+
				"Please ensure the serialization format is supported.", serializationFormat)
	}

	// Turn the current value to a map so it can be resolved more easily
	for _, attr := range bindAttrPath {
		switch resolvedVal := currVal.(type) {
		// map
		case map[string]interface{}:
			tmpVal, exist = resolvedVal[attr.GetStringValue()]
			if !exist {
				return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "key [%v] does not exist in literal %v", attr.GetStringValue(), currVal)
			}
			currVal = tmpVal
		// list
		case []interface{}:
			if int(attr.GetIntValue()) >= len(resolvedVal) {
				return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "index [%v] is out of range of %v", attr.GetIntValue(), currVal)
			}
			currVal = resolvedVal[attr.GetIntValue()]
		}
	}

	// After resolve, convert the interface to literal
	literal, err := convertJSONToLiteral(nodeID, currVal, serializationFormat)

	return literal, err
}

// convertStructToLiteral converts the protobuf struct (e.g. dataclass) to literal
func convertStructToLiteral(nodeID string, obj interface{}) (*core.Literal, error) {

	literal := &core.Literal{}

	switch obj := obj.(type) {
	case map[string]interface{}:
		newSt, err := structpb.NewStruct(obj)
		if err != nil {
			return nil, err
		}
		literal.Value = &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Generic{
					Generic: newSt,
				},
			},
		}
	case []interface{}:
		literals := []*core.Literal{}
		for _, v := range obj {
			// recursively convert the interface to literal
			literal, err := convertStructToLiteral(nodeID, v)
			if err != nil {
				return nil, err
			}
			literals = append(literals, literal)
		}
		literal.Value = &core.Literal_Collection{
			Collection: &core.LiteralCollection{
				Literals: literals,
			},
		}
	case interface{}:
		scalar, err := convertInterfaceToLiteralScalar(nodeID, obj)
		if err != nil {
			return nil, err
		}
		literal.Value = scalar
	}

	return literal, nil
}

// convertJSONToLiteral converts the JSON (e.g. dataclass) to literal
func convertJSONToLiteral(nodeID string, obj interface{}, serializationFormat string) (*core.Literal, error) {
	literal := &core.Literal{}

	switch obj := obj.(type) {
	case map[string]interface{}:
		if serializationFormat == "msgpack" {
			msgpackBytes, err := msgpack.Marshal(obj)
			if err != nil {
				return nil, err
			}
			literal.Value = &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Json{
						Json: &core.Json{
							Value:               msgpackBytes,
							SerializationFormat: serializationFormat,
						},
					},
				},
			}
		} else {
			return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID,
				"Unsupported serialization format '%v' found for literal value.\n"+
					"Please ensure the serialization format is supported.", serializationFormat)
		}
	case []interface{}:
		literals := []*core.Literal{}
		for _, v := range obj {
			// recursively convert the interface to literal
			literal, err := convertJSONToLiteral(nodeID, v, serializationFormat)
			if err != nil {
				return nil, err
			}
			literals = append(literals, literal)
		}
		literal.Value = &core.Literal_Collection{
			Collection: &core.LiteralCollection{
				Literals: literals,
			},
		}
	case interface{}:
		scalar, err := convertInterfaceToLiteralScalar(nodeID, obj)
		if err != nil {
			return nil, err
		}
		literal.Value = scalar
	}

	return literal, nil
}

// TODO: We should support Blob Types in the future.
// convertInterfaceToLiteralScalar converts the a single value to a literal scalar
func convertInterfaceToLiteralScalar(nodeID string, obj interface{}) (*core.Literal_Scalar, error) {
	value := &core.Primitive{}

	switch obj := obj.(type) {
	case string:
		value.Value = &core.Primitive_StringValue{StringValue: obj}
	case int:
		value.Value = &core.Primitive_Integer{Integer: int64(obj)}
	case int8:
		value.Value = &core.Primitive_Integer{Integer: int64(obj)}
	case int64:
		value.Value = &core.Primitive_Integer{Integer: obj}
	case float64:
		value.Value = &core.Primitive_FloatValue{FloatValue: obj}
	case bool:
		value.Value = &core.Primitive_Boolean{Boolean: obj}
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
