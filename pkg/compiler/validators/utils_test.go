package validators

import (
	"testing"

	flyte "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestJoinVariableMapsUniqueKeys(t *testing.T) {
	intType := &flyte.LiteralType{
		Type: &flyte.LiteralType_Simple{
			Simple: flyte.SimpleType_INTEGER,
		},
	}

	strType := &flyte.LiteralType{
		Type: &flyte.LiteralType_Simple{
			Simple: flyte.SimpleType_STRING,
		},
	}

	t.Run("Simple", func(t *testing.T) {
		m1 := map[string]*flyte.Variable{
			"x": {
				Type: intType,
			},
		}

		m2 := map[string]*flyte.Variable{
			"y": {
				Type: intType,
			},
		}

		res, err := UnionDistinctVariableMaps(m1, m2)
		assert.NoError(t, err)
		assert.Len(t, res, 2)
	})

	t.Run("No type collision", func(t *testing.T) {
		m1 := map[string]*flyte.Variable{
			"x": {
				Type: intType,
			},
		}

		m2 := map[string]*flyte.Variable{
			"x": {
				Type: intType,
			},
		}

		res, err := UnionDistinctVariableMaps(m1, m2)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("Type collision", func(t *testing.T) {
		m1 := map[string]*flyte.Variable{
			"x": {
				Type: intType,
			},
		}

		m2 := map[string]*flyte.Variable{
			"x": {
				Type: strType,
			},
		}

		_, err := UnionDistinctVariableMaps(m1, m2)
		assert.Error(t, err)
	})
}
