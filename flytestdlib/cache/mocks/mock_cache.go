package mocks

import (
	"context"
	"fmt"

	"github.com/eko/gocache/lib/v4/store"
)

// MockCache is a simple mock implementation of cache.CacheInterface[T]
// that always returns cache misses. When created with alwaysMiss=true,
// Get always returns an error (cache miss), Set/Delete/Invalidate/Clear are no-ops.
type MockCache[T any] struct {
	alwaysMiss bool
}

func NewMockCache[T any](alwaysMiss bool) *MockCache[T] {
	return &MockCache[T]{alwaysMiss: alwaysMiss}
}

func (m *MockCache[T]) Get(_ context.Context, _ any) (T, error) {
	var zero T
	if m.alwaysMiss {
		return zero, fmt.Errorf("cache miss")
	}
	return zero, nil
}

func (m *MockCache[T]) Set(_ context.Context, _ any, _ T, _ ...store.Option) error {
	return nil
}

func (m *MockCache[T]) Delete(_ context.Context, _ any) error {
	return nil
}

func (m *MockCache[T]) Invalidate(_ context.Context, _ ...store.InvalidateOption) error {
	return nil
}

func (m *MockCache[T]) Clear(_ context.Context) error {
	return nil
}

func (m *MockCache[T]) GetType() string {
	return "mock"
}
