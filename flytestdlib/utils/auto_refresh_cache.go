package utils

import (
	"context"
	"time"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	lru "github.com/hashicorp/golang-lru"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/util/wait"
)

//go:generate mockery -all -case=underscore

// AutoRefreshCache with regular GetOrCreate and Delete along with background asynchronous refresh. Caller provides
// callbacks for create, refresh and delete item.
// The cache doesn't provide apis to update items.
// Deprecated: This utility is deprecated, it has been refactored and moved into `cache` package.
type AutoRefreshCache interface {
	// starts background refresh of items
	Start(ctx context.Context)

	// Get item by id if exists else null
	Get(id string) CacheItem

	// Get object if exists else create it
	GetOrCreate(item CacheItem) (CacheItem, error)
}

// Deprecated: This utility is deprecated, it has been refactored and moved into `cache` package.
type CacheItem interface {
	ID() string
}

// Possible actions for the cache to take as a result of running the sync function on any given cache item
// Deprecated: This utility is deprecated, it has been refactored and moved into `cache` package.
type CacheSyncAction int

const (
	Unchanged CacheSyncAction = iota

	// The item returned has been updated and should be updated in the cache
	Update

	// The item should be removed from the cache
	Delete
)

// CacheSyncItem is a func type. Your implementation of this function for your cache instance is responsible for returning
// The new CacheItem what action should be taken.  The sync function has no insight into your object, and needs to be
// told explicitly if the new item is different from the old one.
// Deprecated: This utility is deprecated, it has been refactored and moved into `cache` package.
type CacheSyncItem func(ctx context.Context, obj CacheItem) (
	newItem CacheItem, result CacheSyncAction, err error)

func getEvictionFunction(counter prometheus.Counter) func(key interface{}, value interface{}) {
	return func(_ interface{}, _ interface{}) {
		counter.Inc()
	}
}

// Deprecated: This utility is deprecated, it has been refactored and moved into `cache` package.
func NewAutoRefreshCache(syncCb CacheSyncItem, syncRateLimiter RateLimiter, resyncPeriod time.Duration,
	size int, scope promutils.Scope) (AutoRefreshCache, error) {

	// If a scope is specified, we'll add a function to log a metric when an object gets evicted
	var evictionFunction func(key interface{}, value interface{})
	if scope != nil {
		counter := scope.MustNewCounter("lru_evictions", "Counter for evictions from LRU")
		evictionFunction = getEvictionFunction(counter)
	}
	lruCache, err := lru.NewWithEvict(size, evictionFunction)
	if err != nil {
		return nil, err
	}

	cache := &autoRefreshCache{
		syncCb:          syncCb,
		lruMap:          lruCache,
		syncRateLimiter: syncRateLimiter,
		resyncPeriod:    resyncPeriod,
		scope:           scope,
	}

	return cache, nil
}

// Thread-safe general purpose auto-refresh cache that watches for updates asynchronously for the keys after they are added to
// the cache. An item can be inserted only once.
//
// Get reads from sync.map while refresh is invoked on a snapshot of keys. Cache eventually catches up on deleted items.
//
// Sync is run as a fixed-interval-scheduled-task, and is skipped if sync from previous cycle is still running.
type autoRefreshCache struct {
	syncCb          CacheSyncItem
	lruMap          *lru.Cache
	syncRateLimiter RateLimiter
	resyncPeriod    time.Duration
	scope           promutils.Scope
}

func (w *autoRefreshCache) Start(ctx context.Context) {
	go wait.Until(func() { w.sync(ctx) }, w.resyncPeriod, ctx.Done())
}

func (w *autoRefreshCache) Get(id string) CacheItem {
	if val, ok := w.lruMap.Get(id); ok {
		return val.(CacheItem)
	}
	return nil
}

// Return the item if exists else create it.
// Create should be invoked only once. recreating the object is not supported.
func (w *autoRefreshCache) GetOrCreate(item CacheItem) (CacheItem, error) {
	if val, ok := w.lruMap.Get(item.ID()); ok {
		return val.(CacheItem), nil
	}

	w.lruMap.Add(item.ID(), item)
	return item, nil
}

// This function is called internally by its own timer. Roughly, it will list keys and for each of the keys, call
// syncCb, which tells us if the item has been updated. If it has, then do a remove followed by an add. We can get away
// with this because it is guaranteed that this loop will run to completion before the next one begins.
//
// What happens when the number of things that a user is trying to keep track of exceeds the size
// of the cache?  Trivial case where the cache is size 1 and we're trying to keep track of two things.
// * Plugin asks for update on item 1 - cache evicts item 2, stores 1 and returns it unchanged
// * Plugin asks for update on item 2 - cache evicts item 1, stores 2 and returns it unchanged
// * Sync loop updates item 2, repeat
func (w *autoRefreshCache) sync(ctx context.Context) {
	keys := w.lruMap.Keys()
	for _, k := range keys {
		// If not ok, it means evicted between the item was evicted between getting the keys and this update loop
		// which is fine, we can just ignore.
		if value, ok := w.lruMap.Peek(k); ok {
			newItem, result, err := w.syncCb(ctx, value.(CacheItem))
			if err != nil {
				logger.Errorf(ctx, "failed to get latest copy of the item %v, error: %s", k, err)
			}

			if result == Update {
				w.lruMap.Add(k, newItem)
			} else if result == Delete {
				w.lruMap.Remove(k)
			}
		}
	}
}
