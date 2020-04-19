package utils

import (
	"reflect"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
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
	return nil, errors.Errorf("Failed to convert to a known primitive type. Input Type [%v] not supported", reflect.TypeOf(v).String())
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

func MakeLiteral(v interface{}) (*core.Literal, error) {
	if v == nil {
		return &core.Literal{
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_NoneType{
						NoneType: nil,
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
	if typ != nil {
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
				// case core.SimpleType_ERROR:
				// case core.SimpleType_STRUCT:
			}
			return nil, errors.Errorf("Not yet implemented. Default creation is not yet implemented. ")

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
			// case *core.LiteralType_Schema:
		}
	}
	return nil, errors.Errorf("Failed to convert to a known Literal. Input Type [%v] not supported", typ.String())
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
