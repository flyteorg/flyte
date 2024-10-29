package common

import (
	"google.golang.org/protobuf/types/known/structpb"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestLiteralTypeToStr(t *testing.T) {
	dataclassType := &core.LiteralType{
		Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRUCT},
		Structure: &core.TypeStructure{
			DataclassType: map[string]*core.LiteralType{
				"a": {
					Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
				},
			},
		},
		Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{
			"key": {Kind: &structpb.Value_StringValue{StringValue: "a"}},
		}},
	}
	assert.Equal(t, LiteralTypeToStr(nil), "None")
	assert.Equal(t, LiteralTypeToStr(dataclassType), "simple: STRUCT structure{dataclass_type:{key:a value:{simple:INTEGER}}")
	assert.NotEqual(t, LiteralTypeToStr(dataclassType), dataclassType.String())

	// Test for SimpleType
	simpleType := &core.LiteralType{
		Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
	}
	assert.Equal(t, LiteralTypeToStr(simpleType), "simple:INTEGER")
	assert.Equal(t, LiteralTypeToStr(simpleType), simpleType.String())
}
