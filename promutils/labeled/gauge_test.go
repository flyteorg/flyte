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
		SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey, contextutils.LaunchPlanIDKey)
	})

	scope := promutils.NewScope("testscope")
	ctx := context.Background()
	ctx = contextutils.WithProjectDomain(ctx, "flyte", "dev")
	// Make sure we will not register the same metrics key again.
	option := AdditionalLabelsOption{Labels: []string{contextutils.ProjectKey.String(), contextutils.DomainKey.String()}}
	g := NewGauge("unittest", "some desc", scope, option)
	assert.NotNil(t, g)

	g.Inc(ctx)

	const header = `
		# HELP testscope:unittest some desc
        # TYPE testscope:unittest gauge
	`
	var expected = `
        testscope:unittest{domain="dev",lp="",project="flyte",task="",wf=""} 1
	`
	err := testutil.CollectAndCompare(g.GaugeVec, strings.NewReader(header+expected))
	assert.NoError(t, err)

	g.Set(ctx, 42)
	expected = `
        testscope:unittest{domain="dev",lp="",project="flyte",task="",wf=""} 42
	`
	err = testutil.CollectAndCompare(g.GaugeVec, strings.NewReader(header+expected))
	assert.NoError(t, err)

	g.Add(ctx, 1)
	expected = `
        testscope:unittest{domain="dev",lp="",project="flyte",task="",wf=""} 43
	`
	err = testutil.CollectAndCompare(g.GaugeVec, strings.NewReader(header+expected))
	assert.NoError(t, err)

	g.Dec(ctx)
	expected = `
        testscope:unittest{domain="dev",lp="",project="flyte",task="",wf=""} 42
	`
	err = testutil.CollectAndCompare(g.GaugeVec, strings.NewReader(header+expected))
	assert.NoError(t, err)

	g.Sub(ctx, 1)
	expected = `
        testscope:unittest{domain="dev",lp="",project="flyte",task="",wf=""} 41
	`
	err = testutil.CollectAndCompare(g.GaugeVec, strings.NewReader(header+expected))
	assert.NoError(t, err)

	g.SetToCurrentTime(ctx)
}

func TestWithAdditionalLabels(t *testing.T) {
	UnsetMetricKeys()
	assert.NotPanics(t, func() {
		SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey, contextutils.LaunchPlanIDKey)
	})

	scope := promutils.NewScope("testscope")
	ctx := context.Background()
	ctx = contextutils.WithProjectDomain(ctx, "flyte", "dev")
	g := NewGauge("unittestlabeled", "some desc", scope, AdditionalLabelsOption{Labels: []string{"bearing"}})
	assert.NotNil(t, g)

	const header = `
		# HELP testscope:unittestlabeled some desc
		# TYPE testscope:unittestlabeled gauge
	`

	g.Inc(ctx)
	var expected = `
        testscope:unittestlabeled{bearing="", domain="dev",lp="",project="flyte",task="",wf=""} 1
	`
	err := testutil.CollectAndCompare(g.GaugeVec, strings.NewReader(header+expected))
	assert.NoError(t, err)

	bearingKey := contextutils.Key("bearing")
	ctx = context.WithValue(ctx, bearingKey, "123")
	g.Set(ctx, 42)
	expected = `
		testscope:unittestlabeled{bearing="", domain="dev",lp="",project="flyte",task="",wf=""} 1
		testscope:unittestlabeled{bearing="123", domain="dev",lp="",project="flyte",task="",wf=""} 42
	`
	err = testutil.CollectAndCompare(g.GaugeVec, strings.NewReader(header+expected))
	assert.NoError(t, err)

	g.Add(ctx, 1)
	expected = `
		testscope:unittestlabeled{bearing="", domain="dev",lp="",project="flyte",task="",wf=""} 1
		testscope:unittestlabeled{bearing="123", domain="dev",lp="",project="flyte",task="",wf=""} 43
	`
	err = testutil.CollectAndCompare(g.GaugeVec, strings.NewReader(header+expected))
	assert.NoError(t, err)

	g.Dec(ctx)
	expected = `
		testscope:unittestlabeled{bearing="", domain="dev",lp="",project="flyte",task="",wf=""} 1
		testscope:unittestlabeled{bearing="123", domain="dev",lp="",project="flyte",task="",wf=""} 42
	`
	err = testutil.CollectAndCompare(g.GaugeVec, strings.NewReader(header+expected))
	assert.NoError(t, err)

	g.Sub(ctx, 42)
	expected = `
		testscope:unittestlabeled{bearing="", domain="dev",lp="",project="flyte",task="",wf=""} 1
		testscope:unittestlabeled{bearing="123", domain="dev",lp="",project="flyte",task="",wf=""} 0
	`
	err = testutil.CollectAndCompare(g.GaugeVec, strings.NewReader(header+expected))
	assert.NoError(t, err)
}
