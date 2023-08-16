package array

import (
	"testing"

	idlcore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"
)

func TestAppendLiteral(t *testing.T) {
	outputLiterals := make(map[string]*idlcore.Literal)
	literalMaps := []map[string]*idlcore.Literal{
		map[string]*idlcore.Literal{
			"foo": nilLiteral,
			"bar": nilLiteral,
		},
		map[string]*idlcore.Literal{
			"foo": nilLiteral,
			"bar": nilLiteral,
		},
	}

	for _, m := range literalMaps {
		for k, v := range m {
			appendLiteral(k, v, outputLiterals, len(literalMaps))
		}
	}

	for _, v := range outputLiterals {
		collection, ok := v.Value.(*idlcore.Literal_Collection)
		assert.True(t, ok)

		assert.Equal(t, 2, len(collection.Collection.Literals))
	}
}
