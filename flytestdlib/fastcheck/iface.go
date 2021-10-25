package fastcheck

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/flyteorg/flytestdlib/promutils"
)

// Filter provides an interface to check if a Key of type []byte was ever seen.
// The size of the cache is dependent on the id size and the initialization. It may also vary based on the implementation
// For example an LRU cache, may have an overhead because of the use of a HashMap with loading factor and collision
// resolution
// The Data-structure is thread-safe and can be accessed by multiple threads concurrently.

//go:generate mockery -name Filter -case=underscore

type Filter interface {
	// Contains returns a True if the id was previously seen or false otherwise
	// It may return a false, even if a item may have previously occurred.
	Contains(ctx context.Context, id []byte) bool

	// Adds the element id to the Filter and returns if a previous object was evicted in the process
	Add(ctx context.Context, id []byte) (evicted bool)
}

// Every implementation of the Filter Interface provides these metrics
type Metrics struct {
	// Indicates if the item was found in the cache
	Hit prometheus.Counter
	// Indicates if the item was not found in the cache
	Miss prometheus.Counter
}

func newMetrics(scope promutils.Scope) Metrics {
	return Metrics{
		Hit:  scope.MustNewCounter("cache_hit", "Indicates that the item was found in the cache"),
		Miss: scope.MustNewCounter("cache_miss", "Indicates that the item was found in the cache"),
	}
}
