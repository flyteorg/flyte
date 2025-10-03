package utils

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInfiniteRateLimiter(t *testing.T) {
	infRateLimiter := NewRateLimiter("test_rate_limiter", math.MaxFloat64, 0)
	start := time.Now()

	for i := 0; i < 100; i++ {
		err := infRateLimiter.Wait(context.Background())
		assert.NoError(t, err, "unexpected failure in wait")
	}
	assert.True(t, time.Since(start) < 100*time.Millisecond)
}

func TestRateLimiter(t *testing.T) {
	rateLimiter := NewRateLimiter("test_rate_limiter", 1, 1)
	start := time.Now()

	for i := 0; i < 5; i++ {
		err := rateLimiter.Wait(context.Background())
		assert.Nil(t, err)
	}
	assert.True(t, time.Since(start) > 3*time.Second)
	assert.True(t, time.Since(start) < 5*time.Second)
}

func TestInvalidRateLimitConfig(t *testing.T) {
	rateLimiter := NewRateLimiter("test_rate_limiter", 1, 0)
	err := rateLimiter.Wait(context.Background())
	assert.NotNil(t, err)
}
