package labeled

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/v2/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

type HistogramStopWatch struct {
	*promutils.HistogramStopWatchVec
	promutils.HistogramStopWatch

	labels []contextutils.Key
}

// Start creates a new Instance of the HistogramStopWatch called a Timer that is closeable/stoppable.
func (c HistogramStopWatch) Start(ctx context.Context) Timer {
	w, err := c.GetMetricWith(contextutils.Values(ctx, c.labels...))
	if err != nil {
		panic(err.Error())
	}

	if c.Observer == nil {
		return w.Start()
	}

	return timer{
		Timers: []Timer{
			w.Start(),
			c.HistogramStopWatch.Start(),
		},
	}
}

// Observe observes specified duration between the start and end time. The data point will be labeled with values from context.
// See labeled.SetMetricsKeys for information about how to configure that.
func (c HistogramStopWatch) Observe(ctx context.Context, start, end time.Time) {
	w, err := c.GetMetricWith(contextutils.Values(ctx, c.labels...))
	if err != nil {
		panic(err.Error())
	}
	w.Observe(start, end)

	if c.Observer != nil {
		c.HistogramStopWatch.Observe(start, end)
	}
}

// Time observes the elapsed duration since the creation of the timer. The timer is created using a StopWatch.
// The data point will be labeled with values from context. See labeled.SetMetricsKeys for information about to
// configure that.
func (c HistogramStopWatch) Time(ctx context.Context, f func()) {
	t := c.Start(ctx)
	f()
	t.Stop()
}

// NewHistogramStopWatch creates a new labeled HistogramStopWatch. Label keys must be set before instantiating a counter. See labeled.SetMetricsKeys
// for information about how to configure that.
func NewHistogramStopWatch(name, description string, scope promutils.Scope, opts ...MetricOption) HistogramStopWatch {
	if len(metricKeys) == 0 {
		panic(ErrNeverSet)
	}

	sw := HistogramStopWatch{}

	name = promutils.SanitizeMetricName(name)
	for _, opt := range opts {
		if _, emitUnableMetric := opt.(EmitUnlabeledMetricOption); emitUnableMetric {
			sw.HistogramStopWatch = scope.MustNewHistogramStopWatch(GetUnlabeledMetricName(name), description)
		} else if additionalLabels, casted := opt.(AdditionalLabelsOption); casted {
			// compute unique labels
			labelSet := sets.NewString(metricStringKeys...)
			labelSet.Insert(additionalLabels.Labels...)
			labels := labelSet.List()

			sw.HistogramStopWatchVec = scope.MustNewHistogramStopWatchVec(name, description, labels...)
			sw.labels = contextutils.MetricKeysFromStrings(labels)
		}
	}

	if sw.HistogramStopWatchVec == nil {
		sw.HistogramStopWatchVec = scope.MustNewHistogramStopWatchVec(name, description, metricStringKeys...)
		sw.labels = metricKeys
	}

	return sw
}
