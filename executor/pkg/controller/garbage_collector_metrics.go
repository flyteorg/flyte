package controller

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
)

const gcMeterName = "taskaction-gc"

// Values for the "outcome" attribute on taskaction.gc.deletions.
const (
	gcOutcomeDeleted     = "deleted"
	gcOutcomeAlreadyGone = "already_gone"
	gcOutcomeFailed      = "failed"
)

// gcMetrics holds OTel instruments for the garbage collector.
type gcMetrics struct {
	deletions     metric.Int64Counter
	sweepDuration metric.Float64Histogram
}

// newGCMetrics builds the garbage collector instruments on the given provider.
// Returns nil when metrics are disabled (noop provider).
func newGCMetrics(provider metric.MeterProvider) (*gcMetrics, error) {
	if _, ok := provider.(metricnoop.MeterProvider); ok {
		return nil, nil
	}

	meter := provider.Meter(gcMeterName)

	deletions, err := meter.Int64Counter(
		"taskaction.gc.deletions",
		metric.WithDescription("TaskAction delete attempts by the garbage collector, labeled by outcome"),
	)
	if err != nil {
		return nil, err
	}

	sweepDuration, err := meter.Float64Histogram(
		"taskaction.gc.sweep.duration",
		metric.WithDescription("Duration of one garbage collection sweep, labeled by error"),
		metric.WithUnit("ms"),
		// The SDK's default buckets top out at 10s, but a sweep over a large
		// backlog runs for minutes, these keep the slow tail measurable.
		metric.WithExplicitBucketBoundaries(100, 500, 1000, 5000, 10000, 30000, 60000, 300000),
	)
	if err != nil {
		return nil, err
	}

	return &gcMetrics{deletions: deletions, sweepDuration: sweepDuration}, nil
}

// recordDeletion counts one delete attempt by outcome: deleted, already_gone
// (removed by cascade before the GC got to it), or failed. No-op when metrics
// are disabled.
func (m *gcMetrics) recordDeletion(ctx context.Context, outcome string) {
	if m == nil || m.deletions == nil {
		return
	}
	m.deletions.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", outcome)))
}

// recordSweep records the duration of one sweep, labeled by whether the sweep
// failed. No-op when metrics are disabled.
func (m *gcMetrics) recordSweep(ctx context.Context, start time.Time, err error) {
	if m == nil || m.sweepDuration == nil {
		return
	}
	m.sweepDuration.Record(ctx, float64(time.Since(start).Microseconds())/1000.0,
		metric.WithAttributes(attribute.Bool("error", err != nil)))
}
