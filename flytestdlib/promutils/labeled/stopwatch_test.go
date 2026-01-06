package labeled

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

func ExampleStopWatch_Start() {
	ctx := context.Background()
	stopWatch := NewStopWatch("test", "this is an example stopwatch", time.Millisecond, promutils.NewTestScope())
	{
		timer := stopWatch.Start(ctx)
		defer timer.Stop()

		// An operation you want to measure the time for.
		time.Sleep(time.Second)
	}
}

func TestLabeledStopWatch(t *testing.T) {
	UnsetMetricKeys()
	assert.NotPanics(t, func() {
		SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
	})

	t.Run("Labeled", func(t *testing.T) {
		scope := promutils.NewScope("testscope_stopwatch")
		s := NewStopWatch("s1", "some desc", time.Minute, scope)
		assert.NotNil(t, s)
		metricName := scope.CurrentScope() + "s1_m"

		ctx := context.TODO()
		const header = `
			# HELP testscope_stopwatch:s1_m some desc
			# TYPE testscope_stopwatch:s1_m summary`

		w := s.Start(ctx)
		w.Stop()
		expectedMetrics := map[string]any{
			`testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.5"}`:  0.0,
			`testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.9"}`:  0.0,
			`testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.99"}`: 0.0,
			`testscope_stopwatch:s1_m_sum{domain="",project="",task="",wf=""}`:             0.0,
			`testscope_stopwatch:s1_m_count{domain="",project="",task="",wf=""}`:           1,
		}
		assertMetrics(t, s.StopWatchVec, metricName, header, expectedMetrics)

		ctx = contextutils.WithProjectDomain(ctx, "project", "domain")
		w = s.Start(ctx)
		w.Stop()
		expectedMetrics = map[string]any{
			`testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.5"}`:               0.0,
			`testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.9"}`:               0.0,
			`testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.99"}`:              0.0,
			`testscope_stopwatch:s1_m_sum{domain="",project="",task="",wf=""}`:                          0.0,
			`testscope_stopwatch:s1_m_count{domain="",project="",task="",wf=""}`:                        1,
			`testscope_stopwatch:s1_m{domain="domain",project="project",task="",wf="",quantile="0.5"}`:  0.0,
			`testscope_stopwatch:s1_m{domain="domain",project="project",task="",wf="",quantile="0.9"}`:  0.0,
			`testscope_stopwatch:s1_m{domain="domain",project="project",task="",wf="",quantile="0.99"}`: 0.0,
			`testscope_stopwatch:s1_m_sum{domain="domain",project="project",task="",wf=""}`:             0.0,
			`testscope_stopwatch:s1_m_count{domain="domain",project="project",task="",wf=""}`:           1,
		}
		assertMetrics(t, s.StopWatchVec, metricName, header, expectedMetrics)

		now := time.Now()
		s.Observe(ctx, now, now.Add(time.Minute))
		expectedMetrics = map[string]any{
			`testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.5"}`:               0.0,
			`testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.9"}`:               0.0,
			`testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.99"}`:              0.0,
			`testscope_stopwatch:s1_m_sum{domain="",project="",task="",wf=""}`:                          0.0,
			`testscope_stopwatch:s1_m_count{domain="",project="",task="",wf=""}`:                        1,
			`testscope_stopwatch:s1_m{domain="domain",project="project",task="",wf="",quantile="0.5"}`:  0.0,
			`testscope_stopwatch:s1_m{domain="domain",project="project",task="",wf="",quantile="0.9"}`:  1,
			`testscope_stopwatch:s1_m{domain="domain",project="project",task="",wf="",quantile="0.99"}`: 1,
			`testscope_stopwatch:s1_m_sum{domain="domain",project="project",task="",wf=""}`:             1.0,
			`testscope_stopwatch:s1_m_count{domain="domain",project="project",task="",wf=""}`:           2,
		}
		assertMetrics(t, s.StopWatchVec, metricName, header, expectedMetrics)

		s.Time(ctx, func() {
			// Do nothing
		})
		expectedMetrics = map[string]any{
			`testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.5"}`:               0.0,
			`testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.9"}`:               0.0,
			`testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.99"}`:              0.0,
			`testscope_stopwatch:s1_m_sum{domain="",project="",task="",wf=""}`:                          0.0,
			`testscope_stopwatch:s1_m_count{domain="",project="",task="",wf=""}`:                        1,
			`testscope_stopwatch:s1_m{domain="domain",project="project",task="",wf="",quantile="0.5"}`:  0.0,
			`testscope_stopwatch:s1_m{domain="domain",project="project",task="",wf="",quantile="0.9"}`:  1,
			`testscope_stopwatch:s1_m{domain="domain",project="project",task="",wf="",quantile="0.99"}`: 1,
			`testscope_stopwatch:s1_m_sum{domain="domain",project="project",task="",wf=""}`:             1.0,
			`testscope_stopwatch:s1_m_count{domain="domain",project="project",task="",wf=""}`:           3,
		}
		assertMetrics(t, s.StopWatchVec, metricName, header, expectedMetrics)
	})

	t.Run("Unlabeled", func(t *testing.T) {
		scope := promutils.NewScope("testscope_stopwatch")
		s := NewStopWatch("s2", "some desc", time.Minute, scope, EmitUnlabeledMetric)
		assert.NotNil(t, s)

		ctx := context.TODO()
		const header = `
			# HELP testscope_stopwatch:s2_unlabeled_m some desc
			# TYPE testscope_stopwatch:s2_unlabeled_m summary
		`

		w := s.Start(ctx)
		w.Stop()
		expectedMetrics := map[string]any{
			`testscope_stopwatch:s2_unlabeled_m{quantile="0.5"}`:  0.0,
			`testscope_stopwatch:s2_unlabeled_m{quantile="0.9"}`:  0.0,
			`testscope_stopwatch:s2_unlabeled_m{quantile="0.99"}`: 0.0,
			`testscope_stopwatch:s2_unlabeled_m_sum`:              0.0,
			`testscope_stopwatch:s2_unlabeled_m_count`:            1,
		}
		assertMetrics(t, s.Observer.(prometheus.Summary), "testscope_stopwatch:s2_unlabeled_m", header, expectedMetrics)
	})

	t.Run("AdditionalLabels", func(t *testing.T) {
		scope := promutils.NewScope("testscope_stopwatch")
		opts := AdditionalLabelsOption{Labels: []string{contextutils.ProjectKey.String(), contextutils.ExecIDKey.String()}}
		s := NewStopWatch("s3", "some desc", time.Minute, scope, opts)
		assert.NotNil(t, s)
		metricName := scope.CurrentScope() + "s3_m"

		ctx := context.TODO()
		const header = `
			# HELP testscope_stopwatch:s3_m some desc
			# TYPE testscope_stopwatch:s3_m summary
		`

		w := s.Start(ctx)
		w.Stop()
		expectedMetrics := map[string]any{
			`testscope_stopwatch:s3_m{domain="",exec_id="",project="",task="",wf="",quantile="0.5"}`:  0.0,
			`testscope_stopwatch:s3_m{domain="",exec_id="",project="",task="",wf="",quantile="0.9"}`:  0.0,
			`testscope_stopwatch:s3_m{domain="",exec_id="",project="",task="",wf="",quantile="0.99"}`: 0.0,
			`testscope_stopwatch:s3_m_sum{domain="",exec_id="",project="",task="",wf=""}`:             0.0,
			`testscope_stopwatch:s3_m_count{domain="",exec_id="",project="",task="",wf=""}`:           1,
		}
		assertMetrics(t, s.StopWatchVec, metricName, header, expectedMetrics)

		ctx = contextutils.WithProjectDomain(ctx, "project", "domain")
		w = s.Start(ctx)
		w.Stop()
		expectedMetrics = map[string]any{
			`testscope_stopwatch:s3_m{domain="",exec_id="",project="",task="",wf="",quantile="0.5"}`:               0.0,
			`testscope_stopwatch:s3_m{domain="",exec_id="",project="",task="",wf="",quantile="0.9"}`:               0.0,
			`testscope_stopwatch:s3_m{domain="",exec_id="",project="",task="",wf="",quantile="0.99"}`:              0.0,
			`testscope_stopwatch:s3_m_sum{domain="",exec_id="",project="",task="",wf=""}`:                          0.0,
			`testscope_stopwatch:s3_m_count{domain="",exec_id="",project="",task="",wf=""}`:                        1,
			`testscope_stopwatch:s3_m{domain="domain",exec_id="",project="project",task="",wf="",quantile="0.5"}`:  0.0,
			`testscope_stopwatch:s3_m{domain="domain",exec_id="",project="project",task="",wf="",quantile="0.9"}`:  0.0,
			`testscope_stopwatch:s3_m{domain="domain",exec_id="",project="project",task="",wf="",quantile="0.99"}`: 0.0,
			`testscope_stopwatch:s3_m_sum{domain="domain",exec_id="",project="project",task="",wf=""}`:             0.0,
			`testscope_stopwatch:s3_m_count{domain="domain",exec_id="",project="project",task="",wf=""}`:           1,
		}
		assertMetrics(t, s.StopWatchVec, metricName, header, expectedMetrics)

		ctx = contextutils.WithExecutionID(ctx, "exec_id")
		w = s.Start(ctx)
		w.Stop()
		expectedMetrics = map[string]any{
			`testscope_stopwatch:s3_m{domain="",exec_id="",project="",task="",wf="",quantile="0.5"}`:                      0.0,
			`testscope_stopwatch:s3_m{domain="",exec_id="",project="",task="",wf="",quantile="0.9"}`:                      0.0,
			`testscope_stopwatch:s3_m{domain="",exec_id="",project="",task="",wf="",quantile="0.99"}`:                     0.0,
			`testscope_stopwatch:s3_m_sum{domain="",exec_id="",project="",task="",wf=""}`:                                 0.0,
			`testscope_stopwatch:s3_m_count{domain="",exec_id="",project="",task="",wf=""}`:                               1,
			`testscope_stopwatch:s3_m{domain="domain",exec_id="",project="project",task="",wf="",quantile="0.5"}`:         0.0,
			`testscope_stopwatch:s3_m{domain="domain",exec_id="",project="project",task="",wf="",quantile="0.9"}`:         0.0,
			`testscope_stopwatch:s3_m{domain="domain",exec_id="",project="project",task="",wf="",quantile="0.99"}`:        0.0,
			`testscope_stopwatch:s3_m_sum{domain="domain",exec_id="",project="project",task="",wf=""}`:                    0.0,
			`testscope_stopwatch:s3_m_count{domain="domain",exec_id="",project="project",task="",wf=""}`:                  1,
			`testscope_stopwatch:s3_m{domain="domain",exec_id="exec_id",project="project",task="",wf="",quantile="0.5"}`:  0.0,
			`testscope_stopwatch:s3_m{domain="domain",exec_id="exec_id",project="project",task="",wf="",quantile="0.9"}`:  0.0,
			`testscope_stopwatch:s3_m{domain="domain",exec_id="exec_id",project="project",task="",wf="",quantile="0.99"}`: 0.0,
			`testscope_stopwatch:s3_m_sum{domain="domain",exec_id="exec_id",project="project",task="",wf=""}`:             0.0,
			`testscope_stopwatch:s3_m_count{domain="domain",exec_id="exec_id",project="project",task="",wf=""}`:           1,
		}
		assertMetrics(t, s.StopWatchVec, metricName, header, expectedMetrics)
	})
}
