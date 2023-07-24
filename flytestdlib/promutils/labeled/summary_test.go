package labeled

import (
	"context"
	"strings"
	"testing"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"
)

func TestLabeledSummary(t *testing.T) {
	UnsetMetricKeys()
	assert.NotPanics(t, func() {
		SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
	})

	t.Run("Labeled", func(t *testing.T) {
		scope := promutils.NewScope("testscope_summary")
		s := NewSummary("s1", "some desc", scope)
		assert.NotNil(t, s)

		ctx := context.TODO()
		const header = `
			# HELP testscope_summary:s1 some desc
			# TYPE testscope_summary:s1 summary
		`

		s.Observe(ctx, 10)
		var expected = `
			testscope_summary:s1{domain="",project="",task="",wf="",quantile="0.5"} 10
			testscope_summary:s1{domain="",project="",task="",wf="",quantile="0.9"} 10
			testscope_summary:s1{domain="",project="",task="",wf="",quantile="0.99"} 10
			testscope_summary:s1_sum{domain="",project="",task="",wf=""} 10
			testscope_summary:s1_count{domain="",project="",task="",wf=""} 1
		`
		err := testutil.CollectAndCompare(s.SummaryVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		s.Observe(ctx, 100)
		s.Observe(ctx, 1000)
		expected = `
			testscope_summary:s1{domain="",project="",task="",wf="",quantile="0.5"} 100
			testscope_summary:s1{domain="",project="",task="",wf="",quantile="0.9"} 1000
			testscope_summary:s1{domain="",project="",task="",wf="",quantile="0.99"} 1000
			testscope_summary:s1_sum{domain="",project="",task="",wf=""} 1110
			testscope_summary:s1_count{domain="",project="",task="",wf=""} 3
		`
		err = testutil.CollectAndCompare(s.SummaryVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		ctx = contextutils.WithProjectDomain(ctx, "project", "domain")
		s.Observe(ctx, 10)
		expected = `
			testscope_summary:s1{domain="",project="",task="",wf="",quantile="0.5"} 100
			testscope_summary:s1{domain="",project="",task="",wf="",quantile="0.9"} 1000
			testscope_summary:s1{domain="",project="",task="",wf="",quantile="0.99"} 1000
			testscope_summary:s1_sum{domain="",project="",task="",wf=""} 1110
			testscope_summary:s1_count{domain="",project="",task="",wf=""} 3
			testscope_summary:s1{domain="domain",project="project",task="",wf="",quantile="0.5"} 10
			testscope_summary:s1{domain="domain",project="project",task="",wf="",quantile="0.9"} 10
			testscope_summary:s1{domain="domain",project="project",task="",wf="",quantile="0.99"} 10
			testscope_summary:s1_sum{domain="domain",project="project",task="",wf=""} 10
			testscope_summary:s1_count{domain="domain",project="project",task="",wf=""} 1
		`
		err = testutil.CollectAndCompare(s.SummaryVec, strings.NewReader(header+expected))
		assert.NoError(t, err)
	})

	t.Run("Unlabeled", func(t *testing.T) {
		scope := promutils.NewScope("testscope_summary")
		s := NewSummary("s2", "some desc", scope, EmitUnlabeledMetric)
		assert.NotNil(t, s)

		ctx := context.TODO()
		const header = `
			# HELP testscope_summary:s2_unlabeled some desc
			# TYPE testscope_summary:s2_unlabeled summary
		`

		s.Observe(ctx, 10)
		var expected = `
			testscope_summary:s2_unlabeled{quantile="0.5"} 10
			testscope_summary:s2_unlabeled{quantile="0.9"} 10
			testscope_summary:s2_unlabeled{quantile="0.99"} 10
			testscope_summary:s2_unlabeled_sum 10
			testscope_summary:s2_unlabeled_count 1
		`
		err := testutil.CollectAndCompare(s.Summary, strings.NewReader(header+expected))
		assert.NoError(t, err)
	})

	/*t.Run("extra labels", func(t *testing.T) {
		scope := promutils.NewScope("testscope_summary")
		ctx := context.Background()
		s := NewSummary("s12", "some desc", scope, AdditionalLabelsOption{Labels: []string{"method"}})
		assert.NotNil(t, s)

		methodKey := contextutils.Key("method")
		s.Observe(context.WithValue(ctx, methodKey, "GET"), 10)
		s.Observe(context.WithValue(ctx, methodKey, "POST"), 100)

		const header = `
			# HELP testscope_summary:s12 some desc
			# TYPE testscope_summary:s12 summary
		`
		var expected = `
			testscope_summary:s12{domain="",lp="",method="GET",project="",task="",wf="",quantile="0.5"} 10
			testscope_summary:s12{domain="",lp="",method="GET",project="",task="",wf="",quantile="0.9"} 10
			testscope_summary:s12{domain="",lp="",method="GET",project="",task="",wf="",quantile="0.99"} 10
			testscope_summary:s12_sum{domain="",lp="",method="GET",project="",task="",wf=""} 10
			testscope_summary:s12_count{domain="",lp="",method="GET",project="",task="",wf=""} 1
			testscope_summary:s12{domain="",lp="",method="POST",project="",task="",wf="",quantile="0.5"} 100
			testscope_summary:s12{domain="",lp="",method="POST",project="",task="",wf="",quantile="0.9"} 100
			testscope_summary:s12{domain="",lp="",method="POST",project="",task="",wf="",quantile="0.99"} 100
			testscope_summary:s12_sum{domain="",lp="",method="POST",project="",task="",wf=""} 100
			testscope_summary:s12_count{domain="",lp="",method="POST",project="",task="",wf=""} 1
		`
		err := testutil.CollectAndCompare(s.SummaryVec, strings.NewReader(header+expected))
		assert.NoError(t, err)
	})*/

	t.Run("AdditionalLabels", func(t *testing.T) {
		scope := promutils.NewScope("testscope_summary")
		opts := AdditionalLabelsOption{Labels: []string{contextutils.ProjectKey.String(), contextutils.ExecIDKey.String()}}
		s := NewSummary("s3", "some desc", scope, opts)
		assert.NotNil(t, s)

		ctx := context.TODO()
		const header = `
			# HELP testscope_summary:s3 some desc
			# TYPE testscope_summary:s3 summary
		`

		s.Observe(ctx, 10)
		var expected = `
			testscope_summary:s3{domain="",exec_id="",project="",task="",wf="",quantile="0.5"} 10
			testscope_summary:s3{domain="",exec_id="",project="",task="",wf="",quantile="0.9"} 10
			testscope_summary:s3{domain="",exec_id="",project="",task="",wf="",quantile="0.99"} 10
			testscope_summary:s3_sum{domain="",exec_id="",project="",task="",wf=""} 10
			testscope_summary:s3_count{domain="",exec_id="",project="",task="",wf=""} 1
		`
		err := testutil.CollectAndCompare(s.SummaryVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		ctx = contextutils.WithProjectDomain(ctx, "project", "domain")
		s.Observe(ctx, 10)
		expected = `
			testscope_summary:s3{domain="",exec_id="",project="",task="",wf="",quantile="0.5"} 10
			testscope_summary:s3{domain="",exec_id="",project="",task="",wf="",quantile="0.9"} 10
			testscope_summary:s3{domain="",exec_id="",project="",task="",wf="",quantile="0.99"} 10
			testscope_summary:s3_sum{domain="",exec_id="",project="",task="",wf=""} 10
			testscope_summary:s3_count{domain="",exec_id="",project="",task="",wf=""} 1
			testscope_summary:s3{domain="domain",exec_id="",project="project",task="",wf="",quantile="0.5"} 10
			testscope_summary:s3{domain="domain",exec_id="",project="project",task="",wf="",quantile="0.9"} 10
			testscope_summary:s3{domain="domain",exec_id="",project="project",task="",wf="",quantile="0.99"} 10
			testscope_summary:s3_sum{domain="domain",exec_id="",project="project",task="",wf=""} 10
			testscope_summary:s3_count{domain="domain",exec_id="",project="project",task="",wf=""} 1
		`
		err = testutil.CollectAndCompare(s.SummaryVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		ctx = contextutils.WithExecutionID(ctx, "exec_id")
		s.Observe(ctx, 10)
		expected = `
			testscope_summary:s3{domain="",exec_id="",project="",task="",wf="",quantile="0.5"} 10
			testscope_summary:s3{domain="",exec_id="",project="",task="",wf="",quantile="0.9"} 10
			testscope_summary:s3{domain="",exec_id="",project="",task="",wf="",quantile="0.99"} 10
			testscope_summary:s3_sum{domain="",exec_id="",project="",task="",wf=""} 10
			testscope_summary:s3_count{domain="",exec_id="",project="",task="",wf=""} 1
			testscope_summary:s3{domain="domain",exec_id="",project="project",task="",wf="",quantile="0.5"} 10
			testscope_summary:s3{domain="domain",exec_id="",project="project",task="",wf="",quantile="0.9"} 10
			testscope_summary:s3{domain="domain",exec_id="",project="project",task="",wf="",quantile="0.99"} 10
			testscope_summary:s3_sum{domain="domain",exec_id="",project="project",task="",wf=""} 10
			testscope_summary:s3_count{domain="domain",exec_id="",project="project",task="",wf=""} 1
			testscope_summary:s3{domain="domain",exec_id="exec_id",project="project",task="",wf="",quantile="0.5"} 10
			testscope_summary:s3{domain="domain",exec_id="exec_id",project="project",task="",wf="",quantile="0.9"} 10
			testscope_summary:s3{domain="domain",exec_id="exec_id",project="project",task="",wf="",quantile="0.99"} 10
			testscope_summary:s3_sum{domain="domain",exec_id="exec_id",project="project",task="",wf=""} 10
			testscope_summary:s3_count{domain="domain",exec_id="exec_id",project="project",task="",wf=""} 1
		`
		err = testutil.CollectAndCompare(s.SummaryVec, strings.NewReader(header+expected))
		assert.NoError(t, err)
	})
}
