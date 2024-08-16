package cache

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"

	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type ItemID = string
type Batch = []ItemWrapper

const (
	ErrNotFound errors.ErrorCode = "NOT_FOUND"
)

//go:generate mockery -all

// AutoRefresh with regular GetOrCreate and Delete along with background asynchronous refresh. Caller provides
// callbacks for create, refresh and delete item.
// The cache doesn't provide apis to update items.
type AutoRefresh interface {
	// Starts background refresh of items. To shutdown the cache, cancel the context.
	Start(ctx context.Context) error

	// Get item by id.
	Get(id ItemID) (Item, error)

	// Get object if exists else create it.
	GetOrCreate(id ItemID, item Item) (Item, error)

	// DeleteDelayed queues an item for deletion. It Will get deleted as part of the next Sync cycle. Until the next sync
	// cycle runs, Get and GetOrCreate will continue to return the Item in its previous state.
	DeleteDelayed(id ItemID) error
}

type metrics struct {
	SyncErrors  prometheus.Counter
	Evictions   prometheus.Counter
	SyncLatency promutils.StopWatch
	CacheHit    prometheus.Counter
	CacheMiss   prometheus.Counter
	Size        prometheus.Gauge
	scope       promutils.Scope
}

type Item interface {
	IsTerminal() bool
}

// Items are wrapped inside an ItemWrapper to be stored in the cache.
type ItemWrapper interface {
	GetID() ItemID
	GetItem() Item
}

// Represents the response for the sync func
type ItemSyncResponse struct {
	ID     ItemID
	Item   Item
	Action SyncAction
}

// Possible actions for the cache to take as a result of running the sync function on any given cache item
type SyncAction int

const (
	Unchanged SyncAction = iota

	// The item returned has been updated and should be updated in the cache
	Update
)

// SyncFunc func type. Your implementation of this function for your cache instance is responsible for returning
// The new Item and what action should be taken.  The sync function has no insight into your object, and needs to be
// told explicitly if the new item is different from the old one.
type SyncFunc func(ctx context.Context, batch Batch) (
	updatedBatch []ItemSyncResponse, err error)

// CreateBatchesFunc is a func type. Your implementation of this function for your cache instance is responsible for
// subdividing the list of cache items into batches.
type CreateBatchesFunc func(ctx context.Context, snapshot []ItemWrapper) (batches []Batch, err error)

type itemWrapper struct {
	id   ItemID
	item Item
}

func (i itemWrapper) GetID() ItemID {
	return i.id
}

func (i itemWrapper) GetItem() Item {
	return i.item
}

// Thread-safe general purpose auto-refresh cache that watches for updates asynchronously for the keys after they are added to
// the cache. An item can be inserted only once.
//
// Get reads from sync.map while refresh is invoked on a snapshot of keys. Cache eventually catches up on deleted items.
//
// Sync is run as a fixed-interval-scheduled-task, and is skipped if sync from previous cycle is still running.
type autoRefresh struct {
	name            string
	metrics         metrics
	syncCb          SyncFunc
	createBatchesCb CreateBatchesFunc
	lruMap          *lru.Cache
	// Items that are currently being processed are in the processing set.
	// It will prevent the same item from being processed multiple times by different workers.
	processing  *sync.Map
	toDelete    *syncSet
	syncPeriod  time.Duration
	workqueue   workqueue.RateLimitingInterface
	parallelizm int
	lock        sync.RWMutex
}

func getEvictionFunction(counter prometheus.Counter) func(key interface{}, value interface{}) {
	return func(_ interface{}, _ interface{}) {
		counter.Inc()
	}
}

func SingleItemBatches(_ context.Context, snapshot []ItemWrapper) (batches []Batch, err error) {
	res := make([]Batch, 0, len(snapshot))
	for _, item := range snapshot {
		res = append(res, Batch{item})
	}

	return res, nil
}

func newMetrics(scope promutils.Scope) metrics {
	return metrics{
		SyncErrors:  scope.MustNewCounter("sync_errors", "Counter for sync errors."),
		Evictions:   scope.MustNewCounter("lru_evictions", "Counter for evictions from LRU."),
		SyncLatency: scope.MustNewStopWatch("latency", "Latency for sync operations.", time.Millisecond),
		CacheHit:    scope.MustNewCounter("cache_hit", "Counter for cache hits."),
		CacheMiss:   scope.MustNewCounter("cache_miss", "Counter for cache misses."),
		Size:        scope.MustNewGauge("size", "Current size of the cache"),
		scope:       scope,
	}
}

func (w *autoRefresh) Start(ctx context.Context) error {
	for i := 0; i < w.parallelizm; i++ {
		go func(ctx context.Context) {
			err := w.sync(ctx)
			if err != nil {
				logger.Errorf(ctx, "Failed to sync. Error: %v", err)
			}
		}(contextutils.WithGoroutineLabel(ctx, fmt.Sprintf("%v-worker-%v", w.name, i)))
	}

	enqueueCtx := contextutils.WithGoroutineLabel(ctx, fmt.Sprintf("%v-enqueue", w.name))

	go wait.Until(func() {
		err := w.enqueueBatches(enqueueCtx)
		if err != nil {
			logger.Errorf(enqueueCtx, "Failed to enqueue. Error: %v", err)
		}
	}, w.syncPeriod, enqueueCtx.Done())

	return nil
}

// Update updates the item only if it exists in the cache, return true if we updated the item.
func (w *autoRefresh) Update(id ItemID, item Item) (ok bool) {
	w.lock.Lock()
	defer w.lock.Unlock()
	ok = w.lruMap.Contains(id)
	if ok {
		w.lruMap.Add(id, item)
	}
	return ok
}

// Delete deletes the item from the cache if it exists.
func (w *autoRefresh) Delete(key interface{}) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.toDelete.Remove(key)
	w.lruMap.Remove(key)
}

func (w *autoRefresh) Get(id ItemID) (Item, error) {
	if val, ok := w.lruMap.Get(id); ok {
		w.metrics.CacheHit.Inc()
		return val.(Item), nil
	}

	w.metrics.CacheMiss.Inc()
	return nil, errors.Errorf(ErrNotFound, "Item with id [%v] not found.", id)
}

// Return the item if exists else create it.
// Create should be invoked only once. recreating the object is not supported.
func (w *autoRefresh) GetOrCreate(id ItemID, item Item) (Item, error) {
	if val, ok := w.lruMap.Get(id); ok {
		w.metrics.CacheHit.Inc()
		return val.(Item), nil
	}

	w.lruMap.Add(id, item)
	w.metrics.CacheMiss.Inc()

	// It fixes cold start issue in the AutoRefreshCache by adding the item to the workqueue when it is created.
	// This way, the item will be processed without waiting for the next sync cycle (30s by default).
	batch := make([]ItemWrapper, 0, 1)
	batch = append(batch, itemWrapper{id: id, item: item})
	w.workqueue.AddRateLimited(&batch)
	w.processing.Store(id, time.Now())
	return item, nil
}

// DeleteDelayed queues an item for deletion. It Will get deleted as part of the next Sync cycle. Until the next sync
// cycle runs, Get and GetOrCreate will continue to return the Item in its previous state.
func (w *autoRefresh) DeleteDelayed(id ItemID) error {
	w.toDelete.Insert(id)
	return nil
}

// This function is called internally by its own timer. Roughly, it will list keys, create batches of keys based on
// createBatchesCb and, enqueue all the batches into the workqueue.
func (w *autoRefresh) enqueueBatches(ctx context.Context) error {
	keys := w.lruMap.Keys()
	w.metrics.Size.Set(float64(len(keys)))

	snapshot := make([]ItemWrapper, 0, len(keys))
	for _, k := range keys {
		if w.toDelete.Contains(k) {
			w.Delete(k)
			continue
		}
		// If not ok, it means evicted between the item was evicted between getting the keys and this update loop
		// which is fine, we can just ignore.
		if value, ok := w.lruMap.Peek(k); ok {
			if item, ok := value.(Item); !ok || (ok && !item.IsTerminal() && !w.inProcessing(k)) {
				snapshot = append(snapshot, itemWrapper{
					id:   k.(ItemID),
					item: value.(Item),
				})
			}
		}
	}

	batches, err := w.createBatchesCb(ctx, snapshot)
	if err != nil {
		return err
	}

	for _, batch := range batches {
		b := batch
		w.workqueue.AddRateLimited(&b)
		for i := 1; i < len(b); i++ {
			w.processing.Store(b[i].GetID(), time.Now())
		}
	}

	return nil
}

// There are w.parallelizm instances of this function running all the time, each one will:
// - Retrieve an item from the workqueue
// - For each batch of the keys, call syncCb, which tells us if the items have been updated
// -- If any has, then overwrite the item in the cache.
//
// What happens when the number of things that a user is trying to keep track of exceeds the size
// of the cache?  Trivial case where the cache is size 1 and we're trying to keep track of two things.
// * Plugin asks for update on item 1 - cache evicts item 2, stores 1 and returns it unchanged
// * Plugin asks for update on item 2 - cache evicts item 1, stores 2 and returns it unchanged
// * Sync loop updates item 2, repeat
func (w *autoRefresh) sync(ctx context.Context) (err error) {
	defer func() {
		var isErr bool
		rVal := recover()
		if rVal == nil {
			return
		}

		if err, isErr = rVal.(error); isErr {
			err = fmt.Errorf("worker panic'd and is shutting down. Error: %w with Stack: %v", err, string(debug.Stack()))
		} else {
			err = fmt.Errorf("worker panic'd and is shutting down. Panic value: %v with Stack: %v", rVal, string(debug.Stack()))
		}

		logger.Error(ctx, err)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			batch, shutdown := w.workqueue.Get()
			if shutdown {
				logger.Debugf(ctx, "Shutting down worker")
				return nil
			}
			// Since we create batches every time we sync, we will just remove the item from the queue here
			// regardless of whether it succeeded the sync or not.
			w.workqueue.Forget(batch)
			w.workqueue.Done(batch)

			newBatch := make(Batch, 0, len(*batch.(*Batch)))
			for _, b := range *batch.(*Batch) {
				itemID := b.GetID()
				w.processing.Delete(itemID)
				item, ok := w.lruMap.Get(itemID)
				if !ok {
					logger.Debugf(ctx, "item with id [%v] not found in cache", itemID)
					continue
				}
				if item.(Item).IsTerminal() {
					logger.Debugf(ctx, "item with id [%v] is terminal", itemID)
					continue
				}
				newBatch = append(newBatch, b)
			}
			if len(newBatch) == 0 {
				continue
			}

			t := w.metrics.SyncLatency.Start()
			updatedBatch, err := w.syncCb(ctx, newBatch)

			if err != nil {
				w.metrics.SyncErrors.Inc()
				logger.Errorf(ctx, "failed to get latest copy of a batch. Error: %v", err)
				t.Stop()
				continue
			}

			for _, item := range updatedBatch {
				if item.Action == Update {
					// Updates an existing item.
					w.Update(item.ID, item.Item)
				}
			}

			w.toDelete.Range(func(key interface{}) bool {
				w.Delete(key)
				return true
			})

			t.Stop()
		}
	}
}

// Checks if the item is currently being processed and returns false if the item has been in processing for too long
func (w *autoRefresh) inProcessing(key interface{}) bool {
	item, found := w.processing.Load(key)
	if found {
		// handle potential race conditions where the item is in processing but not in the workqueue
		if timeItem, ok := item.(time.Time); ok && time.Since(timeItem) > (w.syncPeriod*5) {
			w.processing.Delete(key)
			return false
		}
		return true
	}
	return false
}

// Instantiates a new AutoRefresh Cache that syncs items in batches.
func NewAutoRefreshBatchedCache(name string, createBatches CreateBatchesFunc, syncCb SyncFunc, syncRateLimiter workqueue.RateLimiter,
	resyncPeriod time.Duration, parallelizm, size int, scope promutils.Scope) (AutoRefresh, error) {

	metrics := newMetrics(scope)
	lruCache, err := lru.NewWithEvict(size, getEvictionFunction(metrics.Evictions))
	if err != nil {
		return nil, err
	}

	cache := &autoRefresh{
		name:            name,
		metrics:         metrics,
		parallelizm:     parallelizm,
		createBatchesCb: createBatches,
		syncCb:          syncCb,
		lruMap:          lruCache,
		processing:      &sync.Map{},
		toDelete:        newSyncSet(),
		syncPeriod:      resyncPeriod,
		workqueue:       workqueue.NewNamedRateLimitingQueue(syncRateLimiter, scope.CurrentScope()),
	}

	return cache, nil
}

// Instantiates a new AutoRefresh Cache that syncs items periodically.
func NewAutoRefreshCache(name string, syncCb SyncFunc, syncRateLimiter workqueue.RateLimiter, resyncPeriod time.Duration,
	parallelizm, size int, scope promutils.Scope) (AutoRefresh, error) {

	return NewAutoRefreshBatchedCache(name, SingleItemBatches, syncCb, syncRateLimiter, resyncPeriod, parallelizm, size, scope)
}
