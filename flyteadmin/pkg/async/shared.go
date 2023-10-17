package async

import (
	"context"
	"net/http"
	"time"

	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// RetryDelay indicates how long to wait between restarting a subscriber connection in the case of network failures.
var RetryDelay = 30 * time.Second

func RetryOnSpecificErrorCodes(attempts int, delay time.Duration, f func() (*http.Response, error), IsErrorCodeRetryable func(*http.Response) bool) error {
	var err error
	var resp *http.Response
	for attempt := 0; attempt <= attempts; attempt++ {
		resp, err = f()
		if err != nil {
			return err
		}
		if !IsErrorCodeRetryable(resp) {
			return err
		}
		logger.Warningf(context.Background(),
			"Failed status code %v on attempt %d of %d", resp.StatusCode, attempt, attempts)
		time.Sleep(delay)
	}
	return err
}

func RetryOnSpecificErrors(attempts int, delay time.Duration, f func() error, IsErrorRetryable func(error) bool) error {
	var err error
	for attempt := 0; attempt <= attempts; attempt++ {
		err = f()
		if err == nil {
			return nil
		}
		if !IsErrorRetryable(err) {
			return err
		}
		logger.Warningf(context.Background(),
			"Failed [%v] on attempt %d of %d", err, attempt, attempts)
		time.Sleep(delay)
	}
	return err
}

func retryOnAllErrors(err error) bool {
	return true
}

func Retry(attempts int, delay time.Duration, f func() error) error {
	return RetryOnSpecificErrors(attempts, delay, f, retryOnAllErrors)
}
