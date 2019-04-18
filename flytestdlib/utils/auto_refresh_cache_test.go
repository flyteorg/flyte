package utils

import (
	"context"
	"sync"
	"testing"
	"time"

	atomic2 "sync/atomic"

	"github.com/lyft/flytestdlib/atomic"
	"github.com/stretchr/testify/assert"
)

type testCacheItem struct {
	val          int
	deleted      atomic.Bool
	resyncPeriod time.Duration
	wg           sync.WaitGroup
}

func (m *testCacheItem) ID() string {
	return "id"
}

func (m *testCacheItem) moveNext() {
	// change value and spare enough time for cache to process the change.
	m.val++
	time.Sleep(m.resyncPeriod * 5)
}

func (m *testCacheItem) syncItem(ctx context.Context, obj CacheItem) (CacheItem, error) {
	defer func() { m.wg.Done() }()

	if m.deleted.Load() {
		return nil, nil
	}
	return m, nil
}

type testAutoIncrementItem struct {
	val int32
}

func (a *testAutoIncrementItem) ID() string {
	return "autoincrement"
}

func (a *testAutoIncrementItem) syncItem(ctx context.Context, obj CacheItem) (CacheItem, error) {
	atomic2.AddInt32(&a.val, 1)
	return a, nil
}

func TestCache(t *testing.T) {
	testResyncPeriod := time.Millisecond
	rateLimiter := NewRateLimiter("mockLimiter", 100, 1)

	wg := sync.WaitGroup{}
	wg.Add(1)
	item := &testCacheItem{
		val:          0,
		resyncPeriod: testResyncPeriod,
		deleted:      atomic.NewBool(false),
		wg:           wg,}
	cache := NewAutoRefreshCache(item.syncItem, rateLimiter, testResyncPeriod)

	ctx, cancel := context.WithCancel(context.Background())
	cache.Start(ctx)

	// create
	_, err := cache.GetOrCreate(item)
	assert.NoError(t, err, "unexpected GetOrCreate failure")

	// synced?
	item.moveNext()
	m := cache.Get(item.ID()).(*testCacheItem)
	assert.Equal(t, 1, m.val)

	// synced again?
	item.moveNext()
	m = cache.Get(item.ID()).(*testCacheItem)
	assert.Equal(t, 2, m.val)

	// removed?
	item.moveNext()
	item.deleted.Store(true)
	wg.Wait()
	time.Sleep(testResyncPeriod * 2) // spare enough time to process remove!
	val := cache.Get(item.ID())

	assert.Nil(t, val)
	cancel()
}

func TestCacheContextCancel(t *testing.T) {
	testResyncPeriod := time.Millisecond
	rateLimiter := NewRateLimiter("mockLimiter", 10000, 1)

	item := &testAutoIncrementItem{val: 0}
	cache := NewAutoRefreshCache(item.syncItem, rateLimiter, testResyncPeriod)

	ctx, cancel := context.WithCancel(context.Background())
	cache.Start(ctx)
	_, err := cache.GetOrCreate(item)
	assert.NoError(t, err, "failed to add item to cache")
	time.Sleep(testResyncPeriod * 10) // spare enough time to process remove!
	cancel()

	// Get item
	m, err := cache.GetOrCreate(item)
	val1 := m.(*testAutoIncrementItem).val
	assert.NoError(t, err, "unexpected GetOrCreate failure")

	// wait a few more resync periods and check that nothings has changed as auto-refresh is stopped
	time.Sleep(testResyncPeriod * 20)
	val2 := m.(*testAutoIncrementItem).val
	assert.Equal(t, val1, val2)
}
