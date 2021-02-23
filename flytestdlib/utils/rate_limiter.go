package utils

import (
	"context"

	"github.com/flyteorg/flytestdlib/logger"
	"golang.org/x/time/rate"
)

// Interface to use rate limiter
type RateLimiter interface {
	Wait(ctx context.Context) error
}

type rateLimiter struct {
	name    string
	limiter rate.Limiter
}

// Blocking method which waits for the next token as per the tps and burst values defined
func (r *rateLimiter) Wait(ctx context.Context) error {
	logger.Debugf(ctx, "Waiting for a token from rate limiter %s", r.name)
	if err := r.limiter.Wait(ctx); err != nil {
		return err
	}
	return nil
}

// Create a new rate-limiter with the tps and burst values
func NewRateLimiter(rateLimiterName string, tps float64, burst int) RateLimiter {
	return &rateLimiter{
		name:    rateLimiterName,
		limiter: *rate.NewLimiter(rate.Limit(tps), burst),
	}
}
