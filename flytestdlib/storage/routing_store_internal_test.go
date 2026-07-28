package storage

import (
	"context"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

// fakeRawStore is a minimal RawStore that records which scheme served the last Head call.
type fakeRawStore struct {
	copyImpl
	scheme   string
	mu       sync.Mutex
	lastHead DataReference
}

func (f *fakeRawStore) GetBaseContainerFQN(context.Context) DataReference {
	return DataReference(f.scheme + "://")
}
func (f *fakeRawStore) CreateSignedURL(context.Context, DataReference, SignedURLProperties) (SignedURLResponse, error) {
	return SignedURLResponse{}, nil
}
func (f *fakeRawStore) Head(_ context.Context, ref DataReference) (Metadata, error) {
	f.mu.Lock()
	f.lastHead = ref
	f.mu.Unlock()
	return MemoryMetadata{exists: true}, nil
}
func (f *fakeRawStore) List(context.Context, DataReference, int, Cursor) ([]DataReference, Cursor, error) {
	return nil, NewCursorAtEnd(), nil
}
func (f *fakeRawStore) ReadRaw(context.Context, DataReference) (io.ReadCloser, error) {
	return nil, nil
}
func (f *fakeRawStore) WriteRaw(context.Context, DataReference, int64, Options, io.Reader) error {
	return nil
}
func (f *fakeRawStore) Delete(context.Context, DataReference) error { return nil }

// newTestRoutingStore builds a routingStore with fake factories so creation is observable and free
// of any network/credential dependency.
func newTestRoutingStore(t *testing.T, schemes ...string) (*routingStore, map[string]*int32) {
	t.Helper()
	metrics := newDataStoreMetrics(promutils.NewTestScope())
	counts := map[string]*int32{}
	registry := map[string]backendFactory{}
	for _, sc := range schemes {

		var n int32
		counts[sc] = &n
		registry[sc] = func(_ context.Context, scheme string, _ DataReference, _ *Config, _ *http.Client, m *dataStoreMetrics) (RawStore, error) {
			atomic.AddInt32(counts[scheme], 1)
			fs := &fakeRawStore{scheme: scheme}
			fs.copyImpl = newCopyImpl(fs, m.copyMetrics)
			return fs, nil
		}
	}

	primary := &fakeRawStore{scheme: schemes[0]}
	primary.copyImpl = newCopyImpl(primary, metrics.copyMetrics)
	rs := &routingStore{
		cfg:          &Config{},
		metrics:      metrics,
		registry:     registry,
		primary:      schemes[0],
		primaryStore: primary,
	}
	rs.live.Store(schemes[0], primary)
	rs.copyImpl = newCopyImpl(rs, metrics.copyMetrics)
	return rs, counts
}

func TestRoutingStore_PrimaryNotRebuilt(t *testing.T) {
	rs, counts := newTestRoutingStore(t, "foo", "bar")

	_, err := rs.Head(context.TODO(), "foo://c/k")
	require.NoError(t, err)

	// The primary scheme is pre-seeded, so its factory is never invoked.
	assert.Equal(t, int32(0), atomic.LoadInt32(counts["foo"]))
}

func TestRoutingStore_LazyCreateAndMemoize(t *testing.T) {
	rs, counts := newTestRoutingStore(t, "foo", "bar")

	for i := 0; i < 5; i++ {
		_, err := rs.Head(context.TODO(), "bar://c/k")
		require.NoError(t, err)
	}

	// Secondary scheme dialed exactly once across repeated references.
	assert.Equal(t, int32(1), atomic.LoadInt32(counts["bar"]))

	store, ok := rs.live.Load("bar")
	require.True(t, ok)
	assert.Equal(t, "bar", store.(*fakeRawStore).scheme)
}

func TestRoutingStore_UnknownSchemeFallsBackToPrimary(t *testing.T) {
	rs, counts := newTestRoutingStore(t, "foo", "bar")

	// "baz" has no factory; it should be served by the primary ("foo") store, not error.
	_, err := rs.Head(context.TODO(), "baz://c/k")
	require.NoError(t, err)
	assert.Equal(t, int32(0), atomic.LoadInt32(counts["bar"]))

	primary := rs.primaryStore.(*fakeRawStore)
	primary.mu.Lock()
	defer primary.mu.Unlock()
	assert.Equal(t, DataReference("baz://c/k"), primary.lastHead)
}

func TestRoutingStore_ConcurrentFirstUseDialsOnce(t *testing.T) {
	rs, counts := newTestRoutingStore(t, "foo", "bar")

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = rs.Head(context.TODO(), "bar://c/k")
		}()
	}
	wg.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(counts["bar"]))
}

// newTestRedisRoutingStore builds a routingStore whose only registered factory is a counting redis
// factory, with the given cfg. The primary is an unrelated fake store.
func newTestRedisRoutingStore(t *testing.T, cfg *Config) (*routingStore, *int32) {
	t.Helper()
	metrics := newDataStoreMetrics(promutils.NewTestScope())
	var count int32
	registry := map[string]backendFactory{
		TypeRedis: func(_ context.Context, scheme string, _ DataReference, _ *Config, _ *http.Client, m *dataStoreMetrics) (RawStore, error) {
			atomic.AddInt32(&count, 1)
			fs := &fakeRawStore{scheme: scheme}
			fs.copyImpl = newCopyImpl(fs, m.copyMetrics)
			return fs, nil
		},
	}
	primary := &fakeRawStore{scheme: TypeMemory}
	primary.copyImpl = newCopyImpl(primary, metrics.copyMetrics)
	rs := &routingStore{
		cfg:          cfg,
		metrics:      metrics,
		registry:     registry,
		primary:      TypeMemory,
		primaryStore: primary,
	}
	rs.live.Store(TypeMemory, primary)
	rs.copyImpl = newCopyImpl(rs, metrics.copyMetrics)
	return rs, &count
}

func TestRoutingStore_RedisKeyedByHostWithoutAddr(t *testing.T) {
	rs, count := newTestRedisRoutingStore(t, &Config{}) // no configured redis address

	// Distinct hosts must dial distinct backends (not silently reuse the first one).
	for _, ref := range []DataReference{"redis://hostA:6379/k", "redis://hostB:6379/k"} {
		_, err := rs.Head(context.TODO(), ref)
		require.NoError(t, err)
	}
	assert.Equal(t, int32(2), atomic.LoadInt32(count))

	// The same host is memoized — no additional dial.
	_, err := rs.Head(context.TODO(), "redis://hostA:6379/other")
	require.NoError(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(count))
}

func TestRoutingStore_RedisSharedWhenAddrConfigured(t *testing.T) {
	rs, count := newTestRedisRoutingStore(t, &Config{Redis: RedisConfig{Addr: "configured:6379"}})

	// With an authoritative configured address, distinct reference hosts share one backend.
	for _, ref := range []DataReference{"redis://hostA:6379/k", "redis://hostB:6379/k"} {
		_, err := rs.Head(context.TODO(), ref)
		require.NoError(t, err)
	}
	assert.Equal(t, int32(1), atomic.LoadInt32(count))
}
