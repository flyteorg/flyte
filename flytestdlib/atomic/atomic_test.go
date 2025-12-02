package atomic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	atom := NewBool(false)
	assert.False(t, atom.Toggle(), "Expected swap to return False.")
	assert.True(t, atom.Load(), "Unexpected state after swap. Expected True")

	assert.True(t, atom.CompareAndSwap(true, true), "CAS should swap when old matches")
	assert.True(t, atom.Load(), "previous swap should have no effect")
	assert.True(t, atom.CompareAndSwap(true, false), "CAS should swap when old matches")
	assert.False(t, atom.Load(), "Post swap the value should be true")
	assert.False(t, atom.CompareAndSwap(true, false), "CAS should fail on old mismatch")
	assert.False(t, atom.Load(), "CAS should not have modified the value")

	atom.Store(false)
	assert.False(t, atom.Load(), "Unexpected state after store.")

	prev := atom.Swap(false)
	assert.False(t, prev, "Expected Swap to return previous value.")

	prev = atom.Swap(true)
	assert.False(t, prev, "Expected Swap to return previous value.")
}

func TestInt32(t *testing.T) {
	atom := NewInt32(2)
	assert.False(t, atom.CompareAndSwap(3, 4), "Expected swap to return False.")
	assert.Equal(t, int32(2), atom.Load(), "Unexpected state after swap. Expected True")

	assert.True(t, atom.CompareAndSwap(2, 2), "CAS should swap when old matches")
	assert.Equal(t, int32(2), atom.Load(), "previous swap should have no effect")
	assert.True(t, atom.CompareAndSwap(2, 4), "CAS should swap when old matches")
	assert.Equal(t, int32(4), atom.Load(), "Post swap the value should be true")
	assert.False(t, atom.CompareAndSwap(2, 3), "CAS should fail on old mismatch")
	assert.Equal(t, int32(4), atom.Load(), "CAS should not have modified the value")

	atom.Store(5)
	assert.Equal(t, int32(5), atom.Load(), "Unexpected state after store.")
}

func TestUint32(t *testing.T) {
	atom := NewUint32(2)
	assert.False(t, atom.CompareAndSwap(3, 4), "Expected swap to return False.")
	assert.Equal(t, uint32(2), atom.Load(), "Unexpected state after swap. Expected True")

	assert.True(t, atom.CompareAndSwap(2, 2), "CAS should swap when old matches")
	assert.Equal(t, uint32(2), atom.Load(), "previous swap should have no effect")
	assert.True(t, atom.CompareAndSwap(2, 4), "CAS should swap when old matches")
	assert.Equal(t, uint32(4), atom.Load(), "Post swap the value should be true")
	assert.False(t, atom.CompareAndSwap(2, 3), "CAS should fail on old mismatch")
	assert.Equal(t, uint32(4), atom.Load(), "CAS should not have modified the value")

	atom.Store(5)
	assert.Equal(t, uint32(5), atom.Load(), "Unexpected state after store.")
}
