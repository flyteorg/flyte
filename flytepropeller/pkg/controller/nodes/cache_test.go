package nodes

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	catalogmocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	executorsmocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	interfacesmocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces/mocks"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var (
	currentAttempt       = uint32(0)
	nodeID               = "baz"
	nodeOutputDir        = storage.DataReference("output_directory")
	parentUniqueID       = "bar"
	parentCurrentAttempt = uint32(1)
	uniqueID             = "foo"
)

type mockTaskReader struct {
	taskTemplate *core.TaskTemplate
}

func (t mockTaskReader) Read(ctx context.Context) (*core.TaskTemplate, error) {
	return t.taskTemplate, nil
}
func (t mockTaskReader) GetTaskType() v1alpha1.TaskType { return "" }
func (t mockTaskReader) GetTaskID() *core.Identifier    { return nil }

func setupCacheableNodeExecutionContext(dataStore *storage.DataStore, taskTemplate *core.TaskTemplate) *nodeExecContext {
	mockNode := &mocks.ExecutableNode{}
	mockNode.OnGetIDMatch(mock.Anything).Return(nodeID)
	mockNode.OnIsCacheableMatch().Return(nil)
	mockNode.OnGetCacheVersionMatch().Return(nil)
	mockNode.OnIsCacheSerializableMatch().Return(nil)

	mockNodeStatus := &mocks.ExecutableNodeStatus{}
	mockNodeStatus.OnGetAttemptsMatch().Return(currentAttempt)
	mockNodeStatus.OnGetOutputDir().Return(nodeOutputDir)

	mockParentInfo := &executorsmocks.ImmutableParentInfo{}
	mockParentInfo.OnCurrentAttemptMatch().Return(parentCurrentAttempt)
	mockParentInfo.OnGetUniqueIDMatch().Return(uniqueID)

	mockExecutionContext := &executorsmocks.ExecutionContext{}
	mockExecutionContext.OnGetParentInfo().Return(mockParentInfo)
	mockExecutionContext.OnGetExecutionConfig().Return(v1alpha1.ExecutionConfig{})

	mockNodeExecutionMetadata := &interfacesmocks.NodeExecutionMetadata{}
	mockNodeExecutionMetadata.OnGetOwnerID().Return(
		types.NamespacedName{
			Name: parentUniqueID,
		},
	)
	mockNodeExecutionMetadata.OnGetNodeExecutionIDMatch().Return(
		&core.NodeExecutionIdentifier{
			NodeId: nodeID,
		},
	)
	mockNodeExecutionMetadata.OnGetConsoleURL().Return("")

	mockNodeLookup := &executorsmocks.NodeLookup{}
	mockNodeLookup.OnGetNodeMatch(mock.Anything).Return(nil, false)

	var taskReader interfaces.TaskReader
	if taskTemplate != nil {
		taskReader = mockTaskReader{
			taskTemplate: taskTemplate,
		}
	}

	return &nodeExecContext{
		ic:         mockExecutionContext,
		md:         mockNodeExecutionMetadata,
		nl:         mockNodeLookup,
		node:       mockNode,
		nodeStatus: mockNodeStatus,
		store:      dataStore,
		tr:         taskReader,
	}
}

func TestComputeCatalogReservationOwnerID(t *testing.T) {
	nCtx := setupCacheableNodeExecutionContext(nil, nil)

	ownerID, err := computeCatalogReservationOwnerID(nCtx)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%s-%s-%d-%s-%d", parentUniqueID, uniqueID, parentCurrentAttempt, nodeID, currentAttempt), ownerID)
}

func TestUpdatePhaseCacheInfo(t *testing.T) {
	cacheStatus := catalog.NewStatus(core.CatalogCacheStatus_CACHE_MISS, nil)
	reservationStatus := core.CatalogReservation_RESERVATION_EXISTS

	tests := []struct {
		name              string
		cacheStatus       *catalog.Status
		reservationStatus *core.CatalogReservation_Status
	}{
		{"BothEmpty", nil, nil},
		{"CacheStatusOnly", &cacheStatus, nil},
		{"ReservationStatusOnly", nil, &reservationStatus},
		{"BothPopulated", &cacheStatus, &reservationStatus},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			phaseInfo := handler.PhaseInfoUndefined
			phaseInfo = updatePhaseCacheInfo(phaseInfo, test.cacheStatus, test.reservationStatus)

			// do not create ExecutionInfo object if neither cacheStatus or reservationStatus exists
			if test.cacheStatus == nil && test.reservationStatus == nil {
				assert.Nil(t, phaseInfo.GetInfo())
			}

			// ensure cache and reservation status' are being set correctly
			if test.cacheStatus != nil {
				assert.Equal(t, cacheStatus.GetCacheStatus(), phaseInfo.GetInfo().TaskNodeInfo.TaskNodeMetadata.CacheStatus)
			}

			if test.reservationStatus != nil {
				assert.Equal(t, reservationStatus, phaseInfo.GetInfo().TaskNodeInfo.TaskNodeMetadata.ReservationStatus)
			}
		})
	}
}

func TestCheckCatalogCache(t *testing.T) {
	tests := []struct {
		name                string
		cacheEntry          catalog.Entry
		cacheError          error
		catalogKey          catalog.Key
		expectedCacheStatus core.CatalogCacheStatus
		assertOutputFile    bool
		outputFileExists    bool
	}{
		{
			"CacheMiss",
			catalog.Entry{},
			status.Error(codes.NotFound, ""),
			catalog.Key{},
			core.CatalogCacheStatus_CACHE_MISS,
			false,
			false,
		},
		{
			"CacheHitWithOutputs",
			catalog.NewCatalogEntry(
				ioutils.NewInMemoryOutputReader(&core.LiteralMap{}, nil, nil),
				catalog.NewStatus(core.CatalogCacheStatus_CACHE_HIT, nil),
			),
			nil,
			catalog.Key{
				TypedInterface: core.TypedInterface{
					Outputs: &core.VariableMap{
						Variables: map[string]*core.Variable{
							"foo": nil,
						},
					},
				},
			},
			core.CatalogCacheStatus_CACHE_HIT,
			true,
			true,
		},
		{
			"CacheHitWithoutOutputs",
			catalog.NewCatalogEntry(
				nil,
				catalog.NewStatus(core.CatalogCacheStatus_CACHE_HIT, nil),
			),
			nil,
			catalog.Key{},
			core.CatalogCacheStatus_CACHE_HIT,
			true,
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testScope := promutils.NewTestScope()
			metrics := &nodeMetrics{
				catalogHitCount:                labeled.NewCounter("discovery_hit_count", "Task cached in Discovery", testScope),
				catalogMissCount:               labeled.NewCounter("discovery_miss_count", "Task not cached in Discovery", testScope),
				catalogSkipCount:               labeled.NewCounter("discovery_skip_count", "Task cached skipped in Discovery", testScope),
				catalogPutSuccessCount:         labeled.NewCounter("discovery_put_success_count", "Discovery Put success count", testScope),
				catalogPutFailureCount:         labeled.NewCounter("discovery_put_failure_count", "Discovery Put failure count", testScope),
				catalogGetFailureCount:         labeled.NewCounter("discovery_get_failure_count", "Discovery Get failure count", testScope),
				reservationGetFailureCount:     labeled.NewCounter("reservation_get_failure_count", "Reservation GetOrExtend failure count", testScope),
				reservationGetSuccessCount:     labeled.NewCounter("reservation_get_success_count", "Reservation GetOrExtend success count", testScope),
				reservationReleaseFailureCount: labeled.NewCounter("reservation_release_failure_count", "Reservation Release failure count", testScope),
				reservationReleaseSuccessCount: labeled.NewCounter("reservation_release_success_count", "Reservation Release success count", testScope),
			}

			cacheableHandler := &interfacesmocks.CacheableNodeHandler{}
			cacheableHandler.OnGetCatalogKeyMatch(mock.Anything, mock.Anything).Return(test.catalogKey, nil)

			catalogClient := &catalogmocks.Client{}
			catalogClient.OnGetMatch(mock.Anything, mock.Anything).Return(test.cacheEntry, test.cacheError)

			dataStore, err := storage.NewDataStore(
				&storage.Config{
					Type: storage.TypeMemory,
				},
				testScope.NewSubScope("data_store"),
			)
			assert.NoError(t, err)

			nodeExecutor := &nodeExecutor{
				cache:   catalogClient,
				metrics: metrics,
			}
			nCtx := setupCacheableNodeExecutionContext(dataStore, nil)

			// execute catalog cache check
			cacheEntry, err := nodeExecutor.CheckCatalogCache(context.TODO(), nCtx, cacheableHandler)
			assert.NoError(t, err)

			// validate the result cache entry status
			assert.Equal(t, test.expectedCacheStatus, cacheEntry.GetStatus().GetCacheStatus())

			if test.assertOutputFile {
				// assert the outputs file exists
				outputFile := v1alpha1.GetOutputsFile(nCtx.NodeStatus().GetOutputDir())
				metadata, err := nCtx.DataStore().Head(context.TODO(), outputFile)
				assert.NoError(t, err)
				assert.Equal(t, test.outputFileExists, metadata.Exists())
			}
		})
	}
}

func TestIsCacheableWithOverrides(t *testing.T) {
	cacheable := true
	cacheSerializable := true

	// initialize mocks
	cacheableHandler := &interfacesmocks.CacheableNodeHandler{}
	cacheableHandler.OnIsCacheableMatch(mock.Anything, mock.Anything, mock.Anything).Return(false, false, nil)

	nCtx := &interfacesmocks.NodeExecutionContext{}

	mockNode := &mocks.ExecutableNode{}
	mockNode.OnIsCacheableMatch().Return(&cacheable)
	mockNode.OnIsCacheSerializableMatch().Return(&cacheSerializable)
	nCtx.OnNode().Return(mockNode)

	// evaluate isCacheableWithOverrides to override the cacheable and cacheSerializable values
	cacheableResult, cacheSerializableResult, err := isCacheableWithOverrides(context.TODO(), nCtx, cacheableHandler)
	assert.NoError(t, err)

	// validate the result
	assert.Equal(t, cacheable, cacheableResult)
	assert.Equal(t, cacheSerializable, cacheSerializableResult)
}

func TestGetCatalogKeyWithOverrides(t *testing.T) {
	cacheVersion := "1"
	catalogKey := catalog.Key{
		CacheVersion: "0",
	}

	// initialize mocks
	cacheableHandler := &interfacesmocks.CacheableNodeHandler{}
	cacheableHandler.OnGetCatalogKeyMatch(mock.Anything, mock.Anything, mock.Anything).Return(catalogKey, nil)

	nCtx := &interfacesmocks.NodeExecutionContext{}

	mockNode := &mocks.ExecutableNode{}
	mockNode.OnGetCacheVersionMatch().Return(&cacheVersion)
	nCtx.OnNode().Return(mockNode)

	// evaluate getCatalogKey to override the CacheVersion value
	catalogKeyResult, err := getCatalogKeyWithOverrides(context.TODO(), nCtx, cacheableHandler)
	assert.NoError(t, err)

	// validate the result
	assert.Equal(t, cacheVersion, catalogKeyResult.CacheVersion)
}

func TestGetOrExtendCatalogReservation(t *testing.T) {

	heartBeatInterval := time.Second * 30
	tests := []struct {
		name                      string
		reservationOwnerID        string
		expectedReservationStatus core.CatalogReservation_Status
		reservationCacheStatus    catalog.ReservationCache
		notExpectClientCall       bool
	}{
		{
			name:                      "Acquired",
			reservationOwnerID:        "bar-foo-1-baz-0",
			expectedReservationStatus: core.CatalogReservation_RESERVATION_ACQUIRED,
			reservationCacheStatus:    catalog.ReservationCache{},
		},
		{
			name:                      "Exists",
			reservationOwnerID:        "some-other-owner",
			expectedReservationStatus: core.CatalogReservation_RESERVATION_EXISTS,
			reservationCacheStatus:    catalog.ReservationCache{},
		},
		{
			name: "Cache hit - acquired",
			reservationCacheStatus: catalog.ReservationCache{
				Timestamp:         time.Now(),
				ReservationStatus: core.CatalogReservation_RESERVATION_ACQUIRED,
			},
			expectedReservationStatus: core.CatalogReservation_RESERVATION_ACQUIRED,
			notExpectClientCall:       true,
		},
		{
			name: "Cache hit - exists",
			reservationCacheStatus: catalog.ReservationCache{
				Timestamp:         time.Now(),
				ReservationStatus: core.CatalogReservation_RESERVATION_EXISTS,
			},
			expectedReservationStatus: core.CatalogReservation_RESERVATION_EXISTS,
			notExpectClientCall:       true,
		},
		{
			name:               "Cache expired",
			reservationOwnerID: "bar-foo-1-baz-0",
			reservationCacheStatus: catalog.ReservationCache{
				Timestamp:         time.Now().Add(-3 * heartBeatInterval),
				ReservationStatus: core.CatalogReservation_RESERVATION_EXISTS,
			},
			expectedReservationStatus: core.CatalogReservation_RESERVATION_ACQUIRED,
			notExpectClientCall:       false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testScope := promutils.NewTestScope()
			metrics := &nodeMetrics{
				catalogHitCount:                labeled.NewCounter("discovery_hit_count", "Task cached in Discovery", testScope),
				catalogMissCount:               labeled.NewCounter("discovery_miss_count", "Task not cached in Discovery", testScope),
				catalogSkipCount:               labeled.NewCounter("discovery_skip_count", "Task cached skipped in Discovery", testScope),
				catalogPutSuccessCount:         labeled.NewCounter("discovery_put_success_count", "Discovery Put success count", testScope),
				catalogPutFailureCount:         labeled.NewCounter("discovery_put_failure_count", "Discovery Put failure count", testScope),
				catalogGetFailureCount:         labeled.NewCounter("discovery_get_failure_count", "Discovery Get failure count", testScope),
				reservationGetFailureCount:     labeled.NewCounter("reservation_get_failure_count", "Reservation GetOrExtend failure count", testScope),
				reservationGetSuccessCount:     labeled.NewCounter("reservation_get_success_count", "Reservation GetOrExtend success count", testScope),
				reservationReleaseFailureCount: labeled.NewCounter("reservation_release_failure_count", "Reservation Release failure count", testScope),
				reservationReleaseSuccessCount: labeled.NewCounter("reservation_release_success_count", "Reservation Release success count", testScope),
			}

			cacheableHandler := &interfacesmocks.CacheableNodeHandler{}
			cacheableHandler.OnGetCatalogKeyMatch(mock.Anything, mock.Anything).Return(catalog.Key{}, nil)

			catalogClient := &catalogmocks.Client{}
			catalogClient.OnGetOrExtendReservationMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
				&datacatalog.Reservation{
					OwnerId: test.reservationOwnerID,
				},
				nil,
			)
			catalogClient.OnGetReservationCache(mock.Anything).Return(test.reservationCacheStatus)
			catalogClient.On("UpdateReservationCache", mock.Anything, mock.Anything).Return()

			nodeExecutor := &nodeExecutor{
				cache:   catalogClient,
				metrics: metrics,
			}
			nCtx := setupCacheableNodeExecutionContext(nil, &core.TaskTemplate{})

			// execute catalog cache check
			reservationStatus, err := nodeExecutor.GetOrExtendCatalogReservation(context.TODO(), nCtx, cacheableHandler, heartBeatInterval)
			assert.NoError(t, err)

			// validate the result cache entry status
			assert.Equal(t, test.expectedReservationStatus, reservationStatus)

			if test.notExpectClientCall {
				catalogClient.AssertNotCalled(t, "GetOrExtendReservation", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			} else {
				catalogClient.AssertCalled(t, "GetOrExtendReservation", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}

func TestReleaseCatalogReservation(t *testing.T) {
	tests := []struct {
		name                      string
		releaseError              error
		expectedReservationStatus core.CatalogReservation_Status
	}{
		{
			"Success",
			nil,
			core.CatalogReservation_RESERVATION_RELEASED,
		},
		{
			"Failure",
			errors.New("failed to release"),
			core.CatalogReservation_RESERVATION_FAILURE,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testScope := promutils.NewTestScope()
			metrics := &nodeMetrics{
				catalogHitCount:                labeled.NewCounter("discovery_hit_count", "Task cached in Discovery", testScope),
				catalogMissCount:               labeled.NewCounter("discovery_miss_count", "Task not cached in Discovery", testScope),
				catalogSkipCount:               labeled.NewCounter("discovery_skip_count", "Task cached skipped in Discovery", testScope),
				catalogPutSuccessCount:         labeled.NewCounter("discovery_put_success_count", "Discovery Put success count", testScope),
				catalogPutFailureCount:         labeled.NewCounter("discovery_put_failure_count", "Discovery Put failure count", testScope),
				catalogGetFailureCount:         labeled.NewCounter("discovery_get_failure_count", "Discovery Get failure count", testScope),
				reservationGetFailureCount:     labeled.NewCounter("reservation_get_failure_count", "Reservation GetOrExtend failure count", testScope),
				reservationGetSuccessCount:     labeled.NewCounter("reservation_get_success_count", "Reservation GetOrExtend success count", testScope),
				reservationReleaseFailureCount: labeled.NewCounter("reservation_release_failure_count", "Reservation Release failure count", testScope),
				reservationReleaseSuccessCount: labeled.NewCounter("reservation_release_success_count", "Reservation Release success count", testScope),
			}

			cacheableHandler := &interfacesmocks.CacheableNodeHandler{}
			cacheableHandler.OnGetCatalogKeyMatch(mock.Anything, mock.Anything).Return(catalog.Key{}, nil)

			catalogClient := &catalogmocks.Client{}
			catalogClient.OnReleaseReservationMatch(mock.Anything, mock.Anything, mock.Anything).Return(test.releaseError)

			nodeExecutor := &nodeExecutor{
				cache:   catalogClient,
				metrics: metrics,
			}
			nCtx := setupCacheableNodeExecutionContext(nil, &core.TaskTemplate{})

			// execute catalog cache check
			reservationEntry, err := nodeExecutor.ReleaseCatalogReservation(context.TODO(), nCtx, cacheableHandler)
			if test.releaseError == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}

			// validate the result cache entry status
			assert.Equal(t, test.expectedReservationStatus, reservationEntry.GetStatus())
		})
	}
}

func TestWriteCatalogCache(t *testing.T) {
	tests := []struct {
		name                string
		cacheStatus         catalog.Status
		cacheError          error
		catalogKey          catalog.Key
		expectedCacheStatus core.CatalogCacheStatus
	}{
		{
			"NoOutputs",
			catalog.NewStatus(core.CatalogCacheStatus_CACHE_DISABLED, nil),
			nil,
			catalog.Key{},
			core.CatalogCacheStatus_CACHE_DISABLED,
		},
		{
			"OutputsExist",
			catalog.NewStatus(core.CatalogCacheStatus_CACHE_POPULATED, nil),
			nil,
			catalog.Key{
				TypedInterface: core.TypedInterface{
					Outputs: &core.VariableMap{
						Variables: map[string]*core.Variable{
							"foo": nil,
						},
					},
				},
			},
			core.CatalogCacheStatus_CACHE_POPULATED,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testScope := promutils.NewTestScope()
			metrics := &nodeMetrics{
				catalogHitCount:                labeled.NewCounter("discovery_hit_count", "Task cached in Discovery", testScope),
				catalogMissCount:               labeled.NewCounter("discovery_miss_count", "Task not cached in Discovery", testScope),
				catalogSkipCount:               labeled.NewCounter("discovery_skip_count", "Task cached skipped in Discovery", testScope),
				catalogPutSuccessCount:         labeled.NewCounter("discovery_put_success_count", "Discovery Put success count", testScope),
				catalogPutFailureCount:         labeled.NewCounter("discovery_put_failure_count", "Discovery Put failure count", testScope),
				catalogGetFailureCount:         labeled.NewCounter("discovery_get_failure_count", "Discovery Get failure count", testScope),
				reservationGetFailureCount:     labeled.NewCounter("reservation_get_failure_count", "Reservation GetOrExtend failure count", testScope),
				reservationGetSuccessCount:     labeled.NewCounter("reservation_get_success_count", "Reservation GetOrExtend success count", testScope),
				reservationReleaseFailureCount: labeled.NewCounter("reservation_release_failure_count", "Reservation Release failure count", testScope),
				reservationReleaseSuccessCount: labeled.NewCounter("reservation_release_success_count", "Reservation Release success count", testScope),
			}

			cacheableHandler := &interfacesmocks.CacheableNodeHandler{}
			cacheableHandler.OnGetCatalogKeyMatch(mock.Anything, mock.Anything).Return(test.catalogKey, nil)

			catalogClient := &catalogmocks.Client{}
			catalogClient.OnPutMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(test.cacheStatus, nil)

			dataStore, err := storage.NewDataStore(
				&storage.Config{
					Type: storage.TypeMemory,
				},
				testScope.NewSubScope("data_store"),
			)
			assert.NoError(t, err)

			nodeExecutor := &nodeExecutor{
				cache:   catalogClient,
				metrics: metrics,
			}
			nCtx := setupCacheableNodeExecutionContext(dataStore, &core.TaskTemplate{})

			// execute catalog cache check
			cacheStatus, err := nodeExecutor.WriteCatalogCache(context.TODO(), nCtx, cacheableHandler)
			assert.NoError(t, err)

			// validate the result cache entry status
			assert.Equal(t, test.expectedCacheStatus, cacheStatus.GetCacheStatus())
		})
	}
}
