// Contains convenience methods for constructing core types.
package coreutils

import (
	"fmt"
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
		//case *core.LiteralType_Schema:
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
