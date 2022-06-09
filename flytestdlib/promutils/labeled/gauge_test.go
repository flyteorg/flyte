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

func TestLabeledGauge(t *testing.T) {
	UnsetMetricKeys()
	assert.NotPanics(t, func() {
		SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
	})

	t.Run("Labeled", func(t *testing.T) {
		scope := promutils.NewScope("testscope_gauge")
		g := NewGauge("g1", "some desc", scope)
		assert.NotNil(t, g)

		ctx := context.TODO()
		const header = `
			# HELP testscope_gauge:g1 some desc
			# TYPE testscope_gauge:g1 gauge
		`

		g.Inc(ctx)
		var expected = `
			testscope_gauge:g1{domain="",project="",task="",wf=""} 1
		`
		err := testutil.CollectAndCompare(g.GaugeVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		g.Dec(ctx)
		expected = `
			testscope_gauge:g1{domain="",project="",task="",wf=""} 0
		`
		err = testutil.CollectAndCompare(g.GaugeVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		ctx = contextutils.WithProjectDomain(ctx, "project", "domain")
		g.Set(ctx, 42)
		expected = `
			testscope_gauge:g1{domain="",project="",task="",wf=""} 0
			testscope_gauge:g1{domain="domain",project="project",task="",wf=""} 42
		`
		err = testutil.CollectAndCompare(g.GaugeVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		ctx = contextutils.WithTaskID(ctx, "task")
		g.Add(ctx, 1)
		expected = `
			testscope_gauge:g1{domain="",project="",task="",wf=""} 0
			testscope_gauge:g1{domain="domain",project="project",task="",wf=""} 42
			testscope_gauge:g1{domain="domain",project="project",task="task",wf=""} 1
		`
		err = testutil.CollectAndCompare(g.GaugeVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		g.Sub(ctx, 1)
		expected = `
			testscope_gauge:g1{domain="",project="",task="",wf=""} 0
			testscope_gauge:g1{domain="domain",project="project",task="",wf=""} 42
			testscope_gauge:g1{domain="domain",project="project",task="task",wf=""} 0
		`
		err = testutil.CollectAndCompare(g.GaugeVec, strings.NewReader(header+expected))
		assert.NoError(t, err)
	})

	t.Run("Unlabeled", func(t *testing.T) {
		scope := promutils.NewScope("testscope_gauge")
		g := NewGauge("g2", "some desc", scope, EmitUnlabeledMetric)
		assert.NotNil(t, g)

		ctx := context.TODO()
		const header = `
			# HELP testscope_gauge:g2_unlabeled some desc
			# TYPE testscope_gauge:g2_unlabeled gauge
		`

		g.Inc(ctx)
		var expected = `
			testscope_gauge:g2_unlabeled 1
		`
		err := testutil.CollectAndCompare(g.Gauge, strings.NewReader(header+expected))
		assert.NoError(t, err)

		g.Dec(ctx)
		expected = `
			testscope_gauge:g2_unlabeled 0
		`
		err = testutil.CollectAndCompare(g.Gauge, strings.NewReader(header+expected))
		assert.NoError(t, err)
	})

	t.Run("AdditionalLabels", func(t *testing.T) {
		scope := promutils.NewScope("testscope_gauge")
		opts := AdditionalLabelsOption{Labels: []string{contextutils.ProjectKey.String(), contextutils.ExecIDKey.String()}}
		g := NewGauge("g3", "some desc", scope, opts)
		assert.NotNil(t, g)

		ctx := context.TODO()
		const header = `
			# HELP testscope_gauge:g3 some desc
			# TYPE testscope_gauge:g3 gauge
		`

		g.Inc(ctx)
		var expected = `
			testscope_gauge:g3{domain="",exec_id="",project="",task="",wf=""} 1
		`
		err := testutil.CollectAndCompare(g.GaugeVec, strings.NewReader(header+expected))
		assert.NoError(t, err)

		ctx = contextutils.WithExecutionID(ctx, "exec_id")
		g.Inc(ctx)
		expected = `
			testscope_gauge:g3{domain="",exec_id="",project="",task="",wf=""} 1
			testscope_gauge:g3{domain="",exec_id="exec_id",project="",task="",wf=""} 1
		`
		err = testutil.CollectAndCompare(g.GaugeVec, strings.NewReader(header+expected))
		assert.NoError(t, err)
	})
}
