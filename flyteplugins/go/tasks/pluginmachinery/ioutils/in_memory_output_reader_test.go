package ioutils

import (
	"context"
	"testing"

	flyteIdlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryOutputReader(t *testing.T) {
	deckPath := storage.DataReference("s3://bucket/key")
	lt := map[string]*flyteIdlCore.Literal{
		"results": {
			Value: &flyteIdlCore.Literal_Scalar{
				Scalar: &flyteIdlCore.Scalar{
					Value: &flyteIdlCore.Scalar_Primitive{
						Primitive: &flyteIdlCore.Primitive{Value: &flyteIdlCore.Primitive_Integer{Integer: 3}},
					},
				},
			},
		},
	}
	or := NewInMemoryOutputReader(&flyteIdlCore.LiteralMap{Literals: lt}, &deckPath, nil)

	assert.Equal(t, &deckPath, or.DeckPath)
	ctx := context.TODO()

	ok, err := or.IsError(ctx)
	assert.False(t, ok)
	assert.NoError(t, err)

	assert.False(t, or.IsFile(ctx))

	ok, err = or.Exists(ctx)
	assert.True(t, ok)
	assert.NoError(t, err)

	literalMap, executionErr, err := or.Read(ctx)
	assert.Equal(t, lt, literalMap.Literals)
	assert.Nil(t, executionErr)
	assert.NoError(t, err)
}
