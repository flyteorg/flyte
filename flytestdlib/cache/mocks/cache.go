package mocks

import (
	"context"

	"github.com/eko/gocache/lib/v4/store"
)

type MockCache[T any] struct {
	store map[string]T
	noop  bool
}

func (m *MockCache[T]) Get(ctx context.Context, key any) (T, error) {
	var zeroValue T
	if value, exists := m.store[key.(string)]; exists {
		return value, nil
	}

	return zeroValue, store.NotFound{}
}

func (m *MockCache[T]) Set(ctx context.Context, key any, object T, options ...store.Option) error {
	if m.noop {
		return nil
	}

	m.store[key.(string)] = object
	return nil
}

func (m *MockCache[T]) Delete(ctx context.Context, key any) error {
	delete(m.store, key.(string))
	return nil
}

func (m *MockCache[T]) Invalidate(ctx context.Context, options ...store.InvalidateOption) error {
	return nil
}

func (m *MockCache[T]) Clear(ctx context.Context) error {
	m.store = make(map[string]T)
	return nil
}

func (m *MockCache[T]) GetType() string {
	return "mock_cache"
}

func NewMockCache[T any](noop bool) *MockCache[T] {
	return &MockCache[T]{
		noop:  noop,
		store: make(map[string]T),
	}
}
