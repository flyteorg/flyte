package webapi

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/webapi"
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
