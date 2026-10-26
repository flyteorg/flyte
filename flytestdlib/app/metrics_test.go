package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/flytestdlib/otelutils"
)

func TestMetricsEndpoint(t *testing.T) {
	for _, exporter := range []string{otelutils.NoopExporter, otelutils.OtlpGrpcExporter, otelutils.PrometheusExporter} {
		t.Run(exporter, func(t *testing.T) {
			mux := http.NewServeMux()
			if exporter == otelutils.PrometheusExporter {
				require.NoError(t, otelutils.RegisterProvidersWithContext(context.Background(), "app-test", &otelutils.Config{ExporterType: exporter}))
				counter, err := otelutils.GetMeterProvider("app-test").Meter("app-test").Int64Counter("app.test.requests")
				require.NoError(t, err)
				counter.Add(context.Background(), 7)
			}
			registerMetricsHandler(mux, &otelutils.Config{ExporterType: exporter})
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
			if exporter == otelutils.PrometheusExporter {
				require.Equal(t, http.StatusOK, response.Code)
				require.Contains(t, response.Body.String(), "go_goroutines")
				require.Contains(t, response.Body.String(), "app_test_requests_total")
				require.Contains(t, response.Body.String(), `service_name="app-test"`)
			} else {
				require.Equal(t, http.StatusNotFound, response.Code)
			}
		})
	}
}
