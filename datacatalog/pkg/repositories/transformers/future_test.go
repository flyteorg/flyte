package transformers

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestFromDynamicJobSpecToLiteral(t *testing.T) {
	djSpec := &core.DynamicJobSpec{
		Nodes: []*core.Node{
			{
				Id: "node1",
			},
		},
		MinSuccesses: 1,
	}

	literal, err := FromDynamicJobSpecToLiteral(djSpec)
	assert.NoError(t, err, "Should not return an error")
	assert.NotNil(t, literal, "Returned Literal should not be nil")

	assert.NotNil(t, literal.GetScalar(), "Scalar should not be nil")
	assert.NotNil(t, literal.GetScalar().GetBinary(), "Binary should not be nil")
	assert.NotEmpty(t, literal.GetScalar().GetBinary().GetValue(), "Binary value should not be empty")

	expectedBinary, _ := proto.Marshal(djSpec)
	assert.Equal(t, expectedBinary, literal.GetScalar().GetBinary().GetValue(), "Binary data should match the expected value")
}

func TestFromFutureLiteralToDynamicJobSpec(t *testing.T) {
	originalDjSpec := &core.DynamicJobSpec{
		Nodes: []*core.Node{
			{
				Id: "node1",
			},
		},
		MinSuccesses: 1,
	}

	literal, err := FromDynamicJobSpecToLiteral(originalDjSpec)
	assert.NoError(t, err, "Should not return an error when converting to Literal")

	ctx := context.Background()
	reconstructedDjSpec, err := FromFutureLiteralToDynamicJobSpec(ctx, literal)
	assert.NoError(t, err, "Should not return an error")
	assert.NotNil(t, reconstructedDjSpec, "Returned DynamicJobSpec should not be nil")

	assert.True(t, proto.Equal(originalDjSpec, reconstructedDjSpec), "Original and deserialized DynamicJobSpec should be equal")
}
