package labeled

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

func ExampleHistogramStopWatch_Start() {
	ctx := context.Background()
	stopWatch := NewHistogramStopWatch("test", "this is an example histogram stopwatch", promutils.NewTestScope())
	{
		timer := stopWatch.Start(ctx)
		defer timer.Stop()

		// An operation you want to measure the time for.
		time.Sleep(time.Second)
	}
}

func TestLabeledHistogramStopWatch(t *testing.T) {
	UnsetMetricKeys()
	assert.NotPanics(t, func() {
		SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
	})

	t.Run("Labeled", func(t *testing.T) {
		scope := promutils.NewScope("testscope_hist_stopwatch")
		s := NewHistogramStopWatch("s1", "some desc", scope)
		assert.NotNil(t, s)
		metricName := scope.CurrentScope() + "s1"

		ctx := context.TODO()
		const header = `
			# HELP testscope_hist_stopwatch:s1 some desc
			# TYPE testscope_hist_stopwatch:s1 histogram`

		w := s.Start(ctx)
		w.Stop()
		expectedMetrics := map[string]any{
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.005"}`: 1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.01"}`:  1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.025"}`: 1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.05"}`:  1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.1"}`:   1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.25"}`:  1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.5"}`:   1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="1"}`:     1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="2.5"}`:   1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="5"}`:     1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="10"}`:    1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="+Inf"}`:  1,
			`testscope_hist_stopwatch:s1_sum{domain="",project="",task="",wf=""}`:               0.0,
			`testscope_hist_stopwatch:s1_count{domain="",project="",task="",wf=""}`:             1,
		}
		assertMetrics(t, s.HistogramStopWatchVec, metricName, header, expectedMetrics)

		ctx = contextutils.WithProjectDomain(ctx, "project", "domain")
		w = s.Start(ctx)
		w.Stop()

		expectedMetrics = map[string]any{
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.005"}`:              1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.01"}`:               1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.025"}`:              1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.05"}`:               1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.1"}`:                1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.25"}`:               1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.5"}`:                1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="1"}`:                  1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="2.5"}`:                1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="5"}`:                  1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="10"}`:                 1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="+Inf"}`:               1,
			`testscope_hist_stopwatch:s1_sum{domain="",project="",task="",wf=""}`:                            0.0,
			`testscope_hist_stopwatch:s1_count{domain="",project="",task="",wf=""}`:                          1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.005"}`: 1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.01"}`:  1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.025"}`: 1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.05"}`:  1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.1"}`:   1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.25"}`:  1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.5"}`:   1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="1"}`:     1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="2.5"}`:   1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="5"}`:     1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="10"}`:    1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="+Inf"}`:  1,
			`testscope_hist_stopwatch:s1_sum{domain="domain",project="project",task="",wf=""}`:               0.0,
			`testscope_hist_stopwatch:s1_count{domain="domain",project="project",task="",wf=""}`:             1,
		}
		assertMetrics(t, s.HistogramStopWatchVec, metricName, header, expectedMetrics)

		now := time.Now()
		s.Observe(ctx, now, now.Add(time.Minute))

		expectedMetrics = map[string]any{
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.005"}`:              1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.01"}`:               1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.025"}`:              1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.05"}`:               1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.1"}`:                1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.25"}`:               1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.5"}`:                1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="1"}`:                  1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="2.5"}`:                1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="5"}`:                  1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="10"}`:                 1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="+Inf"}`:               1,
			`testscope_hist_stopwatch:s1_sum{domain="",project="",task="",wf=""}`:                            0.0,
			`testscope_hist_stopwatch:s1_count{domain="",project="",task="",wf=""}`:                          1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.005"}`: 1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.01"}`:  1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.025"}`: 1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.05"}`:  1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.1"}`:   1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.25"}`:  1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.5"}`:   1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="1"}`:     1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="2.5"}`:   1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="5"}`:     1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="10"}`:    1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="+Inf"}`:  2,
			`testscope_hist_stopwatch:s1_sum{domain="domain",project="project",task="",wf=""}`:               60.0,
			`testscope_hist_stopwatch:s1_count{domain="domain",project="project",task="",wf=""}`:             2,
		}
		assertMetrics(t, s.HistogramStopWatchVec, metricName, header, expectedMetrics)

		s.Time(ctx, func() {
			// Do nothing
		})

		expectedMetrics = map[string]any{
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.005"}`:              1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.01"}`:               1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.025"}`:              1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.05"}`:               1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.1"}`:                1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.25"}`:               1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.5"}`:                1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="1"}`:                  1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="2.5"}`:                1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="5"}`:                  1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="10"}`:                 1,
			`testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="+Inf"}`:               1,
			`testscope_hist_stopwatch:s1_sum{domain="",project="",task="",wf=""}`:                            0.0,
			`testscope_hist_stopwatch:s1_count{domain="",project="",task="",wf=""}`:                          1,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.005"}`: 2,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.01"}`:  2,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.025"}`: 2,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.05"}`:  2,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.1"}`:   2,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.25"}`:  2,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.5"}`:   2,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="1"}`:     2,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="2.5"}`:   2,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="5"}`:     2,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="10"}`:    2,
			`testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="+Inf"}`:  3,
			`testscope_hist_stopwatch:s1_sum{domain="domain",project="project",task="",wf=""}`:               60.0,
			`testscope_hist_stopwatch:s1_count{domain="domain",project="project",task="",wf=""}`:             3,
		}
		assertMetrics(t, s.HistogramStopWatchVec, metricName, header, expectedMetrics)
	})

	t.Run("Unlabeled", func(t *testing.T) {
		scope := promutils.NewScope("testscope_hist_stopwatch")
		s := NewHistogramStopWatch("s2", "some desc", scope, EmitUnlabeledMetric)
		assert.NotNil(t, s)

		ctx := context.TODO()
		const header = `
			# HELP testscope_hist_stopwatch:s2_unlabeled some desc
			# TYPE testscope_hist_stopwatch:s2_unlabeled histogram`

		w := s.Start(ctx)
		w.Stop()
		expectedMetrics := map[string]any{
			`testscope_hist_stopwatch:s2_unlabeled_bucket{le="0.005"}`: 1,
			`testscope_hist_stopwatch:s2_unlabeled_bucket{le="0.01"}`:  1,
			`testscope_hist_stopwatch:s2_unlabeled_bucket{le="0.025"}`: 1,
			`testscope_hist_stopwatch:s2_unlabeled_bucket{le="0.05"}`:  1,
			`testscope_hist_stopwatch:s2_unlabeled_bucket{le="0.1"}`:   1,
			`testscope_hist_stopwatch:s2_unlabeled_bucket{le="0.25"}`:  1,
			`testscope_hist_stopwatch:s2_unlabeled_bucket{le="0.5"}`:   1,
			`testscope_hist_stopwatch:s2_unlabeled_bucket{le="1"}`:     1,
			`testscope_hist_stopwatch:s2_unlabeled_bucket{le="2.5"}`:   1,
			`testscope_hist_stopwatch:s2_unlabeled_bucket{le="5"}`:     1,
			`testscope_hist_stopwatch:s2_unlabeled_bucket{le="10"}`:    1,
			`testscope_hist_stopwatch:s2_unlabeled_bucket{le="+Inf"}`:  1,
			`testscope_hist_stopwatch:s2_unlabeled_sum`:                0.0,
			`testscope_hist_stopwatch:s2_unlabeled_count`:              1,
		}
		assertMetrics(t, s.HistogramStopWatch.Observer.(prometheus.Histogram), "testscope_hist_stopwatch:s2_unlabeled", header, expectedMetrics)
	})

	t.Run("AdditionalLabels", func(t *testing.T) {
		scope := promutils.NewScope("testscope_hist_stopwatch")
		opts := AdditionalLabelsOption{Labels: []string{contextutils.ProjectKey.String(), contextutils.ExecIDKey.String()}}
		s := NewHistogramStopWatch("s3", "some desc", scope, opts)
		assert.NotNil(t, s)
		metricName := scope.CurrentScope() + "s3"

		ctx := context.TODO()
		const header = `
			# HELP testscope_hist_stopwatch:s3 some desc
			# TYPE testscope_hist_stopwatch:s3 histogram`

		w := s.Start(ctx)
		w.Stop()
		expectedMetrics := map[string]any{
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.005"}`: 1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.01"}`:  1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.025"}`: 1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.05"}`:  1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.1"}`:   1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.25"}`:  1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.5"}`:   1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="1"}`:     1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="2.5"}`:   1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="5"}`:     1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="10"}`:    1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="+Inf"}`:  1,
			`testscope_hist_stopwatch:s3_sum{domain="",exec_id="",project="",task="",wf=""}`:               0.0,
			`testscope_hist_stopwatch:s3_count{domain="",exec_id="",project="",task="",wf=""}`:             1,
		}
		assertMetrics(t, s.HistogramStopWatchVec, metricName, header, expectedMetrics)

		ctx = contextutils.WithProjectDomain(ctx, "project", "domain")
		w = s.Start(ctx)
		w.Stop()
		expectedMetrics = map[string]any{
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.005"}`:              1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.01"}`:               1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.025"}`:              1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.05"}`:               1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.1"}`:                1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.25"}`:               1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.5"}`:                1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="1"}`:                  1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="2.5"}`:                1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="5"}`:                  1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="10"}`:                 1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="+Inf"}`:               1,
			`testscope_hist_stopwatch:s3_sum{domain="",exec_id="",project="",task="",wf=""}`:                            0.0,
			`testscope_hist_stopwatch:s3_count{domain="",exec_id="",project="",task="",wf=""}`:                          1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.005"}`: 1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.01"}`:  1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.025"}`: 1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.05"}`:  1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.1"}`:   1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.25"}`:  1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.5"}`:   1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="1"}`:     1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="2.5"}`:   1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="5"}`:     1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="10"}`:    1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="+Inf"}`:  1,
			`testscope_hist_stopwatch:s3_sum{domain="domain",exec_id="",project="project",task="",wf=""}`:               0.0,
			`testscope_hist_stopwatch:s3_count{domain="domain",exec_id="",project="project",task="",wf=""}`:             1,
		}
		assertMetrics(t, s.HistogramStopWatchVec, metricName, header, expectedMetrics)

		ctx = contextutils.WithExecutionID(ctx, "exec_id")
		w = s.Start(ctx)
		w.Stop()
		expectedMetrics = map[string]any{
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.005"}`:                     1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.01"}`:                      1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.025"}`:                     1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.05"}`:                      1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.1"}`:                       1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.25"}`:                      1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.5"}`:                       1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="1"}`:                         1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="2.5"}`:                       1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="5"}`:                         1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="10"}`:                        1,
			`testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="+Inf"}`:                      1,
			`testscope_hist_stopwatch:s3_sum{domain="",exec_id="",project="",task="",wf=""}`:                                   0.0,
			`testscope_hist_stopwatch:s3_count{domain="",exec_id="",project="",task="",wf=""}`:                                 1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.005"}`:        1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.01"}`:         1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.025"}`:        1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.05"}`:         1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.1"}`:          1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.25"}`:         1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.5"}`:          1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="1"}`:            1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="2.5"}`:          1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="5"}`:            1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="10"}`:           1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="+Inf"}`:         1,
			`testscope_hist_stopwatch:s3_sum{domain="domain",exec_id="",project="project",task="",wf=""}`:                      0.0,
			`testscope_hist_stopwatch:s3_count{domain="domain",exec_id="",project="project",task="",wf=""}`:                    1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="0.005"}`: 1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="0.01"}`:  1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="0.025"}`: 1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="0.05"}`:  1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="0.1"}`:   1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="0.25"}`:  1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="0.5"}`:   1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="1"}`:     1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="2.5"}`:   1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="5"}`:     1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="10"}`:    1,
			`testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="+Inf"}`:  1,
			`testscope_hist_stopwatch:s3_sum{domain="domain",exec_id="exec_id",project="project",task="",wf=""}`:               0.0,
			`testscope_hist_stopwatch:s3_count{domain="domain",exec_id="exec_id",project="project",task="",wf=""}`:             1,
		}
		assertMetrics(t, s.HistogramStopWatchVec, metricName, header, expectedMetrics)
	})
}

func assertMetrics(t *testing.T, c prometheus.Collector, metricName, expectedHeader string, expectedMetrics map[string]any) {
	t.Helper()
	metricBytes, err := testutil.CollectAndFormat(c, expfmt.TypeTextPlain, metricName)
	require.NoError(t, err)
	require.NotEmptyf(t, metricBytes, "empty `%q` metric", metricName)

	actual := strings.Split(strings.TrimSpace(string(metricBytes)), "\n")
	n := len(actual)

	expected := strings.Split(strings.TrimSpace(expectedHeader), "\n")
	require.Len(t, expected, 2, "wrong number of expected header lines")

	for i := 0; i < n; i++ {
		line := actual[i]

		if strings.HasPrefix(line, "#") {
			if i != 0 && i != 1 {
				require.Failf(t, "wrong format", "comment line %q on wrong place", line)
			}
			assert.Equal(t, strings.TrimSpace(expected[i]), actual[i])
			continue
		}

		lineSplt := strings.Split(line, " ")
		if len(lineSplt) != 2 {
			require.Failf(t, "metric line has wrong format", "metric %s has line %q with wrong format", metricName, line)
		}

		key := lineSplt[0]
		expectedValue, ok := expectedMetrics[key]
		require.Truef(t, ok, "missing expected %q metric", key)

		switch expectedValue.(type) {
		case int, int8, int16, int32, int64:
			actualValue, err := strconv.Atoi(lineSplt[1])
			require.NoError(t, err)
			assert.Equal(t, expectedValue, actualValue)
		case float32, float64:
			actualValue, err := strconv.ParseFloat(lineSplt[1], 64)
			require.NoError(t, err)
			assert.InDeltaf(t, expectedValue, actualValue, 0.001, "metric %q has wrong value", key)
			assert.Greaterf(t, actualValue, expectedValue, "actual value of %q should be slightly greater than expected", key)
		default:
			require.Fail(t, "unsupported expected value type")
		}
	}
}
