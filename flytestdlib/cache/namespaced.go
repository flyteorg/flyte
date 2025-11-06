package cache

import (
	"context"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
)

// NamespacedCache is a wrapper around a cache that adds a namespace to all keys
type NamespacedCache[T any] struct {
	cache.CacheInterface[T]
	namespace string
}

func (n NamespacedCache[T]) getNamespacedKey(key any) string {
	return n.namespace + ":" + key.(string)
}

func (n NamespacedCache[T]) Get(ctx context.Context, key any) (T, error) {
	return n.CacheInterface.Get(ctx, n.getNamespacedKey(key))
}

func (n NamespacedCache[T]) Set(ctx context.Context, key any, object T, options ...store.Option) error {
	return n.CacheInterface.Set(ctx, n.getNamespacedKey(key), object, options...)
}

func (n NamespacedCache[T]) Delete(ctx context.Context, key any) error {
	return n.CacheInterface.Delete(ctx, n.getNamespacedKey(key))
}

func (n NamespacedCache[T]) GetType() string {
	return "namespaced"
}

// NewNamespacedCache creates a new namespaced cache that prefixes any key with the given namespace
func NewNamespacedCache[T any](namespace string, underlying cache.CacheInterface[T]) NamespacedCache[T] {
	return NamespacedCache[T]{
		CacheInterface: underlying,
		namespace:      namespace,
	}
}
