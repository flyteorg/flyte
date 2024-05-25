package plugin

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"

	"github.com/unionai/flyte/fasttask/plugin/pb"
)

var maxPendingOwnersPerQueue = 100

//go:generate mockery -all -case=underscore

// FastTaskService defines the interface for managing assignment and management of task executions
type FastTaskService interface {
	CheckStatus(ctx context.Context, taskID, queueID, workerID string) (core.Phase, string, error)
	Cleanup(ctx context.Context, taskID, queueID, workerID string) error
	OfferOnQueue(ctx context.Context, queueID, taskID, namespace, workflowID string, cmd []string) (string, error)
}

// fastTaskServiceImpl is a gRPC service that manages assignment and management of task executions
// with respect to fasttask workers.
type fastTaskServiceImpl struct {
	pb.UnimplementedFastTaskServer
	enqueueOwner core.EnqueueOwner

	queues     map[string]*Queue
	queuesLock sync.RWMutex

	// A map of pending owners by queue. When a new worker becomes available, use this to enqueue owners for reevaluation.
	// Note, this is an optimistic approach and may not include all pending owners.
	pendingTaskOwners     map[string]map[string]types.NamespacedName // map[queueID]map[taskID]ownerID
	pendingTaskOwnersLock sync.RWMutex

	taskStatusChannels sync.Map // map[taskID]chan *WorkerTaskStatus
	metrics            metrics
}

// Queue is a collection of Workers that are capable of executing similar tasks.
type Queue struct {
	lock    sync.RWMutex
	workers map[string]*Worker
}

// Worker represents a fasttask worker.
type Worker struct {
	workerID     string
	capacity     *pb.Capacity
	responseChan chan<- *pb.HeartbeatResponse
}

// workerTaskStatus represents the status of a task as reported by a worker.
type workerTaskStatus struct {
	workerID   string
	taskStatus *pb.TaskStatus
}

type metrics struct {
	taskNoWorkersAvailable  prometheus.Counter
	taskNoCapacityAvailable prometheus.Counter
	taskAssigned            prometheus.Counter

	queues  *prometheus.Desc
	workers *prometheus.Desc
}

func newMetrics(scope promutils.Scope) metrics {
	return metrics{
		taskNoWorkersAvailable:  scope.MustNewCounter("task_no_workers_available", "Count of task assignment attempts with no workers available"),
		taskNoCapacityAvailable: scope.MustNewCounter("task_no_capacity_available", "Count of task assignment attempts with no capacity available"),
		taskAssigned:            scope.MustNewCounter("task_assigned", "Count of task assignments"),
		queues:                  prometheus.NewDesc(scope.NewScopedMetricName("queue"), "Current number of queues", nil, nil),
		workers:                 prometheus.NewDesc(scope.NewScopedMetricName("workers"), "Current number of workers", nil, nil),
	}
}

func (f *fastTaskServiceImpl) Describe(ch chan<- *prometheus.Desc) {
	ch <- f.metrics.queues
	ch <- f.metrics.workers
}

func (f *fastTaskServiceImpl) Collect(ch chan<- prometheus.Metric) {
	f.queuesLock.RLock()
	defer f.queuesLock.RUnlock()

	queues := len(f.queues)
	workers := 0
	for _, queue := range f.queues {
		queue.lock.RLock()
		workers += len(queue.workers)
		queue.lock.RUnlock()
	}

	ch <- prometheus.MustNewConstMetric(f.metrics.queues, prometheus.GaugeValue, float64(queues))
	ch <- prometheus.MustNewConstMetric(f.metrics.workers, prometheus.GaugeValue, float64(workers))
}

// Heartbeat is a gRPC stream that manages the heartbeat of a fasttask worker. This includes
// receiving task status updates and sending task assignments.
func (f *fastTaskServiceImpl) Heartbeat(stream pb.FastTask_HeartbeatServer) error {
	workerID := ""

	// recv initial heartbeat request
	heartbeatRequest, err := stream.Recv()
	if heartbeatRequest != nil {
		workerID = heartbeatRequest.GetWorkerId()
	}

	if err == io.EOF || heartbeatRequest == nil {
		logger.Debugf(context.Background(), "heartbeat stream closed for worker %s", workerID)
		return nil
	} else if err != nil {
		return err
	}

	logger.Debugf(context.Background(), "received initial heartbeat for worker %s", workerID)

	// create worker
	responseChan := make(chan *pb.HeartbeatResponse, GetConfig().HeartbeatBufferSize)
	worker := &Worker{
		workerID:     workerID,
		capacity:     heartbeatRequest.GetCapacity(),
		responseChan: responseChan,
	}

	// register worker with queue
	queue := f.addWorkerToQueue(heartbeatRequest.GetQueueId(), worker)

	// cleanup worker on exit
	defer func() {
		f.removeWorkerFromQueue(heartbeatRequest.GetQueueId(), workerID)
	}()

	// start go routine to handle heartbeat responses
	go func() {
		for {
			select {
			case message := <-responseChan:
				if err := stream.Send(message); err != nil {
					logger.Warnf(context.Background(), "failed to send heartbeat response %+v", message)
				}
			case <-stream.Context().Done():
				return
			}
		}
	}()

	// new worker available, enqueue owners
	f.enqueuePendingOwners(heartbeatRequest.GetQueueId())

	// handle heartbeat requests
	for {
		heartbeatRequest, err := stream.Recv()
		if err == io.EOF || heartbeatRequest == nil {
			logger.Debugf(context.Background(), "heartbeat stream closed for worker %s", workerID)
			break
		} else if err != nil {
			logger.Warnf(context.Background(), "failed to recv heartbeat request %+v", err)
			continue
		}

		// update worker capacity
		queue.lock.Lock()
		worker.capacity = heartbeatRequest.GetCapacity()
		queue.lock.Unlock()

		for _, taskStatus := range heartbeatRequest.GetTaskStatuses() {
			// if the taskContext exists then send the taskStatus to the statusChannel
			// if it does not exist, then this plugin has restarted and we rely on the `CheckStatus` to create a new TaskContext.
			// this is because if `CheckStatus` is called, then the task is active and will be cleaned up on completion. If we
			// created it here, then a worker could be reporting a status for a task that has already completed and the TaskContext
			// cleanup would require a separate GC process.
			if taskStatusChannelResult, exists := f.taskStatusChannels.Load(taskStatus.GetTaskId()); exists {
				taskStatusChannel := taskStatusChannelResult.(chan *workerTaskStatus)
				taskStatusChannel <- &workerTaskStatus{
					workerID:   worker.workerID,
					taskStatus: taskStatus,
				}
			}

			// if taskStatus is complete then enqueueOwner for fast feedback
			phase := core.Phase(taskStatus.GetPhase())
			if phase == core.PhaseSuccess || phase == core.PhaseRetryableFailure {
				if err := f.enqueueOwner(types.NamespacedName{
					Namespace: taskStatus.GetNamespace(),
					Name:      taskStatus.GetWorkflowId(),
				}); err != nil {
					logger.Warnf(context.Background(), "failed to enqueue owner for task %s: %+v", taskStatus.GetTaskId(), err)
				}
			}
		}
	}

	return nil
}

// addWorkerToQueue adds a worker to the queue. If the queue does not exist, it is created.
func (f *fastTaskServiceImpl) addWorkerToQueue(queueID string, worker *Worker) *Queue {
	f.queuesLock.Lock()
	defer f.queuesLock.Unlock()

	queue, exists := f.queues[queueID]
	if !exists {
		queue = &Queue{
			workers: make(map[string]*Worker),
		}
		f.queues[queueID] = queue
	}

	queue.lock.Lock()
	defer queue.lock.Unlock()

	queue.workers[worker.workerID] = worker
	return queue
}

// removeWorkerFromQueue removes a worker from the queue. If the queue is empty, it is deleted.
func (f *fastTaskServiceImpl) removeWorkerFromQueue(queueID, workerID string) {
	f.queuesLock.Lock()
	defer f.queuesLock.Unlock()

	queue, exists := f.queues[queueID]
	if !exists {
		return
	}

	queue.lock.Lock()
	defer queue.lock.Unlock()

	delete(queue.workers, workerID)
	if len(queue.workers) == 0 {
		delete(f.queues, queueID)
	}
}

// addPendingOwner adds to the pending owners list for the queue, if not already full
func (f *fastTaskServiceImpl) addPendingOwner(queueID, taskID string, ownerID types.NamespacedName) {
	f.pendingTaskOwnersLock.Lock()
	defer f.pendingTaskOwnersLock.Unlock()

	owners, exists := f.pendingTaskOwners[queueID]
	if !exists {
		owners = make(map[string]types.NamespacedName)
		f.pendingTaskOwners[queueID] = owners
	}

	if len(owners) >= maxPendingOwnersPerQueue {
		return
	}
	owners[taskID] = ownerID
}

// removePendingOwner removes the pending owner from the list if still there
func (f *fastTaskServiceImpl) removePendingOwner(queueID, taskID string) {
	f.pendingTaskOwnersLock.Lock()
	defer f.pendingTaskOwnersLock.Unlock()

	owners, exists := f.pendingTaskOwners[queueID]
	if !exists {
		return
	}

	delete(owners, taskID)
	if len(owners) == 0 {
		delete(f.pendingTaskOwners, queueID)
	}
}

// enqueuePendingOwners drains the pending owners list for the queue and enqueues them for reevaluation
func (f *fastTaskServiceImpl) enqueuePendingOwners(queueID string) {
	f.pendingTaskOwnersLock.Lock()
	defer f.pendingTaskOwnersLock.Unlock()

	owners, exists := f.pendingTaskOwners[queueID]
	if !exists {
		return
	}

	enqueued := make(map[types.NamespacedName]bool)
	for _, ownerID := range owners {
		if _, ok := enqueued[ownerID]; ok {
			continue
		}
		if err := f.enqueueOwner(ownerID); err != nil {
			logger.Warnf(context.Background(), "failed to enqueue owner %s: %+v", ownerID, err)
		}
		enqueued[ownerID] = true
	}

	delete(f.pendingTaskOwners, queueID)
}

// OfferOnQueue offers a task to a worker on a specific queue. If no workers are available, an
// empty string is returned.
func (f *fastTaskServiceImpl) OfferOnQueue(ctx context.Context, queueID, taskID, namespace, workflowID string, cmd []string) (string, error) {
	f.queuesLock.RLock()
	defer f.queuesLock.RUnlock()

	queue, exists := f.queues[queueID]
	if !exists {
		f.addPendingOwner(queueID, taskID, types.NamespacedName{Namespace: namespace, Name: workflowID})
		f.metrics.taskNoWorkersAvailable.Inc()
		return "", nil // no workers available
	}

	// retrieve random worker with capacity
	queue.lock.Lock()
	defer queue.lock.Unlock()

	preferredWorkers := make([]*Worker, 0)
	acceptedWorkers := make([]*Worker, 0)
	for _, worker := range queue.workers {
		if worker.capacity.GetExecutionLimit()-worker.capacity.GetExecutionCount() > 0 {
			preferredWorkers = append(preferredWorkers, worker)
		} else if worker.capacity.GetBacklogLimit()-worker.capacity.GetBacklogCount() > 0 {
			acceptedWorkers = append(acceptedWorkers, worker)
		}
	}

	var worker *Worker
	if len(preferredWorkers) > 0 {
		worker = preferredWorkers[rand.Intn(len(preferredWorkers))]
		worker.capacity.ExecutionCount++
	} else if len(acceptedWorkers) > 0 {
		worker = acceptedWorkers[rand.Intn(len(acceptedWorkers))]
		worker.capacity.BacklogCount++
	} else {
		// No workers available. Note, we do not add to pending owners at this time as we are optimizing for the worker
		// startup case. The worker backlog should be sufficient to keep the worker busy without needing to proactively
		// enqueue owners when capacity becomes available.
		f.metrics.taskNoCapacityAvailable.Inc()
		return "", nil
	}

	// send assign message to worker
	f.metrics.taskAssigned.Inc()
	worker.responseChan <- &pb.HeartbeatResponse{
		TaskId:     taskID,
		Namespace:  namespace,
		WorkflowId: workflowID,
		Cmd:        cmd,
		Operation:  pb.HeartbeatResponse_ASSIGN,
	}

	// create task status channel
	f.taskStatusChannels.Store(taskID, make(chan *workerTaskStatus, GetConfig().TaskStatusBufferSize))
	return worker.workerID, nil
}

// CheckStatus checks the status of a task on a specific queue and worker.
func (f *fastTaskServiceImpl) CheckStatus(ctx context.Context, taskID, queueID, workerID string) (core.Phase, string, error) {
	taskStatusChannelResult, exists := f.taskStatusChannels.Load(taskID)
	if !exists {
		// if this plugin restarts then TaskContexts may not exist for tasks that are still active. we can
		// create a TaskContext here because we ensure it will be cleaned up when the task completes.
		f.taskStatusChannels.Store(taskID, make(chan *workerTaskStatus, GetConfig().TaskStatusBufferSize))
		return core.PhaseUndefined, "", fmt.Errorf("task context not found")
	}

	taskStatusChannel := taskStatusChannelResult.(chan *workerTaskStatus)

	var latestWorkerTaskStatus *workerTaskStatus
Loop:
	for {
		select {
		case x := <-taskStatusChannel:
			// ensure we retrieve the latest status from the worker that is currently assigned to the task
			if x.workerID == workerID {
				latestWorkerTaskStatus = x
			}
		default:
			break Loop
		}
	}

	if latestWorkerTaskStatus == nil {
		return core.PhaseUndefined, "", fmt.Errorf("unable to find task status update: %w", statusUpdateNotFoundError)
	}

	taskStatus := latestWorkerTaskStatus.taskStatus
	phase := core.Phase(taskStatus.GetPhase())

	// if not completed need to send ACK on taskID to worker
	if phase != core.PhaseSuccess && phase != core.PhaseRetryableFailure {
		f.queuesLock.RLock()
		defer f.queuesLock.RUnlock()

		// if here it should be impossible for the queue not to exist, but left for safety
		if queue, exists := f.queues[queueID]; exists {
			queue.lock.RLock()
			defer queue.lock.RUnlock()

			if worker, exists := queue.workers[workerID]; exists {
				worker.responseChan <- &pb.HeartbeatResponse{
					TaskId:    taskID,
					Operation: pb.HeartbeatResponse_ACK,
				}
			}
		}
	}

	return phase, taskStatus.GetReason(), nil
}

// Cleanup is used to indicate a task is no longer being tracked by the worker and delete the
// associated task context.
func (f *fastTaskServiceImpl) Cleanup(ctx context.Context, taskID, queueID, workerID string) error {
	// send delete taskID message to worker
	f.queuesLock.RLock()
	defer f.queuesLock.RUnlock()

	if queue, exists := f.queues[queueID]; exists {
		queue.lock.RLock()
		defer queue.lock.RUnlock()

		if worker, exists := queue.workers[workerID]; exists {
			worker.responseChan <- &pb.HeartbeatResponse{
				TaskId:    taskID,
				Operation: pb.HeartbeatResponse_DELETE,
			}
		}
	}

	// delete task context
	f.taskStatusChannels.Delete(taskID)

	// remove pending owner
	f.removePendingOwner(queueID, taskID)

	return nil
}

// newFastTaskService creates a new fastTaskServiceImpl.
func newFastTaskService(enqueueOwner core.EnqueueOwner, scope promutils.Scope) *fastTaskServiceImpl {
	return &fastTaskServiceImpl{
		enqueueOwner:      enqueueOwner,
		queues:            make(map[string]*Queue),
		pendingTaskOwners: make(map[string]map[string]types.NamespacedName),
		metrics:           newMetrics(scope),
	}
}
