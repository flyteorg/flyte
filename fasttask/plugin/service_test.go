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
				activeTasks: &sync.Map{},
				workers:     &workersMap,
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
		activeTasks: &sync.Map{},
		workers:     &workers1,
	})

	env2 := newEnvironmentStore()
	workers2 := sync.Map{}
	workers2.Store("w0", &workerImpl{id: "w0"})
	workers2.Store("w1", &workerImpl{id: "w1"})
	env2.GetOrCreate("foo", &environmentImpl{
		activeTasks: &sync.Map{},
		workers:     &workers2,
	})

	env3 := newEnvironmentStore()
	workers3 := sync.Map{}
	workers3.Store("w0", &workerImpl{id: "w0"})
	workers3.Store("w1", &workerImpl{id: "w1"})
	env3.GetOrCreate("foo", &environmentImpl{
		activeTasks: &sync.Map{},
		workers:     &workers3,
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
			name:                    "WorkerDoesNotExist",
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
		activeTasks: &sync.Map{},
		workers:     &workers1,
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
		activeTasks: &sync.Map{},
		workers:     &workers2,
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
		activeTasks: &sync.Map{},
		workers:     &workers3,
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

			var taskCountBefore int
			if env := test.envStore.Get(test.queueID); env != nil {
				taskCountBefore = env.GetActiveTaskCount()
			}

			worker, err := fastTaskService.OfferTaskToEnvironment(ctx, execID, test.queueID, test.taskID, test.namespace, test.workflowID, []string{}, make(map[string]string), enqueueLabels)
			if test.expectedError != nil {
				assert.Nil(t, worker)
				assert.Equal(t, test.expectedError, err)
				return
			}
			assert.Equal(t, test.expectedWorkerID, worker.ID())
			assert.Equal(t, test.expectedError, err)

			if env := test.envStore.Get(test.queueID); env != nil {
				assert.Equal(t, taskCountBefore+1, env.GetActiveTaskCount())
			}

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

func newTestEnvironment(workers ...interfaces.Worker) *environmentImpl {
	workersMap := &sync.Map{}
	for _, w := range workers {
		workersMap.Store(w.ID(), w)
	}
	return &environmentImpl{
		activeTasks: &sync.Map{},
		workers:     workersMap,
	}
}

func newTestService(store interfaces.EnvironmentStore, enqueueOwner func(map[string]string) error) *fastTaskServiceImpl {
	scope := promutils.NewTestScope()
	kubeClientImpl := &pluginsCoreMock.KubeClient{}
	kubeClientImpl.OnGetClient().Return(&kubeClient{})
	kubeClientImpl.OnGetCache().Return(&kubeCache{})
	builder := &environmentBuilderImpl{
		kubeClient:  kubeClientImpl,
		store:       store,
		metrics:     newBuilderMetrics(scope),
		randSource:  rand.New(rand.NewSource(time.Now().UnixNano())),
		scaleUpChan: make(chan string, 1),
	}
	return newFastTaskService(enqueueOwner, builder, store, scope, testConfig)
}

var noopEnqueueOwner = func(labels map[string]string) error { return nil }

func TestProcessHeartbeatAfterCleanupDoesNotLeak(t *testing.T) {
	ctx := context.TODO()
	queueID := "env-1"
	taskID := "task-1"
	workerID := "w0"

	worker := &workerImpl{
		id:           workerID,
		capacity:     &pb.Capacity{ExecutionCount: 0, ExecutionLimit: 1},
		responseChan: make(chan *pb.HeartbeatResponse, 4),
	}

	store := newEnvironmentStore()
	env := newTestEnvironment(worker)
	store.GetOrCreate(queueID, env)

	svc := newTestService(store, noopEnqueueOwner)

	// Offer task — registers it and creates status channel
	execID := &coreIdl.WorkflowExecutionIdentifier{Project: "p", Domain: "d", Name: "n"}
	_, err := svc.OfferTaskToEnvironment(ctx, execID, queueID, taskID, "ns", "wf", nil, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, env.GetActiveTaskCount())

	// Cleanup — unregisters task and deletes status channel
	err = svc.Cleanup(ctx, taskID, queueID, workerID)
	assert.NoError(t, err)
	assert.Equal(t, 0, env.GetActiveTaskCount())

	// Drain the initial DELETE sent by Cleanup itself.
	drainResponses(worker.responseChan)

	// Simulate a post-Cleanup heartbeat that still reports the old task.
	// This must NOT re-register the task (would permanently inflate demand).
	hbReq := &pb.HeartbeatRequest{
		QueueId:  queueID,
		WorkerId: workerID,
		TaskStatuses: []*pb.TaskStatus{
			{
				TaskId: taskID,
				Phase:  int32(core.PhaseSuccess),
			},
		},
	}
	svc.processHeartbeatTaskStatuses(hbReq, env, env.GetWorker(workerID))

	assert.Equal(t, 0, env.GetActiveTaskCount(), "post-Cleanup heartbeat must not re-register task")

	// Verify DELETE was re-sent because the tombstone is still alive.
	resp := drainResponses(worker.responseChan)
	assert.Len(t, resp, 1, "should re-send DELETE on first post-Cleanup heartbeat")
	assert.Equal(t, pb.HeartbeatResponse_DELETE, resp[0].Operation)
	assert.Equal(t, taskID, resp[0].TaskId)

	// Send a second heartbeat — DELETE should be re-sent again (not one-shot).
	svc.processHeartbeatTaskStatuses(hbReq, env, env.GetWorker(workerID))

	resp = drainResponses(worker.responseChan)
	assert.Len(t, resp, 1, "should re-send DELETE on subsequent heartbeats while tombstone is alive")
	assert.Equal(t, pb.HeartbeatResponse_DELETE, resp[0].Operation)
}

// cleanupRacingEnv wraps an Environment and simulates Cleanup running during
// RegisterTask — the exact interleaving the double-check guards against.
type cleanupRacingEnv struct {
	interfaces.Environment
	cleanupOnce     sync.Once
	simulateCleanup func()
}

func (e *cleanupRacingEnv) RegisterTask(taskID string) {
	e.cleanupOnce.Do(e.simulateCleanup)
	e.Environment.RegisterTask(taskID)
}

func TestProcessHeartbeatDoubleCheckPreventsLeak(t *testing.T) {
	queueID := "env-1"
	taskID := "task-1"
	workerID := "w0"

	worker := &workerImpl{
		id:           workerID,
		capacity:     &pb.Capacity{ExecutionCount: 0, ExecutionLimit: 1},
		responseChan: make(chan *pb.HeartbeatResponse, 4),
	}

	realEnv := newTestEnvironment(worker)
	store := newEnvironmentStore()
	store.GetOrCreate(queueID, realEnv)
	svc := newTestService(store, noopEnqueueOwner)

	// Set up active task state: channel exists, task registered
	svc.taskStatusChannels.Store(taskID, make(chan *workerTaskStatus, 1))
	realEnv.RegisterTask(taskID)
	assert.Equal(t, 1, realEnv.GetActiveTaskCount())

	// Create env wrapper that simulates Cleanup running during RegisterTask.
	raceEnv := &cleanupRacingEnv{
		Environment: realEnv,
		simulateCleanup: func() {
			svc.taskStatusChannels.Delete(taskID)
			realEnv.UnregisterTask(taskID)
		},
	}

	svc.processHeartbeatTaskStatuses(&pb.HeartbeatRequest{
		QueueId:  queueID,
		WorkerId: workerID,
		TaskStatuses: []*pb.TaskStatus{
			{
				TaskId: taskID,
				Phase:  int32(core.PhaseSuccess),
			},
		},
	}, raceEnv, worker)

	assert.Equal(t, 0, realEnv.GetActiveTaskCount(),
		"double-check must undo RegisterTask when Cleanup deletes the channel concurrently")
}

func TestCleanedUpTaskTombstoneExpiry(t *testing.T) {
	ctx := context.TODO()
	queueID := "env-1"
	taskID := "task-1"
	workerID := "w0"

	worker := &workerImpl{
		id:           workerID,
		capacity:     &pb.Capacity{ExecutionCount: 0, ExecutionLimit: 1},
		responseChan: make(chan *pb.HeartbeatResponse, 4),
	}

	store := newEnvironmentStore()
	env := newTestEnvironment(worker)
	store.GetOrCreate(queueID, env)

	svc := newTestService(store, noopEnqueueOwner)

	// Offer and cleanup
	execID := &coreIdl.WorkflowExecutionIdentifier{Project: "p", Domain: "d", Name: "n"}
	_, err := svc.OfferTaskToEnvironment(ctx, execID, queueID, taskID, "ns", "wf", nil, nil, nil)
	assert.NoError(t, err)
	err = svc.Cleanup(ctx, taskID, queueID, workerID)
	assert.NoError(t, err)
	drainResponses(worker.responseChan)

	hbReq := &pb.HeartbeatRequest{
		QueueId:  queueID,
		WorkerId: workerID,
		TaskStatuses: []*pb.TaskStatus{
			{
				TaskId: taskID,
				Phase:  int32(core.PhaseSuccess),
			},
		},
	}

	// Tombstone alive — DELETE re-sent
	svc.processHeartbeatTaskStatuses(hbReq, env, env.GetWorker(workerID))
	resp := drainResponses(worker.responseChan)
	assert.Len(t, resp, 1, "DELETE should be re-sent while tombstone is alive")

	// Simulate tombstone expiry
	svc.cleanedUpTasks.Delete(taskID)

	// Tombstone gone — DELETE must NOT be re-sent
	svc.processHeartbeatTaskStatuses(hbReq, env, env.GetWorker(workerID))
	resp = drainResponses(worker.responseChan)
	assert.Len(t, resp, 0, "DELETE should not be re-sent after tombstone expires")

	assert.Equal(t, 0, env.GetActiveTaskCount(), "task must not be re-registered")
}

func TestOfferTaskConcurrentDoesNotOverAssign(t *testing.T) {
	ctx := context.TODO()
	queueID := "env-1"
	workerID := "w0"

	worker := &workerImpl{
		id:           workerID,
		capacity:     &pb.Capacity{ExecutionCount: 0, ExecutionLimit: 1},
		responseChan: make(chan *pb.HeartbeatResponse, 200),
	}

	store := newEnvironmentStore()
	env := newTestEnvironment(worker)
	store.GetOrCreate(queueID, env)
	svc := newTestService(store, noopEnqueueOwner)

	execID := &coreIdl.WorkflowExecutionIdentifier{Project: "p", Domain: "d", Name: "n"}
	const goroutines = 100
	results := make(chan error, goroutines)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		go func(idx int) {
			defer wg.Done()
			taskID := fmt.Sprintf("task-%d", idx)
			_, err := svc.OfferTaskToEnvironment(ctx, execID, queueID, taskID, "ns", "wf", nil, nil, nil)
			results <- err
		}(i)
	}
	wg.Wait()
	close(results)

	successes := 0
	failures := 0
	for err := range results {
		if err == nil {
			successes++
		} else {
			assert.ErrorIs(t, err, noCapacityAvailableError)
			failures++
		}
	}
	assert.Equal(t, 1, successes, "exactly one goroutine should succeed")
	assert.Equal(t, goroutines-1, failures)
}

func TestOfferTaskAccountsForPendingReservations(t *testing.T) {
	ctx := context.TODO()
	queueID := "env-1"
	workerID := "w0"

	worker := &workerImpl{
		id:           workerID,
		capacity:     &pb.Capacity{ExecutionCount: 0, ExecutionLimit: 2},
		responseChan: make(chan *pb.HeartbeatResponse, 10),
	}

	store := newEnvironmentStore()
	env := newTestEnvironment(worker)
	store.GetOrCreate(queueID, env)
	svc := newTestService(store, noopEnqueueOwner)

	execID := &coreIdl.WorkflowExecutionIdentifier{Project: "p", Domain: "d", Name: "n"}

	// First and second offers succeed (pending=1, then pending=2)
	_, err := svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-1", "ns", "wf", nil, nil, nil)
	assert.NoError(t, err)
	_, err = svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-2", "ns", "wf", nil, nil, nil)
	assert.NoError(t, err)

	// Third offer fails (reported 0 + pending 2 >= limit 2)
	_, err = svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-3", "ns", "wf", nil, nil, nil)
	assert.ErrorIs(t, err, noCapacityAvailableError)

	// Cleanup first task (removes from pending → pending=1)
	err = svc.Cleanup(ctx, "task-1", queueID, workerID)
	assert.NoError(t, err)

	// Fourth offer succeeds (reported 0 + pending 1 < limit 2)
	_, err = svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-4", "ns", "wf", nil, nil, nil)
	assert.NoError(t, err)
}

func TestHeartbeatDoesNotBreakReservations(t *testing.T) {
	ctx := context.TODO()
	queueID := "env-1"
	workerID := "w0"

	worker := &workerImpl{
		id:           workerID,
		capacity:     &pb.Capacity{ExecutionCount: 0, ExecutionLimit: 1},
		responseChan: make(chan *pb.HeartbeatResponse, 4),
	}

	store := newEnvironmentStore()
	env := newTestEnvironment(worker)
	store.GetOrCreate(queueID, env)
	svc := newTestService(store, noopEnqueueOwner)

	execID := &coreIdl.WorkflowExecutionIdentifier{Project: "p", Domain: "d", Name: "n"}

	// OfferTask succeeds (pending=1)
	_, err := svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-1", "ns", "wf", nil, nil, nil)
	assert.NoError(t, err)

	// Heartbeat arrives with exec_count=0 (worker hasn't processed ASSIGN yet).
	// This overwrites worker capacity but must NOT erase the pending reservation.
	worker.SetCapacity(&pb.Capacity{ExecutionCount: 0, ExecutionLimit: 1})

	// Second OfferTask must fail (reported 0 + pending 1 >= limit 1)
	_, err = svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-2", "ns", "wf", nil, nil, nil)
	assert.ErrorIs(t, err, noCapacityAvailableError, "heartbeat overwrite must not create double-booking")
}

func TestPendingReservationRemovedOnCleanup(t *testing.T) {
	ctx := context.TODO()
	queueID := "env-1"
	workerID := "w0"

	worker := &workerImpl{
		id:           workerID,
		capacity:     &pb.Capacity{ExecutionCount: 0, ExecutionLimit: 1},
		responseChan: make(chan *pb.HeartbeatResponse, 4),
	}

	store := newEnvironmentStore()
	env := newTestEnvironment(worker)
	store.GetOrCreate(queueID, env)
	svc := newTestService(store, noopEnqueueOwner)

	execID := &coreIdl.WorkflowExecutionIdentifier{Project: "p", Domain: "d", Name: "n"}

	_, err := svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-1", "ns", "wf", nil, nil, nil)
	assert.NoError(t, err)

	// Pending=1, slot is full
	_, err = svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-2", "ns", "wf", nil, nil, nil)
	assert.ErrorIs(t, err, noCapacityAvailableError)

	// Cleanup removes pending reservation
	err = svc.Cleanup(ctx, "task-1", queueID, workerID)
	assert.NoError(t, err)

	// Slot is free again
	_, err = svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-3", "ns", "wf", nil, nil, nil)
	assert.NoError(t, err)
}

func TestHeartbeatConfirmationRemovesPendingReservation(t *testing.T) {
	ctx := context.TODO()
	queueID := "env-1"
	workerID := "w0"

	worker := &workerImpl{
		id:           workerID,
		capacity:     &pb.Capacity{ExecutionCount: 0, ExecutionLimit: 1},
		responseChan: make(chan *pb.HeartbeatResponse, 4),
	}

	store := newEnvironmentStore()
	env := newTestEnvironment(worker)
	store.GetOrCreate(queueID, env)
	svc := newTestService(store, noopEnqueueOwner)

	execID := &coreIdl.WorkflowExecutionIdentifier{Project: "p", Domain: "d", Name: "n"}

	_, err := svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-1", "ns", "wf", nil, nil, nil)
	assert.NoError(t, err)
	drainResponses(worker.responseChan)

	// Pending=1, slot is full
	_, err = svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-2", "ns", "wf", nil, nil, nil)
	assert.ErrorIs(t, err, noCapacityAvailableError)

	// Worker reports status for task-1 → confirms receipt, removes pending reservation.
	// Update worker capacity to reflect it's now running the task.
	worker.SetCapacity(&pb.Capacity{ExecutionCount: 1, ExecutionLimit: 1})
	svc.processHeartbeatTaskStatuses(&pb.HeartbeatRequest{
		QueueId:  queueID,
		WorkerId: workerID,
		TaskStatuses: []*pb.TaskStatus{
			{TaskId: "task-1", Phase: int32(core.PhaseRunning)},
		},
	}, env, worker)

	// Verify pending reservation was removed
	svc.pendingReservationsLock.Lock()
	pendingCount := len(svc.pendingReservations[queueID][workerID])
	svc.pendingReservationsLock.Unlock()
	assert.Equal(t, 0, pendingCount, "heartbeat status should confirm and remove pending reservation")

	// Slot still full because worker is actually running the task (exec_count=1)
	_, err = svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-3", "ns", "wf", nil, nil, nil)
	assert.ErrorIs(t, err, noCapacityAvailableError)
}

func TestOfferTaskSkipsFullWorkerPicksNext(t *testing.T) {
	ctx := context.TODO()
	queueID := "env-1"

	worker0 := &workerImpl{
		id:           "w0",
		capacity:     &pb.Capacity{ExecutionCount: 0, ExecutionLimit: 1},
		responseChan: make(chan *pb.HeartbeatResponse, 4),
	}
	worker1 := &workerImpl{
		id:           "w1",
		capacity:     &pb.Capacity{ExecutionCount: 0, ExecutionLimit: 1},
		responseChan: make(chan *pb.HeartbeatResponse, 4),
	}

	store := newEnvironmentStore()
	env := newTestEnvironment(worker0, worker1)
	store.GetOrCreate(queueID, env)
	svc := newTestService(store, noopEnqueueOwner)

	execID := &coreIdl.WorkflowExecutionIdentifier{Project: "p", Domain: "d", Name: "n"}

	// First offer goes to w0 (sorted alphabetically)
	w, err := svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-1", "ns", "wf", nil, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "w0", w.ID())

	// Second offer should skip w0 (pending=1) and go to w1
	w, err = svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-2", "ns", "wf", nil, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "w1", w.ID())
}

func TestDisconnectPreservesReservationsOrphanedPreventsAssignment(t *testing.T) {
	ctx := context.TODO()
	queueID := "env-1"
	workerID := "w0"

	worker := &workerImpl{
		id:           workerID,
		capacity:     &pb.Capacity{ExecutionCount: 0, ExecutionLimit: 1},
		responseChan: make(chan *pb.HeartbeatResponse, 4),
	}

	store := newEnvironmentStore()
	env := newTestEnvironment(worker)
	store.GetOrCreate(queueID, env)
	svc := newTestService(store, noopEnqueueOwner)

	execID := &coreIdl.WorkflowExecutionIdentifier{Project: "p", Domain: "d", Name: "n"}

	_, err := svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-1", "ns", "wf", nil, nil, nil)
	assert.NoError(t, err)

	// Simulate heartbeat disconnect: worker goes ORPHANED but reservations stay.
	// This prevents stale ASSIGN + new ASSIGN from causing over-assignment on reconnect.
	worker.SetState(interfaces.ORPHANED, "")

	svc.pendingReservationsLock.Lock()
	pendingCount := len(svc.pendingReservations[queueID][workerID])
	svc.pendingReservationsLock.Unlock()
	assert.Equal(t, 1, pendingCount, "pending reservation must survive disconnect")

	// ORPHANED worker is skipped by OfferTask — no capacity available
	_, err = svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-2", "ns", "wf", nil, nil, nil)
	assert.ErrorIs(t, err, noCapacityAvailableError, "ORPHANED worker must not receive new tasks")
}

func TestOfferTaskIdempotentRetry(t *testing.T) {
	ctx := context.TODO()
	queueID := "env-1"

	worker := &workerImpl{
		id:           "w0",
		capacity:     &pb.Capacity{ExecutionCount: 0, ExecutionLimit: 1},
		responseChan: make(chan *pb.HeartbeatResponse, 10),
	}

	store := newEnvironmentStore()
	env := newTestEnvironment(worker)
	store.GetOrCreate(queueID, env)
	svc := newTestService(store, noopEnqueueOwner)

	execID := &coreIdl.WorkflowExecutionIdentifier{Project: "p", Domain: "d", Name: "n"}

	// First offer succeeds and picks w0
	w, err := svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-1", "ns", "wf", nil, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "w0", w.ID())

	// Retry with the same taskID (simulates failure between offer and state persistence).
	// Must return the same worker, not fail with noCapacity.
	w, err = svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-1", "ns", "wf", nil, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "w0", w.ID(), "idempotent retry must return the same worker")

	// Verify only one reservation exists (no duplicate)
	svc.pendingReservationsLock.Lock()
	pendingCount := len(svc.pendingReservations[queueID]["w0"])
	svc.pendingReservationsLock.Unlock()
	assert.Equal(t, 1, pendingCount, "should have exactly one reservation, not a duplicate")
}

func TestOfferTaskIdempotentRetryStaleWorker(t *testing.T) {
	ctx := context.TODO()
	queueID := "env-1"

	worker0 := &workerImpl{
		id:           "w0",
		capacity:     &pb.Capacity{ExecutionCount: 0, ExecutionLimit: 1},
		responseChan: make(chan *pb.HeartbeatResponse, 10),
	}
	worker1 := &workerImpl{
		id:           "w1",
		capacity:     &pb.Capacity{ExecutionCount: 0, ExecutionLimit: 1},
		responseChan: make(chan *pb.HeartbeatResponse, 10),
	}

	store := newEnvironmentStore()
	env := newTestEnvironment(worker0, worker1)
	store.GetOrCreate(queueID, env)
	svc := newTestService(store, noopEnqueueOwner)

	execID := &coreIdl.WorkflowExecutionIdentifier{Project: "p", Domain: "d", Name: "n"}

	// First offer picks w0
	w, err := svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-1", "ns", "wf", nil, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "w0", w.ID())

	// w0 goes unhealthy
	worker0.SetState(interfaces.ORPHANED, "")

	// Retry same taskID — stale reservation on w0 should be removed, w1 selected
	w, err = svc.OfferTaskToEnvironment(ctx, execID, queueID, "task-1", "ns", "wf", nil, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "w1", w.ID(), "should select healthy worker after stale reservation removed")

	// Verify reservation moved to w1
	svc.pendingReservationsLock.Lock()
	w0Count := len(svc.pendingReservations[queueID]["w0"])
	w1Count := len(svc.pendingReservations[queueID]["w1"])
	svc.pendingReservationsLock.Unlock()
	assert.Equal(t, 0, w0Count, "stale reservation on w0 should be removed")
	assert.Equal(t, 1, w1Count, "reservation should now be on w1")
}

func TestCrossQueueReservationIsolation(t *testing.T) {
	ctx := context.TODO()
	queueA := "env-a"
	queueB := "env-b"

	// Simulate two queues that happen to share a worker ID (extremely unlikely in
	// production but the scoping must handle it correctly).
	workerA := &workerImpl{
		id:           "shared-worker-id",
		capacity:     &pb.Capacity{ExecutionCount: 0, ExecutionLimit: 1},
		responseChan: make(chan *pb.HeartbeatResponse, 10),
	}
	workerB := &workerImpl{
		id:           "shared-worker-id",
		capacity:     &pb.Capacity{ExecutionCount: 0, ExecutionLimit: 1},
		responseChan: make(chan *pb.HeartbeatResponse, 10),
	}

	store := newEnvironmentStore()
	envA := newTestEnvironment(workerA)
	envB := newTestEnvironment(workerB)
	store.GetOrCreate(queueA, envA)
	store.GetOrCreate(queueB, envB)
	svc := newTestService(store, noopEnqueueOwner)

	execID := &coreIdl.WorkflowExecutionIdentifier{Project: "p", Domain: "d", Name: "n"}

	// Fill queue A's worker
	_, err := svc.OfferTaskToEnvironment(ctx, execID, queueA, "task-a1", "ns", "wf", nil, nil, nil)
	assert.NoError(t, err)

	// Queue B's worker with the same ID must still have capacity
	_, err = svc.OfferTaskToEnvironment(ctx, execID, queueB, "task-b1", "ns", "wf", nil, nil, nil)
	assert.NoError(t, err, "reservation in queue A must not block queue B's identically-named worker")
}

// drainResponses reads all pending messages from a heartbeat response channel.
func drainResponses(ch chan *pb.HeartbeatResponse) []*pb.HeartbeatResponse {
	var out []*pb.HeartbeatResponse
	for {
		select {
		case r := <-ch:
			out = append(out, r)
		default:
			return out
		}
	}
}
