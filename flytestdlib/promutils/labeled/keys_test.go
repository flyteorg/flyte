package labeled

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flytestdlib/contextutils"
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
