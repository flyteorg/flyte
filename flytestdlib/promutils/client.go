package promutils

import (
	"context"
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/tools/metrics"
)

func init() {
	requestMetrics := newRequestMetricsProvider()
	rateLimiterMetrics := newRateLimiterMetricsAdapter()
	metrics.Register(metrics.RegisterOpts{
		RequestLatency:     &requestMetrics,
		RequestResult:      &requestMetrics,
		RateLimiterLatency: &rateLimiterMetrics,
	})
}

var (
	scope          = NewScope("v2")
	latencyBuckets = []float64{.0005, .001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
)

type requestMetricsProvider struct {
	requestLatency *prometheus.HistogramVec
	requestResult  *prometheus.CounterVec
}

func (r *requestMetricsProvider) Observe(ctx context.Context, verb string, _ url.URL, latency time.Duration) {
	r.requestLatency.WithLabelValues(verb).Observe(latency.Seconds())
}

func (r *requestMetricsProvider) Increment(ctx context.Context, code string, method string, _ string) {
	r.requestResult.WithLabelValues(code, method).Inc()
}

func newRequestMetricsProvider() requestMetricsProvider {
	requestLatency := scope.MustNewHistogramVecWithOptions(
		"k8s_client_request_latency",
		"Kubernetes client request latency in seconds",
		HistogramOptions{Buckets: latencyBuckets},
		"verb",
	)
	requestResult := scope.MustNewCounterVec(
		"k8s_client_request_total",
		"Kubernetes client request total",
		"code", "method",
	)
	return requestMetricsProvider{
		requestLatency,
		requestResult,
	}
}

type rateLimiterMetricsProvider struct {
	rateLimiterLatency *prometheus.HistogramVec
}

func (r *rateLimiterMetricsProvider) Observe(ctx context.Context, verb string, _ url.URL, latency time.Duration) {
	r.rateLimiterLatency.WithLabelValues(verb).Observe(latency.Seconds())
}

func newRateLimiterMetricsAdapter() rateLimiterMetricsProvider {
	rateLimiterLatency := scope.MustNewHistogramVecWithOptions(
		"k8s_client_rate_limiter_latency",
		"Kubernetes client rate limiter latency in seconds",
		HistogramOptions{Buckets: latencyBuckets},
		"verb",
	)
	return rateLimiterMetricsProvider{
		rateLimiterLatency,
	}
}
