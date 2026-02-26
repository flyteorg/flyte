package mocks

import (
	"context"
	"fmt"
	"sync"

	"github.com/eko/gocache/lib/v4/store"
)

// MockCache is a simple in-memory mock implementation of CacheInterface[T].
type MockCache[T any] struct {
	mu      sync.RWMutex
	data    map[string]T
	enabled bool
}

// NewMockCache creates a new MockCache. When enabled is true, the cache stores and retrieves values.
// When enabled is false, Get always returns an error (cache miss).
func NewMockCache[T any](enabled bool) *MockCache[T] {
	return &MockCache[T]{
		data:    make(map[string]T),
		enabled: enabled,
	}
}

func (m *MockCache[T]) Get(_ context.Context, key any) (T, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	k := fmt.Sprintf("%v", key)
	if !m.enabled {
		var zero T
		return zero, fmt.Errorf("cache disabled")
	}
	if v, ok := m.data[k]; ok {
		return v, nil
	}
	var zero T
	return zero, fmt.Errorf("key not found: %s", k)
}

func (m *MockCache[T]) Set(_ context.Context, key any, object T, _ ...store.Option) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.enabled {
		return nil
	}
	k := fmt.Sprintf("%v", key)
	m.data[k] = object
	return nil
}

func (m *MockCache[T]) Delete(_ context.Context, key any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	k := fmt.Sprintf("%v", key)
	delete(m.data, k)
	return nil
}

func (m *MockCache[T]) Invalidate(_ context.Context, _ ...store.InvalidateOption) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string]T)
	return nil
}

func (m *MockCache[T]) Clear(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string]T)
	return nil
}

func (m *MockCache[T]) GetType() string {
	return "mock"
}
