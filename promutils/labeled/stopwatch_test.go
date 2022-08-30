package labeled

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"
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

		ctx := context.TODO()
		const header = `
			# HELP testscope_stopwatch:s1_m some desc
			# TYPE testscope_stopwatch:s1_m summary
		`

		w := s.Start(ctx)
		w.Stop()
		var expected = `
			testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.5"} 0
            testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.9"} 0
            testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.99"} 0
            testscope_stopwatch:s1_m_sum{domain="",project="",task="",wf=""} 0
            testscope_stopwatch:s1_m_count{domain="",project="",task="",wf=""} 1
		`
		err := testutil.CollectAndCompare(s.StopWatchVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		ctx = contextutils.WithProjectDomain(ctx, "project", "domain")
		w = s.Start(ctx)
		w.Stop()
		expected = `
			testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.5"} 0
            testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.9"} 0
            testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.99"} 0
            testscope_stopwatch:s1_m_sum{domain="",project="",task="",wf=""} 0
            testscope_stopwatch:s1_m_count{domain="",project="",task="",wf=""} 1
			testscope_stopwatch:s1_m{domain="domain",project="project",task="",wf="",quantile="0.5"} 0
            testscope_stopwatch:s1_m{domain="domain",project="project",task="",wf="",quantile="0.9"} 0
            testscope_stopwatch:s1_m{domain="domain",project="project",task="",wf="",quantile="0.99"} 0
            testscope_stopwatch:s1_m_sum{domain="domain",project="project",task="",wf=""} 0
            testscope_stopwatch:s1_m_count{domain="domain",project="project",task="",wf=""} 1
		`
		err = testutil.CollectAndCompare(s.StopWatchVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		now := time.Now()
		s.Observe(ctx, now, now.Add(time.Minute))
		expected = `
			testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.5"} 0
            testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.9"} 0
            testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.99"} 0
            testscope_stopwatch:s1_m_sum{domain="",project="",task="",wf=""} 0
            testscope_stopwatch:s1_m_count{domain="",project="",task="",wf=""} 1
			testscope_stopwatch:s1_m{domain="domain",project="project",task="",wf="",quantile="0.5"} 0
            testscope_stopwatch:s1_m{domain="domain",project="project",task="",wf="",quantile="0.9"} 1
            testscope_stopwatch:s1_m{domain="domain",project="project",task="",wf="",quantile="0.99"} 1
            testscope_stopwatch:s1_m_sum{domain="domain",project="project",task="",wf=""} 1
            testscope_stopwatch:s1_m_count{domain="domain",project="project",task="",wf=""} 2
		`
		err = testutil.CollectAndCompare(s.StopWatchVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		s.Time(ctx, func() {
			// Do nothing
		})
		expected = `
			testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.5"} 0
            testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.9"} 0
            testscope_stopwatch:s1_m{domain="",project="",task="",wf="",quantile="0.99"} 0
            testscope_stopwatch:s1_m_sum{domain="",project="",task="",wf=""} 0
            testscope_stopwatch:s1_m_count{domain="",project="",task="",wf=""} 1
			testscope_stopwatch:s1_m{domain="domain",project="project",task="",wf="",quantile="0.5"} 0
            testscope_stopwatch:s1_m{domain="domain",project="project",task="",wf="",quantile="0.9"} 1
            testscope_stopwatch:s1_m{domain="domain",project="project",task="",wf="",quantile="0.99"} 1
            testscope_stopwatch:s1_m_sum{domain="domain",project="project",task="",wf=""} 1
            testscope_stopwatch:s1_m_count{domain="domain",project="project",task="",wf=""} 3
		`
		err = testutil.CollectAndCompare(s.StopWatchVec, strings.NewReader(header+expected))
		assert.NoError(t, err)
	})

	t.Run("Unlabeled", func(t *testing.T) {
		scope := promutils.NewScope("testscope_stopwatch")
		s := NewStopWatch("s2", "some desc", time.Minute, scope, EmitUnlabeledMetric)
		assert.NotNil(t, s)

		ctx := context.TODO()
		/*const header = `
			# HELP testscope_stopwatch:s2_m some desc
			# TYPE testscope_stopwatch:s2_m summary
		`*/

		w := s.Start(ctx)
		w.Stop()
		// promutils.StopWatch does not implement prometheus.Collector
		/*var expected = `
			testscope_stopwatch:s2_m{quantile="0.5"} 0
			testscope_stopwatch:s2_m{quantile="0.9"} 0
			testscope_stopwatch:s2_m{quantile="0.99"} 0
			testscope_stopwatch:s2_m_sum 0
			testscope_stopwatch:s2_m_count 1
		`
		err := testutil.CollectAndCompare(s.StopWatch, strings.NewReader(header+expected))
		assert.NoError(t, err)*/
	})

	t.Run("AdditionalLabels", func(t *testing.T) {
		scope := promutils.NewScope("testscope_stopwatch")
		opts := AdditionalLabelsOption{Labels: []string{contextutils.ProjectKey.String(), contextutils.ExecIDKey.String()}}
		s := NewStopWatch("s3", "some desc", time.Minute, scope, opts)
		assert.NotNil(t, s)

		ctx := context.TODO()
		const header = `
			# HELP testscope_stopwatch:s3_m some desc
			# TYPE testscope_stopwatch:s3_m summary
		`

		w := s.Start(ctx)
		w.Stop()
		var expected = `
			testscope_stopwatch:s3_m{domain="",exec_id="",project="",task="",wf="",quantile="0.5"} 0
            testscope_stopwatch:s3_m{domain="",exec_id="",project="",task="",wf="",quantile="0.9"} 0
            testscope_stopwatch:s3_m{domain="",exec_id="",project="",task="",wf="",quantile="0.99"} 0
            testscope_stopwatch:s3_m_sum{domain="",exec_id="",project="",task="",wf=""} 0
            testscope_stopwatch:s3_m_count{domain="",exec_id="",project="",task="",wf=""} 1
		`
		err := testutil.CollectAndCompare(s.StopWatchVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		ctx = contextutils.WithProjectDomain(ctx, "project", "domain")
		w = s.Start(ctx)
		w.Stop()
		expected = `
			testscope_stopwatch:s3_m{domain="",exec_id="",project="",task="",wf="",quantile="0.5"} 0
            testscope_stopwatch:s3_m{domain="",exec_id="",project="",task="",wf="",quantile="0.9"} 0
            testscope_stopwatch:s3_m{domain="",exec_id="",project="",task="",wf="",quantile="0.99"} 0
            testscope_stopwatch:s3_m_sum{domain="",exec_id="",project="",task="",wf=""} 0
            testscope_stopwatch:s3_m_count{domain="",exec_id="",project="",task="",wf=""} 1
			testscope_stopwatch:s3_m{domain="domain",exec_id="",project="project",task="",wf="",quantile="0.5"} 0
            testscope_stopwatch:s3_m{domain="domain",exec_id="",project="project",task="",wf="",quantile="0.9"} 0
            testscope_stopwatch:s3_m{domain="domain",exec_id="",project="project",task="",wf="",quantile="0.99"} 0
            testscope_stopwatch:s3_m_sum{domain="domain",exec_id="",project="project",task="",wf=""} 0
            testscope_stopwatch:s3_m_count{domain="domain",exec_id="",project="project",task="",wf=""} 1
		`
		err = testutil.CollectAndCompare(s.StopWatchVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		ctx = contextutils.WithExecutionID(ctx, "exec_id")
		w = s.Start(ctx)
		w.Stop()
		expected = `
			testscope_stopwatch:s3_m{domain="",exec_id="",project="",task="",wf="",quantile="0.5"} 0
            testscope_stopwatch:s3_m{domain="",exec_id="",project="",task="",wf="",quantile="0.9"} 0
            testscope_stopwatch:s3_m{domain="",exec_id="",project="",task="",wf="",quantile="0.99"} 0
            testscope_stopwatch:s3_m_sum{domain="",exec_id="",project="",task="",wf=""} 0
            testscope_stopwatch:s3_m_count{domain="",exec_id="",project="",task="",wf=""} 1
			testscope_stopwatch:s3_m{domain="domain",exec_id="",project="project",task="",wf="",quantile="0.5"} 0
            testscope_stopwatch:s3_m{domain="domain",exec_id="",project="project",task="",wf="",quantile="0.9"} 0
            testscope_stopwatch:s3_m{domain="domain",exec_id="",project="project",task="",wf="",quantile="0.99"} 0
            testscope_stopwatch:s3_m_sum{domain="domain",exec_id="",project="project",task="",wf=""} 0
            testscope_stopwatch:s3_m_count{domain="domain",exec_id="",project="project",task="",wf=""} 1
			testscope_stopwatch:s3_m{domain="domain",exec_id="exec_id",project="project",task="",wf="",quantile="0.5"} 0
            testscope_stopwatch:s3_m{domain="domain",exec_id="exec_id",project="project",task="",wf="",quantile="0.9"} 0
            testscope_stopwatch:s3_m{domain="domain",exec_id="exec_id",project="project",task="",wf="",quantile="0.99"} 0
            testscope_stopwatch:s3_m_sum{domain="domain",exec_id="exec_id",project="project",task="",wf=""} 0
            testscope_stopwatch:s3_m_count{domain="domain",exec_id="exec_id",project="project",task="",wf=""} 1
		`
		err = testutil.CollectAndCompare(s.StopWatchVec, strings.NewReader(header+expected))
		assert.NoError(t, err)
	})
}
