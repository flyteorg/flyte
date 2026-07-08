package controller

import (
	"context"
	"errors"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func ev(name string) *workflow.ActionEvent { return &workflow.ActionEvent{} }

// Every enqueued event is delivered to the client exactly once, no batch exceeds
// the cap, and concurrent callers coalesce into fewer-than-N batches.
func TestEventBatcher_DeliversAllAndCoalesces(t *testing.T) {
	fake := &batchTestClient{}
	b := newEventBatcher(fake)

	const n = 300
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			assert.NoError(t, b.Record(context.Background(), ev("e")))
		}()
	}
	wg.Wait()

	fake.mu.Lock()
	defer fake.mu.Unlock()
	assert.Equal(t, n, fake.total, "every event must reach the client exactly once")
	assert.Less(t, len(fake.batches), n, "concurrent events should coalesce into fewer batches")
	for _, size := range fake.batches {
		assert.LessOrEqual(t, size, eventBatchMaxSize, "no batch may exceed the cap")
	}
}

// A flush error is returned to the caller (so its reconcile requeues and re-emits).
func TestEventBatcher_PropagatesError(t *testing.T) {
	fake := &batchTestClient{err: errors.New("boom")}
	b := newEventBatcher(fake)
	err := b.Record(context.Background(), ev("e"))
	require.Error(t, err)
}

// A caller whose context is cancelled unblocks instead of waiting on the batch.
func TestEventBatcher_RespectsContextCancel(t *testing.T) {
	fake := &batchTestClient{}
	b := newEventBatcher(fake)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := b.Record(ctx, ev("e"))
	require.ErrorIs(t, err, context.Canceled)
}
