package async

import (
	"context"
	"time"

	"github.com/flyteorg/flytestdlib/logger"
)

// RetryDelay indicates how long to wait between restarting a subscriber connection in the case of network failures.
var RetryDelay = 30 * time.Second

func Retry(attempts int, delay time.Duration, f func() error) error {
	var err error
	for attempt := 0; attempt <= attempts; attempt++ {
		err = f()
		if err == nil {
			return nil
		}
		logger.Warningf(context.Background(),
			"Failed [%v] on attempt %d of %d", err, attempt, attempts)
		time.Sleep(delay)
	}
	return err
}
