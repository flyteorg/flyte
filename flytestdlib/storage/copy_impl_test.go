package storage

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"testing"

	"github.com/flyteorg/flytestdlib/errors"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flytestdlib/ioutils"
)

type notSeekerReader struct {
	bytesCount int
}

func (notSeekerReader) Close() error {
	return nil
}

func (r *notSeekerReader) Read(p []byte) (n int, err error) {
	if len(p) < 1 {
		return 0, nil
	}

	p[0] = byte(10)

	r.bytesCount--
	if r.bytesCount <= 0 {
		return 0, io.EOF
	}

	return 1, nil
}

func newNotSeekerReader(bytesCount int) *notSeekerReader {
	return &notSeekerReader{
		bytesCount: bytesCount,
	}
}

func TestCopyRaw(t *testing.T) {
	t.Run("Called", func(t *testing.T) {
		readerCalled := false
		writerCalled := false
		store := dummyStore{
			ReadRawCb: func(ctx context.Context, reference DataReference) (closer io.ReadCloser, e error) {
				readerCalled = true
				return ioutils.NewBytesReadCloser([]byte{}), nil
			},
			WriteRawCb: func(ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader) error {
				writerCalled = true
				return nil
			},
		}

		copier := newCopyImpl(&store, metrics.copyMetrics)
		assert.NoError(t, copier.CopyRaw(context.Background(), DataReference("source.pb"), DataReference("dest.pb"), Options{}))
		assert.True(t, readerCalled)
		assert.True(t, writerCalled)
	})

	t.Run("Not Seeker", func(t *testing.T) {
		readerCalled := false
		writerCalled := false
		store := dummyStore{
			ReadRawCb: func(ctx context.Context, reference DataReference) (closer io.ReadCloser, e error) {
				readerCalled = true
				return newNotSeekerReader(10), nil
			},
			WriteRawCb: func(ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader) error {
				writerCalled = true
				return nil
			},
		}

		copier := newCopyImpl(&store, metrics.copyMetrics)
		assert.NoError(t, copier.CopyRaw(context.Background(), DataReference("source.pb"), DataReference("dest.pb"), Options{}))
		assert.True(t, readerCalled)
		assert.True(t, writerCalled)
	})
}

func TestCopyRaw_CachingErrorHandling(t *testing.T) {
	t.Run("CopyRaw with Caching Error", func(t *testing.T) {
		readerCalled := false
		writerCalled := false
		bigD := make([]byte, 1.5*1024*1024)
		// #nosec G404
		_, err := rand.Read(bigD)
		assert.NoError(t, err)
		dummyErrorMsg := "Dummy caching error"

		store := dummyStore{
			ReadRawCb: func(ctx context.Context, reference DataReference) (closer io.ReadCloser, e error) {
				readerCalled = true
				return ioutils.NewBytesReadCloser(bigD), errors.Wrapf(ErrFailedToWriteCache, fmt.Errorf(dummyErrorMsg), "Failed to Cache the metadata")
			},
			WriteRawCb: func(ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader) error {
				writerCalled = true
				return errors.Wrapf(ErrFailedToWriteCache, fmt.Errorf(dummyErrorMsg), "Failed to Cache the metadata")
			},
		}

		copier := newCopyImpl(&store, metrics.copyMetrics)
		assert.NoError(t, copier.CopyRaw(context.Background(), DataReference("source.pb"), DataReference("dest.pb"), Options{}))
		assert.True(t, readerCalled)
		assert.True(t, writerCalled)
	})

	t.Run("CopyRaw with Hard Error", func(t *testing.T) {
		readerCalled := false
		writerCalled := false
		bigD := make([]byte, 1.5*1024*1024)
		// #nosec G404
		_, err := rand.Read(bigD)
		assert.NoError(t, err)
		dummyErrorMsg := "Dummy non-caching error"

		store := dummyStore{
			ReadRawCb: func(ctx context.Context, reference DataReference) (closer io.ReadCloser, e error) {
				readerCalled = true
				return ioutils.NewBytesReadCloser(bigD), fmt.Errorf(dummyErrorMsg)
			},
			WriteRawCb: func(ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader) error {
				writerCalled = true
				return fmt.Errorf(dummyErrorMsg)
			},
		}

		copier := newCopyImpl(&store, metrics.copyMetrics)
		err = copier.CopyRaw(context.Background(), DataReference("source.pb"), DataReference("dest.pb"), Options{})
		assert.Error(t, err)
		assert.True(t, readerCalled)
		// writerCalled should be false because CopyRaw should error out right after c.rawstore.ReadRaw() when the underlying error is a hard error
		assert.False(t, writerCalled)
		assert.False(t, IsFailedWriteToCache(err))
	})
}
