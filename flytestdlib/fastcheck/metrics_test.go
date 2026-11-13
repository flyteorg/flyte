package fastcheck

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestMetrics(t *testing.T) {
	constructors := map[string]func(string, metric.MeterProvider) Filter{
		"lru": func(scope string, provider metric.MeterProvider) Filter {
			f, err := NewLRUCacheFilter(1, scope, provider)
			require.NoError(t, err)
			return f
		},
		"oppobloom": func(scope string, provider metric.MeterProvider) Filter {
			f, err := NewOppoBloomFilter(1, scope, provider)
			require.NoError(t, err)
			return f
		},
	}
	for name, newFilter := range constructors {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			reader := sdkmetric.NewManualReader()
			provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
			t.Cleanup(func() { require.NoError(t, provider.Shutdown(ctx)) })
			scope := "test"
			f := newFilter(scope, provider)
			check := func(want int64) {
				t.Helper()
				var data metricdata.ResourceMetrics
				require.NoError(t, reader.Collect(ctx, &data))
				require.Len(t, data.ScopeMetrics, 1)
				require.Len(t, data.ScopeMetrics[0].Metrics, 1)
				m := data.ScopeMetrics[0].Metrics[0]
				require.Equal(t, "fastcheck.size", m.Name)
				g, ok := m.Data.(metricdata.Gauge[int64])
				require.True(t, ok)
				require.Len(t, g.DataPoints, 1)
				require.Equal(t, want, g.DataPoints[0].Value)
				implementation, ok := g.DataPoints[0].Attributes.Value("implementation")
				require.True(t, ok)
				require.Equal(t, name, implementation.AsString())
				filterScope, ok := g.DataPoints[0].Attributes.Value("scope")
				require.True(t, ok)
				require.Equal(t, scope, filterScope.AsString())
			}
			check(0)
			f.Add(ctx, []byte("first"))
			check(1)
			f.Add(ctx, []byte("first"))
			check(1)
			f.Add(ctx, []byte("second")) // Capacity one forces eviction/collision.
			check(1)
			f.Remove(ctx, []byte("first"))
			check(1)
			f.Remove(ctx, []byte("second"))
			check(0)
			f.Remove(ctx, []byte("second"))
			check(0)
			f.Add(ctx, nil) // Even an empty ID occupies a slot.
			check(1)
			f.Remove(ctx, nil)
			check(0)

			var wg sync.WaitGroup
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < 100; j++ {
						f.Add(ctx, []byte("shared"))
						var data metricdata.ResourceMetrics
						require.NoError(t, reader.Collect(ctx, &data))
						f.Remove(ctx, []byte("shared"))
					}
				}()
			}
			wg.Wait()
			check(0)
			// Only Contains contributes to hit/miss counters.
			require.False(t, f.Contains(ctx, []byte("missing")))
			f.Add(ctx, []byte("present"))
			require.True(t, f.Contains(ctx, []byte("present")))
			require.True(t, f.Contains(ctx, []byte("present")))
			f.Remove(ctx, []byte("present"))
			require.False(t, f.Contains(ctx, []byte("present")))
			var data metricdata.ResourceMetrics
			require.NoError(t, reader.Collect(ctx, &data))
			require.Len(t, data.ScopeMetrics, 1)
			require.Len(t, data.ScopeMetrics[0].Metrics, 3)
			for _, m := range data.ScopeMetrics[0].Metrics {
				if m.Name == "fastcheck.size" {
					continue
				}
				require.Contains(t, []string{"fastcheck.cache_hit", "fastcheck.cache_miss"}, m.Name)
				sum, ok := m.Data.(metricdata.Sum[int64])
				require.True(t, ok)
				require.True(t, sum.IsMonotonic)
				require.Len(t, sum.DataPoints, 1)
				require.EqualValues(t, 2, sum.DataPoints[0].Value)
				implementation, ok := sum.DataPoints[0].Attributes.Value("implementation")
				require.True(t, ok)
				require.Equal(t, name, implementation.AsString())
				filterScope, ok := sum.DataPoints[0].Attributes.Value("scope")
				require.True(t, ok)
				require.Equal(t, scope, filterScope.AsString())
			}
		})
	}
}
