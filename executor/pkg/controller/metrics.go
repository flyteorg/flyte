package controller

import (
	"context"
	"encoding/json"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
// gauge is observed asynchronously by listing TaskActions from the controller cache.
func registerTaskActionMetrics(provider metric.MeterProvider, c client.Client) (*taskActionMetrics, error) {
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
		var list flyteorgv1.TaskActionList
		if err := c.List(ctx, &list); err != nil {
			return err
		}
		counts := make(map[string]int64, 8)
		for i := range list.Items {
			phase := list.Items[i].Status.PluginPhase
			if phase == "" {
				phase = "Unknown"
			}
			counts[phase]++
		}
		for phase, n := range counts {
			o.ObserveInt64(active, n, metric.WithAttributes(attribute.String("phase", phase)))
		}
		return nil
	}, active)
	if err != nil {
		return nil, err
	}

	return &taskActionMetrics{crdSizeBytes: crdSize, crdOpDuration: crdOp}, nil
}

// observeCRDSize records the serialized size of a TaskAction CRD. No-op if metrics
// registration failed (m == nil).
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
// No-op when metrics registration failed (m == nil), so callers can record
// unconditionally.
func (m *taskActionMetrics) recordK8sOp(ctx context.Context, op string, start time.Time, err error) {
	if m == nil || m.crdOpDuration == nil {
		return
	}
	m.crdOpDuration.Record(ctx, float64(time.Since(start).Microseconds())/1000.0,
		metric.WithAttributes(attribute.String("op", op), attribute.Bool("error", err != nil)))
}
