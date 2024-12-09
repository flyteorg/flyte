package cache

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/util/workqueue"
	testingclock "k8s.io/utils/clock/testing"

	"github.com/flyteorg/flyte/flytestdlib/atomic"
	"github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

const fakeCacheItemValueLimit = 10

type fakeCacheItem struct {
	val        int
	isTerminal bool
}

func (f fakeCacheItem) IsTerminal() bool {
	return f.isTerminal
}

type terminalCacheItem struct {
	val int
}

func (t terminalCacheItem) IsTerminal() bool {
	return true
}

func syncFakeItem(_ context.Context, batch Batch) ([]ItemSyncResponse, error) {
	items := make([]ItemSyncResponse, 0, len(batch))
	for _, obj := range batch {
		item := obj.GetItem().(fakeCacheItem)
		if item.val == fakeCacheItemValueLimit {
			// After the item has gone through ten update cycles, leave it unchanged
			continue
		}
		isTerminal := false
		if item.val == fakeCacheItemValueLimit-1 {
			isTerminal = true
		}
		items = append(items, ItemSyncResponse{
			ID: obj.GetID(),
			Item: fakeCacheItem{
				val:        item.val + 1,
				isTerminal: isTerminal,
			},
			Action: Update,
		})
	}

	return items, nil
}

func syncTerminalItem(_ context.Context, batch Batch) ([]ItemSyncResponse, error) {
	panic("This should never be called")
}

type panickingSyncer struct {
	callCount atomic.Int32
}

func (p *panickingSyncer) sync(_ context.Context, _ Batch) ([]ItemSyncResponse, error) {
	p.callCount.Inc()
	panic("testing")
}

func TestCacheFour(t *testing.T) {
	testResyncPeriod := 5 * time.Second
	rateLimiter := workqueue.DefaultControllerRateLimiter()
	fakeClock := testingclock.NewFakeClock(time.Now())

	t.Run("normal operation", func(t *testing.T) {
		// the size of the cache is at least as large as the number of items we're storing
		cache, err := newAutoRefreshCacheWithClock("fake1", syncFakeItem, rateLimiter, testResyncPeriod, 10, 10, promutils.NewTestScope(), fakeClock)
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		assert.NoError(t, cache.Start(ctx))

		// Create ten items in the cache
		for i := 1; i <= 10; i++ {
			_, err := cache.GetOrCreate(fmt.Sprintf("%d", i), fakeCacheItem{
				val: 0,
			})
			assert.NoError(t, err)
		}

		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			// trigger periodic sync
			fakeClock.Step(testResyncPeriod)

			for i := 1; i <= 10; i++ {
				item, err := cache.Get(fmt.Sprintf("%d", i))
				assert.NoError(c, err)
				assert.Equal(c, 10, item.(fakeCacheItem).val)
			}
		}, 3*time.Second, time.Millisecond)
		cancel()
	})

	t.Run("Not Found", func(t *testing.T) {
		// the size of the cache is at least as large as the number of items we're storing
		cache, err := newAutoRefreshCacheWithClock("fake2", syncFakeItem, rateLimiter, testResyncPeriod, 10, 2, promutils.NewTestScope(), fakeClock)
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		assert.NoError(t, cache.Start(ctx))

		// Create ten items in the cache
		for i := 1; i <= 10; i++ {
			_, err := cache.GetOrCreate(fmt.Sprintf("%d", i), fakeCacheItem{
				val: 0,
			})
			assert.NoError(t, err)
		}

		notFound := 0
		for i := 1; i <= 10; i++ {
			_, err := cache.Get(fmt.Sprintf("%d", i))
			if err != nil && errors.IsCausedBy(err, ErrNotFound) {
				notFound++
			}
		}

		assert.Equal(t, 8, notFound)

		cancel()
	})

	t.Run("Enqueue nothing", func(t *testing.T) {
		cache, err := newAutoRefreshCacheWithClock("fake3", syncTerminalItem, rateLimiter, testResyncPeriod, 10, 2, promutils.NewTestScope(), fakeClock)
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		assert.NoError(t, cache.Start(ctx))

		// Create ten items in the cache
		for i := 1; i <= 10; i++ {
			_, err := cache.GetOrCreate(fmt.Sprintf("%d", i), terminalCacheItem{
				val: 0,
			})
			assert.NoError(t, err)
		}

		// Enqueue first batch
		fakeClock.Step(testResyncPeriod)
		// If the cache tries to enqueue the item, a panic will be thrown.
		fakeClock.Step(testResyncPeriod)

		cancel()
	})

	t.Run("Test update and delete cache", func(t *testing.T) {
		cache, err := newAutoRefreshCacheWithClock("fake3", syncTerminalItem, rateLimiter, testResyncPeriod, 10, 2, promutils.NewTestScope(), fakeClock)
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		assert.NoError(t, cache.Start(ctx))

		itemID := "dummy_id"
		_, err = cache.GetOrCreate(itemID, terminalCacheItem{
			val: 0,
		})
		assert.NoError(t, err)

		// If the cache tries to enqueue the item, a panic will be thrown.
		fakeClock.Step(testResyncPeriod)

		err = cache.DeleteDelayed(itemID)
		assert.NoError(t, err)

		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			// trigger a sync
			fakeClock.Step(testResyncPeriod)

			item, err := cache.Get(itemID)
			assert.Nil(c, item)
			assert.Error(c, err)
		}, 3*time.Second, time.Millisecond)

		cancel()
	})

	t.Run("Test panic on sync and shutdown", func(t *testing.T) {
		syncer := &panickingSyncer{}
		cache, err := newAutoRefreshCacheWithClock("fake3", syncer.sync, rateLimiter, testResyncPeriod, 10, 2, promutils.NewTestScope(), fakeClock)
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		assert.NoError(t, cache.Start(ctx))

		itemID := "dummy_id"
		_, err = cache.GetOrCreate(itemID, fakeCacheItem{
			val: 0,
		})
		assert.NoError(t, err)

		// wait for all workers to run
		assert.Eventually(t, func() bool {
			// trigger a sync
			fakeClock.Step(testResyncPeriod)

			return syncer.callCount.Load() == int32(10)
		}, 5*time.Second, time.Millisecond)

		cancel()
	})
}

func TestQueueBuildUp(t *testing.T) {
	testResyncPeriod := time.Hour
	rateLimiter := workqueue.DefaultControllerRateLimiter()
	fakeClock := testingclock.NewFakeClock(time.Now())

	syncCount := atomic.NewInt32(0)
	m := sync.Map{}
	alwaysFailing := func(ctx context.Context, batch Batch) (
		updatedBatch []ItemSyncResponse, err error) {
		assert.Len(t, batch, 1)
		_, existing := m.LoadOrStore(batch[0].GetID(), 0)
		assert.False(t, existing, "Saw %v before", batch[0].GetID())
		if existing {
			t.FailNow()
		}

		syncCount.Inc()
		return nil, fmt.Errorf("expected error")
	}

	size := 100
	cache, err := newAutoRefreshCacheWithClock("fake2", alwaysFailing, rateLimiter, testResyncPeriod, 10, size, promutils.NewTestScope(), fakeClock)
	assert.NoError(t, err)

	ctx := context.Background()
	ctx, cancelNow := context.WithCancel(ctx)
	defer cancelNow()

	for i := 0; i < size; i++ {
		_, err := cache.GetOrCreate(strconv.Itoa(i), fakeCacheItem{val: 3})
		assert.NoError(t, err)
	}

	assert.NoError(t, cache.Start(ctx))

	// wait for all workers to run
	assert.Eventually(t, func() bool {
		// trigger a sync and unlock the work queue
		fakeClock.Step(time.Millisecond)

		return syncCount.Load() == int32(size)
	}, 5*time.Second, time.Millisecond)
}

func TestInProcessing(t *testing.T) {

	syncPeriod := time.Millisecond
	fakeClock := testingclock.NewFakeClock(time.Now())
	cache := &autoRefresh{
		processing: &sync.Map{},
		syncPeriod: syncPeriod,
		clock:      fakeClock,
	}

	assert.False(t, cache.inProcessing("test"))

	cache.processing.Store("test", time.Now())
	assert.True(t, cache.inProcessing("test"))

	cache.processing.Store("test1", time.Now().Add(syncPeriod*-11))
	assert.False(t, cache.inProcessing("test1"))
	_, found := cache.processing.Load("test1")
	assert.False(t, found)
}
