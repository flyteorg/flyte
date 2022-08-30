package promutils

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestDurationToString(t *testing.T) {
	assert.Equal(t, "m", DurationToString(time.Minute))
	assert.Equal(t, "m", DurationToString(time.Minute*10))
	assert.Equal(t, "h", DurationToString(time.Hour))
	assert.Equal(t, "h", DurationToString(time.Hour*10))
	assert.Equal(t, "s", DurationToString(time.Second))
	assert.Equal(t, "s", DurationToString(time.Second*10))
	assert.Equal(t, "us", DurationToString(time.Microsecond*10))
	assert.Equal(t, "us", DurationToString(time.Microsecond))
	assert.Equal(t, "ms", DurationToString(time.Millisecond*10))
	assert.Equal(t, "ms", DurationToString(time.Millisecond))
	assert.Equal(t, "ns", DurationToString(1))
}

func ExampleStopWatch_Start() {
	scope := NewTestScope()
	stopWatch, _ := scope.NewStopWatch("test", "This is a test stop watch", time.Millisecond)

	{
		timer := stopWatch.Start()
		defer timer.Stop()

		// Do the operation you want to measure
		time.Sleep(time.Second)
	}
}

func TestNewScope(t *testing.T) {
	assert.Panics(t, func() {
		NewScope("")
	})
	s := NewScope("test")
	assert.Equal(t, "test:", s.CurrentScope())
	assert.Equal(t, "test:hello:", s.NewSubScope("hello").CurrentScope())
	assert.Panics(t, func() {
		s.NewSubScope("")
	})
	assert.Equal(t, "test:timer_x", s.NewScopedMetricName("timer_x"))
	assert.Equal(t, "test:hello:timer:", s.NewSubScope("hello").NewSubScope("timer").CurrentScope())
	assert.Equal(t, "test:hello:timer:", s.NewSubScope("hello").NewSubScope("timer:").CurrentScope())
	assert.Equal(t, "test:k8s_array:test_1:", s.NewSubScope("k8s-array").NewSubScope("test-1:").CurrentScope())
}

func TestMetricsScope(t *testing.T) {
	s := NewScope("test")
	const description = "some x"
	if !assert.NotNil(t, prometheus.DefaultRegisterer) {
		assert.Fail(t, "Prometheus registrar failed")
	}
	t.Run("Counter", func(t *testing.T) {
		m := s.MustNewCounter("xc", description)
		assert.Equal(t, `Desc{fqName: "test:xc", help: "some x", constLabels: {}, variableLabels: []}`, m.Desc().String())
		mv := s.MustNewCounterVec("xcv", description)
		assert.NotNil(t, mv)
		assert.Panics(t, func() {
			_ = s.MustNewCounter("xc", description)
		})
		assert.Panics(t, func() {
			_ = s.MustNewCounterVec("xcv", description)
		})
	})

	t.Run("Histogram", func(t *testing.T) {
		m := s.MustNewHistogram("xh", description)
		assert.Equal(t, `Desc{fqName: "test:xh", help: "some x", constLabels: {}, variableLabels: []}`, m.Desc().String())
		mv := s.MustNewHistogramVec("xhv", description)
		assert.NotNil(t, mv)
		assert.Panics(t, func() {
			_ = s.MustNewHistogram("xh", description)
		})
		assert.Panics(t, func() {
			_ = s.MustNewHistogramVec("xhv", description)
		})
	})

	t.Run("Summary", func(t *testing.T) {
		m := s.MustNewSummary("xs", description)
		assert.Equal(t, `Desc{fqName: "test:xs", help: "some x", constLabels: {}, variableLabels: []}`, m.Desc().String())
		mco, err := s.NewSummaryWithOptions("xsco", description, SummaryOptions{Objectives: map[float64]float64{0.5: 0.05, 1.0: 0.0}})
		assert.Nil(t, err)
		assert.Equal(t, `Desc{fqName: "test:xsco", help: "some x", constLabels: {}, variableLabels: []}`, mco.Desc().String())
		mv := s.MustNewSummaryVec("xsv", description)
		assert.NotNil(t, mv)
		assert.Panics(t, func() {
			_ = s.MustNewSummary("xs", description)
		})
		assert.Panics(t, func() {
			_ = s.MustNewSummaryVec("xsv", description)
		})
	})

	t.Run("Gauge", func(t *testing.T) {
		m := s.MustNewGauge("xg", description)
		assert.Equal(t, `Desc{fqName: "test:xg", help: "some x", constLabels: {}, variableLabels: []}`, m.Desc().String())
		mv := s.MustNewGaugeVec("xgv", description)
		assert.NotNil(t, mv)
		assert.Panics(t, func() {
			m = s.MustNewGauge("xg", description)
		})
		assert.Panics(t, func() {
			_ = s.MustNewGaugeVec("xgv", description)
		})
	})

	t.Run("Timer", func(t *testing.T) {
		m := s.MustNewStopWatch("xt", description, time.Second)
		asDesc, ok := m.Observer.(prometheus.Metric)
		assert.True(t, ok)
		assert.Equal(t, `Desc{fqName: "test:xt_s", help: "some x", constLabels: {}, variableLabels: []}`, asDesc.Desc().String())
		assert.Panics(t, func() {
			_ = s.MustNewStopWatch("xt", description, time.Second)
		})
	})

}

func TestStopWatch_Start(t *testing.T) {
	scope := NewTestScope()
	s, e := scope.NewStopWatch("yt"+rand.String(3), "timer", time.Millisecond)
	assert.NoError(t, e)
	assert.Equal(t, time.Millisecond, s.outputScale)
	i := s.Start()
	assert.Equal(t, time.Millisecond, i.outputScale)
	assert.NotNil(t, i.start)
	i.Stop()
}

func TestStopWatch_Observe(t *testing.T) {
	scope := NewTestScope()
	s, e := scope.NewStopWatch("yt"+rand.String(3), "timer", time.Millisecond)
	assert.NoError(t, e)
	assert.Equal(t, time.Millisecond, s.outputScale)
	s.Observe(time.Now(), time.Now().Add(time.Second))
}

func TestStopWatch_Time(t *testing.T) {
	scope := NewTestScope()
	s, e := scope.NewStopWatch("yt"+rand.String(3), "timer", time.Millisecond)
	assert.NoError(t, e)
	assert.Equal(t, time.Millisecond, s.outputScale)
	s.Time(func() {
	})
}

func TestStopWatchVec_WithLabelValues(t *testing.T) {
	scope := NewTestScope()
	v, e := scope.NewStopWatchVec("yt"+rand.String(3), "timer", time.Millisecond, "workflow", "label")
	assert.NoError(t, e)
	assert.Equal(t, time.Millisecond, v.outputScale)
	s := v.WithLabelValues("my_wf", "something")
	assert.NotNil(t, s)
	i := s.Start()
	assert.Equal(t, time.Millisecond, i.outputScale)
	assert.NotNil(t, i.start)
	i.Stop()
}
