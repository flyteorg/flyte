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

// AdditionalLabelsOption instructs the labeled metric to expect additional labels scoped for this just this metric
// in the context passed.
type AdditionalLabelsOption struct {
	// A collection of labels to look for in the passed context.
	Labels []string
}

func (AdditionalLabelsOption) isMetricOption() {}
