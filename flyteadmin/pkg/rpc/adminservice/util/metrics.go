package util

import (
	"context"
	"fmt"
	"time"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
)

const maxGRPCStatusCode = 17 // From _maxCode in "google.golang.org/grpc/codes"

type responseCodeMetrics struct {
	scope                promutils.Scope
	responseCodeCounters map[codes.Code]prometheus.Counter
}

// RequestMetrics per-endpoint request metrics.
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
		logger.Warnf(context.TODO(), "encountered unexpected error code [%s]", code.String())
		return
	}
	counter.Inc()
}

func (m *RequestMetrics) Success() {
	m.Record(codes.OK)
}

func newResponseCodeMetrics(scope promutils.Scope) responseCodeMetrics {
	responseCodeCounters := make(map[codes.Code]prometheus.Counter)
	for i := 0; i < maxGRPCStatusCode; i++ {
		code := codes.Code(i)
		responseCodeCounters[code] = scope.MustNewCounter(code.String(),
			fmt.Sprintf("count of responses returning: %s", code.String()))
	}
	return responseCodeMetrics{
		scope:                scope,
		responseCodeCounters: responseCodeCounters,
	}
}

func NewRequestMetrics(scope promutils.Scope, method string) RequestMetrics {
	methodScope := scope.NewSubScope(method)
	responseCodeMetrics := newResponseCodeMetrics(methodScope.NewSubScope("codes"))

	return RequestMetrics{
		scope: methodScope,

		requestDuration: methodScope.MustNewStopWatch("duration",
			"recorded response time duration for endpoint", time.Millisecond),
		errCount:      methodScope.MustNewCounter("errors", "count of errors returned by endpoint"),
		successCount:  methodScope.MustNewCounter("success", "count of successful responses returned by endpoint"),
		responseCodes: responseCodeMetrics,
	}
}
