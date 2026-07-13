package controller

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

// Event batching bounds. Coalescing concurrent reconciles' action events into a
// few multi-event Record RPCs (each becoming one multi-row INSERT server-side)
// amortizes the per-event DB commit cost that otherwise dominates reconcile
// latency at high held-action counts. Callers still block until their batch
// commits, so delivery semantics are unchanged versus a direct Record.
const (
	eventBatchMaxSize  = 400
	eventBatchMaxDelay = 25 * time.Millisecond
	eventFlushWorkers  = 32
	eventQueueDepth    = 8192
	eventFlushTimeout  = 30 * time.Second
)

// eventBatcher coalesces ActionEvents from concurrent reconciles into batched
// Record RPCs. Record blocks until the event's batch has been recorded (or
// errored); on error the caller's reconcile requeues and re-emits, so this
// preserves the synchronous at-least-once behaviour of a direct Record while
// collapsing thousands of one-row commits into a few multi-row ones. DB write
// concurrency is bounded to eventFlushWorkers regardless of reconcile fan-out.
type eventBatcher struct {
	client workflowconnect.EventsProxyServiceClient
	tracer trace.Tracer
	queue  chan *eventReq
}

type eventReq struct {
	event   *workflow.ActionEvent
	spanCtx trace.SpanContext
	done    chan error
}

func newEventBatcher(client workflowconnect.EventsProxyServiceClient, tp trace.TracerProvider) *eventBatcher {
	b := &eventBatcher{
		client: client,
		tracer: tp.Tracer("executor/event_batcher"),
		queue:  make(chan *eventReq, eventQueueDepth),
	}
	go b.collect()
	return b
}

// Record enqueues event and blocks until its batch is flushed or ctx is done.
func (b *eventBatcher) Record(ctx context.Context, event *workflow.ActionEvent) error {
	req := &eventReq{event: event, spanCtx: trace.SpanContextFromContext(ctx), done: make(chan error, 1)}
	select {
	case b.queue <- req:
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case err := <-req.done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// collect coalesces queued events into batches (bounded by size and delay) and
// hands them to a fixed pool of flush workers, capping DB write concurrency.
func (b *eventBatcher) collect() {
	batches := make(chan []*eventReq, eventFlushWorkers)
	for i := 0; i < eventFlushWorkers; i++ {
		go func() {
			for batch := range batches {
				b.flush(batch)
			}
		}()
	}
	for {
		first := <-b.queue
		batch := []*eventReq{first}
		timer := time.NewTimer(eventBatchMaxDelay)
	fill:
		for len(batch) < eventBatchMaxSize {
			select {
			case req := <-b.queue:
				batch = append(batch, req)
			case <-timer.C:
				break fill
			}
		}
		timer.Stop()
		batches <- batch
	}
}

// flush records the whole batch in one RPC (one multi-row INSERT server-side)
// and unblocks every caller with the shared result.
func (b *eventBatcher) flush(batch []*eventReq) {
	events := make([]*workflow.ActionEvent, len(batch))
	links := make([]trace.Link, 0, len(batch))
	seen := make(map[trace.SpanID]struct{}, len(batch))
	for i, r := range batch {
		events[i] = r.event
		if r.spanCtx.IsValid() {
			if _, dup := seen[r.spanCtx.SpanID()]; !dup {
				seen[r.spanCtx.SpanID()] = struct{}{}
				links = append(links, trace.Link{SpanContext: r.spanCtx})
			}
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), eventFlushTimeout)
	defer cancel()
	ctx, span := b.tracer.Start(ctx, "eventBatcher.flush",
		trace.WithLinks(links...),
		trace.WithAttributes(attribute.Int("events.batch_size", len(batch))))
	defer span.End()
	_, err := b.client.Record(ctx, connect.NewRequest(&workflow.RecordRequest{
		Events: events,
	}))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
	}
	for _, r := range batch {
		r.done <- err
	}
}
