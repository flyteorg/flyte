package sets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type GenericVal string

func (g GenericVal) GetID() string {
	return string(g)
}

func TestGenericSet(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, NewGeneric(GenericVal("a"), GenericVal("b")).ListKeys())
	assert.Equal(t, []string{"a", "b"}, NewGeneric(GenericVal("a"), GenericVal("b"), GenericVal("a")).ListKeys())

	{
		g1 := NewGeneric(GenericVal("a"), GenericVal("b"))
		g2 := NewGeneric(GenericVal("a"), GenericVal("b"), GenericVal("c"))
		{
			g := g1.Intersection(g2)
			assert.Equal(t, []string{"a", "b"}, g.ListKeys())
		}
		{
			g := g2.Intersection(g1)
			assert.Equal(t, []string{"a", "b"}, g.ListKeys())
		}
	}
	{
		g1 := NewGeneric(GenericVal("a"), GenericVal("b"))
		g2 := NewGeneric(GenericVal("a"), GenericVal("b"), GenericVal("c"))
		g := g2.Difference(g1)
		assert.Equal(t, []string{"c"}, g.ListKeys())
	}
	{
		g1 := NewGeneric(GenericVal("a"), GenericVal("b"))
		g2 := NewGeneric(GenericVal("a"), GenericVal("b"), GenericVal("c"))
		g := g1.Difference(g2)
		assert.Equal(t, []string{}, g.ListKeys())
	}
	{
		g1 := NewGeneric(GenericVal("a"), GenericVal("b"))
		assert.True(t, g1.Has(GenericVal("a")))
		assert.False(t, g1.Has(GenericVal("c")))
	}
	{
		g1 := NewGeneric(GenericVal("a"), GenericVal("b"))
		assert.False(t, g1.HasAll(GenericVal("a"), GenericVal("b"), GenericVal("c")))
		assert.True(t, g1.HasAll(GenericVal("a"), GenericVal("b")))
	}

	{
		g1 := NewGeneric(GenericVal("a"), GenericVal("b"))
		g2 := NewGeneric(GenericVal("a"), GenericVal("b"), GenericVal("c"))
		g := g1.Union(g2)
		assert.Equal(t, []string{"a", "b", "c"}, g.ListKeys())
	}

	{
		g1 := NewGeneric(GenericVal("a"), GenericVal("b"))
		g2 := NewGeneric(GenericVal("a"), GenericVal("b"), GenericVal("c"))
		assert.True(t, g2.IsSuperset(g1))
		assert.False(t, g1.IsSuperset(g2))
	}
	{
		g1 := NewGeneric(GenericVal("a"), GenericVal("b"))
		g2 := g1.UnsortedListKeys()
		assert.Equal(t, g1.Len(), len(g2))

		for _, key := range g2 {
			assert.True(t, g1.Has(GenericVal(key)))
		}
	}
	{
		g1 := NewGeneric(GenericVal("a"), GenericVal("b"))
		assert.True(t, g1.Has(GenericVal("a")))
		assert.False(t, g1.Has(GenericVal("c")))
		assert.Equal(t, []SetObject{
			GenericVal("a"),
			GenericVal("b"),
		}, g1.List())
		g2 := NewGeneric(g1.UnsortedList()...)
		assert.True(t, g1.Equal(g2))
	}
	{
		g1 := NewGeneric(GenericVal("a"), GenericVal("b"))
		g2 := NewGeneric(GenericVal("a"), GenericVal("b"), GenericVal("c"))
		g3 := NewGeneric(GenericVal("a"), GenericVal("b"))
		assert.False(t, g1.Equal(g2))
		assert.True(t, g1.Equal(g3))

		assert.Equal(t, 2, g1.Len())
		g1.Insert(GenericVal("b"))
		assert.Equal(t, 2, g1.Len())
		assert.True(t, g1.Equal(g3))
		g1.Insert(GenericVal("c"))
		assert.Equal(t, 3, g1.Len())
		assert.True(t, g1.Equal(g2))
		assert.True(t, g1.HasAny(GenericVal("a"), GenericVal("d")))
		assert.False(t, g1.HasAny(GenericVal("f"), GenericVal("d")))
		g1.Delete(GenericVal("f"))
		assert.True(t, g1.Equal(g2))
		g1.Delete(GenericVal("c"))
		assert.True(t, g1.Equal(g3))

		{
			p, ok := g1.PopAny()
			assert.NotNil(t, p)
			assert.True(t, ok)
		}
		{
			p, ok := g1.PopAny()
			assert.NotNil(t, p)
			assert.True(t, ok)
		}
		{
			p, ok := g1.PopAny()
			assert.Nil(t, p)
			assert.False(t, ok)
		}
	}
}
