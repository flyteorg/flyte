package controller

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestNewGCMetricsNoop(t *testing.T) {
	m, err := newGCMetrics(metricnoop.NewMeterProvider())
	require.NoError(t, err)
	assert.Nil(t, m)
}

func TestRecordDeletion(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	m, err := newGCMetrics(provider)
	require.NoError(t, err)
	require.NotNil(t, m)

	ctx := context.Background()
	m.recordDeletion(ctx, gcOutcomeDeleted)
	m.recordDeletion(ctx, gcOutcomeDeleted)
	m.recordDeletion(ctx, gcOutcomeAlreadyGone)
	m.recordDeletion(ctx, gcOutcomeFailed)

	sum, ok := collectMetric(t, reader, "taskaction.gc.deletions").Data.(metricdata.Sum[int64])
	require.True(t, ok)

	counts := map[string]int64{}
	for _, dp := range sum.DataPoints {
		v, ok := dp.Attributes.Value(attribute.Key("outcome"))
		require.True(t, ok)
		counts[v.AsString()] = dp.Value
	}
	assert.Equal(t, int64(2), counts[gcOutcomeDeleted])
	assert.Equal(t, int64(1), counts[gcOutcomeAlreadyGone])
	assert.Equal(t, int64(1), counts[gcOutcomeFailed])

	var nilMetrics *gcMetrics
	assert.NotPanics(t, func() { nilMetrics.recordDeletion(ctx, gcOutcomeDeleted) })
}

func TestRecordSweep(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	m, err := newGCMetrics(provider)
	require.NoError(t, err)
	require.NotNil(t, m)

	ctx := context.Background()
	// Backdated start so the recorded duration is reliably non-zero.
	m.recordSweep(ctx, time.Now().Add(-50*time.Millisecond), nil)
	m.recordSweep(ctx, time.Now(), errors.New("boom"))

	hist, ok := collectMetric(t, reader, "taskaction.gc.sweep.duration").Data.(metricdata.Histogram[float64])
	require.True(t, ok)

	points := map[bool]metricdata.HistogramDataPoint[float64]{}
	for _, dp := range hist.DataPoints {
		v, ok := dp.Attributes.Value(attribute.Key("error"))
		require.True(t, ok)
		points[v.AsBool()] = dp
	}
	require.Len(t, points, 2)
	assert.Equal(t, uint64(1), points[false].Count)
	assert.GreaterOrEqual(t, points[false].Sum, 50.0)
	assert.Equal(t, uint64(1), points[true].Count)

	var nilMetrics *gcMetrics
	assert.NotPanics(t, func() { nilMetrics.recordSweep(ctx, time.Now(), nil) })
}
