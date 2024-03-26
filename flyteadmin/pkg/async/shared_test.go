package async

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	attemptsRecorded := 0
	err := Retry(3, time.Millisecond, func() error {
		if attemptsRecorded == 3 {
			return nil
		}
		attemptsRecorded++
		return errors.New("foo")
	})
	assert.Nil(t, err)
}

func TestRetry_RetriesExhausted(t *testing.T) {
	attemptsRecorded := 0
	err := Retry(2, time.Millisecond, func() error {
		if attemptsRecorded == 3 {
			return nil
		}
		attemptsRecorded++
		return errors.New("foo")
	})
	assert.EqualError(t, err, "foo")
}

func TestRetryOnlyOnRetryableExceptions(t *testing.T) {
	attemptsRecorded := 0
	err := RetryOnSpecificErrors(3, time.Millisecond, func() error {
		attemptsRecorded++
		if attemptsRecorded <= 1 {
			return errors.New("foo")
		}
		return errors.New("not-foo")
	}, func(err error) bool {
		return err.Error() == "foo"
	})
	assert.EqualValues(t, attemptsRecorded, 2)
	assert.EqualError(t, err, "not-foo")
}
