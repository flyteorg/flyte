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
	List(objectKey watchedObjectKey, createdAfter time.Time) []*eventInfo
}

type controllerRuntimeEventWatcher struct {
	mu          sync.RWMutex
	objectCache map[watchedObjectKey]*eventObjects
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

	watcher := &controllerRuntimeEventWatcher{
		objectCache: make(map[watchedObjectKey]*eventObjects),
	}
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

	for {
		w.mu.RLock()
		eventInfos, ok := w.objectCache[objectKey]
		w.mu.RUnlock()
		// Create a new event object into objectCache.
		if !ok {
			w.mu.Lock()
			eventInfos, ok = w.objectCache[objectKey]
			if !ok {
				eventInfos = &eventObjects{
					eventInfos: make(map[k8stypes.NamespacedName]*eventInfo),
				}
				w.objectCache[objectKey] = eventInfos
			}
			w.mu.Unlock()
		}

		eventInfos.mu.Lock()

		// Revalidate this bucket is still current before writing.
		w.mu.RLock()
		stillCurrent := w.objectCache[objectKey] == eventInfos
		w.mu.RUnlock()
		if !stillCurrent {
			// eventInfos being deleted/changed, we should get the newest object again
			eventInfos.mu.Unlock()
			continue
		}

		eventInfos.eventInfos[eventKey] = &eventInfo{
			Message:    event.Note,
			CreatedAt:  event.CreationTimestamp.Time,
			RecordedAt: time.Now(),
			Reason:     event.Reason,
		}
		eventInfos.mu.Unlock()
		return
	}
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

	w.mu.RLock()
	eventInfos, ok := w.objectCache[objectKey]
	w.mu.RUnlock()
	if !ok {
		return
	}

	eventInfos.mu.Lock()
	defer eventInfos.mu.Unlock()

	delete(eventInfos.eventInfos, eventKey)
	if len(eventInfos.eventInfos) != 0 {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	// Delete only if this objectKey still points to the same event bucket.
	if current, exists := w.objectCache[objectKey]; exists && current == eventInfos && len(eventInfos.eventInfos) == 0 {
		delete(w.objectCache, objectKey)
	}
}

func (w *controllerRuntimeEventWatcher) List(objectKey watchedObjectKey, createdAfter time.Time) []*eventInfo {
	w.mu.RLock()
	eventInfos, ok := w.objectCache[objectKey]
	w.mu.RUnlock()
	if !ok {
		return nil
	}

	eventInfos.mu.RLock()
	defer eventInfos.mu.RUnlock()

	events := make([]*eventInfo, 0, len(eventInfos.eventInfos))
	for _, info := range eventInfos.eventInfos {
		if info.CreatedAt.After(createdAfter) {
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
