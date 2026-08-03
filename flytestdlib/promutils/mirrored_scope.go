package promutils

import (
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// mirroredScope registers each metric under a primary scope and one or more alias scopes.
// All registrations collect from the same metric instance.
type mirroredScope struct {
	primary Scope
	scopes  []string
}

var _ Scope = mirroredScope{}

// NewMirroredScope creates a scope that publishes metrics under the primary scope and each alias.
func NewMirroredScope(primary Scope, aliases ...Scope) Scope {
	if primary == nil {
		panic("primary metric scope cannot be nil")
	}

	scopes := make([]string, 0, len(aliases)+1)
	seen := make(map[string]struct{}, len(aliases)+1)
	for _, scope := range append([]Scope{primary}, aliases...) {
		if scope == nil {
			panic("metric alias scope cannot be nil")
		}
		name := scope.CurrentScope()
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		scopes = append(scopes, name)
	}

	return mirroredScope{primary: primary, scopes: scopes}
}

func (m mirroredScope) register(collector prometheus.Collector) error {
	registered := make([]prometheus.Registerer, 0, len(m.scopes))
	for _, scope := range m.scopes {
		registerer := prometheus.WrapRegistererWithPrefix(scope, prometheus.DefaultRegisterer)
		if err := registerer.Register(collector); err != nil {
			for i := len(registered) - 1; i >= 0; i-- {
				registered[i].Unregister(collector)
			}
			return fmt.Errorf("register metric in scope %q: %w", scope, err)
		}
		registered = append(registered, registerer)
	}
	return nil
}

func metricName(name string) string {
	if name == "" {
		panic("metric name cannot be an empty string")
	}
	return SanitizeMetricName(name)
}

func (m mirroredScope) NewGauge(name, description string) (prometheus.Gauge, error) {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{Name: metricName(name), Help: description})
	return gauge, m.register(gauge)
}

func (m mirroredScope) MustNewGauge(name, description string) prometheus.Gauge {
	gauge, err := m.NewGauge(name, description)
	panicIfError(err)
	return gauge
}

func (m mirroredScope) NewGaugeVec(name, description string, labelNames ...string) (*prometheus.GaugeVec, error) {
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: metricName(name), Help: description}, labelNames)
	return gauge, m.register(gauge)
}

func (m mirroredScope) MustNewGaugeVec(name, description string, labelNames ...string) *prometheus.GaugeVec {
	gauge, err := m.NewGaugeVec(name, description, labelNames...)
	panicIfError(err)
	return gauge
}

func (m mirroredScope) NewSummary(name, description string) (prometheus.Summary, error) {
	return m.NewSummaryWithOptions(name, description, SummaryOptions{Objectives: defaultObjectives})
}

func (m mirroredScope) MustNewSummary(name, description string) prometheus.Summary {
	summary, err := m.NewSummary(name, description)
	panicIfError(err)
	return summary
}

func (m mirroredScope) NewSummaryWithOptions(
	name, description string, options SummaryOptions,
) (prometheus.Summary, error) {
	summary := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       metricName(name),
		Help:       description,
		Objectives: options.Objectives,
	})
	return summary, m.register(summary)
}

func (m mirroredScope) MustNewSummaryWithOptions(name, description string, options SummaryOptions) prometheus.Summary {
	summary, err := m.NewSummaryWithOptions(name, description, options)
	panicIfError(err)
	return summary
}

func (m mirroredScope) NewSummaryVec(name, description string, labelNames ...string) (*prometheus.SummaryVec, error) {
	summary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       metricName(name),
		Help:       description,
		Objectives: defaultObjectives,
	}, labelNames)
	return summary, m.register(summary)
}

func (m mirroredScope) MustNewSummaryVec(name, description string, labelNames ...string) *prometheus.SummaryVec {
	summary, err := m.NewSummaryVec(name, description, labelNames...)
	panicIfError(err)
	return summary
}

func (m mirroredScope) NewHistogram(name, description string) (prometheus.Histogram, error) {
	histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    metricName(name),
		Help:    description,
		Buckets: defaultBuckets,
	})
	return histogram, m.register(histogram)
}

func (m mirroredScope) MustNewHistogram(name, description string) prometheus.Histogram {
	histogram, err := m.NewHistogram(name, description)
	panicIfError(err)
	return histogram
}

func (m mirroredScope) NewHistogramVec(
	name, description string, labelNames ...string,
) (*prometheus.HistogramVec, error) {
	return m.NewHistogramVecWithOptions(name, description, HistogramOptions{Buckets: defaultBuckets}, labelNames...)
}

func (m mirroredScope) MustNewHistogramVec(name, description string, labelNames ...string) *prometheus.HistogramVec {
	histogram, err := m.NewHistogramVec(name, description, labelNames...)
	panicIfError(err)
	return histogram
}

func (m mirroredScope) NewHistogramVecWithOptions(
	name, description string, options HistogramOptions, labelNames ...string,
) (*prometheus.HistogramVec, error) {
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    metricName(name),
		Help:    description,
		Buckets: options.Buckets,
	}, labelNames)
	return histogram, m.register(histogram)
}

func (m mirroredScope) MustNewHistogramVecWithOptions(
	name, description string, options HistogramOptions, labelNames ...string,
) *prometheus.HistogramVec {
	histogram, err := m.NewHistogramVecWithOptions(name, description, options, labelNames...)
	panicIfError(err)
	return histogram
}

func (m mirroredScope) NewCounter(name, description string) (prometheus.Counter, error) {
	counter := prometheus.NewCounter(prometheus.CounterOpts{Name: metricName(name), Help: description})
	return counter, m.register(counter)
}

func (m mirroredScope) MustNewCounter(name, description string) prometheus.Counter {
	counter, err := m.NewCounter(name, description)
	panicIfError(err)
	return counter
}

func (m mirroredScope) NewCounterVec(name, description string, labelNames ...string) (*prometheus.CounterVec, error) {
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{Name: metricName(name), Help: description}, labelNames)
	return counter, m.register(counter)
}

func (m mirroredScope) MustNewCounterVec(name, description string, labelNames ...string) *prometheus.CounterVec {
	counter, err := m.NewCounterVec(name, description, labelNames...)
	panicIfError(err)
	return counter
}

func (m mirroredScope) NewStopWatch(name, description string, scale time.Duration) (StopWatch, error) {
	name = scaledMetricName(name, scale)
	summary, err := m.NewSummary(name, description)
	if err != nil {
		return StopWatch{}, err
	}
	return StopWatch{Observer: summary, outputScale: scale}, nil
}

func (m mirroredScope) MustNewStopWatch(name, description string, scale time.Duration) StopWatch {
	stopwatch, err := m.NewStopWatch(name, description, scale)
	panicIfError(err)
	return stopwatch
}

func (m mirroredScope) NewStopWatchVec(
	name, description string, scale time.Duration, labelNames ...string,
) (*StopWatchVec, error) {
	name = scaledMetricName(name, scale)
	summary, err := m.NewSummaryVec(name, description, labelNames...)
	if err != nil {
		return &StopWatchVec{}, err
	}
	return &StopWatchVec{SummaryVec: summary, outputScale: scale}, nil
}

func (m mirroredScope) MustNewStopWatchVec(
	name, description string, scale time.Duration, labelNames ...string,
) *StopWatchVec {
	stopwatch, err := m.NewStopWatchVec(name, description, scale, labelNames...)
	panicIfError(err)
	return stopwatch
}

func (m mirroredScope) NewHistogramStopWatch(name, description string) (HistogramStopWatch, error) {
	histogram, err := m.NewHistogram(name, description)
	if err != nil {
		return HistogramStopWatch{}, err
	}
	return HistogramStopWatch{StopWatch: StopWatch{Observer: histogram, outputScale: time.Second}}, nil
}

func (m mirroredScope) MustNewHistogramStopWatch(name, description string) HistogramStopWatch {
	stopwatch, err := m.NewHistogramStopWatch(name, description)
	panicIfError(err)
	return stopwatch
}

func (m mirroredScope) NewHistogramStopWatchVec(
	name, description string, labelNames ...string,
) (*HistogramStopWatchVec, error) {
	histogram, err := m.NewHistogramVec(name, description, labelNames...)
	if err != nil {
		return &HistogramStopWatchVec{}, err
	}
	return &HistogramStopWatchVec{HistogramVec: histogram, outputScale: time.Second}, nil
}

func (m mirroredScope) MustNewHistogramStopWatchVec(
	name, description string, labelNames ...string,
) *HistogramStopWatchVec {
	stopwatch, err := m.NewHistogramStopWatchVec(name, description, labelNames...)
	panicIfError(err)
	return stopwatch
}

func (m mirroredScope) NewSubScope(name string) Scope {
	primary := m.primary.NewSubScope(name)
	aliases := make([]Scope, 0, len(m.scopes)-1)
	for _, scope := range m.scopes[1:] {
		aliases = append(aliases, NewScope(scope).NewSubScope(name))
	}
	return NewMirroredScope(primary, aliases...)
}

func (m mirroredScope) CurrentScope() string {
	return m.primary.CurrentScope()
}

func (m mirroredScope) NewScopedMetricName(name string) string {
	return m.primary.NewScopedMetricName(name)
}

func scaledMetricName(name string, scale time.Duration) string {
	if !strings.HasSuffix(name, defaultMetricDelimiterStr) {
		name += defaultMetricDelimiterStr
	}
	return name + DurationToString(scale)
}
