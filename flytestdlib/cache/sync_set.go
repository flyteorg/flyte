package cache

import "sync"

var emptyVal = struct{}{}

// syncSet is a thread-safe Set.
type syncSet struct {
	underlying sync.Map
}

// Contains checks if the key is present in the set.
func (s *syncSet) Contains(key interface{}) bool {
	_, found := s.underlying.Load(key)
	return found
}

// Insert adds a new key to the set if it doesn't already exist.
func (s *syncSet) Insert(key interface{}) {
	s.underlying.Store(key, emptyVal)
}

// Remove deletes a key from the set.
func (s *syncSet) Remove(key interface{}) {
	s.underlying.Delete(key)
}

// Range allows iterating over the set. Deleting the key while iterating is a supported operation.
func (s *syncSet) Range(callback func(key interface{}) bool) {
	s.underlying.Range(func(key, value interface{}) bool {
		return callback(key)
	})
}

// newSyncSet initializes a new thread-safe set.
func newSyncSet() *syncSet {
	return &syncSet{
		underlying: sync.Map{},
	}
}
