package ioutils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrecomputedShardSelector_GetShardPrefix(t *testing.T) {
	ctx := context.TODO()
	t.Run("single-shard", func(t *testing.T) {
		ss := PrecomputedShardSelector{precomputedPrefixes: []string{"x"}, buckets: 1}
		p, err := ss.GetShardPrefix(ctx, []byte("abc"))
		assert.NoError(t, err)
		assert.Equal(t, "x", p)
	})

	t.Run("two-shards", func(t *testing.T) {
		ss := PrecomputedShardSelector{precomputedPrefixes: []string{"x", "y"}, buckets: 2}
		p, err := ss.GetShardPrefix(ctx, []byte("abc"))
		assert.NoError(t, err)
		assert.Equal(t, "y", p)
		p, err = ss.GetShardPrefix(ctx, []byte("xyz"))
		assert.NoError(t, err)
		assert.Equal(t, "x", p)
	})
}

func TestGenerateAlphabet(t *testing.T) {
	var b []rune
	b = GenerateAlphabet(b)

	assert.Equal(t, 26, len(b))
	assert.Equal(t, 'a', b[0])
	assert.Equal(t, 'z', b[25])

	// Additive
	b = GenerateAlphabet(b)

	assert.Equal(t, 52, len(b))
	assert.Equal(t, 'a', b[26])
	assert.Equal(t, 'z', b[51])
}

func TestGenerateArabicNumerals(t *testing.T) {
	var b []rune
	b = GenerateArabicNumerals(b)

	assert.Equal(t, 10, len(b))
	assert.Equal(t, '0', b[0])
	assert.Equal(t, '9', b[9])

	// Additive
	b = GenerateArabicNumerals(b)
	assert.Equal(t, 20, len(b))
	assert.Equal(t, '0', b[0])
	assert.Equal(t, '9', b[9])
	assert.Equal(t, '0', b[10])
	assert.Equal(t, '9', b[19])
}
