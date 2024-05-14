package plugin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sync"

	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"

	"github.com/unionai/flyte/fasttask/plugin/pb"
)

// FastTaskService is a gRPC service that manages assignment and management of task executions with
// respect to fasttask workers.
type FastTaskService struct {
	pb.UnimplementedFastTaskServer
	enqueueOwner       core.EnqueueOwner
	queues             map[string]*Queue
	queuesLock         sync.RWMutex
	taskStatusChannels sync.Map // map[string]chan *WorkerTaskStatus
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

// Heartbeat is a gRPC stream that manages the heartbeat of a fasttask worker. This includes
// receiving task status updates and sending task assignments.
func (f *FastTaskService) Heartbeat(stream pb.FastTask_HeartbeatServer) error {
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
	f.queuesLock.Lock()
	queue, exists := f.queues[heartbeatRequest.GetQueueId()]
	if !exists {
		queue = &Queue{
			workers: make(map[string]*Worker),
		}
		f.queues[heartbeatRequest.GetQueueId()] = queue
	}
	f.queuesLock.Unlock()

	queue.lock.Lock()
	queue.workers[workerID] = worker
	queue.lock.Unlock()

	// cleanup worker on exit
	defer func() {
		f.queuesLock.Lock()
		queue, exists := f.queues[heartbeatRequest.GetQueueId()]
		if exists {
			queue.lock.Lock()
			delete(queue.workers, workerID)
			if len(queue.workers) == 0 {
				delete(f.queues, heartbeatRequest.GetQueueId())
			}
			queue.lock.Unlock()
		}
		f.queuesLock.Unlock()
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
				ownerID := types.NamespacedName{
					Namespace: taskStatus.GetNamespace(),
					Name:      taskStatus.GetWorkflowId(),
				}

				if err := f.enqueueOwner(ownerID); err != nil {
					logger.Warnf(context.Background(), "failed to enqueue owner for task %s: %+v", taskStatus.GetTaskId(), err)
				}
			}
		}
	}

	return nil
}

// OfferOnQueue offers a task to a worker on a specific queue. If no workers are available, an
// empty string is returned.
func (f *FastTaskService) OfferOnQueue(ctx context.Context, queueID, taskID, namespace, workflowID string, cmd []string) (string, error) {
	f.queuesLock.RLock()
	defer f.queuesLock.RUnlock()

	queue, exists := f.queues[queueID]
	if !exists {
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
		return "", nil // no workers available
	}

	// send assign message to worker
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
func (f *FastTaskService) CheckStatus(ctx context.Context, taskID, queueID, workerID string) (core.Phase, string, error) {
	taskStatusChannelResult, exists := f.taskStatusChannels.Load(taskID)
	if !exists {
		// if this plugin restarts then TaskContexts may not exist for tasks that are still active. we can
		// create a TaskContext here because we ensure it will be cleaned up when the task completes.
		f.taskStatusChannels.Store(taskID, make(chan *workerTaskStatus, GetConfig().TaskStatusBufferSize))
		return core.PhaseUndefined, "", errors.New("task context not found")
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

		queue := f.queues[queueID]
		queue.lock.RLock()
		defer queue.lock.RUnlock()

		if worker, exists := queue.workers[workerID]; exists {
			worker.responseChan <- &pb.HeartbeatResponse{
				TaskId:    taskID,
				Operation: pb.HeartbeatResponse_ACK,
			}
		}
	}

	return phase, taskStatus.GetReason(), nil
}

// Cleanup is used to indicate a task is no longer being tracked by the worker and delete the
// associated task context.
func (f *FastTaskService) Cleanup(ctx context.Context, taskID, queueID, workerID string) error {
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
	return nil
}

// NewFastTaskService creates a new FastTaskService.
func NewFastTaskService(enqueueOwner core.EnqueueOwner) *FastTaskService {
	return &FastTaskService{
		enqueueOwner: enqueueOwner,
		queues:       make(map[string]*Queue),
	}
}
