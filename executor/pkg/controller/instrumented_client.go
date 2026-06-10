package controller

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
)

// newInstrumentedClient wraps c so TaskAction CRD operations (Get, Update, and
// Status().Update) are timed under taskaction.k8s.duration. Operations on other
// object types pass through untimed. When metrics registration failed (m == nil)
// it returns c unchanged, so callers can wrap unconditionally.
//
// The reconciler embeds the wrapped client, which makes the instrumentation
// structural: call sites use the idiomatic r.Get/r.Update/r.Status().Update and
// cannot accidentally bypass the timing.
func newInstrumentedClient(c client.Client, m *taskActionMetrics) client.Client {
	if m == nil {
		return c
	}
	return &instrumentedClient{Client: c, metrics: m}
}

type instrumentedClient struct {
	client.Client
	metrics *taskActionMetrics
}

func (c *instrumentedClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if _, ok := obj.(*flyteorgv1.TaskAction); !ok {
		return c.Client.Get(ctx, key, obj, opts...)
	}
	return c.metrics.timeK8sOp(ctx, opGet, func() error { return c.Client.Get(ctx, key, obj, opts...) })
}

func (c *instrumentedClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if _, ok := obj.(*flyteorgv1.TaskAction); !ok {
		return c.Client.Update(ctx, obj, opts...)
	}
	return c.metrics.timeK8sOp(ctx, opUpdate, func() error { return c.Client.Update(ctx, obj, opts...) })
}

func (c *instrumentedClient) Status() client.SubResourceWriter {
	return &instrumentedStatusWriter{SubResourceWriter: c.Client.Status(), metrics: c.metrics}
}

type instrumentedStatusWriter struct {
	client.SubResourceWriter
	metrics *taskActionMetrics
}

func (w *instrumentedStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	if _, ok := obj.(*flyteorgv1.TaskAction); !ok {
		return w.SubResourceWriter.Update(ctx, obj, opts...)
	}
	return w.metrics.timeK8sOp(ctx, opStatusUpdate, func() error { return w.SubResourceWriter.Update(ctx, obj, opts...) })
}
