package labeled

import (
	"context"
	"testing"

	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestLabeledCounter(t *testing.T) {
	UnsetMetricKeys()
	assert.NotPanics(t, func() {
		SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey, contextutils.LaunchPlanIDKey)
	})

	scope := promutils.NewTestScope()
	c := NewCounter("lbl_counter", "help", scope)
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
