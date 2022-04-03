package labeled

import (
	"context"
	"testing"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestLabeledCounter(t *testing.T) {
	UnsetMetricKeys()
	assert.NotPanics(t, func() {
		SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey, contextutils.LaunchPlanIDKey)
	})

	scope := promutils.NewTestScope()
	// Make sure we will not register the same metrics key again.
	option := AdditionalLabelsOption{Labels: []string{contextutils.ProjectKey.String(), contextutils.DomainKey.String()}}
	c := NewCounter("lbl_counter", "help", scope, option)
	assert.NotNil(t, c)
	ctx := context.TODO()
	c.Inc(ctx)
	c.Add(ctx, 1.0)

	ctx = contextutils.WithProjectDomain(ctx, "project", "domain")
	c.Inc(ctx)
	c.Add(ctx, 1.0)

	ctx = contextutils.WithTaskID(ctx, "task")
	c.Inc(ctx)
	c.Add(ctx, 1.0)

	ctx = contextutils.WithLaunchPlanID(ctx, "lp")
	c.Inc(ctx)
	c.Add(ctx, 1.0)
}
