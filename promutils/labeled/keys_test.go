package labeled

import (
	"testing"

	"github.com/lyft/flytestdlib/contextutils"
	"github.com/stretchr/testify/assert"
)

func TestMetricKeys(t *testing.T) {
	input := []contextutils.Key{
		contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey,
	}

	assert.NotPanics(t, func() { SetMetricKeys(input...) })
	assert.Equal(t, input, metricKeys)

	for i, k := range metricKeys {
		assert.Equal(t, k.String(), metricStringKeys[i])
	}

	assert.NotPanics(t, func() { SetMetricKeys(input...) })
	assert.Panics(t, func() { SetMetricKeys(contextutils.DomainKey) })
}
