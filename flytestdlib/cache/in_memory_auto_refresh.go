package cache

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/utils/clock"

	"github.com/flyteorg/flyte/flytestdlib/atomic"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type metrics struct {
	SyncErrors  prometheus.Counter
	Evictions   prometheus.Counter
	SyncLatency promutils.StopWatch
	CacheHit    prometheus.Counter
	CacheMiss   prometheus.Counter
	Size        prometheus.Gauge
	scope       promutils.Scope
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

func getEvictionFunction(counter prometheus.Counter) func(key interface{}, value interface{}) {
	return func(_ interface{}, _ interface{}) {
		counter.Inc()
	}
}

// Options are configurable options for the InMemoryAutoRefresh.
type Options struct {
	clock           clock.WithTicker
	createBatchesCb CreateBatchesFunc
	syncOnCreate    bool
}

// WithClock configures the clock to use for time related operations. Mainly used for unit testing.
func WithClock(clock clock.WithTicker) Option {
	return func(mo *Options) {
		mo.clock = clock
	}
}

// WithCreateBatchesFunc configures how cache items should be batched for refresh. Defaults to single item batching.
func WithCreateBatchesFunc(createBatchesCb CreateBatchesFunc) Option {
	return func(mo *Options) {
		mo.createBatchesCb = createBatchesCb
	}
}

// WithSyncOnCreate configures whether the cache will attempt to sync items upon creation or wait until the next
// sync interval. Disabling this can be useful when the cache is under high load and synchronization both frequently
// and in large batches. Defaults to true.
func WithSyncOnCreate(syncOnCreate bool) Option {
	return func(mo *Options) {
		mo.syncOnCreate = syncOnCreate
	}
}

func defaultOptions() *Options {
	opts := &Options{}
	WithClock(clock.RealClock{})(opts)
	WithCreateBatchesFunc(SingleItemBatches)(opts)
	WithSyncOnCreate(true)(opts)
	return opts
}

// Option for the KeyfuncProvider
type Option func(*Options)

// InMemoryAutoRefresh is an in-memory implementation of the AutoRefresh interface. It is a thread-safe general
// purpose auto-refresh cache that watches for updates asynchronously for the keys after they are added to
// the cache. An item can be inserted only once.
//
// Get reads from sync.map while refresh is invoked on a snapshot of keys. Cache eventually catches up on deleted items.
//
// Sync is run as a fixed-interval-scheduled-task, and is skipped if sync from previous cycle is still running.
type InMemoryAutoRefresh struct {
	name            string
	metrics         metrics
	syncCb          SyncFunc
	createBatchesCb CreateBatchesFunc
	lruMap          *lru.Cache
	// Items that are currently being processed are in the processing set.
	// It will prevent the same item from being processed multiple times by different workers.
	processing         *sync.Map
	toDelete           *syncSet
	syncPeriod         time.Duration
	workqueue          workqueue.RateLimitingInterface
	parallelizm        uint
	lock               sync.RWMutex
	clock              clock.Clock  // pluggable clock for unit testing
	syncCount          atomic.Int32 // internal sync counter for unit testing
	enqueueCount       atomic.Int32 // internal enqueue counter for unit testing
	enqueueLoopRunning atomic.Bool  // internal bool to ensure goroutines are running
	syncOnCreate       bool
}

// NewInMemoryAutoRefresh creates a new InMemoryAutoRefresh
func NewInMemoryAutoRefresh(
	name string,
	syncCb SyncFunc,
	syncRateLimiter workqueue.RateLimiter,
	resyncPeriod time.Duration,
	parallelizm uint,
	size uint,
	scope promutils.Scope,
	options ...Option,
) (*InMemoryAutoRefresh, error) {
	opts := defaultOptions()
	for _, option := range options {
		option(opts)
	}

	metrics := newMetrics(scope)
	// #nosec G115
	lruCache, err := lru.NewWithEvict(int(size), getEvictionFunction(metrics.Evictions))
	if err != nil {
		return nil, fmt.Errorf("creating LRU cache: %w", err)
	}

	cache := &InMemoryAutoRefresh{
		name:            name,
		metrics:         metrics,
		parallelizm:     parallelizm,
		createBatchesCb: opts.createBatchesCb,
		syncCb:          syncCb,
		lruMap:          lruCache,
		processing:      &sync.Map{},
		toDelete:        newSyncSet(),
		syncPeriod:      resyncPeriod,
		workqueue: workqueue.NewRateLimitingQueueWithConfig(syncRateLimiter, workqueue.RateLimitingQueueConfig{
			Name:  scope.CurrentScope(),
			Clock: opts.clock,
		}),
		clock:              opts.clock,
		syncCount:          atomic.NewInt32(0),
		enqueueCount:       atomic.NewInt32(0),
		enqueueLoopRunning: atomic.NewBool(false),
		syncOnCreate:       opts.syncOnCreate,
	}

	return cache, nil
}

func (w *InMemoryAutoRefresh) Start(ctx context.Context) error {
	for i := uint(0); i < w.parallelizm; i++ {
		go func(ctx context.Context) {
			err := w.sync(ctx)
			if err != nil {
				logger.Errorf(ctx, "Failed to sync. Error: %v", err)
			}
		}(contextutils.WithGoroutineLabel(ctx, fmt.Sprintf("%v-worker-%v", w.name, i)))
	}

	enqueueCtx := contextutils.WithGoroutineLabel(ctx, fmt.Sprintf("%v-enqueue", w.name))
	go w.enqueueLoop(enqueueCtx)

	return nil
}

func (w *InMemoryAutoRefresh) enqueueLoop(ctx context.Context) {
	timer := w.clock.NewTimer(w.syncPeriod)
	defer timer.Stop()

	w.enqueueLoopRunning.Store(true)

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C():
			err := w.enqueueBatches(ctx)
			if err != nil {
				logger.Errorf(ctx, "Failed to enqueue. Error: %v", err)
			}
			timer.Reset(w.syncPeriod)
		}
	}
}

// Update updates the item only if it exists in the cache, return true if we updated the item.
func (w *InMemoryAutoRefresh) Update(id ItemID, item Item) (ok bool) {
	w.lock.Lock()
	defer w.lock.Unlock()
	ok = w.lruMap.Contains(id)
	if ok {
		w.lruMap.Add(id, item)
	}
	return ok
}

// Delete deletes the item from the cache if it exists.
func (w *InMemoryAutoRefresh) Delete(key interface{}) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.toDelete.Remove(key)
	w.lruMap.Remove(key)
}

func (w *InMemoryAutoRefresh) Get(id ItemID) (Item, error) {
	if val, ok := w.lruMap.Get(id); ok {
		w.metrics.CacheHit.Inc()
		return val.(Item), nil
	}

	w.metrics.CacheMiss.Inc()
	return nil, errors.Errorf(ErrNotFound, "Item with id [%v] not found.", id)
}

// Return the item if exists else create it.
// Create should be invoked only once. recreating the object is not supported.
func (w *InMemoryAutoRefresh) GetOrCreate(id ItemID, item Item) (Item, error) {
	if val, ok := w.lruMap.Get(id); ok {
		w.metrics.CacheHit.Inc()
		return val.(Item), nil
	}

	w.lruMap.Add(id, item)
	w.metrics.CacheMiss.Inc()

	// It fixes cold start issue in the AutoRefreshCache by adding the item to the workqueue when it is created.
	// This way, the item will be processed without waiting for the next sync cycle (30s by default).
	if w.syncOnCreate {
		batch := make([]ItemWrapper, 0, 1)
		batch = append(batch, itemWrapper{id: id, item: item})
		w.workqueue.AddRateLimited(&batch)
		w.processing.Store(id, w.clock.Now())
	}
	return item, nil
}

// DeleteDelayed queues an item for deletion. It Will get deleted as part of the next Sync cycle. Until the next sync
// cycle runs, Get and GetOrCreate will continue to return the Item in its previous state.
func (w *InMemoryAutoRefresh) DeleteDelayed(id ItemID) error {
	w.toDelete.Insert(id)
	return nil
}

// This function is called internally by its own timer. Roughly, it will list keys, create batches of keys based on
// createBatchesCb and, enqueue all the batches into the workqueue.
func (w *InMemoryAutoRefresh) enqueueBatches(ctx context.Context) error {
	defer w.enqueueCount.Inc()

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
		for i := 0; i < len(b); i++ {
			w.processing.Store(b[i].GetID(), w.clock.Now())
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
func (w *InMemoryAutoRefresh) sync(ctx context.Context) (err error) {
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

		w.syncCount.Inc()
	}
}

// Checks if the item is currently being processed and returns false if the item has been in processing for too long
func (w *InMemoryAutoRefresh) inProcessing(key interface{}) bool {
	item, found := w.processing.Load(key)
	if found {
		// handle potential race conditions where the item is in processing but not in the workqueue
		if timeItem, ok := item.(time.Time); ok && w.clock.Since(timeItem) > (w.syncPeriod*5) {
			w.processing.Delete(key)
			return false
		}
		return true
	}
	return false
}

// Instantiates a new AutoRefresh Cache that syncs items in batches.
func NewAutoRefreshBatchedCache(name string, createBatches CreateBatchesFunc, syncCb SyncFunc, syncRateLimiter workqueue.RateLimiter,
	resyncPeriod time.Duration, parallelizm, size uint, scope promutils.Scope) (AutoRefresh, error) {
	return NewInMemoryAutoRefresh(name, syncCb, syncRateLimiter, resyncPeriod, parallelizm, size, scope, WithCreateBatchesFunc(createBatches))
}

// Instantiates a new AutoRefresh Cache that syncs items periodically.
func NewAutoRefreshCache(name string, syncCb SyncFunc, syncRateLimiter workqueue.RateLimiter, resyncPeriod time.Duration,
	parallelizm, size uint, scope promutils.Scope) (AutoRefresh, error) {
	return NewAutoRefreshBatchedCache(name, SingleItemBatches, syncCb, syncRateLimiter, resyncPeriod, parallelizm, size, scope)
}

// SingleItemBatches is a function that creates n batches of items, each with size 1
func SingleItemBatches(_ context.Context, snapshot []ItemWrapper) (batches []Batch, err error) {
	res := make([]Batch, 0, len(snapshot))
	for _, item := range snapshot {
		res = append(res, Batch{item})
	}

	return res, nil
}

// itemWrapper is an implementation of ItemWrapper
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
