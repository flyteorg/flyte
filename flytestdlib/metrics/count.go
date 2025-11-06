package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type SuccessAndFailureCounter struct {
	success prometheus.Counter
	failure prometheus.Counter
}

func (c SuccessAndFailureCounter) Inc(err *error) {
	if err != nil && *err != nil {
		c.failure.Inc()
	} else {
		c.success.Inc()
	}
}

func NewSuccessAndFailureCounter(scope promutils.Scope) SuccessAndFailureCounter {
	return SuccessAndFailureCounter{
		success: scope.MustNewCounter("success", "Number of successful events handled"),
		failure: scope.MustNewCounter("failure", "Number of failed events handled"),
	}
}

type SuccessAndFailureCounterVec struct {
	success *prometheus.CounterVec
	failure *prometheus.CounterVec
}

func (c SuccessAndFailureCounterVec) Inc(err *error, labels prometheus.Labels) {
	if err != nil && *err != nil {
		c.failure.With(labels).Inc()
	} else {
		c.success.With(labels).Inc()
	}
}

func NewSuccessAndFailureCounterVec(scope promutils.Scope, labels ...string) SuccessAndFailureCounterVec {
	return SuccessAndFailureCounterVec{
		success: scope.MustNewCounterVec("success", "Number of successful events handled", labels...),
		failure: scope.MustNewCounterVec("failure", "Number of failed events handled", labels...),
	}
}
