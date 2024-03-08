package labeled

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
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

		ctx := context.TODO()
		const header = `
			# HELP testscope_hist_stopwatch:s1 some desc
			# TYPE testscope_hist_stopwatch:s1 histogram
		`

		w := s.Start(ctx)
		w.Stop()
		var expected = `
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.005"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.01"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.025"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.05"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.1"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.25"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.5"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="1"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="2.5"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="5"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="10"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="+Inf"} 1
			testscope_hist_stopwatch:s1_sum{domain="",project="",task="",wf=""} 0
			testscope_hist_stopwatch:s1_count{domain="",project="",task="",wf=""} 1
		`
		err := testutil.CollectAndCompare(s.HistogramStopWatchVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		ctx = contextutils.WithProjectDomain(ctx, "project", "domain")
		w = s.Start(ctx)
		w.Stop()
		expected = `
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.005"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.01"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.025"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.05"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.1"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.25"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.5"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="1"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="2.5"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="5"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="10"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="+Inf"} 1
			testscope_hist_stopwatch:s1_sum{domain="",project="",task="",wf=""} 0
			testscope_hist_stopwatch:s1_count{domain="",project="",task="",wf=""} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.005"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.01"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.025"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.05"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.1"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.25"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.5"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="1"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="2.5"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="5"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="10"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="+Inf"} 1
			testscope_hist_stopwatch:s1_sum{domain="domain",project="project",task="",wf=""} 0
			testscope_hist_stopwatch:s1_count{domain="domain",project="project",task="",wf=""} 1
		`
		err = testutil.CollectAndCompare(s.HistogramStopWatchVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		now := time.Now()
		s.Observe(ctx, now, now.Add(time.Minute))
		expected = `
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.005"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.01"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.025"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.05"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.1"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.25"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.5"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="1"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="2.5"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="5"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="10"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="+Inf"} 1
			testscope_hist_stopwatch:s1_sum{domain="",project="",task="",wf=""} 0
			testscope_hist_stopwatch:s1_count{domain="",project="",task="",wf=""} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.005"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.01"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.025"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.05"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.1"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.25"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.5"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="1"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="2.5"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="5"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="10"} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="+Inf"} 2
			testscope_hist_stopwatch:s1_sum{domain="domain",project="project",task="",wf=""} 60
			testscope_hist_stopwatch:s1_count{domain="domain",project="project",task="",wf=""} 2
		`
		err = testutil.CollectAndCompare(s.HistogramStopWatchVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		s.Time(ctx, func() {
			// Do nothing
		})
		expected = `
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.005"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.01"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.025"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.05"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.1"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.25"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="0.5"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="1"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="2.5"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="5"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="10"} 1
			testscope_hist_stopwatch:s1_bucket{domain="",project="",task="",wf="",le="+Inf"} 1
			testscope_hist_stopwatch:s1_sum{domain="",project="",task="",wf=""} 0
			testscope_hist_stopwatch:s1_count{domain="",project="",task="",wf=""} 1
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.005"} 2
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.01"} 2
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.025"} 2
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.05"} 2
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.1"} 2
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.25"} 2
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="0.5"} 2
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="1"} 2
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="2.5"} 2
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="5"} 2
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="10"} 2
			testscope_hist_stopwatch:s1_bucket{domain="domain",project="project",task="",wf="",le="+Inf"} 3
			testscope_hist_stopwatch:s1_sum{domain="domain",project="project",task="",wf=""} 60
			testscope_hist_stopwatch:s1_count{domain="domain",project="project",task="",wf=""} 3
		`
		err = testutil.CollectAndCompare(s.HistogramStopWatchVec, strings.NewReader(header+expected))
		assert.NoError(t, err)
	})

	t.Run("Unlabeled", func(t *testing.T) {
		scope := promutils.NewScope("testscope_hist_stopwatch")
		s := NewHistogramStopWatch("s2", "some desc", scope, EmitUnlabeledMetric)
		assert.NotNil(t, s)

		ctx := context.TODO()
		const header = `
			# HELP testscope_hist_stopwatch:s2_unlabeled some desc
			# TYPE testscope_hist_stopwatch:s2_unlabeled histogram
		`

		w := s.Start(ctx)
		w.Stop()
		var expected = `
			testscope_hist_stopwatch:s2_unlabeled_bucket{le="0.005"} 1
			testscope_hist_stopwatch:s2_unlabeled_bucket{le="0.01"} 1
			testscope_hist_stopwatch:s2_unlabeled_bucket{le="0.025"} 1
			testscope_hist_stopwatch:s2_unlabeled_bucket{le="0.05"} 1
			testscope_hist_stopwatch:s2_unlabeled_bucket{le="0.1"} 1
			testscope_hist_stopwatch:s2_unlabeled_bucket{le="0.25"} 1
			testscope_hist_stopwatch:s2_unlabeled_bucket{le="0.5"} 1
			testscope_hist_stopwatch:s2_unlabeled_bucket{le="1"} 1
			testscope_hist_stopwatch:s2_unlabeled_bucket{le="2.5"} 1
			testscope_hist_stopwatch:s2_unlabeled_bucket{le="5"} 1
			testscope_hist_stopwatch:s2_unlabeled_bucket{le="10"} 1
			testscope_hist_stopwatch:s2_unlabeled_bucket{le="+Inf"} 1
			testscope_hist_stopwatch:s2_unlabeled_sum 0
			testscope_hist_stopwatch:s2_unlabeled_count 1
		`
		err := testutil.CollectAndCompare(s.HistogramStopWatch.Observer.(prometheus.Histogram), strings.NewReader(header+expected))
		assert.NoError(t, err)
	})

	t.Run("AdditionalLabels", func(t *testing.T) {
		scope := promutils.NewScope("testscope_hist_stopwatch")
		opts := AdditionalLabelsOption{Labels: []string{contextutils.ProjectKey.String(), contextutils.ExecIDKey.String()}}
		s := NewHistogramStopWatch("s3", "some desc", scope, opts)
		assert.NotNil(t, s)

		ctx := context.TODO()
		const header = `
			# HELP testscope_hist_stopwatch:s3 some desc
			# TYPE testscope_hist_stopwatch:s3 histogram
		`

		w := s.Start(ctx)
		w.Stop()
		var expected = `
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.005"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.01"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.025"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.05"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.1"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.25"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.5"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="1"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="2.5"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="5"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="10"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="+Inf"} 1
			testscope_hist_stopwatch:s3_sum{domain="",exec_id="",project="",task="",wf=""} 0
			testscope_hist_stopwatch:s3_count{domain="",exec_id="",project="",task="",wf=""} 1
		`
		err := testutil.CollectAndCompare(s.HistogramStopWatchVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		ctx = contextutils.WithProjectDomain(ctx, "project", "domain")
		w = s.Start(ctx)
		w.Stop()
		expected = `
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.005"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.01"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.025"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.05"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.1"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.25"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.5"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="1"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="2.5"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="5"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="10"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="+Inf"} 1
			testscope_hist_stopwatch:s3_sum{domain="",exec_id="",project="",task="",wf=""} 0
			testscope_hist_stopwatch:s3_count{domain="",exec_id="",project="",task="",wf=""} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.005"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.01"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.025"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.05"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.1"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.25"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.5"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="1"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="2.5"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="5"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="10"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="+Inf"} 1
			testscope_hist_stopwatch:s3_sum{domain="domain",exec_id="",project="project",task="",wf=""} 0
			testscope_hist_stopwatch:s3_count{domain="domain",exec_id="",project="project",task="",wf=""} 1
		`
		err = testutil.CollectAndCompare(s.HistogramStopWatchVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		ctx = contextutils.WithExecutionID(ctx, "exec_id")
		w = s.Start(ctx)
		w.Stop()
		expected = `
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.005"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.01"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.025"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.05"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.1"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.25"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="0.5"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="1"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="2.5"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="5"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="10"} 1
			testscope_hist_stopwatch:s3_bucket{domain="",exec_id="",project="",task="",wf="",le="+Inf"} 1
			testscope_hist_stopwatch:s3_sum{domain="",exec_id="",project="",task="",wf=""} 0
			testscope_hist_stopwatch:s3_count{domain="",exec_id="",project="",task="",wf=""} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.005"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.01"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.025"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.05"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.1"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.25"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="0.5"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="1"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="2.5"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="5"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="10"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="",project="project",task="",wf="",le="+Inf"} 1
			testscope_hist_stopwatch:s3_sum{domain="domain",exec_id="",project="project",task="",wf=""} 0
			testscope_hist_stopwatch:s3_count{domain="domain",exec_id="",project="project",task="",wf=""} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="0.005"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="0.01"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="0.025"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="0.05"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="0.1"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="0.25"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="0.5"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="1"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="2.5"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="5"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="10"} 1
			testscope_hist_stopwatch:s3_bucket{domain="domain",exec_id="exec_id",project="project",task="",wf="",le="+Inf"} 1
			testscope_hist_stopwatch:s3_sum{domain="domain",exec_id="exec_id",project="project",task="",wf=""} 0
			testscope_hist_stopwatch:s3_count{domain="domain",exec_id="exec_id",project="project",task="",wf=""} 1
		`
		err = testutil.CollectAndCompare(s.HistogramStopWatchVec, strings.NewReader(header+expected))
		assert.NoError(t, err)
	})
}
