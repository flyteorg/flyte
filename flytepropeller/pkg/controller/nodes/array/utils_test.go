package array

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
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

func TestInferParallelism(t *testing.T) {
	ctx := context.TODO()
	zero := uint32(0)
	one := uint32(1)

	tests := []struct {
		name                   string
		parallelism            *uint32
		parallelismBehavior    string
		remainingParallelism   int
		arrayNodeSize          int
		expectedIncrement      bool
		expectedMaxParallelism int
	}{
		{
			name:                   "NilParallelismWorkflowBehavior",
			parallelism:            nil,
			parallelismBehavior:    "workflow",
			remainingParallelism:   2,
			arrayNodeSize:          3,
			expectedIncrement:      true,
			expectedMaxParallelism: 2,
		},
		{
			name:                   "NilParallelismHybridBehavior",
			parallelism:            nil,
			parallelismBehavior:    "hybrid",
			remainingParallelism:   2,
			arrayNodeSize:          3,
			expectedIncrement:      true,
			expectedMaxParallelism: 2,
		},
		{
			name:                   "NilParallelismUnlimitedBehavior",
			parallelism:            nil,
			parallelismBehavior:    "unlimited",
			remainingParallelism:   2,
			arrayNodeSize:          3,
			expectedIncrement:      false,
			expectedMaxParallelism: 3,
		},
		{
			name:                   "ZeroParallelismWorkflowBehavior",
			parallelism:            &zero,
			parallelismBehavior:    "workflow",
			remainingParallelism:   2,
			arrayNodeSize:          3,
			expectedIncrement:      true,
			expectedMaxParallelism: 2,
		},
		{
			name:                   "ZeroParallelismHybridBehavior",
			parallelism:            &zero,
			parallelismBehavior:    "hybrid",
			remainingParallelism:   2,
			arrayNodeSize:          3,
			expectedIncrement:      false,
			expectedMaxParallelism: 3,
		},
		{
			name:                   "ZeroParallelismUnlimitedBehavior",
			parallelism:            &zero,
			parallelismBehavior:    "unlimited",
			remainingParallelism:   2,
			arrayNodeSize:          3,
			expectedIncrement:      false,
			expectedMaxParallelism: 3,
		},
		{
			name:                   "OneParallelismWorkflowBehavior",
			parallelism:            &one,
			parallelismBehavior:    "workflow",
			remainingParallelism:   2,
			arrayNodeSize:          3,
			expectedIncrement:      false,
			expectedMaxParallelism: 1,
		},
		{
			name:                   "OneParallelismHybridBehavior",
			parallelism:            &one,
			parallelismBehavior:    "hybrid",
			remainingParallelism:   2,
			arrayNodeSize:          3,
			expectedIncrement:      false,
			expectedMaxParallelism: 1,
		},
		{
			name:                   "OneParallelismUnlimitedBehavior",
			parallelism:            &one,
			parallelismBehavior:    "unlimited",
			remainingParallelism:   2,
			arrayNodeSize:          3,
			expectedIncrement:      false,
			expectedMaxParallelism: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			increment, maxParallelism := inferParallelism(ctx, tt.parallelism, tt.parallelismBehavior, tt.remainingParallelism, tt.arrayNodeSize)
			assert.Equal(t, tt.expectedIncrement, increment)
			assert.Equal(t, tt.expectedMaxParallelism, maxParallelism)
		})
	}
}
