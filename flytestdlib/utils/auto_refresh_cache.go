package utils

import (
	"context"
	"sync"
	"time"

	"github.com/lyft/flytestdlib/logger"
	"k8s.io/apimachinery/pkg/util/wait"
)

// AutoRefreshCache with regular GetOrCreate and Delete along with background asynchronous refresh. Caller provides
// callbacks for create, refresh and delete item.
// The cache doesn't provide apis to update items.
type AutoRefreshCache interface {
	// starts background refresh of items
	Start(ctx context.Context)

	// Get item by id if exists else null
	Get(id string) CacheItem

	// Get object if exists else create it
	GetOrCreate(item CacheItem) (CacheItem, error)
}

type CacheItem interface {
	ID() string
}

type CacheSyncItem func(ctx context.Context, obj CacheItem) (CacheItem, error)

func NewAutoRefreshCache(syncCb CacheSyncItem, syncRateLimiter RateLimiter, resyncPeriod time.Duration) AutoRefreshCache {
	cache := &autoRefreshCache{
		syncCb:          syncCb,
		syncRateLimiter: syncRateLimiter,
		resyncPeriod:    resyncPeriod,
	}

	return cache
}

// Thread-safe general purpose auto-refresh cache that watches for updates asynchronously for the keys after they are added to
// the cache. An item can be inserted only once.
//
// Get reads from sync.map while refresh is invoked on a snapshot of keys. Cache eventually catches up on deleted items.
//
// Sync is run as a fixed-interval-scheduled-task, and is skipped if sync from previous cycle is still running.
type autoRefreshCache struct {
	syncCb          CacheSyncItem
	syncMap         sync.Map
	syncRateLimiter RateLimiter
	resyncPeriod    time.Duration
}

func (w *autoRefreshCache) Start(ctx context.Context) {
	go wait.Until(func() { w.sync(ctx) }, w.resyncPeriod, ctx.Done())
}

func (w *autoRefreshCache) Get(id string) CacheItem {
	if val, ok := w.syncMap.Load(id); ok {
		return val.(CacheItem)
	}
	return nil
}

// Return the item if exists else create it.
// Create should be invoked only once. recreating the object is not supported.
func (w *autoRefreshCache) GetOrCreate(item CacheItem) (CacheItem, error) {
	if val, ok := w.syncMap.Load(item.ID()); ok {
		return val.(CacheItem), nil
	}

	w.syncMap.Store(item.ID(), item)
	return item, nil
}

func (w *autoRefreshCache) sync(ctx context.Context) {
	w.syncMap.Range(func(key, value interface{}) bool {
		if w.syncRateLimiter != nil {
			err := w.syncRateLimiter.Wait(ctx)
			if err != nil {
				logger.Warnf(ctx, "unexpected failure in rate-limiter wait %v", key)
				return true
			}
		}
		item, err := w.syncCb(ctx, value.(CacheItem))
		if err != nil {
			logger.Error(ctx, "failed to get latest copy of the item %v", key)
		}

		if item == nil {
			w.syncMap.Delete(key)
		} else {
			w.syncMap.Store(key, item)
		}

		return true
	})
}
