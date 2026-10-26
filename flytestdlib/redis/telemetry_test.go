package redis

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	redisotelnative "github.com/redis/go-redis/extra/redisotel-native/v9"
	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestNewClientTelemetry(t *testing.T) {
	for _, tc := range []struct {
		name             string
		tracing, metrics bool
	}{
		{name: "none"},
		{name: "tracing", tracing: true},
		{name: "metrics", metrics: true},
		{name: "both", tracing: true, metrics: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			t.Cleanup(func() { require.NoError(t, redisotelnative.GetObservabilityInstance().Shutdown()) })
			server := miniredis.RunT(t)
			exporter := tracetest.NewInMemoryExporter()
			tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
			reader := sdkmetric.NewManualReader()
			mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
			t.Cleanup(func() {
				require.NoError(t, tp.Shutdown(ctx))
				require.NoError(t, mp.Shutdown(ctx))
			})
			var opts []Option
			if tc.tracing {
				opts = append(opts, WithTracerProvider(tp))
			}
			if tc.metrics {
				opts = append(opts, WithMeterProvider(mp))
			}
			client, err := (Config{Addr: server.Addr()}).NewClient(ctx, opts...)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, client.Close()) })
			require.NoError(t, client.Ping(ctx).Err())
			require.NoError(t, client.Set(ctx, "private-key", "private-value", 0).Err())
			if tc.tracing {
				require.NotEmpty(t, exporter.GetSpans())
				for _, span := range exporter.GetSpans() {
					for _, attr := range span.Attributes {
						require.NotEqual(t, "db.statement", string(attr.Key))
					}
				}
			} else {
				require.Empty(t, exporter.GetSpans())
			}
			var data metricdata.ResourceMetrics
			require.NoError(t, reader.Collect(ctx, &data))
			if tc.metrics {
				require.NotEmpty(t, data.ScopeMetrics)
				names := make([]string, 0)
				for _, scope := range data.ScopeMetrics {
					for _, m := range scope.Metrics {
						names = append(names, m.Name)
					}
				}
				require.Contains(t, names, redisotelnative.MetricOperationDuration)
			} else {
				require.Empty(t, data.ScopeMetrics)
			}
		})
	}
}

// A component must not replace the provider chosen by the all-components manager.
func TestInitMetricsPreservesFirstProvider(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { require.NoError(t, redisotelnative.GetObservabilityInstance().Shutdown()) })
	reader := sdkmetric.NewManualReader()
	shared := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	componentReader := sdkmetric.NewManualReader()
	component := sdkmetric.NewMeterProvider(sdkmetric.WithReader(componentReader))
	t.Cleanup(func() {
		require.NoError(t, shared.Shutdown(ctx))
		require.NoError(t, component.Shutdown(ctx))
	})
	require.NoError(t, InitMetrics(nil))
	require.NoError(t, InitMetrics(shared))
	server := miniredis.RunT(t)
	client, err := (Config{Addr: server.Addr()}).NewClient(ctx, WithMeterProvider(component))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	require.NoError(t, client.Ping(ctx).Err())
	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(ctx, &data))
	require.NotEmpty(t, data.ScopeMetrics)
	require.NoError(t, componentReader.Collect(ctx, &data))
	require.Empty(t, data.ScopeMetrics)
}
