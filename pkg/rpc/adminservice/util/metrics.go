package util

import (
	"fmt"
	"time"

	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
)

type responseCodeMetrics struct {
	scope                promutils.Scope
	responseCodeCounters map[codes.Code]prometheus.Counter
}

// Per-endpoint request metrics.
type RequestMetrics struct {
	scope promutils.Scope

	requestDuration promutils.StopWatch
	errCount        prometheus.Counter
	successCount    prometheus.Counter
	responseCodes   responseCodeMetrics
}

func (m *RequestMetrics) Time(fn func()) {
	m.requestDuration.Time(fn)
}

func (m *RequestMetrics) Record(code codes.Code) {
	if code == codes.OK {
		m.successCount.Inc()
	} else {
		m.errCount.Inc()
	}
	counter, ok := m.responseCodes.responseCodeCounters[code]
	if !ok {
		m.responseCodes.responseCodeCounters[code] = m.responseCodes.scope.MustNewCounter(code.String(),
			fmt.Sprintf("count of responses returning: %s", code.String()))
		counter = m.responseCodes.responseCodeCounters[code]
	}
	counter.Inc()
}

func (m *RequestMetrics) Success() {
	m.Record(codes.OK)
}

func NewRequestMetrics(scope promutils.Scope, method string) RequestMetrics {
	methodScope := scope.NewSubScope(method)

	responseCodeMetrics := responseCodeMetrics{
		scope:                methodScope.NewSubScope("codes"),
		responseCodeCounters: make(map[codes.Code]prometheus.Counter),
	}

	return RequestMetrics{
		scope: methodScope,

		requestDuration: methodScope.MustNewStopWatch("duration",
			"recorded response time duration for endpoint", time.Millisecond),
		errCount:      methodScope.MustNewCounter("errors", "count of errors returned by endpoint"),
		successCount:  methodScope.MustNewCounter("success", "count of successful responses returned by endpoint"),
		responseCodes: responseCodeMetrics,
	}
}
