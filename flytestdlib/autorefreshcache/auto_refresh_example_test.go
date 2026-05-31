package autorefreshcache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/util/workqueue"

	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

type ExampleItemStatus string

const (
	ExampleStatusNotStarted ExampleItemStatus = "Not-started"
	ExampleStatusStarted    ExampleItemStatus = "Started"
	ExampleStatusSucceeded  ExampleItemStatus = "Completed"
)

type ExampleCacheItem struct {
	status ExampleItemStatus
	id     string
}

func (e *ExampleCacheItem) IsTerminal() bool {
	return e.status == ExampleStatusSucceeded
}

func (e *ExampleCacheItem) ID() string {
	return e.id
}

type ExampleService struct {
	jobStatus map[string]ExampleItemStatus
	lock      sync.RWMutex
}

func newExampleService() *ExampleService {
	return &ExampleService{
		jobStatus: make(map[string]ExampleItemStatus),
		lock:      sync.RWMutex{},
	}
}

// advance the status to next, and return
func (f *ExampleService) getStatus(id string) *ExampleCacheItem {
	f.lock.Lock()
	defer f.lock.Unlock()
	if _, ok := f.jobStatus[id]; !ok {
		f.jobStatus[id] = ExampleStatusStarted
	}

	f.jobStatus[id] = ExampleStatusSucceeded
	return &ExampleCacheItem{f.jobStatus[id], id}
}

func ExampleNewAutoRefreshCache() {
	// This auto-refresh cache can be used for cases where keys are created by caller but processed by
	// an external service and we want to asynchronously keep track of its progress.
	exampleService := newExampleService()

	// define a sync method that the cache can use to auto-refresh in background
	syncItemCb := func(ctx context.Context, batch []ItemWrapper) ([]ItemSyncResponse, error) {
		updatedItems := make([]ItemSyncResponse, 0, len(batch))
		for _, obj := range batch {
			oldItem := obj.GetItem().(*ExampleCacheItem)
			newItem := exampleService.getStatus(oldItem.ID())
			if newItem.status != oldItem.status {
				updatedItems = append(updatedItems, ItemSyncResponse{
					ID:     oldItem.ID(),
					Item:   newItem,
					Action: Update,
				})
			}
		}

		return updatedItems, nil
	}

	// define resync period as time duration we want cache to refresh. We can go as low as we want but cache
	// would still be constrained by time it takes to run Sync call for each item.
	resyncPeriod := time.Millisecond

	// Since number of items in the cache is dynamic, rate limiter is our knob to control resources we spend on
	// sync.
	rateLimiter := workqueue.DefaultControllerRateLimiter()

	// since cache refreshes itself asynchronously, it may not notice that an object has been deleted immediately,
	// so users of the cache should have the delete logic aware of this shortcoming (eg. not-exists may be a valid
	// error during removal if based on status in cache).
	cache, err := NewAutoRefreshCache("my-cache", syncItemCb, rateLimiter, resyncPeriod, 10, 100, promutils.NewTestScope())
	if err != nil {
		panic(err)
	}

	// start the cache with a context that would be to stop the cache by cancelling the context
	ctx, cancel := context.WithCancel(context.Background())
	err = cache.Start(ctx)
	if err != nil {
		panic(err)
	}

	// creating objects that go through a couple of state transitions to reach the final state.
	item1 := &ExampleCacheItem{status: ExampleStatusNotStarted, id: "item1"}
	item2 := &ExampleCacheItem{status: ExampleStatusNotStarted, id: "item2"}
	_, err1 := cache.GetOrCreate(item1.id, item1)
	_, err2 := cache.GetOrCreate(item2.id, item2)
	if err1 != nil || err2 != nil {
		fmt.Printf("unexpected error in create; err1: %v, err2: %v", err1, err2)
	}

	// The cache refreshes asynchronously in the background, so item1 transitions to its terminal
	// status some time after it is created. Right after GetOrCreate it is already retrievable from
	// the cache in its initial state. See TestNewAutoRefreshCache for verification that the
	// background worker eventually advances item1 to its terminal status.
	item, err := cache.Get(item1.ID())
	if err != nil {
		fmt.Printf("unexpected error getting item1: %v", err)
	} else {
		fmt.Printf("item1 is in the cache with id %q", item.(*ExampleCacheItem).ID())
	}

	// stop the cache
	cancel()

	// Output:
	// item1 is in the cache with id "item1"
}

// TestNewAutoRefreshCache verifies that the background worker eventually advances an item to its
// terminal status. This lives in a test (rather than the example above) so that the wait can use
// require.Eventually, which needs a *testing.T that example functions do not have.
func TestNewAutoRefreshCache(t *testing.T) {
	exampleService := newExampleService()

	syncItemCb := func(ctx context.Context, batch []ItemWrapper) ([]ItemSyncResponse, error) {
		updatedItems := make([]ItemSyncResponse, 0, len(batch))
		for _, obj := range batch {
			oldItem := obj.GetItem().(*ExampleCacheItem)
			newItem := exampleService.getStatus(oldItem.ID())
			if newItem.status != oldItem.status {
				updatedItems = append(updatedItems, ItemSyncResponse{
					ID:     oldItem.ID(),
					Item:   newItem,
					Action: Update,
				})
			}
		}
		return updatedItems, nil
	}

	resyncPeriod := time.Millisecond
	rateLimiter := workqueue.DefaultControllerRateLimiter()
	cache, err := NewAutoRefreshCache("my-cache", syncItemCb, rateLimiter, resyncPeriod, 10, 100, promutils.NewTestScope())
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, cache.Start(ctx))

	item1 := &ExampleCacheItem{status: ExampleStatusNotStarted, id: "item1"}
	_, err = cache.GetOrCreate(item1.id, item1)
	require.NoError(t, err)

	// Poll until the background worker refreshes item1 to its terminal state.
	require.Eventually(t, func() bool {
		item, err := cache.Get(item1.ID())
		return err == nil && item.(*ExampleCacheItem).status == ExampleStatusSucceeded
	}, 3*time.Second, resyncPeriod, "expected item1 to reach terminal status %q", ExampleStatusSucceeded)
}
