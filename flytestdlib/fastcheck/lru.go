package fastcheck

import (
	"context"

	"github.com/flyteorg/flytestdlib/promutils"
	cache "github.com/hashicorp/golang-lru"
)

// validate that it conforms to the interface
var _ Filter = LRUCacheFilter{}

// Implements the fastcheck.Filter interface using an underlying LRUCache from cache.Cache
// the underlying lru cache implementation is thread-safe
type LRUCacheFilter struct {
	lru     *cache.Cache
	metrics Metrics
}

// Simply uses Contains from the LRUCacheFilter
func (l LRUCacheFilter) Contains(_ context.Context, id []byte) bool {
	v := l.lru.Contains(string(id))
	if v {
		l.metrics.Hit.Inc()
		return true
	}
	l.metrics.Miss.Inc()
	return false
}

func (l LRUCacheFilter) Add(_ context.Context, id []byte) bool {
	return l.lru.Add(string(id), nil)
}

// Create a new fastcheck.Filter using an LRU cache of type cache.Cache with a fixed size
func NewLRUCacheFilter(size int, scope promutils.Scope) (*LRUCacheFilter, error) {
	c, err := cache.New(size)
	if err != nil {
		return nil, err
	}
	return &LRUCacheFilter{
		lru:     c,
		metrics: newMetrics(scope),
	}, nil
}
