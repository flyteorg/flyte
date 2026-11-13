package fastcheck

import (
	"context"

	cache "github.com/hashicorp/golang-lru"
	"go.opentelemetry.io/otel/metric"
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
func (l LRUCacheFilter) Contains(ctx context.Context, id []byte) bool {
	v := l.lru.Contains(string(id))
	if v {
		l.metrics.Hit.Add(ctx, 1, l.metrics.attributes)
		return true
	}
	l.metrics.Miss.Add(ctx, 1, l.metrics.attributes)
	return false
}

func (l LRUCacheFilter) Add(_ context.Context, id []byte) bool {
	return l.lru.Add(string(id), nil)
}

func (l LRUCacheFilter) Remove(_ context.Context, id []byte) bool {
	return l.lru.Remove(string(id))
}

// NewLRUCacheFilter creates a fixed-capacity LRU filter with OTEL metrics.
// Scope identifies the filter in metrics. A nil provider disables metrics.
func NewLRUCacheFilter(size int, scope string, provider metric.MeterProvider) (*LRUCacheFilter, error) {
	c, err := cache.New(size)
	if err != nil {
		return nil, err
	}
	metrics, err := newMetrics(scope, "lru", func() int64 { return int64(c.Len()) }, provider)
	if err != nil {
		return nil, err
	}
	return &LRUCacheFilter{
		lru:     c,
		metrics: metrics,
	}, nil
}
