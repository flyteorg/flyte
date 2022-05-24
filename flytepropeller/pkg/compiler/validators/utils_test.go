package validators

import (
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestLiteralTypeForLiterals(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		lt := literalTypeForLiterals(nil)
		assert.Equal(t, core.SimpleType_NONE.String(), lt.GetSimple().String())
	})

	t.Run("homogenous", func(t *testing.T) {
		lt := literalTypeForLiterals([]*core.Literal{
			coreutils.MustMakeLiteral(5),
			coreutils.MustMakeLiteral(0),
			coreutils.MustMakeLiteral(5),
		})

		assert.Equal(t, core.SimpleType_INTEGER.String(), lt.GetSimple().String())
	})

	t.Run("non-homogenous", func(t *testing.T) {
		lt := literalTypeForLiterals([]*core.Literal{
			coreutils.MustMakeLiteral("hello"),
			coreutils.MustMakeLiteral(5),
			coreutils.MustMakeLiteral("world"),
			coreutils.MustMakeLiteral(0),
			coreutils.MustMakeLiteral(2),
		})

		assert.Len(t, lt.GetUnionType().Variants, 2)
		assert.Equal(t, core.SimpleType_INTEGER.String(), lt.GetUnionType().Variants[0].GetSimple().String())
		assert.Equal(t, core.SimpleType_STRING.String(), lt.GetUnionType().Variants[1].GetSimple().String())
	})

	t.Run("non-homogenous ensure ordering", func(t *testing.T) {
		lt := literalTypeForLiterals([]*core.Literal{
			coreutils.MustMakeLiteral(5),
			coreutils.MustMakeLiteral("world"),
			coreutils.MustMakeLiteral(0),
			coreutils.MustMakeLiteral(2),
		})

		assert.Len(t, lt.GetUnionType().Variants, 2)
		assert.Equal(t, core.SimpleType_INTEGER.String(), lt.GetUnionType().Variants[0].GetSimple().String())
		assert.Equal(t, core.SimpleType_STRING.String(), lt.GetUnionType().Variants[1].GetSimple().String())
	})
}

func TestJoinVariableMapsUniqueKeys(t *testing.T) {
	intType := &core.LiteralType{
		Type: &core.LiteralType_Simple{
			Simple: core.SimpleType_INTEGER,
		},
	}

	strType := &core.LiteralType{
		Type: &core.LiteralType_Simple{
			Simple: core.SimpleType_STRING,
		},
	}

	t.Run("Simple", func(t *testing.T) {
		m1 := map[string]*core.Variable{
			"x": {
				Type: intType,
			},
		}

		m2 := map[string]*core.Variable{
			"y": {
				Type: intType,
			},
		}

		res, err := UnionDistinctVariableMaps(m1, m2)
		assert.NoError(t, err)
		assert.Len(t, res, 2)
	})

	t.Run("No type collision", func(t *testing.T) {
		m1 := map[string]*core.Variable{
			"x": {
				Type: intType,
			},
		}

		m2 := map[string]*core.Variable{
			"x": {
				Type: intType,
			},
		}

		res, err := UnionDistinctVariableMaps(m1, m2)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("Type collision", func(t *testing.T) {
		m1 := map[string]*core.Variable{
			"x": {
				Type: intType,
			},
		}

		m2 := map[string]*core.Variable{
			"x": {
				Type: strType,
			},
		}

		_, err := UnionDistinctVariableMaps(m1, m2)
		assert.Error(t, err)
	})
}
