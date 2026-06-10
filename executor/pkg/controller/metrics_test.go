package controller

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
)

// newTestMetricsScheme builds a scheme with the TaskAction CRD and core types
// (the latter for asserting non-TaskAction operations pass through untimed).
func newTestMetricsScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	require.NoError(t, flyteorgv1.AddToScheme(s))
	require.NoError(t, corev1.AddToScheme(s))
	return s
}

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
	c := fake.NewClientBuilder().WithScheme(newTestMetricsScheme(t)).WithObjects(
		&flyteorgv1.TaskAction{ObjectMeta: metav1.ObjectMeta{Name: "ta-1", Namespace: "default"},
			Status: flyteorgv1.TaskActionStatus{PluginPhase: "Executing"}},
		&flyteorgv1.TaskAction{ObjectMeta: metav1.ObjectMeta{Name: "ta-2", Namespace: "default"},
			Status: flyteorgv1.TaskActionStatus{PluginPhase: "Executing"}},
		&flyteorgv1.TaskAction{ObjectMeta: metav1.ObjectMeta{Name: "ta-3", Namespace: "default"}},
	).Build()

	m, err := registerTaskActionMetrics(provider, c)
	require.NoError(t, err)
	require.NotNil(t, m)

	// Collection triggers the async gauge callback, which lists from the client
	// and reports active CRD counts by phase ("" maps to Unknown).
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

func TestObserveCRDSize(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	c := fake.NewClientBuilder().WithScheme(newTestMetricsScheme(t)).Build()

	m, err := registerTaskActionMetrics(provider, c)
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

func TestTimeK8sOp(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	c := fake.NewClientBuilder().WithScheme(newTestMetricsScheme(t)).Build()

	m, err := registerTaskActionMetrics(provider, c)
	require.NoError(t, err)

	// Propagates the wrapped op's result and records latency labeled by op/error.
	sentinel := errors.New("boom")
	assert.ErrorIs(t, m.timeK8sOp(context.Background(), opGet, func() error { return sentinel }), sentinel)
	assert.NoError(t, m.timeK8sOp(context.Background(), opUpdate, func() error { return nil }))

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

	// Nil receiver is a transparent pass-through that still runs the op.
	var nilMetrics *taskActionMetrics
	ran := false
	assert.NoError(t, nilMetrics.timeK8sOp(context.Background(), opGet, func() error { ran = true; return nil }))
	assert.True(t, ran)
}

func TestInstrumentedClient(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	ta := &flyteorgv1.TaskAction{ObjectMeta: metav1.ObjectMeta{Name: "ta-1", Namespace: "default"}}
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "cm-1", Namespace: "default"}}
	base := fake.NewClientBuilder().WithScheme(newTestMetricsScheme(t)).WithObjects(ta, cm).
		WithStatusSubresource(&flyteorgv1.TaskAction{}).Build()

	m, err := registerTaskActionMetrics(provider, base)
	require.NoError(t, err)
	c := newInstrumentedClient(base, m)

	ctx := context.Background()
	fetched := &flyteorgv1.TaskAction{}
	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(ta), fetched))
	require.NoError(t, c.Update(ctx, fetched))
	fetched.Status.PluginPhase = "Executing"
	require.NoError(t, c.Status().Update(ctx, fetched))
	// Non-TaskAction operations pass through untimed.
	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(cm), &corev1.ConfigMap{}))

	hist, ok := collectMetric(t, reader, "taskaction.k8s.duration").Data.(metricdata.Histogram[float64])
	require.True(t, ok)
	for _, op := range []string{opGet, opUpdate, opStatusUpdate} {
		assert.Equal(t, uint64(1), findOpDataPoint(t, hist, op).Count, "op=%s", op)
	}
	// Exactly the three TaskAction ops — the ConfigMap Get recorded nothing.
	assert.Len(t, hist.DataPoints, 3)

	// When metrics registration failed, the client is returned unwrapped.
	_, wrapped := newInstrumentedClient(base, nil).(*instrumentedClient)
	assert.False(t, wrapped)
}
