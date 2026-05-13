package data

import (
	"context"
	stderrors "errors"
	"strings"
	"time"

	"github.com/flyteorg/flyte/flytestdlib/logger"
)

const (
	uploadFileRetryMaxAttemptIndex = 5
	uploadFileRetryDelay           = 2 * time.Second
)

const (
	errMsgHTTPServerClosedIdleConn = "http: server closed idle connection"
	errMsgUseOfClosedNetworkConn   = "use of closed network connection"
	errMsgEOF                      = "EOF"
	errMsgWriteBrokenPipe          = "write: broken pipe"
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

func isTransientStorageWriteError(err error) bool {
	if err == nil {
		return false
	}
	if errorChainContainsSubstring(err, errMsgHTTPServerClosedIdleConn) ||
		errorChainContainsSubstring(err, errMsgUseOfClosedNetworkConn) ||
		errorChainContainsSubstring(err, errMsgEOF) ||
		errorChainContainsSubstring(err, errMsgWriteBrokenPipe) {
		return true
	}
	return false
}

func errorChainContainsSubstring(err error, needle string) bool {
	for cur := err; cur != nil; cur = stderrors.Unwrap(cur) {
		if strings.Contains(cur.Error(), needle) {
			return true
		}
	}
	return false
}
