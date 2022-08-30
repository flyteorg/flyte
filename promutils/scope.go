package promutils

import (
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/prometheus/client_golang/prometheus"
)

const defaultScopeDelimiterStr = ":"
const defaultMetricDelimiterStr = "_"

var (
	defaultObjectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	defaultBuckets    = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
)

func panicIfError(err error) {
	if err != nil {
		panic("Failed to register metrics. Error: " + err.Error())
	}
}

// StopWatch implements a stopwatch style interface that works with prometheus summary
// It will scale the output to match the expected time scale (milliseconds, seconds etc)
// NOTE: Do not create a StopWatch object by hand, use a Scope to get a new instance of the StopWatch object
type StopWatch struct {
	prometheus.Observer
	outputScale time.Duration
}

// Start creates a new Instance of the StopWatch called a Timer that is closeable/stoppable.
// Common pattern to time a scope would be
func (s StopWatch) Start() Timer {
	return Timer{
		start:       time.Now(),
		outputScale: s.outputScale,
		timer:       s.Observer,
	}
}

// Observe records a specified duration between the start and end time
func (s StopWatch) Observe(start, end time.Time) {
	observed := end.Sub(start).Nanoseconds()
	outputScaleDuration := s.outputScale.Nanoseconds()
	if outputScaleDuration == 0 {
		s.Observer.Observe(0)
		return
	}
	scaled := float64(observed / outputScaleDuration)
	s.Observer.Observe(scaled)
}

// Time Observes/records the time to execute the given function synchronously
func (s StopWatch) Time(f func()) {
	t := s.Start()
	f()
	t.Stop()
}

// A Simple StopWatch that works with prometheus summary
// It will scale the output to match the expected time scale (milliseconds, seconds etc)
// NOTE: Do not create a StopWatch object by hand, use a Scope to get a new instance of the StopWatch object
type StopWatchVec struct {
	*prometheus.SummaryVec
	outputScale time.Duration
}

// Gets a concrete StopWatch instance that can be used to start a timer and record observations.
func (s StopWatchVec) WithLabelValues(values ...string) StopWatch {
	return StopWatch{
		Observer:    s.SummaryVec.WithLabelValues(values...),
		outputScale: s.outputScale,
	}
}

func (s StopWatchVec) GetMetricWith(labels prometheus.Labels) (StopWatch, error) {
	sVec, err := s.SummaryVec.GetMetricWith(labels)
	if err != nil {
		return StopWatch{}, err
	}
	return StopWatch{
		Observer:    sVec,
		outputScale: s.outputScale,
	}, nil
}

// Timer is a stoppable instance of a StopWatch or a Timer
// A Timer can only be stopped. On stopping it will output the elapsed duration to prometheus
type Timer struct {
	start       time.Time
	outputScale time.Duration
	timer       prometheus.Observer
}

// Stop observes the elapsed duration since the creation of the timer. The timer is created using a StopWatch
func (s Timer) Stop() float64 {
	observed := time.Since(s.start).Nanoseconds()
	outputScaleDuration := s.outputScale.Nanoseconds()
	if outputScaleDuration == 0 {
		s.timer.Observe(0)
		return 0
	}
	scaled := float64(observed / outputScaleDuration)
	s.timer.Observe(scaled)
	return scaled
}

// A SummaryOptions represents a set of options that can be supplied when creating a new prometheus summary metric
type SummaryOptions struct {
	// An Objectives defines the quantile rank estimates with their respective absolute errors.
	// Refer to https://godoc.org/github.com/prometheus/client_golang/prometheus#SummaryOpts for details
	Objectives map[float64]float64
}

// A Scope represents a prefix in Prometheus. It is nestable, thus every metric that is published does not need to
// provide a prefix, but just the name of the metric. As long as the Scope is used to create a new instance of the metric
// The prefix (or scope) is automatically set.
type Scope interface {
	// NewGauge creates new prometheus.Gauge metric with the prefix as the CurrentScope
	// Name is a string that follows prometheus conventions (mostly [_a-z])
	// Refer to https://prometheus.io/docs/concepts/metric_types/ for more information
	NewGauge(name, description string) (prometheus.Gauge, error)
	MustNewGauge(name, description string) prometheus.Gauge

	// NewGaugeVec creates new prometheus.GaugeVec metric with the prefix as the CurrentScope
	// Refer to https://prometheus.io/docs/concepts/metric_types/ for more information
	NewGaugeVec(name, description string, labelNames ...string) (*prometheus.GaugeVec, error)
	MustNewGaugeVec(name, description string, labelNames ...string) *prometheus.GaugeVec

	// NewSummary creates new prometheus.Summary metric with the prefix as the CurrentScope
	// Refer to https://prometheus.io/docs/concepts/metric_types/ for more information
	NewSummary(name, description string) (prometheus.Summary, error)
	MustNewSummary(name, description string) prometheus.Summary

	// NewSummaryWithOptions creates new prometheus.Summary metric with custom options, such as a custom set of objectives (i.e., target quantiles).
	// Refer to https://prometheus.io/docs/concepts/metric_types/ for more information
	NewSummaryWithOptions(name, description string, options SummaryOptions) (prometheus.Summary, error)
	MustNewSummaryWithOptions(name, description string, options SummaryOptions) prometheus.Summary

	// NewSummaryVec creates new prometheus.SummaryVec metric with the prefix as the CurrentScope
	// Refer to https://prometheus.io/docs/concepts/metric_types/ for more information
	NewSummaryVec(name, description string, labelNames ...string) (*prometheus.SummaryVec, error)
	MustNewSummaryVec(name, description string, labelNames ...string) *prometheus.SummaryVec

	// NewHistogram creates new prometheus.Histogram metric with the prefix as the CurrentScope
	// Refer to https://prometheus.io/docs/concepts/metric_types/ for more information
	NewHistogram(name, description string) (prometheus.Histogram, error)
	MustNewHistogram(name, description string) prometheus.Histogram

	// NewHistogramVec creates new prometheus.HistogramVec metric with the prefix as the CurrentScope
	// Refer to https://prometheus.io/docs/concepts/metric_types/ for more information
	NewHistogramVec(name, description string, labelNames ...string) (*prometheus.HistogramVec, error)
	MustNewHistogramVec(name, description string, labelNames ...string) *prometheus.HistogramVec

	// NewCounter creates new prometheus.Counter metric with the prefix as the CurrentScope
	// Refer to https://prometheus.io/docs/concepts/metric_types/ for more information
	// Important to note, counters are not like typical counters. These are ever increasing and cumulative.
	// So if you want to observe counters within buckets use Summary/Histogram
	NewCounter(name, description string) (prometheus.Counter, error)
	MustNewCounter(name, description string) prometheus.Counter

	// NewCounterVec creates new prometheus.GaugeVec metric with the prefix as the CurrentScope
	// Refer to https://prometheus.io/docs/concepts/metric_types/ for more information
	NewCounterVec(name, description string, labelNames ...string) (*prometheus.CounterVec, error)
	MustNewCounterVec(name, description string, labelNames ...string) *prometheus.CounterVec

	// NewStopWatch is a custom wrapper to create a StopWatch object in the current Scope.
	// Duration is to specify the scale of the Timer. For example if you are measuring times in milliseconds
	// pass scale=times.Millisecond
	// https://golang.org/pkg/time/#Duration
	// The metric name is auto-suffixed with the right scale. Refer to DurationToString to understand
	NewStopWatch(name, description string, scale time.Duration) (StopWatch, error)
	MustNewStopWatch(name, description string, scale time.Duration) StopWatch

	// NewStopWatchVec is a custom wrapper to create a StopWatch object in the current Scope.
	// Duration is to specify the scale of the Timer. For example if you are measuring times in milliseconds
	// pass scale=times.Millisecond
	// https://golang.org/pkg/time/#Duration
	// The metric name is auto-suffixed with the right scale. Refer to DurationToString to understand
	NewStopWatchVec(name, description string, scale time.Duration, labelNames ...string) (*StopWatchVec, error)
	MustNewStopWatchVec(name, description string, scale time.Duration, labelNames ...string) *StopWatchVec

	// NewSubScope creates a new subScope in case nesting is desired for metrics. This is generally useful in creating
	// Scoped and SubScoped metrics
	NewSubScope(name string) Scope

	// CurrentScope returns the current ScopeName. Use for creating your own metrics
	CurrentScope() string

	// NewScopedMetricName provides a scoped metric name. Can be used, if you want to directly create your own metric
	NewScopedMetricName(name string) string
}

type metricsScope struct {
	scope string
}

func (m metricsScope) NewGauge(name, description string) (prometheus.Gauge, error) {
	g := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: m.NewScopedMetricName(name),
			Help: description,
		},
	)
	return g, prometheus.Register(g)
}

func (m metricsScope) MustNewGauge(name, description string) prometheus.Gauge {
	g, err := m.NewGauge(name, description)
	panicIfError(err)
	return g
}

func (m metricsScope) NewGaugeVec(name, description string, labelNames ...string) (*prometheus.GaugeVec, error) {
	g := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: m.NewScopedMetricName(name),
			Help: description,
		},
		labelNames,
	)
	return g, prometheus.Register(g)
}

func (m metricsScope) MustNewGaugeVec(name, description string, labelNames ...string) *prometheus.GaugeVec {
	g, err := m.NewGaugeVec(name, description, labelNames...)
	panicIfError(err)
	return g
}

func (m metricsScope) NewSummary(name, description string) (prometheus.Summary, error) {
	return m.NewSummaryWithOptions(name, description, SummaryOptions{Objectives: defaultObjectives})
}

func (m metricsScope) MustNewSummary(name, description string) prometheus.Summary {
	s, err := m.NewSummary(name, description)
	panicIfError(err)
	return s
}

func (m metricsScope) NewSummaryWithOptions(name, description string, options SummaryOptions) (prometheus.Summary, error) {
	s := prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name:       m.NewScopedMetricName(name),
			Help:       description,
			Objectives: options.Objectives,
		},
	)

	return s, prometheus.Register(s)
}

func (m metricsScope) MustNewSummaryWithOptions(name, description string, options SummaryOptions) prometheus.Summary {
	s, err := m.NewSummaryWithOptions(name, description, options)
	panicIfError(err)
	return s
}

func (m metricsScope) NewSummaryVec(name, description string, labelNames ...string) (*prometheus.SummaryVec, error) {
	s := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       m.NewScopedMetricName(name),
			Help:       description,
			Objectives: defaultObjectives,
		},
		labelNames,
	)

	return s, prometheus.Register(s)
}
func (m metricsScope) MustNewSummaryVec(name, description string, labelNames ...string) *prometheus.SummaryVec {
	s, err := m.NewSummaryVec(name, description, labelNames...)
	panicIfError(err)
	return s
}

func (m metricsScope) NewHistogram(name, description string) (prometheus.Histogram, error) {
	h := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    m.NewScopedMetricName(name),
			Help:    description,
			Buckets: defaultBuckets,
		},
	)
	return h, prometheus.Register(h)
}

func (m metricsScope) MustNewHistogram(name, description string) prometheus.Histogram {
	h, err := m.NewHistogram(name, description)
	panicIfError(err)
	return h
}

func (m metricsScope) NewHistogramVec(name, description string, labelNames ...string) (*prometheus.HistogramVec, error) {
	h := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    m.NewScopedMetricName(name),
			Help:    description,
			Buckets: defaultBuckets,
		},
		labelNames,
	)
	return h, prometheus.Register(h)
}

func (m metricsScope) MustNewHistogramVec(name, description string, labelNames ...string) *prometheus.HistogramVec {
	h, err := m.NewHistogramVec(name, description, labelNames...)
	panicIfError(err)
	return h
}

func (m metricsScope) NewCounter(name, description string) (prometheus.Counter, error) {
	c := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: m.NewScopedMetricName(name),
			Help: description,
		},
	)
	return c, prometheus.Register(c)
}

func (m metricsScope) MustNewCounter(name, description string) prometheus.Counter {
	c, err := m.NewCounter(name, description)
	panicIfError(err)
	return c
}

func (m metricsScope) NewCounterVec(name, description string, labelNames ...string) (*prometheus.CounterVec, error) {
	c := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: m.NewScopedMetricName(name),
			Help: description,
		},
		labelNames,
	)
	return c, prometheus.Register(c)
}

func (m metricsScope) MustNewCounterVec(name, description string, labelNames ...string) *prometheus.CounterVec {
	c, err := m.NewCounterVec(name, description, labelNames...)
	panicIfError(err)
	return c
}

func (m metricsScope) NewStopWatch(name, description string, scale time.Duration) (StopWatch, error) {
	if !strings.HasSuffix(name, defaultMetricDelimiterStr) {
		name += defaultMetricDelimiterStr
	}
	name += DurationToString(scale)
	s, err := m.NewSummary(name, description)
	if err != nil {
		return StopWatch{}, err
	}

	return StopWatch{
		Observer:    s,
		outputScale: scale,
	}, nil
}

func (m metricsScope) MustNewStopWatch(name, description string, scale time.Duration) StopWatch {
	s, err := m.NewStopWatch(name, description, scale)
	panicIfError(err)
	return s
}

func (m metricsScope) NewStopWatchVec(name, description string, scale time.Duration, labelNames ...string) (*StopWatchVec, error) {
	if !strings.HasSuffix(name, defaultMetricDelimiterStr) {
		name += defaultMetricDelimiterStr
	}
	name += DurationToString(scale)
	s, err := m.NewSummaryVec(name, description, labelNames...)
	if err != nil {
		return &StopWatchVec{}, err
	}

	return &StopWatchVec{
		SummaryVec:  s,
		outputScale: scale,
	}, nil
}

func (m metricsScope) MustNewStopWatchVec(name, description string, scale time.Duration, labelNames ...string) *StopWatchVec {
	s, err := m.NewStopWatchVec(name, description, scale, labelNames...)
	panicIfError(err)
	return s
}

func (m metricsScope) CurrentScope() string {
	return m.scope
}

// NewScopedMetricName creates a metric name under the scope. Scope will always have a defaultScopeDelimiterRune as the
// last character
func (m metricsScope) NewScopedMetricName(name string) string {
	if name == "" {
		panic("metric name cannot be an empty string")
	}

	return SanitizeMetricName(m.scope + name)
}

func (m metricsScope) NewSubScope(subscopeName string) Scope {
	if subscopeName == "" {
		panic("scope name cannot be an empty string")
	}

	// If the last character of the new subscope is already a defaultScopeDelimiterRune, do not add anything
	if !strings.HasSuffix(subscopeName, defaultScopeDelimiterStr) {
		subscopeName += defaultScopeDelimiterStr
	}

	// Always add a new defaultScopeDelimiterRune to every scope name
	return NewScope(m.scope + subscopeName)
}

// NewScope creates a new scope in the format `name + defaultScopeDelimiterRune`
// If the last character is already a defaultScopeDelimiterRune, then it does not add it to the scope name
func NewScope(name string) Scope {
	if name == "" {
		panic("base scope for a metric cannot be an empty string")
	}

	// If the last character of the new subscope is already a defaultScopeDelimiterRune, do not add anything
	if !strings.HasSuffix(name, defaultScopeDelimiterStr) {
		name += defaultScopeDelimiterStr
	}

	return metricsScope{
		scope: SanitizeMetricName(name),
	}
}

// SanitizeMetricName ensures the generates metric name is compatible with the underlying prometheus library.
func SanitizeMetricName(name string) string {
	out := strings.Builder{}
	for i, b := range name {
		if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_' || b == ':' || (b >= '0' && b <= '9' && i > 0) {
			out.WriteRune(b)
		} else if b == '-' {
			out.WriteRune('_')
		}
	}

	return out.String()
}

// NewTestScope returns a randomly-named scope for use in tests.
// Prometheus requires that metric names begin with a single word, which is generated from the alphabetic testScopeNameCharset.
func NewTestScope() Scope {
	return NewScope("test" + rand.String(6))
}

// DurationToString converts the duration to a string suffix that indicates the scale of the timer.
func DurationToString(duration time.Duration) string {
	if duration >= time.Hour {
		return "h"
	}
	if duration >= time.Minute {
		return "m"
	}
	if duration >= time.Second {
		return "s"
	}
	if duration >= time.Millisecond {
		return "ms"
	}
	if duration >= time.Microsecond {
		return "us"
	}
	return "ns"
}
