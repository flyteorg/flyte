package k8s

import (
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

func TestEventWatcherRecordsTheFirstObservation(t *testing.T) {
	key := watchedObjectKey{Namespace: "ns", Name: "pod", Kind: "Pod"}
	// The write lands a second after the recorder saw the event, so the creation is later
	// than the recorder's own time.
	observedAt := time.Now().Add(-time.Hour).Truncate(time.Second)
	createdAt := observedAt.Add(time.Second)

	tests := []struct {
		name  string
		event func() *eventsv1.Event
		want  time.Time
	}{
		{
			name: "the event time of a recorder on the events API",
			event: func() *eventsv1.Event {
				event := testEvent("event-1", "pod-uid", createdAt)
				event.EventTime = metav1.NewMicroTime(observedAt)
				return event
			},
			want: observedAt,
		},
		{
			name: "the deprecated first timestamp of an older recorder",
			event: func() *eventsv1.Event {
				event := testEvent("event-1", "pod-uid", createdAt)
				event.DeprecatedFirstTimestamp = metav1.NewTime(observedAt)
				event.DeprecatedLastTimestamp = metav1.NewTime(observedAt)
				return event
			},
			want: observedAt,
		},
		{
			name:  "the creation when the event carries neither",
			event: func() *eventsv1.Event { return testEvent("event-1", "pod-uid", createdAt) },
			want:  createdAt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcher := &controllerRuntimeEventWatcher{}
			watcher.OnAdd(tt.event(), false)

			events := watcher.List(key, time.Time{}, time.Time{})
			require.Len(t, events, 1)
			assert.WithinDuration(t, tt.want, events[0].FirstObservedAt, time.Microsecond)
			assert.False(t, events[0].LastObservedAt.Before(events[0].FirstObservedAt),
				"the observed interval is never inverted")
			assert.Equal(t, createdAt.UTC(), events[0].CreatedAt.UTC())
		})
	}
}

func TestEventWatcherOnUpdateKeepsTheFirstObservation(t *testing.T) {
	watcher := &controllerRuntimeEventWatcher{}
	firstObserved := time.Now().Add(-time.Hour).Truncate(time.Second)
	lastObserved := time.Now().Add(-time.Minute)
	key := watchedObjectKey{Namespace: "ns", Name: "pod", Kind: "Pod"}

	first := testEvent("event-1", "pod-uid", firstObserved.Add(time.Second))
	first.EventTime = metav1.NewMicroTime(firstObserved)
	watcher.OnAdd(first, false)

	// An update that reports a later first occurrence, as a rewritten event can, must not
	// move the first observation forward, just as an update never walks the last one back.
	rewritten := testEvent("event-1", "pod-uid", firstObserved.Add(time.Second))
	rewritten.EventTime = metav1.NewMicroTime(lastObserved)
	rewritten.Series = &eventsv1.EventSeries{Count: 2, LastObservedTime: metav1.NewMicroTime(lastObserved)}
	watcher.OnUpdate(first, rewritten)

	events := watcher.List(key, time.Time{}, time.Time{})
	require.Len(t, events, 1)
	assert.WithinDuration(t, firstObserved, events[0].FirstObservedAt, time.Microsecond)
	assert.WithinDuration(t, lastObserved, events[0].LastObservedAt, time.Microsecond)
}

func TestEventWatcherRecordsTheReportingNode(t *testing.T) {
	key := watchedObjectKey{Namespace: "ns", Name: "pod", Kind: "Pod"}
	createdAt := time.Now().Add(-time.Hour)

	tests := []struct {
		name     string
		instance string
		host     string
		want     string
	}{
		{name: "the reporting instance", instance: testNodeName, want: testNodeName},
		{name: "the deprecated source host from an older recorder", host: testNodeName, want: testNodeName},
		{name: "the reporting instance over the host", instance: testNodeName, host: otherNodeName, want: testNodeName},
		{name: "neither", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcher := &controllerRuntimeEventWatcher{}
			event := testEvent("event-1", "pod-uid", createdAt)
			event.ReportingInstance = tt.instance
			event.DeprecatedSource.Host = tt.host
			watcher.OnAdd(event, false)

			events := watcher.List(key, time.Time{}, time.Time{})
			require.Len(t, events, 1)
			assert.Equal(t, tt.want, events[0].ReportingNode)
		})
	}
}

func TestEventWatcherOnUpdateTakesTheReportingNodeAsItIsNow(t *testing.T) {
	watcher := &controllerRuntimeEventWatcher{}
	createdAt := time.Now().Add(-time.Hour)
	key := watchedObjectKey{Namespace: "ns", Name: "pod", Kind: "Pod"}

	first := testEvent("event-1", "pod-uid", createdAt)
	first.ReportingInstance = testNodeName
	watcher.OnAdd(first, false)

	// An event rewritten to name another node is judged on what it says now, not on
	// the node an earlier version named.
	rewritten := testEvent("event-1", "pod-uid", createdAt)
	rewritten.ReportingInstance = otherNodeName
	watcher.OnUpdate(first, rewritten)

	events := watcher.List(key, time.Time{}, time.Time{})
	require.Len(t, events, 1)
	assert.Equal(t, otherNodeName, events[0].ReportingNode)
	assert.Equal(t, createdAt.UTC(), events[0].CreatedAt.UTC())
}
