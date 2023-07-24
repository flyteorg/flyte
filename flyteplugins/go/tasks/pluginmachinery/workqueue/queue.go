package workqueue

import (
	"context"
	"fmt"
	"sync"

	"github.com/flyteorg/flytestdlib/errors"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/flyteorg/flytestdlib/contextutils"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"

	lru "github.com/hashicorp/golang-lru"

	"k8s.io/client-go/util/workqueue"
)

//go:generate mockery -all -case=underscore
//go:generate enumer --type=WorkStatus

type WorkItemID = string
type WorkStatus uint8

const (
	WorkStatusNotDone WorkStatus = iota
	WorkStatusSucceeded
	WorkStatusFailed
)

const (
	ErrNotYetStarted errors.ErrorCode = "NOT_STARTED"
)

func (w WorkStatus) IsTerminal() bool {
	return w == WorkStatusFailed || w == WorkStatusSucceeded
}

// WorkItem is a generic item that can be stored in the work queue.
type WorkItem interface{}

// Represents the result of the work item processing.
type WorkItemInfo interface {
	Item() WorkItem
	ID() WorkItemID
	Status() WorkStatus
	Error() error
}

// Represents the indexed queue semantics. An indexed work queue is a work queue that additionally keeps track of the
// final processing results of work items.
type IndexedWorkQueue interface {
	// Queues the item to be processed. If the item is already in the cache or has been processed before (and is still
	// in-memory), it'll not be added again.
	Queue(ctx context.Context, id WorkItemID, once WorkItem) error

	// Retrieves an item by id.
	Get(id WorkItemID) (info WorkItemInfo, found bool, err error)

	// Start must be called before queuing items into the queue.
	Start(ctx context.Context) error
}

// Represents the processor logic to operate on work items.
type Processor interface {
	Process(ctx context.Context, workItem WorkItem) (WorkStatus, error)
}

type workItemWrapper struct {
	id         WorkItemID
	logFields  map[string]interface{}
	payload    WorkItem
	status     WorkStatus
	retryCount uint
	err        error
}

func (w workItemWrapper) Item() WorkItem {
	return w.payload
}

func (w workItemWrapper) ID() WorkItemID {
	return w.id
}

func (w workItemWrapper) Status() WorkStatus {
	return w.status
}

func (w workItemWrapper) Error() error {
	return w.err
}

func (w workItemWrapper) Clone() workItemWrapper {
	return w
}

type metrics struct {
	CacheHit        prometheus.Counter
	CacheMiss       prometheus.Counter
	ProcessorErrors prometheus.Counter
	Scope           promutils.Scope
}

type queue struct {
	name       string
	metrics    metrics
	wlock      sync.Mutex
	rlock      sync.RWMutex
	workers    int
	maxRetries int
	started    bool
	queue      workqueue.Interface
	index      workItemCache
	processor  Processor
}

type workItemCache struct {
	*lru.Cache
}

func (c workItemCache) Get(id WorkItemID) (item *workItemWrapper, found bool) {
	o, found := c.Cache.Get(id)
	if !found {
		return nil, found
	}

	return o.(*workItemWrapper), true
}

func (c workItemCache) Add(item *workItemWrapper) (evicted bool) {
	return c.Cache.Add(item.id, item)
}

func copyAllowedLogFields(ctx context.Context) map[string]interface{} {
	logFields := contextutils.GetLogFields(ctx)
	delete(logFields, contextutils.RoutineLabelKey.String())
	return logFields
}

func (q *queue) Queue(ctx context.Context, id WorkItemID, once WorkItem) error {
	q.wlock.Lock()
	defer q.wlock.Unlock()

	if !q.started {
		return errors.Errorf(ErrNotYetStarted, "Queue must be started before enqueuing any item.")
	}

	if _, found := q.index.Get(id); found {
		return nil
	}

	wrapper := &workItemWrapper{
		id:        id,
		logFields: copyAllowedLogFields(ctx),
		payload:   once,
	}

	q.index.Add(wrapper)
	q.queue.Add(wrapper)
	return nil
}

func (q *queue) Get(id WorkItemID) (info WorkItemInfo, found bool, err error) {
	q.rlock.Lock()
	defer q.rlock.Unlock()

	wrapper, found := q.index.Get(id)
	if !found {
		q.metrics.CacheMiss.Inc()
		return nil, found, nil
	}

	v := wrapper.Clone()
	q.metrics.CacheHit.Inc()
	return &v, true, nil
}

func contextWithValues(ctx context.Context, fields map[string]interface{}) context.Context {
	for key, value := range fields {
		ctx = context.WithValue(ctx, contextutils.Key(key), value)
	}

	return ctx
}

func (q *queue) Start(ctx context.Context) error {
	q.wlock.Lock()
	defer q.wlock.Unlock()

	if q.started {
		return fmt.Errorf("queue already started")
	}

	for i := 0; i < q.workers; i++ {
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					logger.Debug(ctx, "Context cancelled. Shutting down.")
					return
				default:
					item, shutdown := q.queue.Get()
					if shutdown {
						logger.Debug(ctx, "Work queue is shutting down.")
						return
					}

					wrapperV := item.(*workItemWrapper).Clone()
					wrapper := &wrapperV
					ws := wrapper.status
					var err error

					func() {
						defer func() {
							if e, ok := recover().(error); ok {
								logger.Errorf(ctx, "Worker panic'd while processing item [%v]. Error: %v", wrapper.id, e)
								err = e
							}
						}()

						ctxWithFields := contextWithValues(ctx, wrapper.logFields)
						ws, err = q.processor.Process(ctxWithFields, wrapper.payload)
					}()

					if err != nil {
						q.metrics.ProcessorErrors.Inc()

						wrapper.retryCount++
						wrapper.err = err
						if wrapper.retryCount >= uint(q.maxRetries) {
							logger.Debugf(ctx, "WorkItem [%v] exhausted all retries. Last Error: %v.",
								wrapper.ID(), err)
							wrapper.status = WorkStatusFailed
							ws = WorkStatusFailed
							q.index.Add(wrapper)
							continue
						}
					}

					wrapper.status = ws
					q.index.Add(wrapper)
					if !ws.IsTerminal() {
						q.queue.Add(wrapper)
					}
				}
			}
		}(contextutils.WithGoroutineLabel(ctx, fmt.Sprintf("%v-worker-%v", q.name, i)))
	}

	q.started = true
	return nil
}

func newMetrics(scope promutils.Scope) metrics {
	return metrics{
		CacheHit:        scope.MustNewCounter("cache_hit", "Counter for cache hits."),
		CacheMiss:       scope.MustNewCounter("cache_miss", "Counter for cache misses."),
		ProcessorErrors: scope.MustNewCounter("proc_errors", "Counter for processor errors."),
		Scope:           scope,
	}
}

// Instantiates a new Indexed Work queue.
func NewIndexedWorkQueue(name string, processor Processor, cfg Config, metricsScope promutils.Scope) (IndexedWorkQueue, error) {
	cache, err := lru.New(cfg.IndexCacheMaxItems)
	if err != nil {
		return nil, err
	}

	return &queue{
		name:       name,
		metrics:    newMetrics(metricsScope),
		wlock:      sync.Mutex{},
		rlock:      sync.RWMutex{},
		workers:    cfg.Workers,
		maxRetries: cfg.MaxRetries,
		queue:      workqueue.NewNamed(metricsScope.CurrentScope()),
		index:      workItemCache{Cache: cache},
		processor:  processor,
	}, nil
}
