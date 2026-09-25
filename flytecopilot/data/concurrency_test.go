package data

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeConcurrencyPerCPU(t *testing.T) {
	assert.Equal(t, DefaultConcurrencyPerCPU, sanitizeConcurrencyPerCPU(0))
	assert.Equal(t, DefaultConcurrencyPerCPU, sanitizeConcurrencyPerCPU(-1))
	assert.Equal(t, DefaultConcurrencyPerCPU, sanitizeConcurrencyPerCPU(-100))
	assert.Equal(t, 1, sanitizeConcurrencyPerCPU(1))
	assert.Equal(t, 2, sanitizeConcurrencyPerCPU(2))
	assert.Equal(t, 10, sanitizeConcurrencyPerCPU(10))
}

func TestGetConcurrency(t *testing.T) {
	result := getConcurrency(DefaultConcurrencyPerCPU)
	assert.GreaterOrEqual(t, result, 1)

	result = getConcurrency(0)
	assert.GreaterOrEqual(t, result, 1)

	result = getConcurrency(1)
	assert.GreaterOrEqual(t, result, 1)
	assert.LessOrEqual(t, result, MaxConcurrency)

	result = getConcurrency(100000)
	assert.LessOrEqual(t, result, MaxConcurrency)
}

func TestAcquireSemaphore(t *testing.T) {
	t.Run("acquires successfully", func(t *testing.T) {
		sem := make(chan struct{}, 2)
		ctx := context.Background()

		assert.NoError(t, acquireSemaphore(ctx, sem))
		assert.NoError(t, acquireSemaphore(ctx, sem))
		assert.Len(t, sem, 2)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		sem := make(chan struct{}, 1)
		sem <- struct{}{}

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err := acquireSemaphore(ctx, sem)
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("returns immediately on already canceled context", func(t *testing.T) {
		sem := make(chan struct{}, 1)
		sem <- struct{}{}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := acquireSemaphore(ctx, sem)
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})
}
