package labeled

import (
	"context"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"
)

// Summary represents a summary labeled with values from the context. See labeled.SetMetricsKeys for information about
// how to configure that.
type Summary struct {
	*prometheus.SummaryVec

	prometheus.Summary
	additionalLabels []contextutils.Key
}

// Observe adds a single observation to the summary.
func (s Summary) Observe(ctx context.Context, v float64) {
	summary, err := s.SummaryVec.GetMetricWith(contextutils.Values(ctx, append(metricKeys, s.additionalLabels...)...))
	if err != nil {
		panic(err.Error())
	}
	summary.Observe(v)

	if s.Summary != nil {
		s.Summary.Observe(v)
	}
}

// NewSummary creates a new labeled summary. Label keys must be set before instantiating. If the unlabeled option is
// given, this object will also instantiate and emit another summary with the given name with an _unlabeled suffix.
// See labeled.SetMetricsKeys for information about how to configure that.
func NewSummary(name, description string, scope promutils.Scope, opts ...MetricOption) Summary {
	if len(metricKeys) == 0 {
		panic(ErrNeverSet)
	}

	s := Summary{}
	for _, opt := range opts {
		if _, emitUnlabeledMetric := opt.(EmitUnlabeledMetricOption); emitUnlabeledMetric {
			s.Summary = scope.MustNewSummary(GetUnlabeledMetricName(name), description)
		} else if additionalLabels, casted := opt.(AdditionalLabelsOption); casted {
			s.SummaryVec = scope.MustNewSummaryVec(name, description, append(metricStringKeys, additionalLabels.Labels...)...)
			s.additionalLabels = contextutils.MetricKeysFromStrings(additionalLabels.Labels)
		}
	}

	if s.SummaryVec == nil {
		s.SummaryVec = scope.MustNewSummaryVec(name, description, metricStringKeys...)
	}

	return s
}
