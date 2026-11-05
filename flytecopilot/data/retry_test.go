package data

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var s3LikeError = errors.Wrap(errors.New(
	`Put "https://example.s3.us-west-2.amazonaws.com/key?partNumber=11&uploadId=x": http: server closed idle connection`),
	`PutObject, putting object: MultipartUpload: upload multipart failed
caused by: RequestError: send request failed
caused by: Put "https://example.s3.us-west-2.amazonaws.com/key?partNumber=11&uploadId=x"`)

func TestRetryOnSpecificErrors_SucceedsFirstTry(t *testing.T) {
	ctx := context.Background()
	calls := 0
	err := retryOnSpecificErrors(ctx, 3, time.Millisecond, func() error {
		calls++
		return nil
	}, retryOnAllErrors)
	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestRetryOnSpecificErrors_RetriesThenSucceeds(t *testing.T) {
	ctx := context.Background()
	calls := 0
	err := retryOnSpecificErrors(ctx, 5, time.Millisecond, func() error {
		calls++
		if calls < 3 {
			return s3LikeError
		}
		return nil
	}, retryOnAllErrors)
	assert.NoError(t, err)
	assert.Equal(t, 3, calls)
}
