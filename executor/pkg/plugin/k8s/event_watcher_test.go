package k8s

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

func testEvent(name string, regardingUID k8stypes.UID, createdAt time.Time) *eventsv1.Event {
	return &eventsv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "ns",
			Name:              name,
			CreationTimestamp: metav1.NewTime(createdAt),
		},
		Regarding: corev1.ObjectReference{
			Namespace: "ns",
			Name:      "pod",
			Kind:      "Pod",
			UID:       regardingUID,
		},
		Reason: "GPUXidError",
		Note:   "Xid 79",
	}
}

func TestEventWatcherOnAddRecordsIdentityAndObservation(t *testing.T) {
	watcher := &controllerRuntimeEventWatcher{}
	createdAt := time.Now().Add(-time.Hour)
	lastObserved := time.Now().Add(-time.Minute)

	event := testEvent("event-1", "pod-uid", createdAt)
	event.Series = &eventsv1.EventSeries{Count: 3, LastObservedTime: metav1.NewMicroTime(lastObserved)}
	watcher.OnAdd(event, false)

	events := watcher.List(watchedObjectKey{Namespace: "ns", Name: "pod", Kind: "Pod"}, time.Time{}, time.Time{})
	require.Len(t, events, 1)
	assert.Equal(t, k8stypes.UID("pod-uid"), events[0].RegardingUID)
	assert.Equal(t, createdAt.UTC(), events[0].CreatedAt.UTC())
	assert.WithinDuration(t, lastObserved, events[0].LastObservedAt, time.Microsecond)
}

func TestEventWatcherOnAddFallsBackToCreationTimestamp(t *testing.T) {
	watcher := &controllerRuntimeEventWatcher{}
	createdAt := time.Now().Add(-time.Hour)

	watcher.OnAdd(testEvent("event-1", "pod-uid", createdAt), false)

	events := watcher.List(watchedObjectKey{Namespace: "ns", Name: "pod", Kind: "Pod"}, time.Time{}, time.Time{})
	require.Len(t, events, 1)
	assert.Equal(t, createdAt.UTC(), events[0].LastObservedAt.UTC())
}

func TestEventWatcherOnUpdateRefreshesLastObservedAt(t *testing.T) {
	watcher := &controllerRuntimeEventWatcher{}
	createdAt := time.Now().Add(-time.Hour)
	key := watchedObjectKey{Namespace: "ns", Name: "pod", Kind: "Pod"}

	watcher.OnAdd(testEvent("event-1", "pod-uid", createdAt), false)

	first := watcher.List(key, time.Time{}, time.Time{})
	require.Len(t, first, 1)
	recordedAt := first[0].RecordedAt

	// The same event, seen again: Kubernetes aggregates the repeat into the object it
	// already has, so it reaches the watcher as an update.
	lastObserved := time.Now().Add(-time.Minute)
	repeat := testEvent("event-1", "pod-uid", createdAt)
	repeat.Series = &eventsv1.EventSeries{Count: 2, LastObservedTime: metav1.NewMicroTime(lastObserved)}
	watcher.OnUpdate(nil, repeat)

	events := watcher.List(key, time.Time{}, time.Time{})
	require.Len(t, events, 1, "the repeat refreshes the entry instead of adding one")
	assert.WithinDuration(t, lastObserved, events[0].LastObservedAt, time.Microsecond)
	// The watermarks a caller consumes incrementally have to stay where they were, or
	// the refreshed event would be handed out a second time.
	assert.Equal(t, createdAt.UTC(), events[0].CreatedAt.UTC())
	assert.Equal(t, recordedAt, events[0].RecordedAt)
	assert.Empty(t, watcher.List(key, events[0].CreatedAt, events[0].RecordedAt))
}

func TestEventWatcherOnUpdateKeepsTheFreshestObservation(t *testing.T) {
	watcher := &controllerRuntimeEventWatcher{}
	createdAt := time.Now().Add(-time.Hour)
	lastObserved := time.Now().Add(-time.Minute)
	key := watchedObjectKey{Namespace: "ns", Name: "pod", Kind: "Pod"}

	fresh := testEvent("event-1", "pod-uid", createdAt)
	fresh.Series = &eventsv1.EventSeries{Count: 2, LastObservedTime: metav1.NewMicroTime(lastObserved)}
	watcher.OnAdd(fresh, false)

	// An update that carries nothing newer must not walk the observation back.
	watcher.OnUpdate(nil, testEvent("event-1", "pod-uid", createdAt))

	events := watcher.List(key, time.Time{}, time.Time{})
	require.Len(t, events, 1)
	assert.WithinDuration(t, lastObserved, events[0].LastObservedAt, time.Microsecond)
}

func cachedBuckets(w *controllerRuntimeEventWatcher) int {
	count := 0
	w.objectCache.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

func TestEventWatcherDropsBucketWhenLastEventIsDeleted(t *testing.T) {
	watcher := &controllerRuntimeEventWatcher{}
	createdAt := time.Now()

	watcher.OnAdd(testEvent("event-1", "pod-uid", createdAt), false)
	watcher.OnAdd(testEvent("event-2", "pod-uid", createdAt), false)
	require.Equal(t, 1, cachedBuckets(watcher))

	// The bucket stays while any event for the object is still cached.
	watcher.OnDelete(testEvent("event-1", "pod-uid", createdAt))
	assert.Equal(t, 1, cachedBuckets(watcher))

	// The last deletion takes the bucket with it, so an object whose events have all
	// expired stops costing anything.
	watcher.OnDelete(testEvent("event-2", "pod-uid", createdAt))
	assert.Equal(t, 0, cachedBuckets(watcher))
	assert.Empty(t, watcher.List(watchedObjectKey{Namespace: "ns", Name: "pod", Kind: "Pod"}, time.Time{}, time.Time{}))
}

func TestEventWatcherDoesNotGrowAcrossObjectChurn(t *testing.T) {
	watcher := &controllerRuntimeEventWatcher{}
	createdAt := time.Now()

	// Every object the cluster emits an event about used to leave a bucket behind for
	// the life of the process; the count now follows only what is still live.
	for i := 0; i < 1000; i++ {
		event := testEvent("event-1", "pod-uid", createdAt)
		event.Regarding.Name = fmt.Sprintf("pod-%d", i)
		watcher.OnAdd(event, false)
	}
	require.Equal(t, 1000, cachedBuckets(watcher))

	for i := 0; i < 1000; i++ {
		event := testEvent("event-1", "pod-uid", createdAt)
		event.Regarding.Name = fmt.Sprintf("pod-%d", i)
		watcher.OnDelete(event)
	}
	assert.Equal(t, 0, cachedBuckets(watcher))
}

func TestEventWatcherStoreSurvivesConcurrentEviction(t *testing.T) {
	watcher := &controllerRuntimeEventWatcher{}
	createdAt := time.Now()

	// A store that loses its bucket to a concurrent eviction has to land in a fresh one
	// rather than write into a bucket nothing can reach again. Run under -race.
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			watcher.OnAdd(testEvent("event-keep", "pod-uid", createdAt), false)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			watcher.OnAdd(testEvent("event-churn", "pod-uid", createdAt), false)
			watcher.OnDelete(testEvent("event-churn", "pod-uid", createdAt))
		}()
	}
	wg.Wait()

	// The surviving event must still be reachable: an eviction that raced a store must
	// not have swallowed it.
	watcher.OnAdd(testEvent("event-keep", "pod-uid", createdAt), false)
	events := watcher.List(watchedObjectKey{Namespace: "ns", Name: "pod", Kind: "Pod"}, time.Time{}, time.Time{})
	require.Len(t, events, 1)
	assert.Equal(t, k8stypes.UID("pod-uid"), events[0].RegardingUID)
	assert.Equal(t, 1, cachedBuckets(watcher))
}
