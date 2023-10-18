package async

import (
	"errors"
	"net/http"
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

func TestRetryOnlyOnRetryableErrorCodes(t *testing.T) {
	attemptsRecorded := 0
	var response *http.Response
	err := RetryOnSpecificErrorCodes(3, time.Millisecond, func() (*http.Response, error) {
		attemptsRecorded++
		if attemptsRecorded <= 1 {
			retryResp := http.Response{StatusCode: 500}
			return &retryResp, nil
		}
		okResp := http.Response{StatusCode: http.StatusOK}
		response = &okResp
		return &okResp, errors.New("not an error")
	}, func(resp *http.Response) bool {
		return resp.StatusCode == 500
	})
	assert.Equal(t, 2, attemptsRecorded)
	assert.EqualError(t, err, "not an error")
	assert.Equal(t, http.StatusOK, response.StatusCode)
}
