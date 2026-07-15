package controller

import (
	"context"
	"errors"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// batchTestClient records the size of every Record batch it receives.
type batchTestClient struct {
	mu      sync.Mutex
	batches []int
	total   int
	err     error
}

func (f *batchTestClient) Record(_ context.Context, req *connect.Request[workflow.RecordRequest]) (*connect.Response[workflow.RecordResponse], error) {
	f.mu.Lock()
	n := len(req.Msg.GetEvents())
	f.batches = append(f.batches, n)
	f.total += n
	err := f.err
	f.mu.Unlock()
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&workflow.RecordResponse{}), nil
}

func ev() *workflow.ActionEvent { return &workflow.ActionEvent{} }

// Every enqueued event is delivered to the client exactly once, no batch exceeds
// the cap, and concurrent callers coalesce into fewer-than-N batches.
func TestEventBatcher_DeliversAllAndCoalesces(t *testing.T) {
	fake := &batchTestClient{}
	b := newEventBatcher(fake, noop.NewTracerProvider())

	// n > 2x eventBatchMaxSize so the cap logic must split at least three batches —
	// otherwise the max-size assertion below never exercises a split.
	const n = 1000
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			assert.NoError(t, b.Record(context.Background(), ev()))
		}()
	}
	wg.Wait()

	fake.mu.Lock()
	defer fake.mu.Unlock()
	assert.Equal(t, n, fake.total, "every event must reach the client exactly once")
	assert.Less(t, len(fake.batches), n, "concurrent events should coalesce into fewer batches")
	assert.GreaterOrEqual(t, len(fake.batches), (n+eventBatchMaxSize-1)/eventBatchMaxSize,
		"the size cap must force splits once n exceeds eventBatchMaxSize")
	for _, size := range fake.batches {
		assert.LessOrEqual(t, size, eventBatchMaxSize, "no batch may exceed the cap")
	}
}

// A flush error is returned to the caller (so its reconcile requeues and re-emits).
func TestEventBatcher_PropagatesError(t *testing.T) {
	fake := &batchTestClient{err: errors.New("boom")}
	b := newEventBatcher(fake, noop.NewTracerProvider())
	err := b.Record(context.Background(), ev())
	require.Error(t, err)
}

// A caller whose context is cancelled unblocks instead of waiting on the batch.
func TestEventBatcher_RespectsContextCancel(t *testing.T) {
	fake := &batchTestClient{}
	b := newEventBatcher(fake, noop.NewTracerProvider())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := b.Record(ctx, ev())
	require.ErrorIs(t, err, context.Canceled)
}

// Every flush span must link (not parent) the contributing callers' spans, so
// fan-in event writes stay correlated with their reconcile traces.
func TestEventBatcher_FlushSpanLinksCallers(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	fake := &batchTestClient{}
	b := newEventBatcher(fake, tp)

	ctx, callerSpan := tp.Tracer("test").Start(context.Background(), "reconcile")
	require.NoError(t, b.Record(ctx, ev()))
	callerSpan.End()

	var flushSpans []sdktrace.ReadOnlySpan
	for _, s := range sr.Ended() {
		if s.Name() == "eventBatcher.flush" {
			flushSpans = append(flushSpans, s)
		}
	}
	require.Len(t, flushSpans, 1, "exactly one flush span for one batch")
	fs := flushSpans[0]
	assert.False(t, fs.Parent().IsValid(), "flush span must be a root, not parented under a caller")
	require.Len(t, fs.Links(), 1)
	assert.Equal(t, callerSpan.SpanContext().SpanID(), fs.Links()[0].SpanContext.SpanID(),
		"the caller's reconcile span must be attached as a link")
}
