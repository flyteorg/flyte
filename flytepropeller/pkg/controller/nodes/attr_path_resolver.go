package nodes

import (
	"context"
	"math"

	"github.com/shamaton/msgpack/v2"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/common"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

// resolveAttrPathInPromise resolves the literal with attribute path
// If the promise is chained with attributes (e.g. promise.a["b"][0]), then we need to resolve the promise
func resolveAttrPathInPromise(ctx context.Context, datastore *storage.DataStore, nodeID string, literal *core.Literal, bindAttrPath []*core.PromiseAttribute) (*core.Literal, error) {
	var currVal = literal
	var tmpVal *core.Literal
	var err error
	var exist bool
	index := 0

	for _, attr := range bindAttrPath {
		if currVal.GetOffloadedMetadata() != nil {
			// currVal will be overwritten with the contents of the offloaded data which contains the actual large literal.
			err := common.ReadLargeLiteral(ctx, datastore, currVal)
			if err != nil {
				return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "failed to read offloaded metadata for promise")
			}
		}
		switch currVal.GetValue().(type) {
		case *core.Literal_OffloadedMetadata:
			return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "unexpected offloaded metadata type")
		case *core.Literal_Map:
			tmpVal, exist = currVal.GetMap().GetLiterals()[attr.GetStringValue()]
			if !exist {
				return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "key [%v] does not exist in literal %v", attr.GetStringValue(), currVal.GetMap().GetLiterals())
			}
			currVal = tmpVal
			index++
		case *core.Literal_Collection:
			if int(attr.GetIntValue()) >= len(currVal.GetCollection().GetLiterals()) {
				return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "index [%v] is out of range of %v", attr.GetIntValue(), currVal.GetCollection().GetLiterals())
			}
			currVal = currVal.GetCollection().GetLiterals()[attr.GetIntValue()]
			index++
		default:
			break
		}
	}

	// resolve dataclass and Pydantic BaseModel
	if scalar := currVal.GetScalar(); scalar != nil {
		if binary := scalar.GetBinary(); binary != nil {
			currVal, err = resolveAttrPathInBinary(nodeID, binary, bindAttrPath[index:])
			if err != nil {
				return nil, err
			}
		} else if generic := scalar.GetGeneric(); generic != nil {
			currVal, err = resolveAttrPathInPbStruct(nodeID, generic, bindAttrPath[index:])
			if err != nil {
				return nil, err
			}
		}
	}

	return currVal, nil
}

// resolveAttrPathInPbStruct resolves the protobuf struct (e.g. dataclass) with attribute path
func resolveAttrPathInPbStruct(nodeID string, st *structpb.Struct, bindAttrPath []*core.PromiseAttribute) (*core.Literal, error) {

	var currVal any
	var tmpVal any
	var exist bool

	currVal = st.AsMap()

	// Turn the current value to a map so it can be resolved more easily
	for _, attr := range bindAttrPath {
		switch resolvedVal := currVal.(type) {
		// map
		case map[string]any:
			tmpVal, exist = resolvedVal[attr.GetStringValue()]
			if !exist {
				return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "key [%v] does not exist in literal %v", attr.GetStringValue(), currVal)
			}
			currVal = tmpVal
		// list
		case []any:
			index := int(attr.GetIntValue())
			if index < 0 || index >= len(resolvedVal) {
				return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID,
					"index [%v] is out of range of %v", index, resolvedVal)
			}
			currVal = resolvedVal[attr.GetIntValue()]
		}
	}

	// After resolve, convert the interface to literal
	literal, err := convertInterfaceToLiteral(nodeID, currVal)

	return literal, err
}

// resolveAttrPathInBinary resolves the binary idl object (e.g. dataclass, pydantic basemodel) with attribute path
func resolveAttrPathInBinary(nodeID string, binaryIDL *core.Binary, bindAttrPath []*core.PromiseAttribute) (*core.Literal, error) {

	binaryBytes := binaryIDL.GetValue()
	serializationFormat := binaryIDL.GetTag()

	var currVal any
	var tmpVal any
	var exist bool

	if serializationFormat == coreutils.MESSAGEPACK {
		err := msgpack.Unmarshal(binaryBytes, &currVal)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID,
			"Unsupported format '%v' found for literal value.\n"+
				"Please ensure the serialization format is supported.", serializationFormat)
	}

	// Turn the current value to a map, so it can be resolved more easily
	for _, attr := range bindAttrPath {
		switch resolvedVal := currVal.(type) {
		// map
		case map[any]any:
			// TODO: for cases like Dict[int, Any] in a dataclass, this will fail,
			// will support it in the future when flytekit supports it
			promise, ok := attr.GetValue().(*core.PromiseAttribute_StringValue)
			if !ok {
				return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID,
					"unexpected attribute type [%T] for value %v", attr.GetValue(), attr.GetValue())
			}
			key := promise.StringValue
			tmpVal, exist = resolvedVal[key]
			if !exist {
				return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "key [%v] does not exist in literal %v", attr.GetStringValue(), currVal)
			}
			currVal = tmpVal
		// list
		case []any:
			promise, ok := attr.GetValue().(*core.PromiseAttribute_IntValue)
			if !ok {
				return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID,
					"unexpected attribute type [%T] for value %v", attr.GetValue(), attr.GetValue())
			}
			index := int(promise.IntValue) // convert to int64
			if index < 0 || index >= len(resolvedVal) {
				return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID,
					"index [%v] is out of range of %v", index, resolvedVal)
			}
			currVal = resolvedVal[attr.GetIntValue()]
		default:
			return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "unexpected type [%T] for value %v", currVal, currVal)
		}
	}

	// In arrayNodeHandler, the resolved value should be a literal collection.
	// If the current value is already a collection, convert it to a literal collection.
	// This conversion does not affect how Flytekit processes the resolved value.
	if collection, ok := currVal.([]any); ok {
		literals := make([]*core.Literal, len(collection))
		for i, v := range collection {
			resolvedBinaryBytes, err := msgpack.Marshal(v)
			if err != nil {
				return nil, err
			}
			literals[i] = constructResolvedBinary(resolvedBinaryBytes, serializationFormat)
		}

		return &core.Literal{
			Value: &core.Literal_Collection{
				Collection: &core.LiteralCollection{
					Literals: literals,
				},
			},
		}, nil
	}

	// Check if the current value is a primitive type, and if it is convert that to a literal scalar
	if isPrimitiveType(currVal) {
		primitiveLiteral, err := convertInterfaceToLiteralScalar(nodeID, currVal)
		if err != nil {
			return nil, err
		}
		if primitiveLiteral != nil {
			return &core.Literal{
				Value: primitiveLiteral,
			}, nil
		}
	}

	// Marshal the current value to MessagePack bytes
	resolvedBinaryBytes, err := msgpack.Marshal(currVal)
	if err != nil {
		return nil, err
	}
	// Construct and return the binary-encoded literal
	return constructResolvedBinary(resolvedBinaryBytes, serializationFormat), nil
}

// isPrimitiveType checks if the value is a primitive type
func isPrimitiveType(value any) bool {
	switch value.(type) {
	case string, uint8, uint16, uint32, uint64, uint, int8, int16, int32, int64, int, float32, float64, bool:
		return true
	}
	return false
}

func constructResolvedBinary(resolvedBinaryBytes []byte, serializationFormat string) *core.Literal {
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Binary{
					Binary: &core.Binary{
						Value: resolvedBinaryBytes,
						Tag:   serializationFormat,
					},
				},
			},
		},
	}
}

// convertInterfaceToLiteral converts the protobuf struct (e.g. dataclass) to literal
func convertInterfaceToLiteral(nodeID string, obj interface{}) (*core.Literal, error) {

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
			literal, err := convertInterfaceToLiteral(nodeID, v)
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
		scalar, err := convertInterfaceToLiteralScalarWithNodeID(nodeID, obj)
		if err != nil {
			return nil, err
		}
		literal.Value = scalar
	}

	return literal, nil
}

// convertInterfaceToLiteralScalar converts a single value to a literal scalar
func convertInterfaceToLiteralScalar(nodeID string, obj interface{}) (*core.Literal_Scalar, error) {
	value := &core.Primitive{}

	switch obj := obj.(type) {
	case string:
		value.Value = &core.Primitive_StringValue{StringValue: obj}
	case uint8:
		value.Value = &core.Primitive_Integer{Integer: int64(obj)}
	case uint16:
		value.Value = &core.Primitive_Integer{Integer: int64(obj)}
	case uint32:
		value.Value = &core.Primitive_Integer{Integer: int64(obj)}
	case uint64:
		if obj > math.MaxInt64 {
			return nil, errors.Errorf(errors.InvalidPrimitiveType, nodeID, "uint64 value is too large to be converted to int64")
		}
		value.Value = &core.Primitive_Integer{Integer: int64(obj)} // #nosec G115
	case uint:
		if obj > math.MaxInt64 {
			return nil, errors.Errorf(errors.InvalidPrimitiveType, nodeID, "uint value is too large to be converted to int64")
		}
		value.Value = &core.Primitive_Integer{Integer: int64(obj)} // #nosec G115
	case int8:
		value.Value = &core.Primitive_Integer{Integer: int64(obj)}
	case int16:
		value.Value = &core.Primitive_Integer{Integer: int64(obj)}
	case int32:
		value.Value = &core.Primitive_Integer{Integer: int64(obj)}
	case int64:
		value.Value = &core.Primitive_Integer{Integer: obj}
	case int:
		value.Value = &core.Primitive_Integer{Integer: int64(obj)}
	case float32:
		value.Value = &core.Primitive_FloatValue{FloatValue: float64(obj)}
	case float64:
		value.Value = &core.Primitive_FloatValue{FloatValue: obj}
	case bool:
		value.Value = &core.Primitive_Boolean{Boolean: obj}
	default:
		return nil, errors.Errorf(errors.InvalidPrimitiveType, nodeID, "Failed to resolve interface to literal scalar")
	}

	return &core.Literal_Scalar{
		Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{
				Primitive: value,
			},
		},
	}, nil
}

func convertInterfaceToLiteralScalarWithNodeID(nodeID string, obj interface{}) (*core.Literal_Scalar, error) {
	literal, err := convertInterfaceToLiteralScalar(nodeID, obj)
	if err != nil {
		return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "Failed to resolve interface to literal scalar")
	}
	return literal, nil
}
