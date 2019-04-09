package labeled

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricOption(t *testing.T) {
	var opt MetricOption = &EmitUnlabeledMetric
	_, isMetricOption := opt.(MetricOption)
	assert.True(t, isMetricOption)
}
