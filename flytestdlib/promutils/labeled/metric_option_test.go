package labeled

import (
	"testing"
)

func TestMetricOption(t *testing.T) {
	var _ MetricOption = &EmitUnlabeledMetric
}
