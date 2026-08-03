package promutils

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func useTestRegistry(t *testing.T) *prometheus.Registry {
	t.Helper()
	registerer := prometheus.DefaultRegisterer
	gatherer := prometheus.DefaultGatherer
	registry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = registry
	prometheus.DefaultGatherer = registry
	t.Cleanup(func() {
		prometheus.DefaultRegisterer = registerer
		prometheus.DefaultGatherer = gatherer
	})
	return registry
}

func gatherMetricFamilies(t *testing.T, gatherer prometheus.Gatherer) map[string]*dto.MetricFamily {
	t.Helper()
	families, err := gatherer.Gather()
	require.NoError(t, err)
	result := make(map[string]*dto.MetricFamily, len(families))
	for _, family := range families {
		result[family.GetName()] = family
	}
	return result
}

func TestMirroredScope(t *testing.T) {
	registry := useTestRegistry(t)
	scope := NewMirroredScope(NewScope("primary"), NewScope("legacy"))

	counter := scope.MustNewCounter("counter", "counter")
	counter.Add(2)
	gauge := scope.MustNewGauge("gauge", "gauge")
	gauge.Set(3)
	counterVec := scope.MustNewCounterVec("counter_vec", "counter vec", "kind")
	counterVec.WithLabelValues("test").Add(4)
	gaugeVec := scope.MustNewGaugeVec("gauge_vec", "gauge vec", "kind")
	gaugeVec.WithLabelValues("test").Set(5)
	summary := scope.MustNewSummary("summary", "summary")
	summary.Observe(6)
	summaryWithOptions := scope.MustNewSummaryWithOptions("summary_options", "summary options", SummaryOptions{
		Objectives: map[float64]float64{0.75: 0.01},
	})
	summaryWithOptions.Observe(6.5)
	summaryVec := scope.MustNewSummaryVec("summary_vec", "summary vec", "kind")
	summaryVec.WithLabelValues("test").Observe(7)
	histogram := scope.MustNewHistogram("histogram", "histogram")
	histogram.Observe(8)
	histogramVec := scope.MustNewHistogramVec("histogram_vec", "histogram vec", "kind")
	histogramVec.WithLabelValues("test").Observe(9)
	histogramVecWithOptions := scope.MustNewHistogramVecWithOptions(
		"histogram_vec_options", "histogram vec options", HistogramOptions{Buckets: []float64{1, 10}}, "kind",
	)
	histogramVecWithOptions.WithLabelValues("test").Observe(9.5)
	stopwatch := scope.MustNewStopWatch("stopwatch", "stopwatch", time.Millisecond)
	stopwatch.Observe(time.Unix(0, 0), time.Unix(0, int64(10*time.Millisecond)))
	stopwatchVec := scope.MustNewStopWatchVec("stopwatch_vec", "stopwatch vec", time.Millisecond, "kind")
	stopwatchVec.WithLabelValues("test").Observe(time.Unix(0, 0), time.Unix(0, int64(11*time.Millisecond)))
	histogramStopwatch := scope.MustNewHistogramStopWatch("histogram_stopwatch", "histogram stopwatch")
	histogramStopwatch.Observe(time.Unix(0, 0), time.Unix(0, int64(time.Second)))
	histogramStopwatchVec := scope.MustNewHistogramStopWatchVec(
		"histogram_stopwatch_vec", "histogram stopwatch vec", "kind",
	)
	histogramStopwatchVec.WithLabelValues("test").Observe(time.Unix(0, 0), time.Unix(0, int64(2*time.Second)))
	nested := scope.NewSubScope("nested").MustNewCounter("counter", "nested counter")
	nested.Inc()

	require.Equal(t, "primary:", scope.CurrentScope())
	require.Equal(t, "primary:metric", scope.NewScopedMetricName("metric"))

	families := gatherMetricFamilies(t, registry)
	for _, name := range []string{
		"counter", "gauge", "counter_vec", "gauge_vec", "summary", "summary_options", "summary_vec",
		"histogram", "histogram_vec", "histogram_vec_options", "stopwatch_ms", "stopwatch_vec_ms",
		"histogram_stopwatch", "histogram_stopwatch_vec",
	} {
		require.Contains(t, families, "primary:"+name)
		require.Contains(t, families, "legacy:"+name)
		require.Equal(t, families["primary:"+name].Metric, families["legacy:"+name].Metric, name)
	}
	require.Contains(t, families, "primary:nested:counter")
	require.Contains(t, families, "legacy:nested:counter")
	require.Equal(t, families["primary:nested:counter"].Metric, families["legacy:nested:counter"].Metric)
}

func TestMirroredScopeDeduplicatesScopes(t *testing.T) {
	registry := useTestRegistry(t)
	primary := NewScope("same")
	scope := NewMirroredScope(primary, primary, NewScope("same"))
	scope.MustNewCounter("counter", "counter").Inc()

	families := gatherMetricFamilies(t, registry)
	require.Contains(t, families, "same:counter")
	require.Len(t, families, 1)
}

func TestMirroredScopeSanitizesLeadingDigitMetricNames(t *testing.T) {
	registry := useTestRegistry(t)
	scope := NewMirroredScope(NewScope("primary"), NewScope("legacy"))
	scope.MustNewCounter("2xx", "leading digit").Inc()
	scope.MustNewCounter("xx", "without leading digit").Add(2)

	families := gatherMetricFamilies(t, registry)
	for _, name := range []string{"_2xx", "xx"} {
		require.Contains(t, families, "primary:"+name)
		require.Contains(t, families, "legacy:"+name)
		require.Equal(t, families["primary:"+name].Metric, families["legacy:"+name].Metric)
	}
	require.Len(t, families, 4)
}

func TestMirroredScopeRollsBackPartialRegistration(t *testing.T) {
	registry := useTestRegistry(t)
	registry.MustRegister(prometheus.NewCounter(prometheus.CounterOpts{Name: "legacy:counter", Help: "conflict"}))

	scope := NewMirroredScope(NewScope("primary"), NewScope("legacy"))
	_, err := scope.NewCounter("counter", "counter")
	require.Error(t, err)

	families := gatherMetricFamilies(t, registry)
	require.NotContains(t, families, "primary:counter")
	require.Contains(t, families, "legacy:counter")
}

func TestMirroredScopeRejectsNilScopes(t *testing.T) {
	require.Panics(t, func() { NewMirroredScope(nil) })
	require.Panics(t, func() { NewMirroredScope(NewScope("primary"), nil) })
}
