package k8s

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	informerEventsv1 "k8s.io/client-go/informers/events/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/k8s/eventwatcheriface"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// EventWatcher is a simple wrapper around the informer that keeps track of outstanding object events.
// Event lifetime is controlled by kube-apiserver (see --event-ttl flag) and defaults to one hour. As a result,
// the cache size is bounded by the number of event objects created in the last hour (or otherwise configured ttl).
// Note that cardinality of per object events is relatively low (10s), while they may occur repeatedly. For example
// the ImagePullBackOff event may continue to fire, but this is only backed by a single event.
type EventWatcher struct {
	cache.ResourceEventHandler
	informer    informerEventsv1.EventInformer
	objectCache sync.Map
}

type objectEvents struct {
	mu         sync.RWMutex
	eventInfos map[types.NamespacedName]*eventwatcheriface.EventInfo
}

func (e *EventWatcher) OnAdd(obj interface{}, isInInitialList bool) {
	event := obj.(*eventsv1.Event)
	objectNsName := types.NamespacedName{Namespace: event.Regarding.Namespace, Name: event.Regarding.Name}
	eventNsName := types.NamespacedName{Namespace: event.Namespace, Name: event.Name}
	v, _ := e.objectCache.LoadOrStore(objectNsName, &objectEvents{
		eventInfos: map[types.NamespacedName]*eventwatcheriface.EventInfo{},
	})
	objEvents := v.(*objectEvents)
	objEvents.mu.Lock()
	defer objEvents.mu.Unlock()
	objEvents.eventInfos[eventNsName] = &eventwatcheriface.EventInfo{
		Reason:     event.Reason,
		Note:       event.Note,
		CreatedAt:  event.CreationTimestamp.Time,
		RecordedAt: time.Now(),
	}
}

func (e *EventWatcher) OnUpdate(_, newObj interface{}) {
	// Dropping event updates since we only care about the creation
}

func (e *EventWatcher) OnDelete(obj interface{}) {
	event, casted := obj.(*eventsv1.Event)
	if !casted {
		unknown, casted := obj.(cache.DeletedFinalStateUnknown)
		if !casted {
			logger.Warnf(context.Background(), "Unknown object type [%T] in OnDelete", obj)
		} else {
			logger.Warnf(context.Background(), "Deleted object of unknown key [%v] type [%T] in OnDelete",
				unknown.Key, unknown.Obj)
		}

		return
	}

	objectNsName := types.NamespacedName{Namespace: event.Regarding.Namespace, Name: event.Regarding.Name}
	eventNsName := types.NamespacedName{Namespace: event.Namespace, Name: event.Name}
	v, _ := e.objectCache.LoadOrStore(objectNsName, &objectEvents{})
	objEvents := v.(*objectEvents)
	objEvents.mu.Lock()
	defer objEvents.mu.Unlock()
	delete(objEvents.eventInfos, eventNsName)
	if len(objEvents.eventInfos) == 0 {
		e.objectCache.Delete(objectNsName)
	}
}

// List returns all events for the given object that were created after the given time, sorted by creation time.
func (e *EventWatcher) List(objectNsName types.NamespacedName, createdAfter time.Time) []*eventwatcheriface.EventInfo {
	v, _ := e.objectCache.Load(objectNsName)
	if v == nil {
		return []*eventwatcheriface.EventInfo{}
	}
	objEvents := v.(*objectEvents)
	objEvents.mu.RLock()
	defer objEvents.mu.RUnlock()
	// This logic assumes that cardinality of events per object is relatively low, so iterating over them to find
	// recent ones and sorting the results is not too expensive.
	result := make([]*eventwatcheriface.EventInfo, 0, len(objEvents.eventInfos))
	for _, eventInfo := range objEvents.eventInfos {
		if eventInfo.CreatedAt.After(createdAfter) {
			result = append(result, eventInfo)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt) ||
			(result[i].CreatedAt.Equal(result[j].CreatedAt) && result[i].RecordedAt.Before(result[j].RecordedAt))
	})
	return result
}

func NewEventWatcher(ctx context.Context, gvk schema.GroupVersionKind, kubeClientset kubernetes.Interface) (*EventWatcher, error) {
	objectSelector := func(opts *metav1.ListOptions) {
		opts.FieldSelector = fields.OneTermEqualSelector("regarding.kind", gvk.Kind).String()
	}
	eventInformer := informers.NewSharedInformerFactoryWithOptions(
		kubeClientset, 0, informers.WithTweakListOptions(objectSelector)).Events().V1().Events()
	watcher := &EventWatcher{
		informer: eventInformer,
	}

	_, err := eventInformer.Informer().AddEventHandler(watcher)
	if err != nil {
		return nil, fmt.Errorf("failed to add event handler: %w", err)
	}

	go eventInformer.Informer().Run(ctx.Done())
	logger.Debugf(ctx, "Started informer for [%s] events", gvk.Kind)

	return watcher, nil
}
