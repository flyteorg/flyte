package atomic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNonBlockingLock(t *testing.T) {
	lock := NewNonBlockingLock()
	assert.True(t, lock.TryLock(), "Unexpected lock acquire failure")
	assert.False(t, lock.TryLock(), "already-acquired lock acquired again")
	lock.Release()
	assert.True(t, lock.TryLock(), "Unexpected lock acquire failure")
}
