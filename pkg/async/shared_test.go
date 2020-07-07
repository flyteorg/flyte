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
