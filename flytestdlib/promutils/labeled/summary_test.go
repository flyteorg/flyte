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
		SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey, contextutils.LaunchPlanIDKey)
	})

	t.Run("labeled", func(t *testing.T) {
		scope := promutils.NewScope("testscope_summary")
		ctx := context.Background()
		ctx = contextutils.WithProjectDomain(ctx, "flyte", "dev")
		// Make sure we will not register the same metrics key again.
		option := AdditionalLabelsOption{Labels: []string{contextutils.ProjectKey.String(), contextutils.DomainKey.String()}}
		s := NewSummary("unittest", "some desc", scope, option)
		assert.NotNil(t, s)

		s.Observe(ctx, 10)

		const header = `
			# HELP testscope_summary:unittest some desc
			# TYPE testscope_summary:unittest summary
		`
		var expected = `
			testscope_summary:unittest{domain="dev",lp="",project="flyte",task="",wf="",quantile="0.5"} 10
			testscope_summary:unittest{domain="dev",lp="",project="flyte",task="",wf="",quantile="0.9"} 10
			testscope_summary:unittest{domain="dev",lp="",project="flyte",task="",wf="",quantile="0.99"} 10
			testscope_summary:unittest_sum{domain="dev",lp="",project="flyte",task="",wf=""} 10
			testscope_summary:unittest_count{domain="dev",lp="",project="flyte",task="",wf=""} 1
		`
		err := testutil.CollectAndCompare(s.SummaryVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		s.Observe(ctx, 100)
		s.Observe(ctx, 1000)

		expected = `
			testscope_summary:unittest{domain="dev",lp="",project="flyte",task="",wf="",quantile="0.5"} 100
			testscope_summary:unittest{domain="dev",lp="",project="flyte",task="",wf="",quantile="0.9"} 1000
			testscope_summary:unittest{domain="dev",lp="",project="flyte",task="",wf="",quantile="0.99"} 1000
			testscope_summary:unittest_sum{domain="dev",lp="",project="flyte",task="",wf=""} 1110
			testscope_summary:unittest_count{domain="dev",lp="",project="flyte",task="",wf=""} 3
		`
		err = testutil.CollectAndCompare(s.SummaryVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		s.Observe(contextutils.WithProjectDomain(ctx, "flyte", "prod"), 10)

		expected = `
			testscope_summary:unittest{domain="dev",lp="",project="flyte",task="",wf="",quantile="0.5"} 100
			testscope_summary:unittest{domain="dev",lp="",project="flyte",task="",wf="",quantile="0.9"} 1000
			testscope_summary:unittest{domain="dev",lp="",project="flyte",task="",wf="",quantile="0.99"} 1000
			testscope_summary:unittest_sum{domain="dev",lp="",project="flyte",task="",wf=""} 1110
			testscope_summary:unittest_count{domain="dev",lp="",project="flyte",task="",wf=""} 3
			testscope_summary:unittest{domain="prod",lp="",project="flyte",task="",wf="",quantile="0.5"} 10
			testscope_summary:unittest{domain="prod",lp="",project="flyte",task="",wf="",quantile="0.9"} 10
			testscope_summary:unittest{domain="prod",lp="",project="flyte",task="",wf="",quantile="0.99"} 10
			testscope_summary:unittest_sum{domain="prod",lp="",project="flyte",task="",wf=""} 10
			testscope_summary:unittest_count{domain="prod",lp="",project="flyte",task="",wf=""} 1
		`
		err = testutil.CollectAndCompare(s.SummaryVec, strings.NewReader(header+expected))
		assert.NoError(t, err)
	})

	t.Run("extra labels", func(t *testing.T) {
		scope := promutils.NewScope("testscope_summary")
		ctx := context.Background()
		s := NewSummary("unittest2", "some desc", scope, AdditionalLabelsOption{Labels: []string{"method"}})
		assert.NotNil(t, s)

		methodKey := contextutils.Key("method")
		s.Observe(context.WithValue(ctx, methodKey, "GET"), 10)
		s.Observe(context.WithValue(ctx, methodKey, "POST"), 100)

		const header = `
			# HELP testscope_summary:unittest2 some desc
			# TYPE testscope_summary:unittest2 summary
		`
		var expected = `
			testscope_summary:unittest2{domain="",lp="",method="GET",project="",task="",wf="",quantile="0.5"} 10
			testscope_summary:unittest2{domain="",lp="",method="GET",project="",task="",wf="",quantile="0.9"} 10
			testscope_summary:unittest2{domain="",lp="",method="GET",project="",task="",wf="",quantile="0.99"} 10
			testscope_summary:unittest2_sum{domain="",lp="",method="GET",project="",task="",wf=""} 10
			testscope_summary:unittest2_count{domain="",lp="",method="GET",project="",task="",wf=""} 1
			testscope_summary:unittest2{domain="",lp="",method="POST",project="",task="",wf="",quantile="0.5"} 100
			testscope_summary:unittest2{domain="",lp="",method="POST",project="",task="",wf="",quantile="0.9"} 100
			testscope_summary:unittest2{domain="",lp="",method="POST",project="",task="",wf="",quantile="0.99"} 100
			testscope_summary:unittest2_sum{domain="",lp="",method="POST",project="",task="",wf=""} 100
			testscope_summary:unittest2_count{domain="",lp="",method="POST",project="",task="",wf=""} 1
		`
		err := testutil.CollectAndCompare(s.SummaryVec, strings.NewReader(header+expected))
		assert.NoError(t, err)
	})

	t.Run("unlabeled", func(t *testing.T) {
		scope := promutils.NewScope("testscope_summary")
		ctx := context.Background()
		ctx = contextutils.WithProjectDomain(ctx, "flyte", "dev")
		s := NewSummary("unittest3", "some desc", scope, EmitUnlabeledMetric)
		assert.NotNil(t, s)

		s.Observe(ctx, 10)

		const header = `
			# HELP testscope_summary:unittest3_unlabeled some desc
			# TYPE testscope_summary:unittest3_unlabeled summary
		`
		var expected = `
			testscope_summary:unittest3_unlabeled{quantile="0.5"} 10
			testscope_summary:unittest3_unlabeled{quantile="0.9"} 10
			testscope_summary:unittest3_unlabeled{quantile="0.99"} 10
			testscope_summary:unittest3_unlabeled_sum 10
			testscope_summary:unittest3_unlabeled_count 1
		`
		err := testutil.CollectAndCompare(s.Summary, strings.NewReader(header+expected))
		assert.NoError(t, err)
	})
}
