package labeled

import (
	"context"
	"testing"
	"time"

	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestLabeledStopWatch(t *testing.T) {
	assert.NotPanics(t, func() {
		SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
	})

	t.Run("always labeled", func(t *testing.T) {
		scope := promutils.NewTestScope()
		c := NewStopWatch("lbl_counter", "help", time.Second, scope)
		assert.NotNil(t, c)
		ctx := context.TODO()
		w := c.Start(ctx)
		w.Stop()

		ctx = contextutils.WithProjectDomain(ctx, "project", "domain")
		w = c.Start(ctx)
		w.Stop()

		ctx = contextutils.WithTaskID(ctx, "task")
		w = c.Start(ctx)
		w.Stop()

		c.Observe(ctx, time.Now(), time.Now().Add(time.Second))
		c.Time(ctx, func() {
			// Do nothing
		})
	})

	t.Run("unlabeled", func(t *testing.T) {
		scope := promutils.NewTestScope()
		c := NewStopWatch("lbl_counter_2", "help", time.Second, scope, EmitUnlabeledMetric)
		assert.NotNil(t, c)

		c.Start(context.TODO())
	})
}
