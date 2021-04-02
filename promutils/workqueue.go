// Source: https://raw.githubusercontent.com/kubernetes/kubernetes/3dbbd0bdf44cb07fdde85aa392adf99ea7e95939/pkg/util/workqueue/prometheus/prometheus.go
/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package promutils

import (
	"k8s.io/client-go/util/workqueue"

	"github.com/prometheus/client_golang/prometheus"
)

// Package prometheus sets the workqueue DefaultMetricsFactory to produce
// prometheus metrics. To use this package, you just have to import it.

func init() {
	provider := prometheusMetricsProvider{}
	workqueue.SetProvider(provider)
}

type prometheusMetricsProvider struct{}

func (prometheusMetricsProvider) NewLongestRunningProcessorSecondsMetric(name string) workqueue.SettableGaugeMetric {
	unfinishedWork := prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: name,
		Name:      "longest_running_processor_s",
		Help:      "How many microseconds longest running processor from workqueue" + name + " takes.",
	})

	prometheus.MustRegister(unfinishedWork)
	return unfinishedWork
}

func (prometheusMetricsProvider) NewUnfinishedWorkSecondsMetric(name string) workqueue.SettableGaugeMetric {
	unfinishedWork := prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: name,
		Name:      "unfinished_work_s",
		Help:      "How many seconds of work in progress in workqueue: " + name,
	})
	prometheus.MustRegister(unfinishedWork)
	return unfinishedWork
}

func (prometheusMetricsProvider) NewLongestRunningProcessorMicrosecondsMetric(name string) workqueue.SettableGaugeMetric {
	unfinishedWork := prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: name,
		Name:      "longest_running_processor_us",
		Help:      "How many microseconds longest running processor from workqueue" + name + " takes.",
	})
	prometheus.MustRegister(unfinishedWork)
	return unfinishedWork
}

func (prometheusMetricsProvider) NewDepthMetric(name string) workqueue.GaugeMetric {
	depth := prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: name,
		Name:      "depth",
		Help:      "Current depth of workqueue: " + name,
	})
	prometheus.MustRegister(depth)
	return depth
}

func (prometheusMetricsProvider) NewAddsMetric(name string) workqueue.CounterMetric {
	adds := prometheus.NewCounter(prometheus.CounterOpts{
		Subsystem: name,
		Name:      "adds",
		Help:      "Total number of adds handled by workqueue: " + name,
	})
	prometheus.MustRegister(adds)
	return adds
}

func (prometheusMetricsProvider) NewLatencyMetric(name string) workqueue.HistogramMetric {
	latency := prometheus.NewHistogram(prometheus.HistogramOpts{
		Subsystem: name,
		Name:      "queue_latency_us",
		Help:      "How long an item stays in workqueue" + name + " before being requested.",
	})
	prometheus.MustRegister(latency)
	return latency
}

func (prometheusMetricsProvider) NewWorkDurationMetric(name string) workqueue.HistogramMetric {
	workDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Subsystem: name,
		Name:      "work_duration_us",
		Help:      "How long processing an item from workqueue" + name + " takes.",
	})
	prometheus.MustRegister(workDuration)
	return workDuration
}

func (prometheusMetricsProvider) NewRetriesMetric(name string) workqueue.CounterMetric {
	retries := prometheus.NewCounter(prometheus.CounterOpts{
		Subsystem: name,
		Name:      "retries",
		Help:      "Total number of retries handled by workqueue: " + name,
	})
	prometheus.MustRegister(retries)
	return retries
}
