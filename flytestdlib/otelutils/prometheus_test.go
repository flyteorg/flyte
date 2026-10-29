package otelutils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func TestPrometheusExporter(t *testing.T) {
	// Multiple components in one process must expose the same instruments
	// without duplicate series or registry errors.
	registry := prometheus.NewRegistry()
	// The executor shares this registry with native controller-runtime metrics.
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{Name: "controller_runtime_test_active"})
	registry.MustRegister(gauge)
	gauge.Set(2)
	for _, service := range []string{"prometheus-test-runs", "prometheus-test-actions"} {
		require.NoError(t, RegisterProvidersWithContext(context.Background(), service,
			&Config{ExporterType: PrometheusExporter}, WithPrometheusRegisterer(registry)))
		provider := meterProviders[service]
		t.Cleanup(func() {
			require.NoError(t, provider.Shutdown(context.Background()))
			delete(meterProviders, service)
		})
		require.Equal(t, noopTracerProvider, GetTracerProvider(service))
		meter := GetMeterProvider(service).Meter("test")
		counter, err := meter.Int64Counter("flyte.test.requests")
		require.NoError(t, err)
		counter.Add(context.Background(), 3, metric.WithAttributes(attribute.String("method", "run")))
		histogram, err := meter.Float64Histogram("flyte.test.duration", metric.WithUnit("s"))
		require.NoError(t, err)
		histogram.Record(context.Background(), 0.5)
	}
	response := httptest.NewRecorder()
	promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Contains(t, response.Body.String(), "controller_runtime_test_active 2")
	require.Contains(t, response.Body.String(), "flyte_test_requests_total")
	require.Contains(t, response.Body.String(), "flyte_test_duration_seconds_bucket")
	require.Contains(t, response.Body.String(), `service_name="prometheus-test-runs"`)
	require.Contains(t, response.Body.String(), `service_name="prometheus-test-actions"`)
	require.Contains(t, response.Body.String(), `method="run"`)
	families, err := registry.Gather()
	require.NoError(t, err)
	for _, family := range families {
		if family.GetName() == "flyte_test_requests_total" {
			require.Len(t, family.Metric, 2)
			for _, sample := range family.Metric {
				require.Equal(t, float64(3), sample.GetCounter().GetValue())
			}
		}
	}
}
