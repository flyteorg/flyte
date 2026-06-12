package controller

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
)

// collectMetric drains the manual reader and returns the named metric.
func collectMetric(t *testing.T, reader *sdkmetric.ManualReader, name string) metricdata.Metrics {
	t.Helper()
	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				return m
			}
		}
	}
	t.Fatalf("metric %q not found", name)
	return metricdata.Metrics{}
}

func findOpDataPoint(t *testing.T, h metricdata.Histogram[float64], op string) metricdata.HistogramDataPoint[float64] {
	t.Helper()
	for _, dp := range h.DataPoints {
		if v, ok := dp.Attributes.Value(attribute.Key("op")); ok && v.AsString() == op {
			return dp
		}
	}
	t.Fatalf("no taskaction.k8s.duration datapoint with op=%q", op)
	return metricdata.HistogramDataPoint[float64]{}
}

func TestRegisterTaskActionMetrics(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	// The active gauge reports whatever the injected phase counter returns.
	activeByPhase := func(context.Context) map[string]int64 {
		return map[string]int64{"Executing": 2, "Unknown": 1}
	}

	m, err := registerTaskActionMetrics(provider, activeByPhase)
	require.NoError(t, err)
	require.NotNil(t, m)

	// Collection triggers the async gauge callback, which observes one data point per phase.
	gauge, ok := collectMetric(t, reader, "taskaction.active").Data.(metricdata.Gauge[int64])
	require.True(t, ok)
	counts := map[string]int64{}
	for _, dp := range gauge.DataPoints {
		v, ok := dp.Attributes.Value(attribute.Key("phase"))
		require.True(t, ok)
		counts[v.AsString()] = dp.Value
	}
	assert.Equal(t, map[string]int64{"Executing": 2, "Unknown": 1}, counts)
}

// TestCountByPhase covers the pure tallying used by cachedPhaseCounter: empty phase
// maps to "Unknown" and nil entries (which a raw cache listing may contain) are skipped.
func TestCountByPhase(t *testing.T) {
	items := []*flyteorgv1.TaskAction{
		{ObjectMeta: metav1.ObjectMeta{Name: "ta-1"}, Status: flyteorgv1.TaskActionStatus{PluginPhase: "Executing"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "ta-2"}, Status: flyteorgv1.TaskActionStatus{PluginPhase: "Executing"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "ta-3"}}, // empty phase -> Unknown
		nil,
	}
	assert.Equal(t, map[string]int64{"Executing": 2, "Unknown": 1}, countByPhase(items))
	assert.Equal(t, map[string]int64{}, countByPhase(nil))
}

func TestObserveCRDSize(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	m, err := registerTaskActionMetrics(provider, nil)
	require.NoError(t, err)

	ta := &flyteorgv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{Name: "ta-1", Namespace: "default"},
		Status:     flyteorgv1.TaskActionStatus{PluginPhase: "Executing"},
	}
	m.observeCRDSize(context.Background(), ta)

	b, err := json.Marshal(ta)
	require.NoError(t, err)
	hist, ok := collectMetric(t, reader, "taskaction.crd.size_bytes").Data.(metricdata.Histogram[int64])
	require.True(t, ok)
	require.Len(t, hist.DataPoints, 1)
	assert.Equal(t, uint64(1), hist.DataPoints[0].Count)
	assert.Equal(t, int64(len(b)), hist.DataPoints[0].Sum)

	// Safe no-op when metrics registration failed (nil receiver).
	var nilMetrics *taskActionMetrics
	assert.NotPanics(t, func() {
		nilMetrics.observeCRDSize(context.Background(), &flyteorgv1.TaskAction{})
	})
}

func TestRecordK8sOp(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	m, err := registerTaskActionMetrics(provider, nil)
	require.NoError(t, err)

	// Records latency labeled by op/error.
	m.recordK8sOp(context.Background(), opGet, time.Now(), errors.New("boom"))
	m.recordK8sOp(context.Background(), opUpdate, time.Now(), nil)

	hist, ok := collectMetric(t, reader, "taskaction.k8s.duration").Data.(metricdata.Histogram[float64])
	require.True(t, ok)
	get := findOpDataPoint(t, hist, opGet)
	errAttr, ok := get.Attributes.Value(attribute.Key("error"))
	require.True(t, ok)
	assert.True(t, errAttr.AsBool())
	update := findOpDataPoint(t, hist, opUpdate)
	errAttr, ok = update.Attributes.Value(attribute.Key("error"))
	require.True(t, ok)
	assert.False(t, errAttr.AsBool())

	// Safe no-op when metrics registration failed (nil receiver).
	var nilMetrics *taskActionMetrics
	assert.NotPanics(t, func() {
		nilMetrics.recordK8sOp(context.Background(), opGet, time.Now(), nil)
	})
}
