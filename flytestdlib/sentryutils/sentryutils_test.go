package sentryutils

import (
	"context"
	"errors"
	"math"
	"slices"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureMetrics returns a context whose Sentry hub records every metric
// emitted through it, before batching or sending, and a function that returns
// the metrics recorded so far. The recording is locked, so metrics can be
// emitted on another goroutine.
func captureMetrics(t *testing.T) (context.Context, func() []*sentry.Metric) {
	t.Helper()
	var (
		mu      sync.Mutex
		metrics []*sentry.Metric
	)
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:       "https://key@example.invalid/1",
		Transport: &sentry.MockTransport{},
		BeforeSendMetric: func(m *sentry.Metric) *sentry.Metric {
			mu.Lock()
			defer mu.Unlock()
			metrics = append(metrics, m)
			return nil
		},
	})
	require.NoError(t, err)
	hub := sentry.NewHub(client, sentry.NewScope())
	recorded := func() []*sentry.Metric {
		mu.Lock()
		defer mu.Unlock()
		return slices.Clone(metrics)
	}
	return sentry.SetHubOnContext(context.Background(), hub), recorded
}

func TestGauge(t *testing.T) {
	t.Run("emits a gauge under its own name", func(t *testing.T) {
		ctx, metrics := captureMetrics(t)

		Gauge(ctx, "flyte.active_runs", 3)

		require.Len(t, metrics(), 1)
		m := metrics()[0]
		assert.Equal(t, sentry.MetricTypeGauge, m.Type)
		assert.Equal(t, "flyte.active_runs", m.Name)
		value, ok := m.Value.Float64()
		require.True(t, ok)
		assert.Equal(t, 3.0, value)
		assert.NotContains(t, m.Attributes, "operation")
		assert.NotContains(t, m.Attributes, "status")
	})

	t.Run("keeps a fractional value", func(t *testing.T) {
		ctx, metrics := captureMetrics(t)

		Gauge(ctx, "flyte.active_runs", 2.5)

		require.Len(t, metrics(), 1)
		value, ok := metrics()[0].Value.Float64()
		require.True(t, ok)
		assert.Equal(t, 2.5, value)
	})

	for _, tc := range []struct {
		desc  string
		name  string
		value float64
	}{
		{"empty name", "", 3},
		{"NaN", "flyte.active_runs", math.NaN()},
		{"positive infinity", "flyte.active_runs", math.Inf(1)},
		{"negative infinity", "flyte.active_runs", math.Inf(-1)},
	} {
		t.Run("drops "+tc.desc, func(t *testing.T) {
			ctx, metrics := captureMetrics(t)

			Gauge(ctx, tc.name, tc.value)

			assert.Empty(t, metrics())
		})
	}

	t.Run("no-op without a Sentry client", func(t *testing.T) {
		ctx := sentry.SetHubOnContext(context.Background(), sentry.NewHub(nil, sentry.NewScope()))
		assert.NotPanics(t, func() { Gauge(ctx, "flyte.active_runs", 3) })
	})
}

const testInterval = time.Minute

// startSampling runs SampleGauges in the background and returns a function
// that stops it and reports what it returned. It must be called inside a
// synctest bubble, so the bubble's fake clock drives the ticker.
func startSampling(ctx context.Context, sample func(context.Context) (map[string]float64, error)) (stop func() error) {
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan error)
	go func() { done <- SampleGauges(ctx, testInterval, sample) }()
	return func() error {
		cancel()
		return <-done
	}
}

// gaugeValues returns the name and value of each recorded gauge.
func gaugeValues(t *testing.T, metrics []*sentry.Metric) map[string]float64 {
	t.Helper()
	values := make(map[string]float64, len(metrics))
	for _, m := range metrics {
		require.Equal(t, sentry.MetricTypeGauge, m.Type)
		value, ok := m.Value.Float64()
		require.True(t, ok)
		values[m.Name] = value
	}
	return values
}

func TestSampleGauges(t *testing.T) {
	// captureMetrics is called outside each bubble: the Sentry client starts
	// background goroutines that never exit, which synctest.Test rejects.

	t.Run("emits every value on each tick", func(t *testing.T) {
		ctx, metrics := captureMetrics(t)
		synctest.Test(t, func(t *testing.T) {
			stop := startSampling(ctx, func(context.Context) (map[string]float64, error) {
				return map[string]float64{"test.a": 1, "test.b": 2}, nil
			})

			synctest.Sleep(testInterval)
			assert.Equal(t, map[string]float64{"test.a": 1, "test.b": 2}, gaugeValues(t, metrics()))

			synctest.Sleep(testInterval)
			assert.Len(t, metrics(), 4)

			assert.NoError(t, stop())
		})
	})

	t.Run("emits an explicit zero", func(t *testing.T) {
		ctx, metrics := captureMetrics(t)
		synctest.Test(t, func(t *testing.T) {
			stop := startSampling(ctx, func(context.Context) (map[string]float64, error) {
				return map[string]float64{"test.a": 0}, nil
			})

			synctest.Sleep(testInterval)
			assert.Equal(t, map[string]float64{"test.a": 0}, gaugeValues(t, metrics()))

			assert.NoError(t, stop())
		})
	})

	t.Run("gives each sample a deadline of one interval", func(t *testing.T) {
		ctx, _ := captureMetrics(t)
		synctest.Test(t, func(t *testing.T) {
			// Recorded here and asserted below: require can't stop the test
			// from the SampleGauges goroutine.
			var (
				hasDeadline bool
				timeLeft    time.Duration
			)
			stop := startSampling(ctx, func(ctx context.Context) (map[string]float64, error) {
				var deadline time.Time
				deadline, hasDeadline = ctx.Deadline()
				timeLeft = time.Until(deadline)
				return nil, nil
			})

			synctest.Sleep(testInterval)
			assert.True(t, hasDeadline)
			assert.Equal(t, testInterval, timeLeft)

			assert.NoError(t, stop())
		})
	})

	t.Run("emits nothing before the first tick", func(t *testing.T) {
		ctx, metrics := captureMetrics(t)
		synctest.Test(t, func(t *testing.T) {
			calls := 0
			stop := startSampling(ctx, func(context.Context) (map[string]float64, error) {
				calls++
				return map[string]float64{"test.a": 1}, nil
			})

			synctest.Sleep(testInterval - time.Second)
			assert.Zero(t, calls)
			assert.Empty(t, metrics())

			assert.NoError(t, stop())
		})
	})

	t.Run("emits nothing for a failed sample and keeps going", func(t *testing.T) {
		ctx, metrics := captureMetrics(t)
		synctest.Test(t, func(t *testing.T) {
			calls := 0
			stop := startSampling(ctx, func(context.Context) (map[string]float64, error) {
				calls++
				if calls == 1 {
					return map[string]float64{"test.a": 1}, errors.New("db down")
				}
				return map[string]float64{"test.a": 2}, nil
			})

			synctest.Sleep(testInterval)
			assert.Empty(t, metrics(), "a failed sample must not emit, even values it returned")

			synctest.Sleep(testInterval)
			assert.Equal(t, map[string]float64{"test.a": 2}, gaugeValues(t, metrics()))

			assert.NoError(t, stop())
		})
	})

	t.Run("cancels a hung sample at its deadline and recovers", func(t *testing.T) {
		ctx, metrics := captureMetrics(t)
		synctest.Test(t, func(t *testing.T) {
			calls := 0
			stop := startSampling(ctx, func(ctx context.Context) (map[string]float64, error) {
				calls++
				if calls == 1 {
					<-ctx.Done() // hang until the deadline
					return nil, ctx.Err()
				}
				return map[string]float64{"test.a": 1}, nil
			})

			// The hung sample starts at 1m and is cancelled at 2m, when the
			// next tick is already waiting.
			synctest.Sleep(2 * testInterval)
			assert.Equal(t, 2, calls)
			assert.Equal(t, map[string]float64{"test.a": 1}, gaugeValues(t, metrics()))

			assert.NoError(t, stop())
		})
	})

	t.Run("returns nil when ctx is cancelled before any tick", func(t *testing.T) {
		ctx, metrics := captureMetrics(t)
		synctest.Test(t, func(t *testing.T) {
			calls := 0
			stop := startSampling(ctx, func(context.Context) (map[string]float64, error) {
				calls++
				return map[string]float64{"test.a": 1}, nil
			})

			assert.NoError(t, stop())
			assert.Zero(t, calls)
			assert.Empty(t, metrics())
		})
	})
}
