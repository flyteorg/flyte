package service

import (
	"context"
	"fmt"
	"sync"

	"connectrpc.com/connect"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

// EventsWorker asynchronously forwards action events to InternalRunService.
type EventsWorker struct {
	runClient   workflowconnect.InternalRunServiceClient
	workerCount int
	queue       chan *workflow.ActionEvent

	mu      sync.Mutex
	wg      sync.WaitGroup
	started bool
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewEventsWorker creates a worker pool with a bounded in-memory queue.
func NewEventsWorker(
	runClient workflowconnect.InternalRunServiceClient,
	queueSize int,
	workerCount int,
) (*EventsWorker, error) {
	if runClient == nil {
		return nil, fmt.Errorf("run client is nil")
	}
	if queueSize <= 0 {
		return nil, fmt.Errorf("queue size must be > 0, got %d", queueSize)
	}
	if workerCount <= 0 {
		return nil, fmt.Errorf("worker count must be > 0, got %d", workerCount)
	}

	return &EventsWorker{
		runClient:   runClient,
		workerCount: workerCount,
		queue:       make(chan *workflow.ActionEvent, queueSize),
	}, nil
}

// Start launches background worker goroutines.
func (w *EventsWorker) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.started {
		logger.Warnf(ctx, "events worker already started")
		return nil
	}

	w.ctx, w.cancel = context.WithCancel(ctx)

	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go w.processEvents(w.ctx, i)
	}

	w.started = true
	logger.Infof(ctx, "events worker started with %d workers", w.workerCount)
	return nil
}

// processEvents handles one event per unary request to InternalRunService.
func (w *EventsWorker) processEvents(ctx context.Context, workerID int) {
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-w.queue:
			if !ok {
				// Queue closed by owner; worker should exit.
				return
			}
			if event == nil {
				logger.Warnf(ctx, "events worker[%d]: received nil action event", workerID)
				continue
			}

			recordEventReq := &workflow.RecordActionEventsRequest{
				Events: []*workflow.ActionEvent{event},
			}
			if _, err := w.runClient.RecordActionEvents(ctx, connect.NewRequest(recordEventReq)); err != nil {
				logger.Warnf(ctx, "events worker[%d]: failed to record action event for %s: %v", workerID, event.GetId().GetName(), err)
			}
		}
	}
}

// Enqueue attempts to add an event to the in-memory queue without blocking.
func (w *EventsWorker) Enqueue(event *workflow.ActionEvent) error {
	select {
	case w.queue <- event:
		return nil
	default:
		return fmt.Errorf("events queue is full")
	}
}

// End signals workers to stop and waits for all goroutines to exit.
func (w *EventsWorker) End() {
	w.mu.Lock()
	if !w.started || w.cancel == nil {
		w.mu.Unlock()
		return
	}

	cancel := w.cancel
	w.ctx, w.cancel = nil, nil
	w.mu.Unlock()

	// Send cancel signal and wait until all workers exit.
	cancel()
	w.wg.Wait()

	w.mu.Lock()
	w.started = false
	w.mu.Unlock()
}
