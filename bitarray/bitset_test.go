package bitarray

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleBitSet() {
	s := new(BitSet)
	s.Set(13)
	s.Set(45)
	s.Clear(13)
	fmt.Printf("s.IsSet(13) = %t; s.IsSet(45) = %t; s.IsSet(30) = %t\n",
		s.IsSet(13), s.IsSet(45), s.IsSet(30))
	// Output: s.IsSet(13) = false; s.IsSet(45) = true; s.IsSet(30) = false
}

func TestBitSet_Set(t *testing.T) {
	t.Run("Empty Set", func(t *testing.T) {
		b := new(BitSet)
		b.Set(5)
		assert.True(t, b.IsSet(5))
	})

	t.Run("Auto resize", func(t *testing.T) {
		b := new(BitSet)
		b.Set(2)
		assert.Equal(t, 1, len(*b))
		assert.False(t, b.IsSet(500))
		b.Set(500)
		assert.True(t, b.IsSet(2))
		assert.True(t, b.IsSet(500))
	})
}

func TestNewBitSet(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		b := new(BitSet)
		assert.Equal(t, 0, b.BlockCount())

		b = NewBitSet(0)
		assert.Equal(t, 0, b.BlockCount())
	})

	t.Run("Block size", func(t *testing.T) {
		b := NewBitSet(63)
		assert.Equal(t, 2, b.BlockCount())
	})

	t.Run("Bigger than block size", func(t *testing.T) {
		b := NewBitSet(100)
		assert.Equal(t, 4, b.BlockCount())
	})
}

func TestBitSet_Cap(t *testing.T) {
	t.Run("Cap == size", func(t *testing.T) {
		b := NewBitSet(blockSize * 5)
		assert.Equal(t, int(blockSize*5), int(b.Cap()))
	})

	t.Run("Cap > size", func(t *testing.T) {
		b := NewBitSet(blockSize*2 + 20)
		assert.Equal(t, int(blockSize*3), int(b.Cap()))
	})
}
