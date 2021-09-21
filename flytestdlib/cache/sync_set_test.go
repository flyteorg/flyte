package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func rangeAndRemove(tb testing.TB, s *syncSet, count int) {
	for i := 0; i < count; i++ {
		s.Insert(i)
	}

	s.Range(func(key interface{}) bool {
		s.Remove(key)
		return true
	})

	for i := 0; i < count; i++ {
		assert.False(tb, s.Contains(i))
	}
}

func TestSyncSet_Range(t *testing.T) {
	s := newSyncSet()
	rangeAndRemove(t, s, 1000)
}

func BenchmarkSyncSet_Range(b *testing.B) {
	s := newSyncSet()
	rangeAndRemove(b, s, b.N)
}

func TestSyncSet_Contains(t *testing.T) {
	s := newSyncSet()
	count := 1000
	for i := 0; i < count; i++ {
		s.Insert(i)
	}

	for i := 0; i < count; i++ {
		assert.True(t, s.Contains(i))
		s.Remove(i)
	}

	for i := 0; i < count; i++ {
		assert.False(t, s.Contains(i))
	}
}
