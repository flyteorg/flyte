package labeled

import (
	"context"
	"time"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils"
)

type StopWatch struct {
	*promutils.StopWatchVec

	// We use SummaryVec for emitting StopWatchVec, this computes percentiles per metric tags combination on the client-
	// side. This makes it impossible to aggregate percentiles across tags (e.g. to have system-wide view). When enabled
	// through a flag in the constructor, we initialize this additional untagged stopwatch to compute percentiles
	// across tags.
	promutils.StopWatch

	additionalLabels []contextutils.Key
}

// Start creates a new Instance of the StopWatch called a Timer that is closeable/stoppable.
// Common pattern to time a scope would be
// {
//   timer := stopWatch.Start(ctx)
//   defer timer.Stop()
//   ....
// }
func (c StopWatch) Start(ctx context.Context) Timer {
	w, err := c.StopWatchVec.GetMetricWith(contextutils.Values(ctx, append(metricKeys, c.additionalLabels...)...))
	if err != nil {
		panic(err.Error())
	}

	if c.StopWatch.Observer == nil {
		return w.Start()
	}

	return timer{
		Timers: []Timer{
			w.Start(),
			c.StopWatch.Start(),
		},
	}
}

// Observes specified duration between the start and end time. The data point will be labeled with values from context.
// See labeled.SetMetricsKeys for information about how to configure that.
func (c StopWatch) Observe(ctx context.Context, start, end time.Time) {
	w, err := c.StopWatchVec.GetMetricWith(contextutils.Values(ctx, append(metricKeys, c.additionalLabels...)...))
	if err != nil {
		panic(err.Error())
	}
	w.Observe(start, end)

	if c.StopWatch.Observer != nil {
		c.StopWatch.Observe(start, end)
	}
}

// This method observes the elapsed duration since the creation of the timer. The timer is created using a StopWatch.
// The data point will be labeled with values from context. See labeled.SetMetricsKeys for information about to
// configure that.
func (c StopWatch) Time(ctx context.Context, f func()) {
	t := c.Start(ctx)
	f()
	t.Stop()
}

// Creates a new labeled stopwatch. Label keys must be set before instantiating a counter. See labeled.SetMetricsKeys
// for information about how to configure that.
func NewStopWatch(name, description string, scale time.Duration, scope promutils.Scope, opts ...MetricOption) StopWatch {
	if len(metricKeys) == 0 {
		panic(ErrNeverSet)
	}

	sw := StopWatch{}

	for _, opt := range opts {
		if _, emitUnableMetric := opt.(EmitUnlabeledMetricOption); emitUnableMetric {
			sw.StopWatch = scope.MustNewStopWatch(GetUnlabeledMetricName(name), description, scale)
		} else if additionalLabels, casted := opt.(AdditionalLabelsOption); casted {
			sw.StopWatchVec = scope.MustNewStopWatchVec(name, description, scale,
				append(metricStringKeys, additionalLabels.Labels...)...)
			sw.additionalLabels = contextutils.MetricKeysFromStrings(additionalLabels.Labels)
		}
	}

	if sw.StopWatchVec == nil {
		sw.StopWatchVec = scope.MustNewStopWatchVec(name, description, scale, metricStringKeys...)
	}

	return sw
}
