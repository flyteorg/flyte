package controller

import (
	"context"
	"fmt"
	"runtime/pprof"
	"time"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

type Handler interface {
	// Initialize the Handler
	Initialize(ctx context.Context) error
	// Handle method that should handle the object and try to converge the desired and the actual state
	Handle(ctx context.Context, namespace, key string) error
}

type workerPoolMetrics struct {
	Scope            promutils.Scope
	FreeWorkers      prometheus.Gauge
	PerRoundTimer    promutils.StopWatch
	RoundError       prometheus.Counter
	RoundSuccess     prometheus.Counter
	WorkersRestarted prometheus.Counter
}

type WorkerPool struct {
	workQueue CompositeWorkQueue
	metrics   workerPoolMetrics
	handler   Handler
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the handler.
func (w *WorkerPool) processNextWorkItem(ctx context.Context) bool {
	obj, shutdown := w.workQueue.Get()

	w.metrics.FreeWorkers.Dec()
	defer w.metrics.FreeWorkers.Inc()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer w.workQueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			w.workQueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		t := w.metrics.PerRoundTimer.Start()
		defer t.Stop()

		// Convert the namespace/name string into a distinct namespace and name
		namespace, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			logger.Errorf(ctx, "Unable to split enqueued key into namespace/execId. Error[%v]", err)
			return nil
		}
		ctx = contextutils.WithNamespace(ctx, namespace)
		ctx = contextutils.WithExecutionID(ctx, name)
		// Reconcile the Workflow
		if err := w.handler.Handle(ctx, namespace, name); err != nil {
			w.metrics.RoundError.Inc()
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		w.metrics.RoundSuccess.Inc()

		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		w.workQueue.Forget(obj)
		logger.Infof(ctx, "Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (w *WorkerPool) runWorker(ctx context.Context) {
	logger.Infof(ctx, "Started Worker")
	defer logger.Infof(ctx, "Exiting Worker")
	for w.processNextWorkItem(ctx) {
	}
}

func (w *WorkerPool) Initialize(ctx context.Context) error {
	return w.handler.Initialize(ctx)
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (w *WorkerPool) Run(ctx context.Context, threadiness int, synced ...cache.InformerSynced) error {
	defer runtime.HandleCrash()
	defer w.workQueue.ShutdownAll()

	// Start the informer factories to begin populating the informer caches
	logger.Info(ctx, "Starting FlyteWorkflow controller")
	w.metrics.WorkersRestarted.Inc()

	// Wait for the caches to be synced before starting workers
	logger.Info(ctx, "Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(ctx.Done(), synced...); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	logger.Infof(ctx, "Starting workers [%d]", threadiness)
	// Launch workers to process FlyteWorkflow resources
	for i := 0; i < threadiness; i++ {
		w.metrics.FreeWorkers.Inc()
		logger.Infof(ctx, "Starting worker [%d]", i)
		workerLabel := fmt.Sprintf("worker-%v", i)
		go func() {
			workerCtx := contextutils.WithGoroutineLabel(ctx, workerLabel)
			pprof.SetGoroutineLabels(workerCtx)
			w.runWorker(workerCtx)
		}()
	}

	w.workQueue.Start(ctx)
	logger.Info(ctx, "Started workers")
	<-ctx.Done()
	logger.Info(ctx, "Shutting down workers")

	return nil
}

func NewWorkerPool(ctx context.Context, scope promutils.Scope, workQueue CompositeWorkQueue, handler Handler) *WorkerPool {
	roundScope := scope.NewSubScope("round")
	metrics := workerPoolMetrics{
		Scope:            scope,
		FreeWorkers:      scope.MustNewGauge("free_workers_count", "Number of workers free"),
		PerRoundTimer:    roundScope.MustNewStopWatch("round_total", "Latency per round", time.Millisecond),
		RoundSuccess:     roundScope.MustNewCounter("success_count", "Round succeeded"),
		RoundError:       roundScope.MustNewCounter("error_count", "Round failed"),
		WorkersRestarted: scope.MustNewCounter("workers_restarted", "Propeller worker-pool was restarted"),
	}
	return &WorkerPool{
		workQueue: workQueue,
		metrics:   metrics,
		handler:   handler,
	}
}
