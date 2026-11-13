package fastcheck

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

// Filter provides an interface to check if a Key of type []byte was ever seen.
// The size of the cache is dependent on the id size and the initialization. It may also vary based on the implementation
// For example an LRU cache, may have an overhead because of the use of a HashMap with loading factor and collision
// resolution
// The Data-structure is thread-safe and can be accessed by multiple threads concurrently.

type Filter interface {
	// Contains returns a True if the id was previously seen or false otherwise
	// It may return a false, even if a item may have previously occurred.
	Contains(ctx context.Context, id []byte) bool

	// Adds the element id to the Filter and returns if a previous object was evicted in the process
	Add(ctx context.Context, id []byte) (evicted bool)

	// Remove removes id from the Filter and returns whether it was present.
	Remove(ctx context.Context, id []byte) (removed bool)
}

// Metrics contains the OTEL instruments shared by filter implementations.
type Metrics struct {
	Hit  metric.Int64Counter
	Miss metric.Int64Counter
	// Size reports entries currently stored, not capacity.
	Size       metric.Int64ObservableGauge
	attributes metric.MeasurementOption
}

func newMetrics(scope, implementation string, size func() int64, provider metric.MeterProvider) (Metrics, error) {
	if provider == nil {
		provider = noop.NewMeterProvider()
	}
	meter := provider.Meter("github.com/flyteorg/flyte/v2/flytestdlib/fastcheck")
	attributes := metric.WithAttributes(attribute.String("scope", scope), attribute.String("implementation", implementation))
	hit, err := meter.Int64Counter("fastcheck.cache_hit", metric.WithDescription("Number of filter lookups that found the ID"), metric.WithUnit("{lookup}"))
	if err != nil {
		return Metrics{}, err
	}
	miss, err := meter.Int64Counter("fastcheck.cache_miss", metric.WithDescription("Number of filter lookups that did not find the ID"), metric.WithUnit("{lookup}"))
	if err != nil {
		return Metrics{}, err
	}
	sizeGauge, err := meter.Int64ObservableGauge(
		"fastcheck.size",
		metric.WithDescription("Number of entries currently stored in the filter"),
		metric.WithUnit("{entry}"),
		metric.WithInt64Callback(func(_ context.Context, observer metric.Int64Observer) error {
			observer.Observe(size(), attributes)
			return nil
		}),
	)
	if err != nil {
		return Metrics{}, err
	}
	return Metrics{Hit: hit, Miss: miss, Size: sizeGauge, attributes: attributes}, nil
}
