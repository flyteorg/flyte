package redis

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
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
			if tc.tracing {
				require.NotEmpty(t, exporter.GetSpans())
			} else {
				require.Empty(t, exporter.GetSpans())
			}
			var data metricdata.ResourceMetrics
			require.NoError(t, reader.Collect(ctx, &data))
			if tc.metrics {
				require.NotEmpty(t, data.ScopeMetrics)
			} else {
				require.Empty(t, data.ScopeMetrics)
			}
		})
	}
}
