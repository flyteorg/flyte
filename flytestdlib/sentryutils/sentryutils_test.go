package sentryutils

import (
	"context"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureMetrics returns a context whose Sentry hub records every metric
// emitted through it, before batching or sending.
func captureMetrics(t *testing.T) (context.Context, *[]*sentry.Metric) {
	t.Helper()
	var metrics []*sentry.Metric
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:       "https://key@example.invalid/1",
		Transport: &sentry.MockTransport{},
		BeforeSendMetric: func(m *sentry.Metric) *sentry.Metric {
			metrics = append(metrics, m)
			return nil
		},
	})
	require.NoError(t, err)
	hub := sentry.NewHub(client, sentry.NewScope())
	return sentry.SetHubOnContext(context.Background(), hub), &metrics
}

func TestGauge(t *testing.T) {
	t.Run("emits a gauge under its own name", func(t *testing.T) {
		ctx, metrics := captureMetrics(t)

		Gauge(ctx, "flyte.active_runs", 3)

		require.Len(t, *metrics, 1)
		m := (*metrics)[0]
		assert.Equal(t, sentry.MetricTypeGauge, m.Type)
		assert.Equal(t, "flyte.active_runs", m.Name)
		value, ok := m.Value.Float64()
		require.True(t, ok)
		assert.Equal(t, 3.0, value)
		assert.NotContains(t, m.Attributes, "operation")
		assert.NotContains(t, m.Attributes, "status")
	})

	t.Run("no-op without a Sentry client", func(t *testing.T) {
		ctx := sentry.SetHubOnContext(context.Background(), sentry.NewHub(nil, sentry.NewScope()))
		assert.NotPanics(t, func() { Gauge(ctx, "flyte.active_runs", 3) })
	})
}
