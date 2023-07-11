package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flytestdlib/ioutils"
)

func TestNewCachedStore(t *testing.T) {
	t.Run("CachingDisabled", func(t *testing.T) {
		cfg := &Config{}
		assert.Nil(t, newCachedRawStore(cfg, nil, metrics.cacheMetrics))
		store, err := NewInMemoryRawStore(context.TODO(), cfg, metrics)
		assert.NoError(t, err)
		assert.Equal(t, store, newCachedRawStore(cfg, store, metrics.cacheMetrics))
	})

	t.Run("CachingEnabled", func(t *testing.T) {
		cfg := &Config{
			Cache: CachingConfig{
				MaxSizeMegabytes: 1,
				TargetGCPercent:  20,
			},
		}
		store, err := NewInMemoryRawStore(context.TODO(), cfg, metrics)
		assert.NoError(t, err)
		cStore := newCachedRawStore(cfg, store, metrics.cacheMetrics)
		assert.Equal(t, 20, debug.SetGCPercent(100))
		assert.NotNil(t, cStore)
		assert.NotNil(t, cStore.(*cachedRawStore).cache)
	})
}

func dummyCacheStore(t *testing.T, store RawStore, metrics *cacheMetrics) *cachedRawStore {
	cfg := &Config{
		Cache: CachingConfig{
			MaxSizeMegabytes: 1,
			TargetGCPercent:  20,
		},
	}
	cStore := newCachedRawStore(cfg, store, metrics)
	assert.NotNil(t, cStore)
	return cStore.(*cachedRawStore)
}

type dummyStore struct {
	copyImpl
	HeadCb     func(ctx context.Context, reference DataReference) (Metadata, error)
	ReadRawCb  func(ctx context.Context, reference DataReference) (io.ReadCloser, error)
	WriteRawCb func(ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader) error
	DeleteCb   func(ctx context.Context, reference DataReference) error
}

// CreateSignedURL creates a signed url with the provided properties.
func (d *dummyStore) CreateSignedURL(ctx context.Context, reference DataReference, properties SignedURLProperties) (SignedURLResponse, error) {
	return SignedURLResponse{}, fmt.Errorf("unsupported")
}

func (d *dummyStore) GetBaseContainerFQN(ctx context.Context) DataReference {
	return "dummy"
}

func (d *dummyStore) Head(ctx context.Context, reference DataReference) (Metadata, error) {
	return d.HeadCb(ctx, reference)
}

func (d *dummyStore) ReadRaw(ctx context.Context, reference DataReference) (io.ReadCloser, error) {
	return d.ReadRawCb(ctx, reference)
}

func (d *dummyStore) WriteRaw(ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader) error {
	return d.WriteRawCb(ctx, reference, size, opts, raw)
}

func (d *dummyStore) Delete(ctx context.Context, reference DataReference) error {
	return d.DeleteCb(ctx, reference)
}

func TestCachedRawStore(t *testing.T) {
	ctx := context.TODO()
	k1 := DataReference("k1")
	k2 := DataReference("k2")
	bigK := DataReference("bigK")
	d1 := []byte("abc")
	d2 := []byte("xyz")
	bigD := make([]byte, 1.5*1024*1024)
	// #nosec G404
	_, err := rand.Read(bigD)
	assert.NoError(t, err)
	writeCalled := false
	readCalled := false
	store := &dummyStore{
		HeadCb: func(ctx context.Context, reference DataReference) (Metadata, error) {
			if reference == "k1" {
				return MemoryMetadata{exists: true, size: int64(len(d1))}, nil
			}
			return MemoryMetadata{}, fmt.Errorf("err")
		},
		WriteRawCb: func(ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader) error {
			if writeCalled {
				assert.FailNow(t, "Should not be writeCalled")
			}
			writeCalled = true
			if reference == "k2" {
				b, err := ioutil.ReadAll(raw)
				assert.NoError(t, err)
				assert.Equal(t, d2, b)
				return nil
			} else if reference == "bigK" {
				b, err := ioutil.ReadAll(raw)
				assert.NoError(t, err)
				assert.Equal(t, bigD, b)
				return nil
			}
			return fmt.Errorf("err")
		},
		ReadRawCb: func(ctx context.Context, reference DataReference) (io.ReadCloser, error) {
			if readCalled {
				assert.FailNow(t, "Should not be invoked again")
			}
			readCalled = true
			if reference == "k1" {
				return ioutils.NewBytesReadCloser(d1), nil
			} else if reference == "bigK" {
				return ioutils.NewBytesReadCloser(bigD), nil
			}
			return nil, fmt.Errorf("err")
		},
		DeleteCb: func(ctx context.Context, reference DataReference) error {
			if reference == "k1" {
				return nil
			}
			return fmt.Errorf("err")
		},
	}

	store.copyImpl = newCopyImpl(store, metrics.copyMetrics)

	cStore := dummyCacheStore(t, store, metrics.cacheMetrics)

	t.Run("HeadExists", func(t *testing.T) {
		m, err := cStore.Head(ctx, k1)
		assert.NoError(t, err)
		assert.Equal(t, int64(len(d1)), m.Size())
		assert.True(t, m.Exists())
	})

	t.Run("HeadNotExists", func(t *testing.T) {
		m, err := cStore.Head(ctx, k2)
		assert.Error(t, err)
		assert.False(t, m.Exists())
	})

	t.Run("ReadCachePopulate", func(t *testing.T) {
		o, err := cStore.ReadRaw(ctx, k1)
		assert.NoError(t, err)
		b, err := ioutil.ReadAll(o)
		assert.NoError(t, err)
		assert.Equal(t, d1, b)
		assert.True(t, readCalled)
		readCalled = false
		o, err = cStore.ReadRaw(ctx, k1)
		assert.NoError(t, err)
		b, err = ioutil.ReadAll(o)
		assert.NoError(t, err)
		assert.Equal(t, d1, b)
		assert.False(t, readCalled)
	})

	t.Run("ReadFail", func(t *testing.T) {
		readCalled = false
		_, err := cStore.ReadRaw(ctx, k2)
		assert.Error(t, err)
		assert.True(t, readCalled)
	})

	t.Run("WriteAndRead", func(t *testing.T) {
		readCalled = false
		assert.NoError(t, cStore.WriteRaw(ctx, k2, int64(len(d2)), Options{}, bytes.NewReader(d2)))
		assert.True(t, writeCalled)

		o, err := cStore.ReadRaw(ctx, k2)
		assert.NoError(t, err)
		b, err := ioutil.ReadAll(o)
		assert.NoError(t, err)
		assert.Equal(t, d2, b)
		assert.False(t, readCalled)
	})

	t.Run("WriteAndReadBigData", func(t *testing.T) {
		writeCalled = false
		readCalled = false
		err := cStore.WriteRaw(ctx, bigK, int64(len(bigD)), Options{}, bytes.NewReader(bigD))
		assert.True(t, writeCalled)
		assert.True(t, IsFailedWriteToCache(err))

		o, err := cStore.ReadRaw(ctx, bigK)
		assert.True(t, IsFailedWriteToCache(err))
		b, err := ioutil.ReadAll(o)
		assert.NoError(t, err)
		assert.Equal(t, bigD, b)
		assert.True(t, readCalled)
	})

	t.Run("DeleteExists", func(t *testing.T) {
		err := cStore.Delete(ctx, k1)
		assert.NoError(t, err)
	})

	t.Run("DeleteNotExists", func(t *testing.T) {
		err := cStore.Delete(ctx, k2)
		assert.Error(t, err)
	})
}
