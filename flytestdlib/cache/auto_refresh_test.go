package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"k8s.io/client-go/util/workqueue"

	"github.com/lyft/flytestdlib/errors"

	"github.com/lyft/flytestdlib/promutils"

	"github.com/stretchr/testify/assert"
)

const fakeCacheItemValueLimit = 10

type fakeCacheItem struct {
	val int
}

func syncFakeItem(_ context.Context, batch Batch) ([]ItemSyncResponse, error) {
	items := make([]ItemSyncResponse, 0, len(batch))
	for _, obj := range batch {
		item := obj.GetItem().(fakeCacheItem)
		if item.val == fakeCacheItemValueLimit {
			// After the item has gone through ten update cycles, leave it unchanged
			continue
		}

		items = append(items, ItemSyncResponse{
			ID: obj.GetID(),
			Item: fakeCacheItem{
				val: item.val + 1,
			},
			Action: Update,
		})
	}

	return items, nil
}

func TestCacheTwo(t *testing.T) {
	testResyncPeriod := time.Millisecond
	rateLimiter := workqueue.DefaultControllerRateLimiter()

	t.Run("normal operation", func(t *testing.T) {
		// the size of the cache is at least as large as the number of items we're storing
		cache, err := NewAutoRefreshCache(syncFakeItem, rateLimiter, testResyncPeriod, 10, 10, promutils.NewTestScope())
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

		// Wait half a second for all resync periods to complete
		time.Sleep(500 * time.Millisecond)
		for i := 1; i <= 10; i++ {
			item, err := cache.Get(fmt.Sprintf("%d", i))
			assert.NoError(t, err)
			assert.Equal(t, 10, item.(fakeCacheItem).val)
		}
		cancel()
	})

	t.Run("Not Found", func(t *testing.T) {
		// the size of the cache is at least as large as the number of items we're storing
		cache, err := NewAutoRefreshCache(syncFakeItem, rateLimiter, testResyncPeriod, 10, 2, promutils.NewTestScope())
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
}
