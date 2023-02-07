package controller

import (
	"context"
	"time"

	"github.com/flyteorg/flytepropeller/pkg/controller/config"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
)

// A CompositeWorkQueue can be used in cases where the work is enqueued by two sources. It can be enqueued by either
//  1. Informer for the Primary Object itself. In case of FlytePropeller, this is the workflow object
//  2. Informer or any other process that enqueues the top-level object for re-evaluation in response to one of the
//     sub-objects being ready. In the case of FlytePropeller this is the "Node/Task" updates, will re-enqueue the workflow
//     to be re-evaluated
type CompositeWorkQueue interface {
	workqueue.RateLimitingInterface
	// Specialized interface that should be called to start the migration of work from SubQueue to primaryQueue
	Start(ctx context.Context)
	// Shutsdown all the queues that are in the context
	ShutdownAll()
	// Adds the item explicitly to the subqueue
	AddToSubQueue(item interface{})
	// Adds the item explicitly to the subqueue, using a rate limiter
	AddToSubQueueRateLimited(item interface{})
	// Adds the item explicitly to the subqueue after some duration
	AddToSubQueueAfter(item interface{}, duration time.Duration)
}

// SimpleWorkQueue provides a simple RateLimitingInterface, but ensures that the compositeQueue interface works
// with a default queue.
type SimpleWorkQueue struct {
	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue.RateLimitingInterface
}

func (s *SimpleWorkQueue) Start(ctx context.Context) {
}

func (s *SimpleWorkQueue) ShutdownAll() {
	s.ShutDown()
}

func (s *SimpleWorkQueue) AddToSubQueue(item interface{}) {
	s.Add(item)
}

func (s *SimpleWorkQueue) AddToSubQueueAfter(item interface{}, duration time.Duration) {
	s.AddAfter(item, duration)
}

func (s *SimpleWorkQueue) AddToSubQueueRateLimited(item interface{}) {
	s.AddRateLimited(item)
}

// A BatchingWorkQueue consists of 2 queues and migrates items from sub-queue to parent queue as a batch at a specified
// interval
type BatchingWorkQueue struct {
	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue.RateLimitingInterface

	subQueue         workqueue.RateLimitingInterface
	batchingInterval time.Duration
	batchSize        int
}

func (b *BatchingWorkQueue) Start(ctx context.Context) {
	logger.Infof(ctx, "Batching queue started")
	go wait.Until(func() {
		b.runSubQueueHandler(ctx)
	}, b.batchingInterval, ctx.Done())
}

func (b *BatchingWorkQueue) runSubQueueHandler(ctx context.Context) {
	logger.Debugf(ctx, "Subqueue handler batch round")
	defer logger.Debugf(ctx, "Exiting SubQueue handler batch round")
	if b.subQueue.ShuttingDown() {
		return
	}
	numToRetrieve := b.batchSize
	if b.batchSize == -1 || b.batchSize > b.subQueue.Len() {
		numToRetrieve = b.subQueue.Len()
	}

	logger.Debugf(ctx, "Dynamically configured batch size [%d]", b.batchSize)
	// Run batches forever
	objectsRetrieved := make([]interface{}, numToRetrieve)
	for i := 0; i < numToRetrieve; i++ {
		obj, shutdown := b.subQueue.Get()
		if obj != nil {
			// We expect strings to come off the workqueue. These are of the
			// form namespace/name. We do this as the delayed nature of the
			// workqueue means the items in the informer cache may actually be
			// more up to date that when the item was initially put onto the
			// workqueue.
			if key, ok := obj.(string); ok {
				objectsRetrieved[i] = key
			}
		}
		if shutdown {
			logger.Warningf(ctx, "NodeQ shutdown invoked. Shutting down poller.")
			// We cannot add after shutdown, so just quit!
			return
		}

	}

	for _, obj := range objectsRetrieved {
		b.AddRateLimited(obj)
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		b.subQueue.Forget(obj)
		b.subQueue.Done(obj)
	}

}

func (b *BatchingWorkQueue) ShutdownAll() {
	b.subQueue.ShutDown()
	b.ShutDown()
}

func (b *BatchingWorkQueue) AddToSubQueue(item interface{}) {
	b.subQueue.Add(item)
}

func (b *BatchingWorkQueue) AddToSubQueueAfter(item interface{}, duration time.Duration) {
	b.subQueue.AddAfter(item, duration)
}

func (b *BatchingWorkQueue) AddToSubQueueRateLimited(item interface{}) {
	b.subQueue.AddRateLimited(item)
}

func NewCompositeWorkQueue(ctx context.Context, cfg config.CompositeQueueConfig, scope promutils.Scope) (CompositeWorkQueue, error) {
	workQ, err := NewWorkQueue(ctx, cfg.Queue, scope.NewScopedMetricName("main"))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create WorkQueue in CompositeQueue type Batch")
	}
	switch cfg.Type {
	case config.CompositeQueueBatch:
		subQ, err := NewWorkQueue(ctx, cfg.Sub, scope.NewScopedMetricName("sub"))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create SubQueue in CompositeQueue type Batch")
		}
		return &BatchingWorkQueue{
			RateLimitingInterface: workQ,
			batchSize:             cfg.BatchSize,
			batchingInterval:      cfg.BatchingInterval.Duration,
			subQueue:              subQ,
		}, nil
	case config.CompositeQueueSimple:
		fallthrough
	default:
	}
	return &SimpleWorkQueue{
		RateLimitingInterface: workQ,
	}, nil
}
