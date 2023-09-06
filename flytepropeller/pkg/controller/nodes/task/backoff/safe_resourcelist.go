package backoff

import (
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// SyncResourceList is a thread-safe Map. It's meant to replace v1.ResourceList for concurrency-sensitive
// code.
type SyncResourceList struct {
	sync.Map
}

// Store stores the value in the map overriding existing value or adding a new one of one doesn't exist.
func (s *SyncResourceList) Store(resourceName v1.ResourceName, quantity resource.Quantity) {
	s.Map.Store(resourceName, quantity)
}

// Load loads a resource quantity if one exists.
func (s *SyncResourceList) Load(resourceName v1.ResourceName) (quantity resource.Quantity, found bool) {
	val, found := s.Map.Load(resourceName)
	if !found {
		return
	}

	return val.(resource.Quantity), true
}

// Range iterates over all the entries of the list in a non-sorted non-deterministic order.
func (s *SyncResourceList) Range(visitor func(key v1.ResourceName, value resource.Quantity) bool) {
	s.Map.Range(func(key, value interface{}) bool {
		return visitor(key.(v1.ResourceName), value.(resource.Quantity))
	})
}

// String returns a formatted string of some snapshot of the map ordered by keys.
func (s *SyncResourceList) String() string {
	set := sets.NewString()
	snapshot := map[string]resource.Quantity{}
	s.Range(func(key v1.ResourceName, value resource.Quantity) bool {
		strKey := string(key)
		set.Insert(strKey)
		snapshot[strKey] = value
		return true
	})

	sb := strings.Builder{}
	for _, key := range set.List() {
		value := snapshot[key]
		sb.WriteString(key)
		sb.WriteString(":")
		sb.WriteString(value.String())
		sb.WriteString(", ")
	}

	return sb.String()
}

// AsResourceList serializes a snapshot of the sync map into a v1.ResourceList
func (s *SyncResourceList) AsResourceList() v1.ResourceList {
	lst := v1.ResourceList{}
	s.Range(func(key v1.ResourceName, value resource.Quantity) bool {
		lst[key] = value
		return true
	})

	return lst
}

// AddResourceList stores a list into the sync map.
func (s *SyncResourceList) AddResourceList(list v1.ResourceList) *SyncResourceList {
	for key, value := range list {
		s.Store(key, value)
	}

	return s
}

// NewSyncResourceList creates a thread-safe map to store resource names and resource
// quantities. Equivalent to v1.ResourceList but offering concurrent-safe operations.
func NewSyncResourceList() *SyncResourceList {
	return &SyncResourceList{
		Map: sync.Map{},
	}
}
