package impl

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

// operationLabel is the single, low-cardinality label applied to every repository
// DB metric. It carries the logical operation name (e.g. "get_task"), never row
// IDs, project, domain, or user values.
const operationLabel = "operation"

// dbMetrics holds the Prometheus instruments shared by the runs repository
// implementations. All metrics are vectors labeled by operation name so that a
// single registration serves every DB operation without high label cardinality.
//
// A zero-value dbMetrics (returned when no Scope is provided) is a safe no-op:
// observe still runs the wrapped function and propagates its error, it simply
// records nothing.
type dbMetrics struct {
	calls   *prometheus.CounterVec
	errors  *prometheus.CounterVec
	latency *promutils.StopWatchVec
}

// newDBMetrics builds a dbMetrics under the "db" sub-scope of the provided scope.
// Metrics are registered once here, so each repository should construct its
// metrics exactly once (in its constructor) to avoid duplicate-registration
// panics. A nil scope yields a no-op dbMetrics, keeping constructors usable in
// contexts (e.g. unit tests) that do not wire metrics.
func newDBMetrics(scope promutils.Scope) dbMetrics {
	if scope == nil {
		return dbMetrics{}
	}
	dbScope := scope.NewSubScope("db")
	return dbMetrics{
		calls: dbScope.MustNewCounterVec(
			"calls_total",
			"Total number of DB calls made by the runs repository, by operation.",
			operationLabel,
		),
		errors: dbScope.MustNewCounterVec(
			"errors_total",
			"Total number of failed DB calls made by the runs repository, by operation.",
			operationLabel,
		),
		latency: dbScope.MustNewStopWatchVec(
			"latency",
			"Latency of DB calls made by the runs repository, by operation.",
			time.Millisecond,
			operationLabel,
		),
	}
}

// observe wraps a single DB operation with metrics: it increments the call
// counter, times the call, and increments the error counter when the operation
// returns a non-nil error. The wrapped error is returned unchanged so callers
// keep their existing error handling. It is a no-op recorder when dbMetrics is
// the zero value.
func (m dbMetrics) observe(_ context.Context, operation string, fn func() error) error {
	if m.calls == nil {
		// No-op metrics (nil scope): still execute the operation.
		return fn()
	}
	m.calls.WithLabelValues(operation).Inc()
	timer := m.latency.WithLabelValues(operation).Start()
	defer timer.Stop()

	if err := fn(); err != nil {
		m.errors.WithLabelValues(operation).Inc()
		return err
	}
	return nil
}
