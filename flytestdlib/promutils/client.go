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

var latencyBuckets = []float64{.0005, .001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

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
	requestLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "k8s_client_request_latency",
			Help:    "Kubernetes client request latency in seconds",
			Buckets: latencyBuckets,
		},
		[]string{"verb"})
	prometheus.MustRegister(requestLatency)
	requestResult := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "k8s_client_request_total",
			Help: "Kubernetes client request total",
		},
		[]string{"code", "method"},
	)
	prometheus.MustRegister(requestResult)
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
	rateLimiterLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "k8s_client_rate_limiter_latency",
			Help:    "Kubernetes client rate limiter latency in seconds",
			Buckets: latencyBuckets,
		},
		[]string{"verb"})
	prometheus.MustRegister(rateLimiterLatency)
	return rateLimiterMetricsProvider{
		rateLimiterLatency,
	}
}
