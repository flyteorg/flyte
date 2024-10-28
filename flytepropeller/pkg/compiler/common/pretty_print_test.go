package common

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestLiteralTypeToStr(t *testing.T) {
	assert.Equal(t, LiteralTypeToStr(nil), "None")
	assert.Equal(t, LiteralTypeToStr(&core.LiteralType{
		Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRUCT},
		Structure: &core.TypeStructure{
			DataclassType: map[string]*core.LiteralType{
				"a": {
					Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
				},
			},
		},
	}), "Simple: STRUCT structure{dataclass_type:{key:a value:{simple:INTEGER}}")
}
