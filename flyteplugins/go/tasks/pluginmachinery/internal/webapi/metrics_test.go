package webapi

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	coreMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/webapi"
	cacheMocks "github.com/flyteorg/flyte/v2/flytestdlib/autorefreshcache/mocks"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

func Test_enteredResourcesCreated(t *testing.T) {
	cases := []struct {
		name string
		prev webapi.Phase
		next webapi.Phase
		want bool
	}{
		{"launch creates resource", webapi.PhaseAllocationTokenAcquired, webapi.PhaseResourcesCreated, true},
		{"no-quota launch", webapi.PhaseNotStarted, webapi.PhaseResourcesCreated, true},
		{"still monitoring", webapi.PhaseResourcesCreated, webapi.PhaseResourcesCreated, false},
		{"never created", webapi.PhaseNotStarted, webapi.PhaseAllocationTokenAcquired, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, enteredResourcesCreated(c.prev, c.next))
		})
	}
}

func Test_activeTasksGauge(t *testing.T) {
	m := newMetrics(promutils.NewTestScope())
	assert.Equal(t, 0.0, testutil.ToFloat64(m.ActiveTasks))
	m.ActiveTasks.Inc()
	assert.Equal(t, 1.0, testutil.ToFloat64(m.ActiveTasks))
	m.ActiveTasks.Dec()
	assert.Equal(t, 0.0, testutil.ToFloat64(m.ActiveTasks))
}

// Test_Handle_activeTasksInc is a regression test for the state pointer-aliasing bug.
//
// launch()/allocateToken() receive &incomingState and mutate state.Phase to
// PhaseResourcesCreated in place, returning that same pointer as nextState. The
// active_tasks Inc check therefore MUST compare against the phase captured BEFORE the
// switch runs; reading incomingState.Phase after the switch aliases the already-mutated
// value, so enteredResourcesCreated(prev, next) is always false and the gauge never
// increments. The pure-predicate Test_enteredResourcesCreated above passes regardless of
// this bug, so it does not catch it — this Handle-level test does.
func Test_Handle_activeTasksInc(t *testing.T) {
	ctx := context.Background()

	// State reader returns the zero-value State, i.e. PhaseNotStarted.
	stateReader := &coreMocks.PluginStateReader{}
	stateReader.EXPECT().Get(mock.Anything).Return(uint8(0), nil)

	stateWriter := &coreMocks.PluginStateWriter{}
	stateWriter.EXPECT().Put(mock.Anything, mock.Anything).Return(nil)

	taskID := &coreMocks.TaskExecutionID{}
	taskID.EXPECT().GetGeneratedName().Return("my-id")
	meta := &coreMocks.TaskExecutionMetadata{}
	meta.EXPECT().GetTaskExecutionID().Return(taskID)

	tCtx := &coreMocks.TaskExecutionContext{}
	tCtx.EXPECT().PluginStateReader().Return(stateReader)
	tCtx.EXPECT().PluginStateWriter().Return(stateWriter)
	tCtx.EXPECT().TaskExecutionMetadata().Return(meta)

	cache := &cacheMocks.AutoRefresh{}
	cache.EXPECT().GetOrCreate(mock.Anything, mock.Anything).Return(nil, nil)

	// No ResourceQuotas -> Handle takes the direct launch() path from PhaseNotStarted.
	plgn := newPluginWithProperties(webapi.PluginConfig{})
	plgn.EXPECT().Create(ctx, tCtx).Return("resource-meta", nil, nil)

	c := CorePlugin{
		id:      "test-plugin",
		p:       plgn,
		cache:   cache,
		metrics: newMetrics(promutils.NewTestScope()),
	}

	assert.Equal(t, 0.0, testutil.ToFloat64(c.metrics.ActiveTasks))

	_, err := c.Handle(ctx, tCtx)
	assert.NoError(t, err)

	// NotStarted -> ResourcesCreated must bump the gauge exactly once.
	assert.Equal(t, 1.0, testutil.ToFloat64(c.metrics.ActiveTasks))
}
