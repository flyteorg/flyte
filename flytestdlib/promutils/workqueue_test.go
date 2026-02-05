package promutils

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

var provider = prometheusMetricsProvider{}

func TestPrometheusMetricsProvider(t *testing.T) {
	t.Run("Adds", func(t *testing.T) {
		c := provider.NewAddsMetric("x")
		_, ok := c.(prometheus.Counter)
		assert.True(t, ok)
	})

	t.Run("Depth", func(t *testing.T) {
		c := provider.NewDepthMetric("x")
		_, ok := c.(prometheus.Gauge)
		assert.True(t, ok)
	})

	t.Run("Latency", func(t *testing.T) {
		c := provider.NewLatencyMetric("x")
		_, ok := c.(prometheus.Summary)
		assert.True(t, ok)
	})

	t.Run("Retries", func(t *testing.T) {
		c := provider.NewRetriesMetric("x")
		_, ok := c.(prometheus.Counter)
		assert.True(t, ok)
	})

	t.Run("WorkDuration", func(t *testing.T) {
		c := provider.NewWorkDurationMetric("x")
		_, ok := c.(prometheus.Summary)
		assert.True(t, ok)
	})

	t.Run("NewLongestRunningProcessorSecondsMetric", func(t *testing.T) {
		c := provider.NewLongestRunningProcessorSecondsMetric("x")
		_, ok := c.(prometheus.Gauge)
		assert.True(t, ok)
	})

	t.Run("NewUnfinishedWorkSecondsMetric", func(t *testing.T) {
		c := provider.NewUnfinishedWorkSecondsMetric("x")
		_, ok := c.(prometheus.Gauge)
		assert.True(t, ok)
	})

	t.Run("NewLongestRunningProcessorMicrosecondsMetric", func(t *testing.T) {
		c := provider.NewLongestRunningProcessorMicrosecondsMetric("x")
		_, ok := c.(prometheus.Gauge)
		assert.True(t, ok)
	})
}
