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
}

type objectEventWatcher interface {
	List(objectKey watchedObjectKey, createdAfter time.Time, recordedAfter time.Time) []*eventInfo
}

type controllerRuntimeEventWatcher struct {
	objectCache sync.Map
}

type eventObjects struct {
	mu         sync.RWMutex
	eventInfos map[k8stypes.NamespacedName]*eventInfo
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

	value, _ := w.objectCache.LoadOrStore(objectKey, &eventObjects{
		eventInfos: make(map[k8stypes.NamespacedName]*eventInfo),
	})
	eventInfos := value.(*eventObjects)

	eventInfos.mu.Lock()
	eventInfos.eventInfos[eventKey] = &eventInfo{
		Message:    event.Note,
		CreatedAt:  event.CreationTimestamp.Time,
		RecordedAt: time.Now(),
		Reason:     event.Reason,
	}
	eventInfos.mu.Unlock()
}

func (w *controllerRuntimeEventWatcher) OnUpdate(_, _ interface{}) {
	// Ignore updates; we only need newly observed object events.
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
	// We intentionally do not delete empty buckets from objectCache. This avoids races where
	// a new event is being added to the bucket while the top-level map entry is concurrently removed.
}

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
