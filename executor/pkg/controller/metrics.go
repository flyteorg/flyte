package controller

import (
	"context"
	"encoding/json"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	toolscache "k8s.io/client-go/tools/cache"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
)

const taskActionMeterName = "taskaction-controller"

// Values for the "op" attribute on taskaction.k8s.duration.
const (
	opGet          = "get"
	opUpdate       = "update"
	opStatusUpdate = "status_update"
)

// taskActionMetrics holds OTel instruments for the TaskAction controller.
//
// Reconcile throughput, active workers, and workqueue latency already come from
// the controller-runtime metrics server, and event-proxy send latency comes from
// the otelconnect-wrapped events client (rpc_client_duration). This adds what
// none of those provide: TaskAction CRD count by phase, serialized CRD size, and
// per-operation Kubernetes API read/write latency for the CRD.
type taskActionMetrics struct {
	crdSizeBytes  metric.Int64Histogram
	crdOpDuration metric.Float64Histogram
}

// registerTaskActionMetrics wires the TaskAction OTel meters onto the given meter
// provider (the executor's, registered in executor/setup.go). The active-by-phase
// gauge is observed asynchronously via activeByPhase, which returns the current
// TaskAction count per plugin phase (see cachedPhaseCounter for the cache-backed,
// no-copy implementation). A nil activeByPhase leaves the gauge unobserved.
func registerTaskActionMetrics(provider metric.MeterProvider, activeByPhase func(context.Context) map[string]int64) (*taskActionMetrics, error) {
	if _, ok := provider.(metricnoop.MeterProvider); ok {
		return nil, nil
	}
	meter := provider.Meter(taskActionMeterName)

	crdSize, err := meter.Int64Histogram(
		"taskaction.crd.size_bytes",
		metric.WithDescription("Serialized (JSON) size of a TaskAction CRD"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, err
	}

	crdOp, err := meter.Float64Histogram(
		"taskaction.k8s.duration",
		metric.WithDescription("Latency of TaskAction CRD operations against the Kubernetes API, labeled by op (get/update/status_update)"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, err
	}

	active, err := meter.Int64ObservableGauge(
		"taskaction.active",
		metric.WithDescription("Number of TaskAction CRDs, labeled by plugin phase"),
	)
	if err != nil {
		return nil, err
	}

	_, err = meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		if activeByPhase == nil {
			return nil
		}
		for phase, n := range activeByPhase(ctx) {
			o.ObserveInt64(active, n, metric.WithAttributes(attribute.String("phase", phase)))
		}
		return nil
	}, active)
	if err != nil {
		return nil, err
	}

	return &taskActionMetrics{crdSizeBytes: crdSize, crdOpDuration: crdOp}, nil
}

// countByPhase tallies TaskActions by plugin phase; an empty phase maps to "Unknown".
// nil entries are skipped so callers can pass a raw cache listing without filtering.
func countByPhase(items []*flyteorgv1.TaskAction) map[string]int64 {
	counts := make(map[string]int64, 8)
	for _, ta := range items {
		if ta == nil {
			continue
		}
		phase := ta.Status.PluginPhase
		if phase == "" {
			phase = "Unknown"
		}
		counts[phase]++
	}
	return counts
}

// cachedPhaseCounter returns a function that counts TaskAction CRDs by plugin phase
// straight from the controller's informer cache indexer. Unlike client.List, the
// indexer's List returns the cached object pointers without deep-copying every CRD,
// so a collection cycle costs O(N) pointer reads instead of O(N) full-object copies —
// this is what keeps the active gauge cheap when many TaskActions exist. The indexer
// is resolved lazily on first call, because the cache is not yet started when the
// reconciler is constructed.
func cachedPhaseCounter(c ctrlcache.Cache) func(context.Context) map[string]int64 {
	var indexer toolscache.Indexer
	return func(ctx context.Context) map[string]int64 {
		if indexer == nil {
			if c == nil {
				return nil
			}
			// BlockUntilSynced(false): never stall the metric-collection goroutine; a
			// not-yet-synced cache just yields a partial count the next cycle corrects.
			informer, err := c.GetInformer(ctx, &flyteorgv1.TaskAction{}, ctrlcache.BlockUntilSynced(false))
			if err != nil {
				return nil
			}
			sii, ok := informer.(toolscache.SharedIndexInformer)
			if !ok {
				return nil
			}
			indexer = sii.GetIndexer()
		}
		objs := indexer.List()
		items := make([]*flyteorgv1.TaskAction, 0, len(objs))
		for _, obj := range objs {
			if ta, ok := obj.(*flyteorgv1.TaskAction); ok {
				items = append(items, ta)
			}
		}
		return countByPhase(items)
	}
}

// observeCRDSize records the serialized size of a TaskAction CRD. No-op when
// custom metrics are disabled (m == nil).
func (m *taskActionMetrics) observeCRDSize(ctx context.Context, ta *flyteorgv1.TaskAction) {
	if m == nil || m.crdSizeBytes == nil {
		return
	}
	if b, err := json.Marshal(ta); err == nil {
		m.crdSizeBytes.Record(ctx, int64(len(b)))
	}
}

// recordK8sOp records the latency of a Kubernetes API operation against the
// TaskAction CRD under taskaction.k8s.duration{op,error}. Call sites time the
// operation inline (start := time.Now() before the call) and record right after.
// No-op when custom metrics are disabled (m == nil), so callers can record
// unconditionally.
func (m *taskActionMetrics) recordK8sOp(ctx context.Context, op string, start time.Time, err error) {
	if m == nil || m.crdOpDuration == nil {
		return
	}
	m.crdOpDuration.Record(ctx, float64(time.Since(start).Microseconds())/1000.0,
		metric.WithAttributes(attribute.String("op", op), attribute.Bool("error", err != nil)))
}
