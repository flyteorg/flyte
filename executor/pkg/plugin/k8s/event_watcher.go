package k8s

import (
	"context"
	"sort"
	"sync"
	"time"

	eventsv1 "k8s.io/api/events/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	toolscache "k8s.io/client-go/tools/cache"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
)

type watchedObjectKey struct {
	Namespace string
	Name      string
	Kind      string
}

type eventInfo struct {
	Message    string
	CreatedAt  time.Time
	RecordedAt time.Time
	Reason     string
	// RegardingUID is the UID of the object the event was recorded against. Events are
	// cached by namespace, name and kind, all of which a recreated object reuses, so this
	// is what tells one incarnation of an object from the next.
	RegardingUID k8stypes.UID
	// LastObservedAt is the freshest time the source saw this event. Kubernetes aggregates
	// a repeating event by updating the same object rather than creating a new one, so
	// CreatedAt only says when the first occurrence was seen.
	LastObservedAt time.Time
}

// objectEventWatcher lists the events cached for an object. The bounds filter on the time
// the event was first created and recorded; a caller that cares about a repeating event
// still happening now filters on LastObservedAt itself.
type objectEventWatcher interface {
	List(objectKey watchedObjectKey, createdAfter time.Time, recordedAfter time.Time) []*eventInfo
}

type controllerRuntimeEventWatcher struct {
	objectCache sync.Map
}

type eventObjects struct {
	mu         sync.RWMutex
	eventInfos map[k8stypes.NamespacedName]*eventInfo
	// evicted marks a bucket that has been taken out of objectCache because its last
	// event was deleted. A writer that loaded the bucket before it was removed must not
	// add to it: the bucket is no longer reachable, so the event would be dropped. It
	// sees this flag under the same lock the eviction is done under and retries against
	// a fresh bucket instead.
	evicted bool
}

func newControllerRuntimeEventWatcher(ctx context.Context, cache ctrlcache.Cache) (*controllerRuntimeEventWatcher, error) {
	informer, err := cache.GetInformer(ctx, &eventsv1.Event{})
	if err != nil {
		return nil, err
	}

	watcher := &controllerRuntimeEventWatcher{}
	if _, err := informer.AddEventHandler(watcher); err != nil {
		return nil, err
	}

	return watcher, nil
}

func (w *controllerRuntimeEventWatcher) OnAdd(obj interface{}, _ bool) {
	w.store(obj)
}

func (w *controllerRuntimeEventWatcher) OnUpdate(_, newObj interface{}) {
	// A repeating event is aggregated into the object that already exists, so its later
	// occurrences reach us as updates rather than as adds.
	w.store(newObj)
}

// store records an event, or refreshes the entry an earlier occurrence of the same event
// left behind. A refresh only moves what the newer occurrence actually tells us, the
// identity and the last observed time; the times the entry was first created and first
// recorded stay put, so a caller listing by watermark does not see the event again.
func (w *controllerRuntimeEventWatcher) store(obj interface{}) {
	event, ok := obj.(*eventsv1.Event)
	if !ok || event == nil {
		return
	}

	objectKey := watchedObjectKey{
		Namespace: event.Regarding.Namespace,
		Name:      event.Regarding.Name,
		Kind:      event.Regarding.Kind,
	}
	if objectKey.Name == "" || objectKey.Kind == "" {
		return
	}

	eventKey := k8stypes.NamespacedName{Namespace: event.Namespace, Name: event.Name}

	info := &eventInfo{
		Message:        event.Note,
		CreatedAt:      event.CreationTimestamp.Time,
		RecordedAt:     time.Now(),
		Reason:         event.Reason,
		RegardingUID:   event.Regarding.UID,
		LastObservedAt: lastObservedTime(event),
	}

	// A bucket can be evicted between the load and the lock, so the store is retried
	// until it lands in one that is still reachable. The loop turns at most once per
	// concurrent eviction of this object's bucket, and an eviction only happens when the
	// bucket is empty, so it cannot spin.
	for {
		value, _ := w.objectCache.LoadOrStore(objectKey, &eventObjects{
			eventInfos: make(map[k8stypes.NamespacedName]*eventInfo),
		})
		eventInfos := value.(*eventObjects)

		eventInfos.mu.Lock()
		if eventInfos.evicted {
			eventInfos.mu.Unlock()
			continue
		}

		stored := *info
		if existing, ok := eventInfos.eventInfos[eventKey]; ok {
			stored.CreatedAt = existing.CreatedAt
			stored.RecordedAt = existing.RecordedAt
			if existing.LastObservedAt.After(stored.LastObservedAt) {
				stored.LastObservedAt = existing.LastObservedAt
			}
		}
		// The entry is replaced rather than mutated: List hands out these pointers, and a
		// reader may still be looking at the one it got.
		eventInfos.eventInfos[eventKey] = &stored
		eventInfos.mu.Unlock()
		return
	}
}

// lastObservedTime is the freshest occurrence the event reports. An aggregated event
// carries a series whose last observed time moves with every repeat; a plain one only
// has the time it was recorded at, and older recorders fill in the deprecated field.
func lastObservedTime(event *eventsv1.Event) time.Time {
	candidates := []time.Time{event.EventTime.Time, event.DeprecatedLastTimestamp.Time}
	if event.Series != nil {
		candidates = append(candidates, event.Series.LastObservedTime.Time)
	}

	latest := time.Time{}
	for _, candidate := range candidates {
		if candidate.After(latest) {
			latest = candidate
		}
	}
	if latest.IsZero() {
		return event.CreationTimestamp.Time
	}
	return latest
}

func (w *controllerRuntimeEventWatcher) OnDelete(obj interface{}) {
	event, ok := obj.(*eventsv1.Event)
	if !ok {
		tombstone, ok := obj.(toolscache.DeletedFinalStateUnknown)
		if !ok {
			return
		}
		event, ok = tombstone.Obj.(*eventsv1.Event)
		if !ok {
			return
		}
	}

	objectKey := watchedObjectKey{
		Namespace: event.Regarding.Namespace,
		Name:      event.Regarding.Name,
		Kind:      event.Regarding.Kind,
	}
	if objectKey.Name == "" || objectKey.Kind == "" {
		return
	}

	eventKey := k8stypes.NamespacedName{Namespace: event.Namespace, Name: event.Name}

	value, ok := w.objectCache.Load(objectKey)
	if !ok {
		return
	}
	eventInfos := value.(*eventObjects)

	eventInfos.mu.Lock()
	defer eventInfos.mu.Unlock()

	delete(eventInfos.eventInfos, eventKey)

	// The bucket goes when its last event does, so objectCache holds an entry only for
	// objects with a live event rather than for every object seen since startup.
	//
	// Removing it is safe against a concurrent store because both the flag and the
	// removal happen under this bucket's write lock: a writer that already loaded this
	// bucket blocks here, then sees evicted and retries against a fresh one, and a writer
	// that has not loaded it yet cannot reach it once it is out of the map. Taking the
	// map's lock while holding the bucket's cannot deadlock, because every other path
	// releases the map's lock before it takes a bucket's.
	if len(eventInfos.eventInfos) == 0 {
		eventInfos.evicted = true
		w.objectCache.Delete(objectKey)
	}
}

// List returns the cached events for an object, ordered by when they were created. The
// bounds are on first creation and first recording, which a later occurrence of the same
// event does not move, so an event already reported is not reported again.
func (w *controllerRuntimeEventWatcher) List(objectKey watchedObjectKey, createdAfter time.Time, recordedAfter time.Time) []*eventInfo {
	value, ok := w.objectCache.Load(objectKey)
	if !ok {
		return nil
	}
	eventInfos := value.(*eventObjects)

	eventInfos.mu.RLock()
	defer eventInfos.mu.RUnlock()

	events := make([]*eventInfo, 0, len(eventInfos.eventInfos))
	for _, info := range eventInfos.eventInfos {
		if info.CreatedAt.After(createdAfter) ||
			(info.CreatedAt.Equal(createdAfter) && info.RecordedAt.After(recordedAfter)) {
			events = append(events, info)
		}
	}

	sort.SliceStable(events, func(i, j int) bool {
		if events[i].CreatedAt.Equal(events[j].CreatedAt) {
			return events[i].RecordedAt.Before(events[j].RecordedAt)
		}
		return events[i].CreatedAt.Before(events[j].CreatedAt)
	})

	return events
}
