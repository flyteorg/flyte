// Contains convenience methods for constructing core types.
package coreutils

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/pkg/errors"
)

func MakePrimitive(v interface{}) (*core.Primitive, error) {
	switch p := v.(type) {
	case int:
		return &core.Primitive{
			Value: &core.Primitive_Integer{
				Integer: int64(p),
			},
		}, nil
	case int64:
		return &core.Primitive{
			Value: &core.Primitive_Integer{
				Integer: p,
			},
		}, nil
	case float64:
		return &core.Primitive{
			Value: &core.Primitive_FloatValue{
				FloatValue: p,
			},
		}, nil
	case time.Time:
		t, err := ptypes.TimestampProto(p)
		if err != nil {
			return nil, err
		}
		return &core.Primitive{
			Value: &core.Primitive_Datetime{
				Datetime: t,
			},
		}, nil
	case time.Duration:
		d := ptypes.DurationProto(p)
		return &core.Primitive{
			Value: &core.Primitive_Duration{
				Duration: d,
			},
		}, nil
	case string:
		return &core.Primitive{
			Value: &core.Primitive_StringValue{
				StringValue: p,
			},
		}, nil
	case bool:
		return &core.Primitive{
			Value: &core.Primitive_Boolean{
				Boolean: p,
			},
		}, nil
	}
	return nil, fmt.Errorf("failed to convert to a known primitive type. Input Type [%v] not supported", reflect.TypeOf(v).String())
}

func MustMakePrimitive(v interface{}) *core.Primitive {
	f, err := MakePrimitive(v)
	if err != nil {
		panic(err)
	}
	return f
}

func MakePrimitiveLiteral(v interface{}) (*core.Literal, error) {
	p, err := MakePrimitive(v)
	if err != nil {
		return nil, err
	}
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: p,
				},
			},
		},
	}, nil
}

func MustMakePrimitiveLiteral(v interface{}) *core.Literal {
	p, err := MakePrimitiveLiteral(v)
	if err != nil {
		panic(err)
	}
	return p
}

func MakeLiteralForMap(v map[string]interface{}) (*core.Literal, error) {
	m, err := MakeLiteralMap(v)
	if err != nil {
		return nil, err
	}

	return &core.Literal{
		Value: &core.Literal_Map{
			Map: m,
		},
	}, nil
}

func MakeLiteralForCollection(v []interface{}) (*core.Literal, error) {
	literals := make([]*core.Literal, 0, len(v))
	for _, val := range v {
		l, err := MakeLiteral(val)
		if err != nil {
			return nil, err
		}

		literals = append(literals, l)
	}

	return &core.Literal{
		Value: &core.Literal_Collection{
			Collection: &core.LiteralCollection{
				Literals: literals,
			},
		},
	}, nil
}

func MakeBinaryLiteral(v []byte) *core.Literal {
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Binary{
					Binary: &core.Binary{
						Value: v,
					},
				},
			},
		},
	}
}

func MakeGenericLiteral(v *structpb.Struct) *core.Literal {
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Generic{
					Generic: v,
				},
			},
		}}
}

func MakeLiteral(v interface{}) (*core.Literal, error) {
	if v == nil {
		return &core.Literal{
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_NoneType{
						NoneType: &core.Void{},
					},
				},
			},
		}, nil
	}
	switch o := v.(type) {
	case *core.Literal:
		return o, nil
	case []interface{}:
		return MakeLiteralForCollection(o)
	case map[string]interface{}:
		return MakeLiteralForMap(o)
	case []byte:
		return MakeBinaryLiteral(v.([]byte)), nil
	case *structpb.Struct:
		return MakeGenericLiteral(v.(*structpb.Struct)), nil
	case *core.Error:
		return &core.Literal{
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Error{
						Error: v.(*core.Error),
					},
				},
			},
		}, nil
	default:
		return MakePrimitiveLiteral(o)
	}
}

func MustMakeDefaultLiteralForType(typ *core.LiteralType) *core.Literal {
	if res, err := MakeDefaultLiteralForType(typ); err != nil {
		panic(err)
	} else {
		return res
	}
}

func MakeDefaultLiteralForType(typ *core.LiteralType) (*core.Literal, error) {
	switch t := typ.GetType().(type) {
	case *core.LiteralType_Simple:
		switch t.Simple {
		case core.SimpleType_NONE:
			return MakeLiteral(nil)
		case core.SimpleType_INTEGER:
			return MakeLiteral(int(0))
		case core.SimpleType_FLOAT:
			return MakeLiteral(float64(0))
		case core.SimpleType_STRING:
			return MakeLiteral("")
		case core.SimpleType_BOOLEAN:
			return MakeLiteral(false)
		case core.SimpleType_DATETIME:
			return MakeLiteral(time.Now())
		case core.SimpleType_DURATION:
			return MakeLiteral(time.Second)
		case core.SimpleType_BINARY:
			return MakeLiteral([]byte{})
		case core.SimpleType_ERROR:
			return &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Error{
							Error: &core.Error{
								Message: "Default Error message",
							},
						},
					},
				},
			}, nil
		case core.SimpleType_STRUCT:
			return &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Generic{
							Generic: &structpb.Struct{},
						},
					},
				},
			}, nil
		}
		return nil, errors.Errorf("Not yet implemented. Default creation is not yet implemented for [%s] ", t.Simple.String())
	case *core.LiteralType_Blob:
		return &core.Literal{
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Blob{
						Blob: &core.Blob{
							Metadata: &core.BlobMetadata{
								Type: t.Blob,
							},
							Uri: "/tmp/somepath",
						},
					},
				},
			},
		}, nil
	case *core.LiteralType_CollectionType:
		single, err := MakeDefaultLiteralForType(t.CollectionType)
		if err != nil {
			return nil, err
		}

		return &core.Literal{
			Value: &core.Literal_Collection{
				Collection: &core.LiteralCollection{
					Literals: []*core.Literal{single},
				},
			},
		}, nil
	case *core.LiteralType_MapValueType:
		single, err := MakeDefaultLiteralForType(t.MapValueType)
		if err != nil {
			return nil, err
		}

		return &core.Literal{
			Value: &core.Literal_Map{
				Map: &core.LiteralMap{
					Literals: map[string]*core.Literal{
						"itemKey": single,
					},
				},
			},
		}, nil
	case *core.LiteralType_EnumType:
		return MakeLiteralForType(typ, nil)
	case *core.LiteralType_Schema:
		return MakeLiteralForType(typ, nil)
	case *core.LiteralType_UnionType:
		if len(t.UnionType.Variants) == 0 {
			return nil, errors.Errorf("Union type must have at least one variant")
		}
		// For union types, we just return the default for the first variant
		val, err := MakeDefaultLiteralForType(t.UnionType.Variants[0])
		if err != nil {
			return nil, errors.Errorf("Failed to create default literal for first union type variant [%v]", t.UnionType.Variants[0])
		}
		res := &core.Literal{
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Union{
						Union: &core.Union{
							Type:  t.UnionType.Variants[0],
							Value: val,
						},
					},
				},
			},
		}
		return res, nil
	}

	return nil, fmt.Errorf("failed to convert to a known Literal. Input Type [%v] not supported", typ.String())
}

func MakePrimitiveForType(t core.SimpleType, s string) (*core.Primitive, error) {
	p := &core.Primitive{}
	switch t {
	case core.SimpleType_INTEGER:
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse integer value")
		}
		p.Value = &core.Primitive_Integer{Integer: v}
	case core.SimpleType_FLOAT:
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse Float value")
		}
		p.Value = &core.Primitive_FloatValue{FloatValue: v}
	case core.SimpleType_BOOLEAN:
		v, err := strconv.ParseBool(s)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse Bool value")
		}
		p.Value = &core.Primitive_Boolean{Boolean: v}
	case core.SimpleType_STRING:
		p.Value = &core.Primitive_StringValue{StringValue: s}
	case core.SimpleType_DURATION:
		v, err := time.ParseDuration(s)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse Duration, valid formats: e.g. 300ms, -1.5h, 2h45m")
		}
		p.Value = &core.Primitive_Duration{Duration: ptypes.DurationProto(v)}
	case core.SimpleType_DATETIME:
		v, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse Datetime in RFC3339 format")
		}
		ts, err := ptypes.TimestampProto(v)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert datetime to proto")
		}
		p.Value = &core.Primitive_Datetime{Datetime: ts}
	default:
		return nil, fmt.Errorf("unsupported type %s", t.String())
	}
	return p, nil
}

func MakeLiteralForSimpleType(t core.SimpleType, s string) (*core.Literal, error) {
	s = strings.Trim(s, " \n\t")
	scalar := &core.Scalar{}
	switch t {
	case core.SimpleType_STRUCT:
		st := &structpb.Struct{}
		err := jsonpb.UnmarshalString(s, st)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load generic type as json.")
		}
		scalar.Value = &core.Scalar_Generic{
			Generic: st,
		}
	case core.SimpleType_BINARY:
		scalar.Value = &core.Scalar_Binary{
			Binary: &core.Binary{
				Value: []byte(s),
				// TODO Tag not supported at the moment
			},
		}
	case core.SimpleType_ERROR:
		scalar.Value = &core.Scalar_Error{
			Error: &core.Error{
				Message: s,
			},
		}
	case core.SimpleType_NONE:
		scalar.Value = &core.Scalar_NoneType{
			NoneType: &core.Void{},
		}
	default:
		p, err := MakePrimitiveForType(t, s)
		if err != nil {
			return nil, err
		}
		scalar.Value = &core.Scalar_Primitive{Primitive: p}
	}
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: scalar,
		},
	}, nil
}

func MustMakeLiteral(v interface{}) *core.Literal {
	p, err := MakeLiteral(v)
	if err != nil {
		panic(err)
	}

	return p
}

func MakeLiteralMap(v map[string]interface{}) (*core.LiteralMap, error) {

	literals := make(map[string]*core.Literal, len(v))
	for key, val := range v {
		l, err := MakeLiteral(val)
		if err != nil {
			return nil, err
		}

		literals[key] = l
	}

	return &core.LiteralMap{
		Literals: literals,
	}, nil
}

func MakeLiteralForSchema(path storage.DataReference, columns []*core.SchemaType_SchemaColumn) *core.Literal {
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Schema{
					Schema: &core.Schema{
						Uri: path.String(),
						Type: &core.SchemaType{
							Columns: columns,
						},
					},
				},
			},
		},
	}
}

func MakeLiteralForStructuredDataSet(path storage.DataReference, columns []*core.StructuredDatasetType_DatasetColumn, format string) *core.Literal {
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_StructuredDataset{
					StructuredDataset: &core.StructuredDataset{
						Uri: path.String(),
						Metadata: &core.StructuredDatasetMetadata{
							StructuredDatasetType: &core.StructuredDatasetType{
								Columns: columns,
								Format:  format,
							},
						},
					},
				},
			},
		},
	}
}

func MakeLiteralForBlob(path storage.DataReference, isDir bool, format string) *core.Literal {
	dim := core.BlobType_SINGLE
	if isDir {
		dim = core.BlobType_MULTIPART
	}
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Blob{
					Blob: &core.Blob{
						Uri: path.String(),
						Metadata: &core.BlobMetadata{
							Type: &core.BlobType{
								Dimensionality: dim,
								Format:         format,
							},
						},
					},
				},
			},
		},
	}
}

func MakeLiteralForType(t *core.LiteralType, v interface{}) (*core.Literal, error) {
	l := &core.Literal{}
	switch newT := t.Type.(type) {
	case *core.LiteralType_MapValueType:
		newV, ok := v.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("map value types can only be of type map[string]interface{}, but found %v", reflect.TypeOf(v))
		}

		literals := make(map[string]*core.Literal, len(newV))
		for key, val := range newV {
			lv, err := MakeLiteralForType(newT.MapValueType, val)
			if err != nil {
				return nil, err
			}
			literals[key] = lv
		}
		l.Value = &core.Literal_Map{
			Map: &core.LiteralMap{
				Literals: literals,
			},
		}

	case *core.LiteralType_CollectionType:
		newV, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("collection type expected but found %v", reflect.TypeOf(v))
		}

		literals := make([]*core.Literal, 0, len(newV))
		for _, val := range newV {
			lv, err := MakeLiteralForType(newT.CollectionType, val)
			if err != nil {
				return nil, err
			}
			literals = append(literals, lv)
		}
		l.Value = &core.Literal_Collection{
			Collection: &core.LiteralCollection{
				Literals: literals,
			},
		}

	case *core.LiteralType_Simple:
		strValue := fmt.Sprintf("%v", v)
		if v == nil {
			strValue = ""
		}
		// Note this is to support large integers which by default when passed from an unmarshalled json will be
		// converted to float64 and printed as exponential format by Sprintf.
		// eg : 8888888 get converted to 8.888888e+06 and which causes strconv.ParseInt to fail
		// Inorder to avoid this we explicitly add this check.
		if f, ok := v.(float64); ok && math.Trunc(f) == f {
			strValue = fmt.Sprintf("%.0f", math.Trunc(f))
		}
		if newT.Simple == core.SimpleType_STRUCT {
			if _, isValueStringType := v.(string); !isValueStringType {
				byteValue, err := json.Marshal(v)
				if err != nil {
					return nil, fmt.Errorf("unable to marshal to json string for struct value %v", v)
				}
				strValue = string(byteValue)
			}
		}
		lv, err := MakeLiteralForSimpleType(newT.Simple, strValue)
		if err != nil {
			return nil, err
		}
		return lv, nil

	case *core.LiteralType_Blob:
		isDir := newT.Blob.Dimensionality == core.BlobType_MULTIPART
		lv := MakeLiteralForBlob(storage.DataReference(fmt.Sprintf("%v", v)), isDir, newT.Blob.Format)
		return lv, nil

	case *core.LiteralType_Schema:
		lv := MakeLiteralForSchema(storage.DataReference(fmt.Sprintf("%v", v)), newT.Schema.Columns)
		return lv, nil
	case *core.LiteralType_StructuredDatasetType:
		lv := MakeLiteralForStructuredDataSet(storage.DataReference(fmt.Sprintf("%v", v)), newT.StructuredDatasetType.Columns, newT.StructuredDatasetType.Format)
		return lv, nil

	case *core.LiteralType_EnumType:
		var newV string
		if v == nil {
			if len(t.GetEnumType().Values) == 0 {
				return nil, fmt.Errorf("enum types need atleast one value")
			}
			newV = t.GetEnumType().Values[0]
		} else {
			var ok bool
			newV, ok = v.(string)
			if !ok {
				return nil, fmt.Errorf("cannot convert [%v] to enum representations, only string values are supported in enum literals", reflect.TypeOf(v))
			}
			found := false
			for _, val := range t.GetEnumType().GetValues() {
				if val == newV {
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("incorrect enum value [%s], supported values %+v", newV, t.GetEnumType().GetValues())
			}
		}
		return MakePrimitiveLiteral(newV)

	case *core.LiteralType_UnionType:
		// Try different types in the variants, return the first one matched
		found := false
		for _, subType := range newT.UnionType.Variants {
			lv, err := MakeLiteralForType(subType, v)
			if err == nil {
				l = &core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Union{
								Union: &core.Union{
									Value: lv,
									Type:  subType,
								},
							},
						},
					},
				}
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("incorrect union value [%s], supported values %+v", v, newT.UnionType.Variants)
		}

	default:
		return nil, fmt.Errorf("unsupported type %s", t.String())
	}

	return l, nil
}
