package data

import (
	"context"
	"time"

	"github.com/flyteorg/flyte/flytestdlib/logger"
)

const (
	uploadFileRetryMaxAttemptIndex = 5
	uploadFileRetryDelay           = 2 * time.Second
)

// See flyteadmin/pkg/async.RetryOnSpecificErrors
func retryOnSpecificErrors(ctx context.Context, attempts int, delay time.Duration, f func() error, isErrorRetryable func(error) bool) error {
	var err error
	for attempt := 0; attempt <= attempts; attempt++ {
		err = f()
		if err == nil {
			return nil
		}
		if !isErrorRetryable(err) {
			return err
		}
		if attempt == attempts {
			return err
		}
		logger.Warningf(ctx, "Failed [%v] on attempt %d of %d; retrying after %v", err, attempt, attempts, delay)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return err
}

func retryOnAllErrors(err error) bool {
	return true
}
