package labeled

import (
	"testing"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/stretchr/testify/assert"
)

func TestMetricKeys(t *testing.T) {
	UnsetMetricKeys()
	input := []contextutils.Key{
		contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey, contextutils.LaunchPlanIDKey,
	}

	assert.NotPanics(t, func() { SetMetricKeys(input...) })
	assert.Equal(t, input, metricKeys)

	for i, k := range metricKeys {
		assert.Equal(t, k.String(), metricStringKeys[i])
	}

	assert.NotPanics(t, func() { SetMetricKeys(input...) })
	assert.Panics(t, func() { SetMetricKeys(contextutils.DomainKey) })
}
