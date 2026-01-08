package dynamic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestFromFutureLiteralMapToDynamicJobSpec(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		spec, err := FromFutureLiteralMapToDynamicJobSpec(nil)
		assert.NoError(t, err)
		assert.Nil(t, spec)
	})

	t.Run("empty literals", func(t *testing.T) {
		spec, err := FromFutureLiteralMapToDynamicJobSpec(&core.LiteralMap{
			Literals: map[string]*core.Literal{},
		})
		assert.NoError(t, err)
		assert.Nil(t, spec)
	})

	t.Run("missing future literal", func(t *testing.T) {
		spec, err := FromFutureLiteralMapToDynamicJobSpec(&core.LiteralMap{
			Literals: map[string]*core.Literal{
				"wrong-key": {},
			},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to retrieve Future literal")
		assert.Nil(t, spec)
	})

	t.Run("invalid literal format", func(t *testing.T) {
		spec, err := FromFutureLiteralMapToDynamicJobSpec(&core.LiteralMap{
			Literals: map[string]*core.Literal{
				"future": {},
			},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to get Future Binary from literal")
		assert.Nil(t, spec)
	})

	t.Run("successful conversion", func(t *testing.T) {
		// Create a sample DynamicJobSpec
		expectedSpec := &core.DynamicJobSpec{
			Tasks: []*core.TaskTemplate{
				{
					Id: &core.Identifier{Name: "test-task"},
				},
			},
		}

		// Serialize DynamicJobSpec
		bytes, err := proto.Marshal(expectedSpec)
		assert.NoError(t, err)

		// Create LiteralMap containing serialized data
		literalMap := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"future": {
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Binary{
								Binary: &core.Binary{
									Value: bytes,
								},
							},
						},
					},
				},
			},
		}

		// Test conversion
		actualSpec, err := FromFutureLiteralMapToDynamicJobSpec(literalMap)
		assert.NoError(t, err)
		assert.NotNil(t, actualSpec)
		assert.Equal(t, expectedSpec.GetTasks()[0].GetId().GetName(), actualSpec.GetTasks()[0].GetId().GetName())
	})
}
