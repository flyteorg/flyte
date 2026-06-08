package controller

import (
	"context"
	"encoding/json"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"sigs.k8s.io/controller-runtime/pkg/client"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flytestdlib/otelutils"
)

const taskActionMeterName = "taskaction-controller"

// taskActionMetrics holds OTel instruments for the TaskAction controller.
//
// Reconcile latency and k8s API read/write latency already come from the
// controller-runtime metrics server (controller_runtime_reconcile_time_seconds,
// rest_client_request_duration_seconds), and event-proxy send latency comes from
// the otelconnect-wrapped events client (rpc_client_duration). This only adds what
// neither provides: TaskAction CRD count by phase and serialized CRD size.
type taskActionMetrics struct {
	crdSizeBytes metric.Int64Histogram
}

// registerTaskActionMetrics wires the TaskAction OTel meters onto the "executor"
// meter provider (registered in executor/setup.go). The active-by-phase gauge is
// observed asynchronously by listing TaskActions from the controller cache.
func registerTaskActionMetrics(c client.Client) (*taskActionMetrics, error) {
	meter := otelutils.GetMeterProvider("executor").Meter(taskActionMeterName)

	crdSize, err := meter.Int64Histogram(
		"taskaction.crd.size_bytes",
		metric.WithDescription("Serialized (JSON) size of a TaskAction CRD"),
		metric.WithUnit("By"),
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

	return &taskActionMetrics{crdSizeBytes: crdSize}, nil
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
