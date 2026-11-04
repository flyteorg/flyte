package service

import (
	"context"
	"time"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/sentryutils"
)

const activeRunsSentryMetric = "flyte.active_runs"

// ActiveRunsReporter periodically reports the number of RUNNING runs to Sentry.
type ActiveRunsReporter struct {
	count    func(ctx context.Context) (int64, error)
	interval time.Duration
	report   func(ctx context.Context, name string, value float64)
}

// NewActiveRunsReporter creates a reporter that calls count once per interval.
func NewActiveRunsReporter(count func(ctx context.Context) (int64, error), interval time.Duration) *ActiveRunsReporter {
	return &ActiveRunsReporter{count: count, interval: interval, report: sentryutils.Gauge}
}

// Run reports once per interval until ctx is cancelled.
func (r *ActiveRunsReporter) Run(ctx context.Context) error {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			r.reportOnce(ctx)
		}
	}
}

func (r *ActiveRunsReporter) reportOnce(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, r.interval)
	defer cancel()

	n, err := r.count(ctx)
	if err != nil {
		logger.Warnf(ctx, "active runs reporter: %v", err)
		return
	}
	r.report(ctx, activeRunsSentryMetric, float64(n))
}
