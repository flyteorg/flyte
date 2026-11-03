package retention

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Metrics observes invocations and committed cleanup without run-specific labels.
type Metrics struct {
	duration metric.Float64Histogram
	deleted  metric.Int64Counter
}

func NewMetrics(provider metric.MeterProvider) (*Metrics, error) {
	meter := provider.Meter("flyteadmin-retention")
	duration, err := meter.Float64Histogram("flyteadmin.retention.duration", metric.WithUnit("s"), metric.WithDescription("Duration of retention invocations, including database connection setup"))
	if err != nil {
		return nil, err
	}
	deleted, err := meter.Int64Counter("flyteadmin.retention.deleted", metric.WithUnit("{record}"), metric.WithDescription("Committed records deleted by retention, by entity; excludes dry-runs"))
	if err != nil {
		return nil, err
	}
	return &Metrics{duration: duration, deleted: deleted}, nil
}

// RecordInvocation must be called once at completion, including on failure.
func (m *Metrics) RecordInvocation(ctx context.Context, elapsed time.Duration, result Result, err error) {
	if m == nil {
		return
	}
	outcome := "success"
	if err != nil {
		outcome = "failure"
	} else if result.LockUnavailable {
		outcome = "skipped"
	}
	labels := metric.WithAttributes(attribute.String("result", outcome), attribute.Bool("dry_run", result.DryRun))
	ctx = context.WithoutCancel(ctx)
	m.duration.Record(ctx, elapsed.Seconds(), labels)
}

func (m *Metrics) recordCleanup(ctx context.Context, counts map[string]int64, dryRun bool) {
	if m == nil || dryRun {
		return
	}
	ctx = context.WithoutCancel(ctx)
	for table, count := range counts {
		if count > 0 {
			m.deleted.Add(ctx, count, metric.WithAttributes(attribute.String("entity", table)))
		}
	}
}
