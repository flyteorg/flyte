package labeled

// Defines extra set of options to customize the emitted metric.
type MetricOption interface {
	isMetricOption()
}

// Instructs the metric to emit unlabeled metric (besides the labeled one). This is useful to get overall system
// performance.
type EmitUnlabeledMetricOption struct {
}

func (EmitUnlabeledMetricOption) isMetricOption() {}

var EmitUnlabeledMetric = EmitUnlabeledMetricOption{}
