package plugin

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"

	coreIdl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	pluginsCoreMock "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/transformers/k8s"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/promutils"

	"github.com/unionai/flyte/fasttask/plugin/interfaces"
	"github.com/unionai/flyte/fasttask/plugin/pb"
)

var testConfig = &Config{
	FastTaskExecutionMetric: FastTaskExecutionMetricConfig{
		TTL:             config.Duration{Duration: 5 * time.Second},
		CleanupInterval: config.Duration{Duration: time.Second},
	},
}

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
			enqueueOwner := func(labels map[string]string) error {
				return nil
			}
			scope := promutils.NewTestScope()

			workersMap := sync.Map{}
			queue := &environmentImpl{
				workers: &workersMap,
			}
			store := newEnvironmentStore()
			store.GetOrCreate(test.queueID, queue)
			// setup response channels for env workers
			responseChans := make(map[string]chan *pb.HeartbeatResponse)
			if len(test.queueID) > 0 {
				if len(test.workerID) > 0 {
					responseChan := make(chan *pb.HeartbeatResponse, 1)
					worker := &workerImpl{
						id:           test.workerID,
						responseChan: responseChan,
					}
					queue.workers.Store(test.workerID, worker)

					responseChans[test.workerID] = responseChan
				}

				for _, taskStatus := range test.taskStatuses {
					if taskStatus.workerID != test.workerID {
						responseChan := make(chan *pb.HeartbeatResponse, 1)
						worker := &workerImpl{
							id:           taskStatus.workerID,
							responseChan: responseChan,
						}
						queue.workers.Store(taskStatus.workerID, worker)

						responseChans[taskStatus.workerID] = responseChan
					}
				}
			}

			// initialize InMemoryBuilder
			kubeClient := &kubeClient{}
			kubeCache := &kubeCache{}

			kubeClientImpl := &pluginsCoreMock.KubeClient{}
			kubeClientImpl.OnGetClient().Return(kubeClient)
			kubeClientImpl.OnGetCache().Return(kubeCache)

			// Set up metrics
			metrics := newBuilderMetrics(scope)

			// Create test builder
			builder := &environmentBuilderImpl{
				kubeClient:  kubeClientImpl,
				store:       store,
				metrics:     metrics,
				randSource:  rand.New(rand.NewSource(time.Now().UnixNano())),
				scaleUpChan: make(chan string, 1),
			}

			fastTaskService := newFastTaskService(enqueueOwner, builder, store, scope, testConfig)

			// setup taskStatusChannels
			if test.taskStatuses != nil {
				taskStatusChannel := make(chan *workerTaskStatus, len(test.taskStatuses))
				for _, taskStatus := range test.taskStatuses {
					taskStatusChannel <- taskStatus
				}
				fastTaskService.taskStatusChannels.Store(test.taskID, taskStatusChannel)
			}

			// offer on queue and validate
			taskStatus, err := fastTaskService.CheckStatus(ctx, test.taskID, test.queueID, test.workerID)
			assert.Equal(t, test.expectedPhase, taskStatus.Phase)
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

	env1 := newEnvironmentStore()
	workers1 := sync.Map{}
	workers1.Store("w0", &workerImpl{id: "w0"})
	env1.GetOrCreate("foo", &environmentImpl{
		workers: &workers1,
	})

	env2 := newEnvironmentStore()
	workers2 := sync.Map{}
	workers2.Store("w0", &workerImpl{id: "w0"})
	workers2.Store("w1", &workerImpl{id: "w1"})
	env2.GetOrCreate("foo", &environmentImpl{
		workers: &workers2,
	})

	env3 := newEnvironmentStore()
	workers3 := sync.Map{}
	workers3.Store("w0", &workerImpl{id: "w0"})
	workers3.Store("w1", &workerImpl{id: "w1"})
	env3.GetOrCreate("foo", &environmentImpl{
		workers: &workers3,
	})

	tests := []struct {
		name                    string
		taskID                  string
		queueID                 string
		workerID                string
		envStore                interfaces.EnvironmentStore
		expectedError           error
		pendingOwnerExists      bool
		taskStatusChannelExists bool
	}{
		{
			name:                    "QueueDoesNotExist",
			taskID:                  "bar",
			queueID:                 "foo",
			workerID:                "w1",
			envStore:                newEnvironmentStore(),
			expectedError:           nil,
			pendingOwnerExists:      false,
			taskStatusChannelExists: false,
		},
		{
			name:                    "WorkerDoesNostExist",
			taskID:                  "bar",
			queueID:                 "foo",
			workerID:                "w1",
			envStore:                env1,
			expectedError:           nil,
			pendingOwnerExists:      false,
			taskStatusChannelExists: false,
		},
		{
			name:                    "WorkerExists",
			taskID:                  "bar",
			queueID:                 "foo",
			workerID:                "w1",
			envStore:                env2,
			expectedError:           nil,
			pendingOwnerExists:      false,
			taskStatusChannelExists: false,
		},
		{
			// worker exists and pendingOwner / taskStatusChannel are cleaned up
			name:                    "WorkerExistsCleanupAll",
			taskID:                  "bar",
			queueID:                 "foo",
			workerID:                "w1",
			envStore:                env3,
			expectedError:           nil,
			pendingOwnerExists:      true,
			taskStatusChannelExists: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create fastTaskService
			enqueueOwner := func(labels map[string]string) error {
				return nil
			}
			scope := promutils.NewTestScope()

			// initialize InMemoryBuilder
			kubeClient := &kubeClient{}
			kubeCache := &kubeCache{}

			kubeClientImpl := &pluginsCoreMock.KubeClient{}
			kubeClientImpl.OnGetClient().Return(kubeClient)
			kubeClientImpl.OnGetCache().Return(kubeCache)

			// Set up metrics
			metrics := newBuilderMetrics(scope)

			// Create test builder
			builder := &environmentBuilderImpl{
				kubeClient:  kubeClientImpl,
				store:       test.envStore,
				metrics:     metrics,
				randSource:  rand.New(rand.NewSource(time.Now().UnixNano())),
				scaleUpChan: make(chan string, 1),
			}

			fastTaskService := newFastTaskService(enqueueOwner, builder, test.envStore, scope, testConfig)
			// setup response channels for queue workers
			responseChans := make(map[string]chan *pb.HeartbeatResponse)
			if env := test.envStore.Get(test.queueID); env != nil {
				//for workerID, worker := range queue.workers {
				//	responseChan := make(chan *pb.HeartbeatResponse, 1)
				//	responseChans[workerID] = responseChan
				//	worker.responseChan = responseChan
				//}
				env.RangeWorkers(func(workerID string, worker interfaces.Worker) bool {
					workerImpl, ok := worker.(*workerImpl)
					if !ok {
						return true
					}
					responseChan := make(chan *pb.HeartbeatResponse, 1)
					responseChans[workerID] = responseChan
					workerImpl.responseChan = responseChan
					return true
				})
			}

			workflowID := types.NamespacedName{
				Name: "foo",
			}
			enqueueLabels := map[string]string{
				k8s.WorkflowID: workflowID.String(),
			}

			// initialize pendingTaskOwners and taskStatusChannels if necessary
			if test.pendingOwnerExists {
				fastTaskService.AddPendingOwner(test.queueID, test.taskID, enqueueLabels)
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
					if queue := test.envStore.Get(test.queueID); queue != nil {
						if queue.GetWorker(responseWorkerID) != nil {
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

	env1 := newEnvironmentStore()
	workers1 := sync.Map{}
	workers1.Store("w0", &workerImpl{
		id: "w0",
		capacity: &pb.Capacity{
			ExecutionCount: 0,
			ExecutionLimit: 1,
		},
	})
	workers1.Store("w1", &workerImpl{
		id: "w1",
		capacity: &pb.Capacity{
			ExecutionCount: 1,
			ExecutionLimit: 1,
		},
	})
	env1.GetOrCreate("foo", &environmentImpl{
		workers: &workers1,
	})

	env2 := newEnvironmentStore()
	workers2 := sync.Map{}
	workers2.Store("w0", &workerImpl{
		id: "w0",
		capacity: &pb.Capacity{
			ExecutionCount: 1,
			ExecutionLimit: 1,
			BacklogCount:   1,
			BacklogLimit:   1,
		},
	})
	workers2.Store("w1", &workerImpl{
		id: "w1",
		capacity: &pb.Capacity{
			ExecutionCount: 1,
			ExecutionLimit: 1,
			BacklogCount:   0,
			BacklogLimit:   1,
		},
	})
	env2.GetOrCreate("foo", &environmentImpl{
		workers: &workers2,
	})

	env3 := newEnvironmentStore()
	workers3 := sync.Map{}
	workers3.Store("w0", &workerImpl{
		id: "w0",
		capacity: &pb.Capacity{
			ExecutionCount: 1,
			ExecutionLimit: 1,
			BacklogCount:   1,
			BacklogLimit:   1,
		},
	})
	env3.GetOrCreate("foo", &environmentImpl{
		workers: &workers3,
	})

	tests := []struct {
		name                 string
		queueID              string
		taskID               string
		namespace            string
		workflowID           string
		envStore             interfaces.EnvironmentStore
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
			envStore:             newEnvironmentStore(),
			expectedWorkerID:     "",
			expectedError:        fmt.Errorf("environment 'foo' not found"),
			expectedPendingOwner: true,
		},
		{
			name:             "PreferredWorker",
			queueID:          "foo",
			taskID:           "bar",
			namespace:        "x",
			workflowID:       "y",
			envStore:         env1,
			expectedWorkerID: "w0",
			expectedError:    nil,
		},
		{
			name:             "AcceptedWorker",
			queueID:          "foo",
			taskID:           "bar",
			namespace:        "x",
			workflowID:       "y",
			envStore:         env2,
			expectedWorkerID: "w1",
			expectedError:    nil,
		},
		{
			name:             "NoWorkerAvailable",
			queueID:          "foo",
			taskID:           "bar",
			namespace:        "x",
			workflowID:       "y",
			envStore:         env3,
			expectedWorkerID: "",
			expectedError:    noCapacityAvailableError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create fastTaskService
			enqueueOwner := func(labels map[string]string) error {
				return nil
			}
			scope := promutils.NewTestScope()

			// initialize InMemoryBuilder
			kubeClient := &kubeClient{}
			kubeCache := &kubeCache{}

			kubeClientImpl := &pluginsCoreMock.KubeClient{}
			kubeClientImpl.OnGetClient().Return(kubeClient)
			kubeClientImpl.OnGetCache().Return(kubeCache)

			// Set up metrics
			metrics := newBuilderMetrics(scope)

			// Create test builder
			builder := &environmentBuilderImpl{
				kubeClient:  kubeClientImpl,
				store:       test.envStore,
				metrics:     metrics,
				randSource:  rand.New(rand.NewSource(time.Now().UnixNano())),
				scaleUpChan: make(chan string, 1),
			}

			fastTaskService := newFastTaskService(enqueueOwner, builder, test.envStore, scope, testConfig)

			// setup response channels for queue workers
			responseChans := make(map[string]chan *pb.HeartbeatResponse)
			if env := test.envStore.Get(test.queueID); env != nil {
				env.RangeWorkers(func(workerID string, worker interfaces.Worker) bool {
					workerImpl, ok := worker.(*workerImpl)
					if !ok {
						return true
					}
					responseChan := make(chan *pb.HeartbeatResponse, 1)
					responseChans[workerID] = responseChan
					workerImpl.responseChan = responseChan
					return true
				})
			}

			// pre-execute validation - taskStatusChannel does not exist
			_, exists := fastTaskService.taskStatusChannels.Load(test.taskID)
			assert.False(t, exists)

			// offer on queue and validate
			workflowID := types.NamespacedName{
				Namespace: test.namespace,
				Name:      test.workflowID,
			}

			execID := &coreIdl.WorkflowExecutionIdentifier{
				Org:     "foo",
				Project: "bar",
				Domain:  "dev",
				Name:    "abc",
			}

			enqueueLabels := map[string]string{
				k8s.WorkflowID: workflowID.String(),
			}

			worker, err := fastTaskService.OfferTaskToEnvironment(ctx, execID, test.queueID, test.taskID, test.namespace, test.workflowID, []string{}, make(map[string]string), enqueueLabels)
			if test.expectedError != nil {
				assert.Nil(t, worker)
				assert.Equal(t, test.expectedError, err)
				return
			}
			assert.Equal(t, test.expectedWorkerID, worker.ID())
			assert.Equal(t, test.expectedError, err)

			if len(worker.ID()) > 0 {
				// validate ASSIGN response
				for responseWorkerID, responseChan := range responseChans {
					if responseWorkerID == worker.ID() {
						// the assigned worker should have received an ASSIGN response
						select {
						case response := <-responseChan:
							assert.NotNil(t, response)
							assert.Equal(t, test.taskID, response.GetTaskId())
							assert.Equal(t, pb.HeartbeatResponse_ASSIGN, response.GetOperation())
							assert.Equal(t, execID.GetOrg(), response.GetExecId().GetOrg())
							assert.Equal(t, execID.GetProject(), response.GetExecId().GetProject())
							assert.Equal(t, execID.GetDomain(), response.GetExecId().GetDomain())
							assert.Equal(t, execID.GetName(), response.GetExecId().GetName())
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
	enqueueOwner := func(labels map[string]string) error {
		ownerEnqueueCount++
		return nil
	}
	scope := promutils.NewTestScope()

	// initialize InMemoryBuilder
	kubeClient := &kubeClient{}
	kubeCache := &kubeCache{}

	kubeClientImpl := &pluginsCoreMock.KubeClient{}
	kubeClientImpl.OnGetClient().Return(kubeClient)
	kubeClientImpl.OnGetCache().Return(kubeCache)

	// Set up metrics
	metrics := newBuilderMetrics(scope)

	store := newEnvironmentStore()

	// Create test builder
	builder := &environmentBuilderImpl{
		kubeClient:  kubeClientImpl,
		store:       store,
		metrics:     metrics,
		randSource:  rand.New(rand.NewSource(time.Now().UnixNano())),
		scaleUpChan: make(chan string, 1),
	}

	fastTaskService := newFastTaskService(enqueueOwner, builder, store, scope, testConfig)
	assert.Equal(t, 0, len(fastTaskService.store.List()))

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
		enqueueLabels := map[string]string{
			k8s.WorkflowID: types.NamespacedName{Name: addition.ownerIDName}.String(),
		}

		// Test that AddPendingOwner returns true for newly added tasks
		isNewlyAdded := fastTaskService.AddPendingOwner(addition.queueID, addition.taskID, enqueueLabels)
		assert.True(t, isNewlyAdded, "Task %s should be newly added to queue %s", addition.taskID, addition.queueID)

		assert.Equal(t, addition.expectedQueueOwnersCount, len(fastTaskService.pendingTaskOwners))
		totalOwnerCount := 0
		for _, queueOwners := range fastTaskService.pendingTaskOwners {
			totalOwnerCount += len(queueOwners)
		}
		assert.Equal(t, addition.totalOwnerCount, totalOwnerCount)
	}

	// Test duplicate addition should return false
	for _, addition := range additions {
		isNewlyAdded := fastTaskService.AddPendingOwner(addition.queueID, addition.taskID, nil)
		assert.False(t, isNewlyAdded, "Adding duplicate task should return false")
	}

	// validate overflow management on addPendingOwner
	overflowTestQueueID := "baz"
	for i := 0; i < maxPendingOwnersPerQueue; i++ {
		enqueueLabels := map[string]string{
			k8s.WorkflowID: types.NamespacedName{Name: fmt.Sprintf("%d", i)}.String(),
		}
		isNewlyAdded := fastTaskService.AddPendingOwner(overflowTestQueueID, fmt.Sprintf("%d", i), enqueueLabels)
		assert.True(t, isNewlyAdded, "Task %d should be newly added", i)
	}
	assert.Equal(t, maxPendingOwnersPerQueue, len(fastTaskService.pendingTaskOwners[overflowTestQueueID]))

	enqueueLabels := map[string]string{
		k8s.WorkflowID: types.NamespacedName{Name: "overflow"}.String(),
	}

	// Test that adding to full queue returns false
	isOverflow := fastTaskService.AddPendingOwner(overflowTestQueueID, "overflow", enqueueLabels)
	assert.False(t, isOverflow, "Adding to full queue should return false")
	assert.Equal(t, maxPendingOwnersPerQueue, len(fastTaskService.pendingTaskOwners[overflowTestQueueID]))

	// validate enqueuePendingOwners
	assert.Equal(t, 0, ownerEnqueueCount)

	fastTaskService.EnqueuePendingOwners(overflowTestQueueID, 3)
	assert.Equal(t, 3, ownerEnqueueCount)

	fastTaskService.EnqueuePendingOwners(overflowTestQueueID, maxPendingOwnersPerQueue)
	assert.Equal(t, maxPendingOwnersPerQueue, ownerEnqueueCount)
	assert.Equal(t, 0, len(fastTaskService.pendingTaskOwners[overflowTestQueueID]))

	fastTaskService.EnqueuePendingOwners(overflowTestQueueID, maxPendingOwnersPerQueue) // call a second time to validate on empty queue
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
		fastTaskService.RemovePendingOwner(removal.queueID, removal.taskID)

		assert.Equal(t, removal.expectedQueueOwnersCount, len(fastTaskService.pendingTaskOwners))
		totalOwnerCount := 0
		for _, queueOwners := range fastTaskService.pendingTaskOwners {
			totalOwnerCount += len(queueOwners)
		}
		assert.Equal(t, removal.totalOwnerCount, totalOwnerCount)
	}
}
