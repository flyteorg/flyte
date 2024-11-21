package plugin

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flytestdlib/promutils"

	"github.com/unionai/flyte/fasttask/plugin/pb"
)

func TestCheckStatus(t *testing.T) {
	ctx := context.TODO()
	tests := []struct {
		name          string
		taskID        string
		queueID       string
		workerID      string
		taskStatuses  []*workerTaskStatus
		expectedPhase core.Phase
		expectedError error
	}{
		{
			name:          "ChannelDoesNotExist",
			taskID:        "bar",
			queueID:       "foo",
			workerID:      "w1",
			taskStatuses:  nil,
			expectedPhase: core.PhaseUndefined,
			expectedError: fmt.Errorf("task context not found: %w", taskContextNotFoundError),
		},
		{
			name:          "NoUpdates",
			taskID:        "bar",
			queueID:       "foo",
			workerID:      "w1",
			taskStatuses:  []*workerTaskStatus{},
			expectedPhase: core.PhaseUndefined,
			expectedError: fmt.Errorf("unable to find task status update: %w", statusUpdateNotFoundError),
		},
		{
			name:     "UpdateFromDifferentWorker",
			taskID:   "bar",
			queueID:  "foo",
			workerID: "w1",
			taskStatuses: []*workerTaskStatus{
				&workerTaskStatus{
					workerID: "w2",
					taskStatus: &pb.TaskStatus{
						TaskId: "bar",
						Phase:  int32(core.PhaseRunning),
					},
				},
			},
			expectedPhase: core.PhaseUndefined,
			expectedError: fmt.Errorf("unable to find task status update: %w", statusUpdateNotFoundError),
		},
		{
			name:     "MultipleUpdates",
			taskID:   "bar",
			queueID:  "foo",
			workerID: "w1",
			taskStatuses: []*workerTaskStatus{
				&workerTaskStatus{
					workerID: "w1",
					taskStatus: &pb.TaskStatus{
						TaskId: "bar",
						Phase:  int32(core.PhaseQueued),
					},
				},
				&workerTaskStatus{
					workerID: "w1",
					taskStatus: &pb.TaskStatus{
						TaskId: "bar",
						Phase:  int32(core.PhaseRunning),
					},
				},
			},
			expectedPhase: core.PhaseRunning,
			expectedError: nil,
		},
		{
			name:     "NoAckOnSuccess",
			taskID:   "bar",
			queueID:  "foo",
			workerID: "w1",
			taskStatuses: []*workerTaskStatus{
				&workerTaskStatus{
					workerID: "w1",
					taskStatus: &pb.TaskStatus{
						TaskId: "bar",
						Phase:  int32(core.PhaseSuccess),
					},
				},
			},
			expectedPhase: core.PhaseSuccess,
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create fastTaskService
			enqueueOwner := func(owner types.NamespacedName) error {
				return nil
			}
			scope := promutils.NewTestScope()

			fastTaskService := newFastTaskService(enqueueOwner, scope)

			// setup taskStatusChannels
			if test.taskStatuses != nil {
				taskStatusChannel := make(chan *workerTaskStatus, len(test.taskStatuses))
				for _, taskStatus := range test.taskStatuses {
					taskStatusChannel <- taskStatus
				}
				fastTaskService.taskStatusChannels.Store(test.taskID, taskStatusChannel)
			}

			// setup response channels for queue workers
			queue := &Queue{
				workers: make(map[string]*Worker),
			}
			queues := map[string]*Queue{
				test.queueID: queue,
			}

			responseChans := make(map[string]chan *pb.HeartbeatResponse)
			if len(test.queueID) > 0 {
				if len(test.workerID) > 0 {
					responseChan := make(chan *pb.HeartbeatResponse, 1)
					worker := &Worker{
						workerID:     test.workerID,
						responseChan: responseChan,
					}
					queue.workers[test.workerID] = worker

					responseChans[test.workerID] = responseChan
				}

				for _, taskStatus := range test.taskStatuses {
					if taskStatus.workerID != test.workerID {
						responseChan := make(chan *pb.HeartbeatResponse, 1)
						worker := &Worker{
							workerID:     taskStatus.workerID,
							responseChan: responseChan,
						}
						queue.workers[taskStatus.workerID] = worker

						responseChans[taskStatus.workerID] = responseChan
					}
				}
			}
			fastTaskService.queues = queues

			// offer on queue and validate
			phase, _, err := fastTaskService.CheckStatus(ctx, test.taskID, test.queueID, test.workerID)
			assert.Equal(t, test.expectedPhase, phase)
			assert.Equal(t, test.expectedError, err)

			// validate ACK response
			for responseWorkerID, responseChan := range responseChans {
				expectAck := false
				if test.expectedError == nil && responseWorkerID == test.workerID &&
					test.expectedPhase != core.PhaseSuccess && test.expectedPhase != core.PhaseRetryableFailure {

					expectAck = true
				}

				if expectAck {
					// the assigned worker should have received an ACK response
					select {
					case response := <-responseChan:
						assert.NotNil(t, response)
						assert.Equal(t, test.taskID, response.GetTaskId())
						assert.Equal(t, pb.HeartbeatResponse_ACK, response.GetOperation())
					default:
						assert.Fail(t, "expected response")
					}
				} else {
					// all other workers should have no responses
					select {
					case <-responseChan:
						assert.Fail(t, "unexpected response")
					default:
					}
				}
			}
		})
	}
}

func TestCleanup(t *testing.T) {
	ctx := context.TODO()
	tests := []struct {
		name                    string
		taskID                  string
		queueID                 string
		workerID                string
		queues                  map[string]*Queue
		expectedError           error
		pendingOwnerExists      bool
		taskStatusChannelExists bool
	}{
		{
			name:                    "QueueDoesNotExist",
			taskID:                  "bar",
			queueID:                 "foo",
			workerID:                "w1",
			queues:                  map[string]*Queue{},
			expectedError:           nil,
			pendingOwnerExists:      false,
			taskStatusChannelExists: false,
		},
		{
			name:     "WorkerDoesNostExist",
			taskID:   "bar",
			queueID:  "foo",
			workerID: "w1",
			queues: map[string]*Queue{
				"foo": &Queue{
					workers: map[string]*Worker{
						"w0": &Worker{
							workerID: "w0",
						},
					},
				},
			},
			expectedError:           nil,
			pendingOwnerExists:      false,
			taskStatusChannelExists: false,
		},
		{
			name:     "WorkerExists",
			taskID:   "bar",
			queueID:  "foo",
			workerID: "w1",
			queues: map[string]*Queue{
				"foo": &Queue{
					workers: map[string]*Worker{
						"w0": &Worker{
							workerID: "w0",
						},
						"w1": &Worker{
							workerID: "w1",
						},
					},
				},
			},
			expectedError:           nil,
			pendingOwnerExists:      false,
			taskStatusChannelExists: false,
		},
		{
			// worker exists and pendingOwner / taskStatusChannel are cleaned up
			name:     "WorkerExistsCleanupAll",
			taskID:   "bar",
			queueID:  "foo",
			workerID: "w1",
			queues: map[string]*Queue{
				"foo": &Queue{
					workers: map[string]*Worker{
						"w0": &Worker{
							workerID: "w0",
						},
						"w1": &Worker{
							workerID: "w1",
						},
					},
				},
			},
			expectedError:           nil,
			pendingOwnerExists:      true,
			taskStatusChannelExists: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create fastTaskService
			enqueueOwner := func(owner types.NamespacedName) error {
				return nil
			}
			scope := promutils.NewTestScope()

			fastTaskService := newFastTaskService(enqueueOwner, scope)

			// setup response channels for queue workers
			responseChans := make(map[string]chan *pb.HeartbeatResponse)
			if queue, exists := test.queues[test.queueID]; exists {
				for workerID, worker := range queue.workers {
					responseChan := make(chan *pb.HeartbeatResponse, 1)
					responseChans[workerID] = responseChan
					worker.responseChan = responseChan
				}
			}
			fastTaskService.queues = test.queues

			// initialize pendingTaskOwners and taskStatusChannels if necessary
			if test.pendingOwnerExists {
				fastTaskService.addPendingOwner(test.queueID, test.taskID,
					types.NamespacedName{Name: "foo"})
			} else {
				_, exists := fastTaskService.pendingTaskOwners[test.queueID]
				assert.False(t, exists)
			}

			if test.taskStatusChannelExists {
				fastTaskService.taskStatusChannels.Store(test.taskID, make(chan *pb.TaskStatus, 1))
			} else {
				_, exists := fastTaskService.taskStatusChannels.Load(test.taskID)
				assert.False(t, exists)
			}

			// offer on queue and validate
			err := fastTaskService.Cleanup(ctx, test.taskID, test.queueID, test.workerID)
			assert.Equal(t, test.expectedError, err)

			// validate DELETE response
			for responseWorkerID, responseChan := range responseChans {
				expectDelete := false
				if responseWorkerID == test.workerID {
					if queue, exists := test.queues[test.queueID]; exists {
						if _, exists := queue.workers[responseWorkerID]; exists {
							expectDelete = true
						}
					}
				}

				if expectDelete {
					// the assigned worker should have received an DELETE response
					select {
					case response := <-responseChan:
						assert.NotNil(t, response)
						assert.Equal(t, test.taskID, response.GetTaskId())
						assert.Equal(t, pb.HeartbeatResponse_DELETE, response.GetOperation())
					default:
						assert.Fail(t, "expected response")
					}
				} else {
					// all other workers should have no responses
					select {
					case <-responseChan:
						assert.Fail(t, "unexpected response")
					default:
					}
				}
			}

			// validate pendingTaskOwners and taskStatusChannels are cleaned up
			_, exists := fastTaskService.pendingTaskOwners[test.queueID]
			assert.False(t, exists)

			_, exists = fastTaskService.taskStatusChannels.Load(test.taskID)
			assert.False(t, exists)
		})
	}
}

func TestOfferOnQueue(t *testing.T) {
	ctx := context.TODO()
	tests := []struct {
		name                 string
		queueID              string
		taskID               string
		namespace            string
		workflowID           string
		queues               map[string]*Queue
		expectedWorkerID     string
		expectedError        error
		expectedPendingOwner bool
	}{
		{
			name:                 "QueueDoesNotExist",
			queueID:              "foo",
			taskID:               "bar",
			namespace:            "x",
			workflowID:           "y",
			queues:               map[string]*Queue{},
			expectedWorkerID:     "",
			expectedError:        nil,
			expectedPendingOwner: true,
		},
		{
			name:       "PreferredWorker",
			queueID:    "foo",
			taskID:     "bar",
			namespace:  "x",
			workflowID: "y",
			queues: map[string]*Queue{
				"foo": &Queue{
					workers: map[string]*Worker{
						"w0": &Worker{
							workerID: "w0",
							capacity: &pb.Capacity{
								ExecutionCount: 0,
								ExecutionLimit: 1,
							},
						},
						"w1": &Worker{
							workerID: "w1",
							capacity: &pb.Capacity{
								ExecutionCount: 1,
								ExecutionLimit: 1,
							},
						},
					},
				},
			},
			expectedWorkerID: "w0",
			expectedError:    nil,
		},
		{
			name:       "AcceptedWorker",
			queueID:    "foo",
			taskID:     "bar",
			namespace:  "x",
			workflowID: "y",
			queues: map[string]*Queue{
				"foo": &Queue{
					workers: map[string]*Worker{
						"w0": &Worker{
							workerID: "w0",
							capacity: &pb.Capacity{
								ExecutionCount: 1,
								ExecutionLimit: 1,
								BacklogCount:   1,
								BacklogLimit:   1,
							},
						},
						"w1": &Worker{
							workerID: "w1",
							capacity: &pb.Capacity{
								ExecutionCount: 1,
								ExecutionLimit: 1,
								BacklogCount:   0,
								BacklogLimit:   1,
							},
						},
					},
				},
			},
			expectedWorkerID: "w1",
			expectedError:    nil,
		},
		{
			name:       "NoWorkerAvailable",
			queueID:    "foo",
			taskID:     "bar",
			namespace:  "x",
			workflowID: "y",
			queues: map[string]*Queue{
				"foo": &Queue{
					workers: map[string]*Worker{
						"w0": &Worker{
							workerID: "w0",
							capacity: &pb.Capacity{
								ExecutionCount: 1,
								ExecutionLimit: 1,
								BacklogCount:   1,
								BacklogLimit:   1,
							},
						},
					},
				},
			},
			expectedWorkerID: "",
			expectedError:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create fastTaskService
			enqueueOwner := func(owner types.NamespacedName) error {
				return nil
			}
			scope := promutils.NewTestScope()

			fastTaskService := newFastTaskService(enqueueOwner, scope)

			// setup response channels for queue workers
			responseChans := make(map[string]chan *pb.HeartbeatResponse)
			if queue, exists := test.queues[test.queueID]; exists {
				for workerID, worker := range queue.workers {
					responseChan := make(chan *pb.HeartbeatResponse, 1)
					responseChans[workerID] = responseChan
					worker.responseChan = responseChan
				}
			}
			fastTaskService.queues = test.queues

			// pre-execute validation - taskStatusChannel does not exist
			_, exists := fastTaskService.taskStatusChannels.Load(test.taskID)
			assert.False(t, exists)

			// offer on queue and validate
			workerID, err := fastTaskService.OfferOnQueue(ctx, test.queueID, test.taskID, test.namespace, test.workflowID, []string{}, make(map[string]string))
			assert.Equal(t, test.expectedWorkerID, workerID)
			assert.Equal(t, test.expectedError, err)

			if len(workerID) > 0 {
				// validate ASSIGN response
				for responseWorkerID, responseChan := range responseChans {
					if responseWorkerID == workerID {
						// the assigned worker should have received an ASSIGN response
						select {
						case response := <-responseChan:
							assert.NotNil(t, response)
							assert.Equal(t, test.taskID, response.GetTaskId())
							assert.Equal(t, pb.HeartbeatResponse_ASSIGN, response.GetOperation())
						default:
							assert.Fail(t, "expected response")
						}
					} else {
						// all other workers should have no responses
						select {
						case <-responseChan:
							assert.Fail(t, "unexpected response")
						default:
						}
					}
				}

				// ensure taskStatusChannel now exists
				_, exists := fastTaskService.taskStatusChannels.Load(test.taskID)
				assert.True(t, exists)
			}

			if test.expectedPendingOwner {
				pendingOwners, exists := fastTaskService.pendingTaskOwners[test.queueID]
				assert.True(t, exists)

				_, exists = pendingOwners[test.taskID]
				assert.True(t, exists)
			}
		})
	}
}

func TestPendingOwnerManagement(t *testing.T) {
	// create fastTaskService
	ownerEnqueueCount := 0
	enqueueOwner := func(owner types.NamespacedName) error {
		ownerEnqueueCount++
		return nil
	}
	scope := promutils.NewTestScope()

	fastTaskService := newFastTaskService(enqueueOwner, scope)
	assert.Equal(t, 0, len(fastTaskService.queues))

	// add pending owners
	additions := []struct {
		queueID                  string
		taskID                   string
		ownerIDName              string
		expectedQueueOwnersCount int
		totalOwnerCount          int
	}{
		{
			// add owner to new queue
			queueID:                  "foo",
			taskID:                   "a",
			ownerIDName:              "0",
			expectedQueueOwnersCount: 1,
			totalOwnerCount:          1,
		},
		{
			// add owner to existing queue
			queueID:                  "foo",
			taskID:                   "b",
			ownerIDName:              "1",
			expectedQueueOwnersCount: 1,
			totalOwnerCount:          2,
		},
		{
			// add owner to another new queue
			queueID:                  "bar",
			taskID:                   "c",
			ownerIDName:              "2",
			expectedQueueOwnersCount: 2,
			totalOwnerCount:          3,
		},
	}

	for _, addition := range additions {
		fastTaskService.addPendingOwner(addition.queueID, addition.taskID, types.NamespacedName{Name: addition.ownerIDName})

		assert.Equal(t, addition.expectedQueueOwnersCount, len(fastTaskService.pendingTaskOwners))
		totalOwnerCount := 0
		for _, queueOwners := range fastTaskService.pendingTaskOwners {
			totalOwnerCount += len(queueOwners)
		}
		assert.Equal(t, addition.totalOwnerCount, totalOwnerCount)
	}

	// validate overflow management on addPendingOwner
	overflowTestQueueID := "baz"
	for i := 0; i < maxPendingOwnersPerQueue; i++ {
		fastTaskService.addPendingOwner(overflowTestQueueID, fmt.Sprintf("%d", i), types.NamespacedName{Name: fmt.Sprintf("%d", i)})
	}
	assert.Equal(t, maxPendingOwnersPerQueue, len(fastTaskService.pendingTaskOwners[overflowTestQueueID]))

	fastTaskService.addPendingOwner(overflowTestQueueID, "overflow", types.NamespacedName{Name: "overflow"})
	assert.Equal(t, maxPendingOwnersPerQueue, len(fastTaskService.pendingTaskOwners[overflowTestQueueID]))

	// validate enqueuePendingOwners
	assert.Equal(t, 0, ownerEnqueueCount)

	fastTaskService.enqueuePendingOwners(overflowTestQueueID)
	assert.Equal(t, maxPendingOwnersPerQueue, ownerEnqueueCount)
	assert.Equal(t, 0, len(fastTaskService.pendingTaskOwners[overflowTestQueueID]))

	fastTaskService.enqueuePendingOwners(overflowTestQueueID) // call a second time to validate on empty queue
	assert.Equal(t, maxPendingOwnersPerQueue, ownerEnqueueCount)

	// remove workers
	removals := []struct {
		queueID                  string
		taskID                   string
		expectedQueueOwnersCount int
		totalOwnerCount          int
	}{
		{
			// remove owner from non-existent queue
			queueID:                  "baz",
			taskID:                   "d",
			expectedQueueOwnersCount: 2,
			totalOwnerCount:          3,
		},
		{
			// remove worker from existing queue
			queueID:                  "foo",
			taskID:                   "a",
			expectedQueueOwnersCount: 2,
			totalOwnerCount:          2,
		},
		{
			// remove last worker from queue
			queueID:                  "foo",
			taskID:                   "b",
			expectedQueueOwnersCount: 1,
			totalOwnerCount:          1,
		},
	}

	for _, removal := range removals {
		fastTaskService.removePendingOwner(removal.queueID, removal.taskID)

		assert.Equal(t, removal.expectedQueueOwnersCount, len(fastTaskService.pendingTaskOwners))
		totalOwnerCount := 0
		for _, queueOwners := range fastTaskService.pendingTaskOwners {
			totalOwnerCount += len(queueOwners)
		}
		assert.Equal(t, removal.totalOwnerCount, totalOwnerCount)
	}
}

func TestQueueWorkerManagement(t *testing.T) {
	// create fastTaskService
	enqueueOwner := func(owner types.NamespacedName) error {
		return nil
	}
	scope := promutils.NewTestScope()

	fastTaskService := newFastTaskService(enqueueOwner, scope)
	assert.Equal(t, 0, len(fastTaskService.queues))

	// add workers
	additions := []struct {
		queueID            string
		workerID           string
		expectedQueueCount int
		totalWorkerCount   int
	}{
		{
			// add worker to new queue
			queueID:            "foo",
			workerID:           "a",
			expectedQueueCount: 1,
			totalWorkerCount:   1,
		},
		{
			// add worker to existing queue
			queueID:            "foo",
			workerID:           "b",
			expectedQueueCount: 1,
			totalWorkerCount:   2,
		},
		{
			// add worker to another new queue
			queueID:            "bar",
			workerID:           "c",
			expectedQueueCount: 2,
			totalWorkerCount:   3,
		},
	}

	for _, addition := range additions {
		worker := &Worker{
			workerID: addition.workerID,
		}

		queue := fastTaskService.addWorkerToQueue(addition.queueID, worker)
		assert.NotNil(t, queue)

		assert.Equal(t, addition.expectedQueueCount, len(fastTaskService.queues))
		totalWorkers := 0
		for _, q := range fastTaskService.queues {
			totalWorkers += len(q.workers)
		}
		assert.Equal(t, addition.totalWorkerCount, totalWorkers)
	}

	// remove workers
	removals := []struct {
		queueID            string
		workerID           string
		expectedQueueCount int
		totalWorkerCount   int
	}{
		{
			// remove worker from non-existent queue
			queueID:            "baz",
			workerID:           "d",
			expectedQueueCount: 2,
			totalWorkerCount:   3,
		},
		{
			// remove worker from existing queue
			queueID:            "foo",
			workerID:           "a",
			expectedQueueCount: 2,
			totalWorkerCount:   2,
		},
		{
			// remove last worker from queue
			queueID:            "foo",
			workerID:           "b",
			expectedQueueCount: 1,
			totalWorkerCount:   1,
		},
	}

	for _, removal := range removals {
		fastTaskService.removeWorkerFromQueue(removal.queueID, removal.workerID)

		assert.Equal(t, removal.expectedQueueCount, len(fastTaskService.queues))
		totalWorkers := 0
		for _, q := range fastTaskService.queues {
			totalWorkers += len(q.workers)
		}
		assert.Equal(t, removal.totalWorkerCount, totalWorkers)
	}
}
