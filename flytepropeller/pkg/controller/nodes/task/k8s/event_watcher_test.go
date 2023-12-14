package k8s

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestEventWatcher_OnAdd(t *testing.T) {
	now := time.Now()
	ew := eventWatcher{}
	ew.objectCache.Store(types.NamespacedName{Namespace: "ns1", Name: "name1"}, &objectEvents{
		eventInfos: map[types.NamespacedName]*EventInfo{
			{Namespace: "eventns1", Name: "eventname1"}: {CreatedAt: now.Add(-time.Minute)},
		},
	})
	ew.objectCache.Store(types.NamespacedName{Namespace: "ns2", Name: "name2"}, &objectEvents{
		eventInfos: map[types.NamespacedName]*EventInfo{},
	})

	t.Run("existing event", func(t *testing.T) {
		ew.OnAdd(&eventsv1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:         "eventns1",
				Name:              "eventname1",
				CreationTimestamp: metav1.NewTime(now),
			},
			Regarding: corev1.ObjectReference{
				Namespace: "ns1",
				Name:      "name1",
			},
		}, false)

		v, _ := ew.objectCache.Load(types.NamespacedName{Namespace: "ns1", Name: "name1"})
		assert.NotNil(t, v)
		objEvents := v.(*objectEvents)
		assert.Len(t, objEvents.eventInfos, 1)
		assert.Equal(t, now, objEvents.eventInfos[types.NamespacedName{Namespace: "eventns1", Name: "eventname1"}].CreatedAt)
	})

	t.Run("new event", func(t *testing.T) {
		ew.OnAdd(&eventsv1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:         "eventns2",
				Name:              "eventname2",
				CreationTimestamp: metav1.NewTime(now),
			},
			Regarding: corev1.ObjectReference{
				Namespace: "ns2",
				Name:      "name2",
			},
		}, false)

		v, _ := ew.objectCache.Load(types.NamespacedName{Namespace: "ns2", Name: "name2"})
		assert.NotNil(t, v)
		objEvents := v.(*objectEvents)
		assert.Len(t, objEvents.eventInfos, 1)
		assert.Equal(t, now, objEvents.eventInfos[types.NamespacedName{Namespace: "eventns2", Name: "eventname2"}].CreatedAt)
	})

	t.Run("new object", func(t *testing.T) {
		ew.OnAdd(&eventsv1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:         "eventns3",
				Name:              "eventname3",
				CreationTimestamp: metav1.NewTime(now),
			},
			Regarding: corev1.ObjectReference{
				Namespace: "ns3",
				Name:      "name3",
			},
		}, false)

		v, _ := ew.objectCache.Load(types.NamespacedName{Namespace: "ns3", Name: "name3"})
		assert.NotNil(t, v)
		objEvents := v.(*objectEvents)
		assert.Len(t, objEvents.eventInfos, 1)
		assert.Equal(t, now, objEvents.eventInfos[types.NamespacedName{Namespace: "eventns3", Name: "eventname3"}].CreatedAt)
	})
}

func TestEventWatcher_OnDelete(t *testing.T) {
	now := time.Now()
	ew := eventWatcher{}
	ew.objectCache.Store(types.NamespacedName{Namespace: "ns1", Name: "name1"}, &objectEvents{
		eventInfos: map[types.NamespacedName]*EventInfo{
			{Namespace: "eventns1", Name: "eventname1"}: {CreatedAt: now.Add(-time.Minute)},
		},
	})
	ew.objectCache.Store(types.NamespacedName{Namespace: "ns2", Name: "name2"}, &objectEvents{
		eventInfos: map[types.NamespacedName]*EventInfo{},
	})

	t.Run("existing event", func(t *testing.T) {
		ew.OnDelete(&eventsv1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "eventns1",
				Name:      "eventname1",
			},
			Regarding: corev1.ObjectReference{
				Namespace: "ns1",
				Name:      "name1",
			},
		})

		v, _ := ew.objectCache.Load(types.NamespacedName{Namespace: "ns1", Name: "name1"})
		assert.Nil(t, v)
	})

	t.Run("missing event", func(t *testing.T) {
		ew.OnDelete(&eventsv1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "eventns2",
				Name:      "eventname2",
			},
			Regarding: corev1.ObjectReference{
				Namespace: "ns2",
				Name:      "name2",
			},
		})

		v, _ := ew.objectCache.Load(types.NamespacedName{Namespace: "ns2", Name: "name2"})
		assert.Nil(t, v)
	})

	t.Run("missing object", func(t *testing.T) {
		ew.OnDelete(&eventsv1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "eventns3",
				Name:      "eventname3",
			},
			Regarding: corev1.ObjectReference{
				Namespace: "ns3",
				Name:      "name3",
			},
		})

		v, _ := ew.objectCache.Load(types.NamespacedName{Namespace: "ns3", Name: "name3"})
		assert.Nil(t, v)
	})
}

func TestEventWatcher_List(t *testing.T) {
	now := time.Now()
	ew := eventWatcher{}
	ew.objectCache.Store(types.NamespacedName{Namespace: "ns1", Name: "name1"}, &objectEvents{
		eventInfos: map[types.NamespacedName]*EventInfo{
			{Namespace: "eventns1", Name: "eventname1"}: {CreatedAt: now},
			{Namespace: "eventns2", Name: "eventname2"}: {CreatedAt: now.Add(-time.Hour)},
		},
	})
	ew.objectCache.Store(types.NamespacedName{Namespace: "ns2", Name: "name2"}, &objectEvents{
		eventInfos: map[types.NamespacedName]*EventInfo{},
	})

	t.Run("all events", func(t *testing.T) {
		result := ew.List(types.NamespacedName{Namespace: "ns1", Name: "name1"}, time.Time{})

		assert.Len(t, result, 2)
	})

	t.Run("recent events", func(t *testing.T) {
		result := ew.List(types.NamespacedName{Namespace: "ns1", Name: "name1"}, now.Add(-time.Minute))

		assert.Len(t, result, 1)
	})

	t.Run("no events", func(t *testing.T) {
		result := ew.List(types.NamespacedName{Namespace: "ns2", Name: "name2"}, time.Time{})

		assert.Len(t, result, 0)
	})

	t.Run("no object", func(t *testing.T) {
		result := ew.List(types.NamespacedName{Namespace: "ns3", Name: "name3"}, time.Time{})

		assert.Len(t, result, 0)
	})

	t.Run("sorted", func(t *testing.T) {
		result := ew.List(types.NamespacedName{Namespace: "ns1", Name: "name1"}, time.Time{})

		assert.Equal(t, now.Add(-time.Hour), result[0].CreatedAt)
		assert.Equal(t, now, result[1].CreatedAt)
	})
}
