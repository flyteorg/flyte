package dynamic

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	outputReaderMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	flyteMocks "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	executorMocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors/mocks"
	futureCatalogMocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/catalog/mocks"
	dynamicHandlerMock "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/dynamic/mocks"
	nodeMocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces/mocks"
	lpMocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/subworkflow/launchplan/mocks"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

func TestComputeCatalogReservationOwnerID(t *testing.T) {

	t.Run("Successfully generate ownerID", func(t *testing.T) {
		// Setup test data
		mockCtx := &nodeMocks.NodeExecutionContext{}
		mockExecContext := &executorMocks.ExecutionContext{}
		mockNodeExecMetadata := &nodeMocks.NodeExecutionMetadata{}
		mockNodeExecMetadata.OnGetOwnerID().Return(types.NamespacedName{Namespace: "namespace", Name: "name"})

		// Set mock expectations
		mockParentInfo := &executorMocks.ImmutableParentInfo{}
		mockParentInfo.OnGetUniqueID().Return("id")
		mockParentInfo.OnCurrentAttempt().Return(1)
		mockParentInfo.On("GetExecutionID").Return(&core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		})
		mockExecContext.OnGetParentInfo().Return(mockParentInfo)
		mockCtx.OnExecutionContext().Return(mockExecContext)
		mockCtx.OnNodeID().Return("node-1")
		mockCtx.OnCurrentAttempt().Return(uint32(1))
		mockCtx.OnNodeExecutionMetadata().Return(mockNodeExecMetadata)

		// Execute test
		ownerID, err := computeCatalogReservationOwnerID(mockCtx)

		// Verify results
		assert.NoError(t, err)
		assert.NotEmpty(t, ownerID)
	})
}

func TestCheckCatalogFutureCache(t *testing.T) {
	ctx := context.Background()

	t.Run("getFutureCatalogEntry returns error", func(t *testing.T) {
		h := &dynamicHandlerMock.TaskNodeHandler{}
		expectedErr := errors.New("catalog entry error")
		h.OnGetCatalogKeyMatch(mock.Anything, mock.Anything).Return(catalog.Key{}, expectedErr)

		mockNodeExecutor := &nodeMocks.Node{}
		mockLPLauncher := &lpMocks.Reader{}
		mockFutureCatalog := &futureCatalogMocks.FutureCatalogClient{}

		nCtx := &nodeMocks.NodeExecutionContext{}
		d := New(h, mockNodeExecutor, mockLPLauncher, eventConfig, promutils.NewTestScope(), mockFutureCatalog)

		entry, err := d.(*dynamicNodeTaskNodeHandler).CheckCatalogFutureCache(ctx, nCtx, h)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedErr.Error())
		assert.Equal(t, catalog.Entry{}, entry)
	})

	t.Run("handleCatalogError returns correct error info", func(t *testing.T) {
		h := &dynamicHandlerMock.TaskNodeHandler{}
		mockNodeExecutor := &nodeMocks.Node{}
		mockLPLauncher := &lpMocks.Reader{}
		mockFutureCatalog := &futureCatalogMocks.FutureCatalogClient{}
		scope := promutils.NewTestScope()

		nCtx := &nodeMocks.NodeExecutionContext{}
		d := New(h, mockNodeExecutor, mockLPLauncher, eventConfig, scope, mockFutureCatalog)

		catalogKey := catalog.Key{
			Identifier: core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Name:         "task-1",
			},
			TypedInterface: core.TypedInterface{},
		}
		h.OnGetCatalogKeyMatch(mock.Anything, mock.Anything).Return(catalogKey, nil)

		expectedErr := errors.New("catalog get error")
		mockFutureCatalog.EXPECT().GetFuture(mock.Anything, mock.Anything).Return(catalog.Entry{}, expectedErr)

		entry, err := d.(*dynamicNodeTaskNodeHandler).CheckCatalogFutureCache(ctx, nCtx, h)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to check Catalog for previous results")
		assert.Contains(t, err.Error(), expectedErr.Error())
		assert.Equal(t, catalog.Entry{}, entry)
	})

	t.Run("future not found return correct error and status", func(t *testing.T) {
		h := &dynamicHandlerMock.TaskNodeHandler{}
		mockNodeExecutor := &nodeMocks.Node{}
		mockLPLauncher := &lpMocks.Reader{}
		mockFutureCatalog := &futureCatalogMocks.FutureCatalogClient{}
		scope := promutils.NewTestScope()

		nCtx := &nodeMocks.NodeExecutionContext{}
		d := New(h, mockNodeExecutor, mockLPLauncher, eventConfig, scope, mockFutureCatalog)

		catalogKey := catalog.Key{
			Identifier: core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Name:         "task-1",
			},
			TypedInterface: core.TypedInterface{},
		}
		h.OnGetCatalogKeyMatch(mock.Anything, mock.Anything).Return(catalogKey, nil)

		expectedErr := status.Error(codes.NotFound, "catalog entry not found")
		mockFutureCatalog.EXPECT().GetFuture(mock.Anything, mock.Anything).Return(catalog.Entry{}, expectedErr)

		entry, err := d.(*dynamicNodeTaskNodeHandler).CheckCatalogFutureCache(ctx, nCtx, h)

		assert.NoError(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_MISS, entry.GetStatus().GetCacheStatus())
	})

	t.Run("test readAndConvertOutput return correct outcome", func(t *testing.T) {
		mockReader := &outputReaderMocks.OutputReader{}

		expectDjSpec := &core.DynamicJobSpec{
			Tasks: []*core.TaskTemplate{
				{
					Id: &core.Identifier{
						ResourceType: core.ResourceType_TASK,
						Name:         "task-1",
					},
					Metadata: &core.TaskMetadata{},
					Interface: &core.TypedInterface{
						Inputs: &core.VariableMap{
							Variables: map[string]*core.Variable{
								"input1": {
									Type: &core.LiteralType{
										Type: &core.LiteralType_Simple{
											Simple: core.SimpleType_STRING,
										},
									},
								},
							},
						},
						Outputs: &core.VariableMap{
							Variables: map[string]*core.Variable{
								"output1": {
									Type: &core.LiteralType{
										Type: &core.LiteralType_Simple{
											Simple: core.SimpleType_INTEGER,
										},
									},
								},
							},
						},
					},
					Target: &core.TaskTemplate_Container{
						Container: &core.Container{
							Image: "test-image:latest",
							Args:  []string{"--input", "{{.input}}"},
						},
					},
				},
			},
			Outputs: []*core.Binding{
				{
					Var: "final_output",
					Binding: &core.BindingData{
						Value: &core.BindingData_Promise{
							Promise: &core.OutputReference{
								Var:    "output1",
								NodeId: "task-1",
							},
						},
					},
				},
			},
		}
		expectDjSpecBytes, err := proto.Marshal(expectDjSpec)
		assert.NoError(t, err)

		djSpecLiteralMap := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"future": {
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Binary{
								Binary: &core.Binary{
									Value: expectDjSpecBytes,
								},
							},
						},
					},
				},
			},
		}
		mockReader.OnReadMatch(mock.Anything).Return(djSpecLiteralMap, nil, nil)

		status := catalog.NewStatus(core.CatalogCacheStatus_CACHE_HIT, &core.CatalogMetadata{})
		entry := catalog.NewCatalogEntry(mockReader, status)

		h := &dynamicHandlerMock.TaskNodeHandler{}
		mockNodeExecutor := &nodeMocks.Node{}
		mockLPLauncher := &lpMocks.Reader{}
		mockFutureCatalog := &futureCatalogMocks.FutureCatalogClient{}
		scope := promutils.NewTestScope()

		d := New(h, mockNodeExecutor, mockLPLauncher, eventConfig, scope, mockFutureCatalog)

		djSpec, err := d.(*dynamicNodeTaskNodeHandler).readAndConvertOutput(ctx, entry)

		assert.NoError(t, err)
		assert.NotNil(t, djSpec)
		assert.True(t, proto.Equal(djSpec, expectDjSpec))
	})
}

func setupMockExecutionContext(t *testing.T) (*nodeMocks.NodeExecutionContext, *executorMocks.ExecutionContext) {
	mockCtx := &nodeMocks.NodeExecutionContext{}
	mockExecContext := &executorMocks.ExecutionContext{}
	mockNodeStatus := &flyteMocks.ExecutableNodeStatus{}
	dataStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	mockCtx.OnNodeID().Return("test-node-id")

	mockNodeExecMetadata := &nodeMocks.NodeExecutionMetadata{}
	mockNodeExecMetadata.OnGetOwnerID().Return(types.NamespacedName{Namespace: "test-namespace", Name: "test-name"})

	nodeExecutionID := &core.NodeExecutionIdentifier{
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "execution_name",
		},
		NodeId: "node-1",
	}
	mockNodeExecMetadata.OnGetNodeExecutionID().Return(nodeExecutionID)
	mockCtx.OnNodeExecutionMetadata().Return(mockNodeExecMetadata)

	mockTaskReader := &nodeMocks.TaskReader{}
	mockCtx.OnTaskReader().Return(mockTaskReader)
	mockCtx.OnCurrentAttempt().Return(1)

	taskID := &core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "project",
		Domain:       "domain",
		Name:         "name",
		Version:      "version",
	}

	mockTaskReader.OnGetTaskID().Return(taskID)
	mockTaskReader.OnReadMatch(mock.Anything).Return(&core.TaskTemplate{
		Id: taskID,
	}, nil)

	mockCtx.OnExecutionContext().Return(mockExecContext)
	mockCtx.OnNodeStatus().Return(mockNodeStatus)
	mockNodeStatus.OnGetOutputDir().Return(storage.DataReference("test-output-dir"))
	mockCtx.OnDataStore().Return(dataStore)

	return mockCtx, mockExecContext
}

func TestDynamicNodeTaskNodeHandler_WriteCatalogFutureCache(t *testing.T) {
	ctx := context.Background()

	t.Run("success with overwrite cache", func(t *testing.T) {
		mockCatalog := &futureCatalogMocks.FutureCatalogClient{}
		mockCacheHandler := &dynamicHandlerMock.TaskNodeHandler{}

		mockCtx, mockExecContext := setupMockExecutionContext(t)

		expectedCatalogKey := catalog.Key{
			Identifier: core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "project",
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
		}
		expectedStatus := catalog.NewStatus(core.CatalogCacheStatus_CACHE_HIT, nil)

		mockCacheHandler.On("GetCatalogKey",
			mock.Anything,
			mockCtx,
		).Return(expectedCatalogKey, nil)

		mockExecContext.OnGetExecutionConfig().Return(v1alpha1.ExecutionConfig{OverwriteCache: true})

		mockCatalog.On("UpdateFuture",
			mock.Anything,
			expectedCatalogKey,
			mock.Anything,
			mock.Anything,
		).Return(expectedStatus, nil)

		h := &dynamicNodeTaskNodeHandler{
			catalog: mockCatalog,
			metrics: newMetrics(promutils.NewTestScope()),
		}

		status, err := h.WriteCatalogFutureCache(ctx, mockCtx, mockCacheHandler)

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		mockCatalog.AssertExpectations(t)
		mockCacheHandler.AssertExpectations(t)
	})

	t.Run("success without overwrite cache", func(t *testing.T) {
		mockCtx, mockExecContext := setupMockExecutionContext(t)

		mockCatalog := &futureCatalogMocks.FutureCatalogClient{}
		mockCacheHandler := &dynamicHandlerMock.TaskNodeHandler{}

		expectedCatalogKey := catalog.Key{
			Identifier: core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "project",
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
		}
		expectedStatus := catalog.NewStatus(core.CatalogCacheStatus_CACHE_HIT, nil)

		mockCacheHandler.On("GetCatalogKey",
			mock.Anything,
			mockCtx,
		).Return(expectedCatalogKey, nil)

		mockExecContext.OnGetExecutionConfig().Return(v1alpha1.ExecutionConfig{OverwriteCache: false})

		mockCatalog.On("PutFuture",
			mock.Anything,
			expectedCatalogKey,
			mock.Anything,
			mock.Anything,
		).Return(expectedStatus, nil)

		h := &dynamicNodeTaskNodeHandler{
			catalog: mockCatalog,
			metrics: newMetrics(promutils.NewTestScope()),
		}

		status, err := h.WriteCatalogFutureCache(ctx, mockCtx, mockCacheHandler)

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		mockCatalog.AssertExpectations(t)
		mockCacheHandler.AssertExpectations(t)
	})

	t.Run("get catalog key error", func(t *testing.T) {
		mockCtx, _ := setupMockExecutionContext(t)

		mockCatalog := &futureCatalogMocks.FutureCatalogClient{}
		mockCacheHandler := &dynamicHandlerMock.TaskNodeHandler{}

		expectedError := errors.New("catalog key error")
		mockCacheHandler.On("GetCatalogKey",
			mock.Anything,
			mockCtx,
		).Return(catalog.Key{}, expectedError)

		h := &dynamicNodeTaskNodeHandler{
			catalog: mockCatalog,
			metrics: newMetrics(promutils.NewTestScope()),
		}

		status, err := h.WriteCatalogFutureCache(ctx, mockCtx, mockCacheHandler)

		assert.Error(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_DISABLED, status.GetCacheStatus())
		mockCacheHandler.AssertExpectations(t)
	})

	t.Run("catalog write error", func(t *testing.T) {
		mockCatalog := &futureCatalogMocks.FutureCatalogClient{}
		mockCacheHandler := &dynamicHandlerMock.TaskNodeHandler{}

		mockCtx, mockExecContext := setupMockExecutionContext(t)

		expectedCatalogKey := catalog.Key{
			Identifier: core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "project",
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
		}
		expectedError := errors.New("catalog write error")

		mockCacheHandler.On("GetCatalogKey",
			mock.Anything,
			mockCtx,
		).Return(expectedCatalogKey, nil)

		mockExecContext.OnGetExecutionConfig().Return(v1alpha1.ExecutionConfig{OverwriteCache: false})

		mockCatalog.On("PutFuture",
			mock.Anything,
			expectedCatalogKey,
			mock.Anything,
			mock.Anything,
		).Return(catalog.Status{}, expectedError)

		h := &dynamicNodeTaskNodeHandler{
			catalog: mockCatalog,
			metrics: newMetrics(promutils.NewTestScope()),
		}

		status, err := h.WriteCatalogFutureCache(ctx, mockCtx, mockCacheHandler)

		assert.NoError(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_PUT_FAILURE, status.GetCacheStatus())
		mockCatalog.AssertExpectations(t)
		mockCacheHandler.AssertExpectations(t)
	})
}

func TestGetOrExtendCatalogReservation(t *testing.T) {
	ctx := context.Background()

	testCatalogKey := catalog.Key{
		Identifier: core.Identifier{
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "test-name",
			Version: "test-version",
		},
	}
	testHeartbeatInterval := time.Second * 10

	tests := []struct {
		name           string
		setup          func() (*nodeMocks.NodeExecutionContext, *futureCatalogMocks.FutureCatalogClient, *dynamicHandlerMock.TaskNodeHandler)
		expectedError  error
		expectedStatus core.CatalogReservation_Status
	}{
		{
			name: "Successfully get reservation - new reservation",
			setup: func() (*nodeMocks.NodeExecutionContext, *futureCatalogMocks.FutureCatalogClient, *dynamicHandlerMock.TaskNodeHandler) {
				mockNodeExecutionContext, mockExecContext := setupMockExecutionContext(t)
				mockCatalog := &futureCatalogMocks.FutureCatalogClient{}
				mockCacheHandler := &dynamicHandlerMock.TaskNodeHandler{}

				mockParentInfo := &executorMocks.ImmutableParentInfo{}
				mockParentInfo.On("GetUniqueID").Return("test-unique-id")
				mockParentInfo.On("CurrentAttempt").Return(uint32(1))
				mockParentInfo.On("GetExecutionID").Return(&core.WorkflowExecutionIdentifier{
					Project: "test-project",
					Domain:  "test-domain",
					Name:    "test-name",
				})
				mockExecContext.OnGetParentInfo().Return(mockParentInfo)

				mockCacheHandler.On("GetCatalogKey", mock.Anything, mockNodeExecutionContext).Return(testCatalogKey, nil)

				mockReservation := &datacatalog.Reservation{
					ReservationId: &datacatalog.ReservationID{
						DatasetId: &datacatalog.DatasetID{
							Project: "project",
							Name:    "name",
							Domain:  "domain",
							Version: "version",
						},
						TagName: "tag",
					},
					OwnerId: "owner123",
					HeartbeatInterval: &durationpb.Duration{
						Seconds: 60,
					},
					ExpiresAt: &timestamppb.Timestamp{
						Seconds: time.Now().Add(time.Hour).Unix(),
					},
				}
				mockCatalog.On("GetOrExtendReservation",
					mock.Anything,
					testCatalogKey,
					mock.Anything,
					testHeartbeatInterval,
				).Return(mockReservation, nil)

				return mockNodeExecutionContext, mockCatalog, mockCacheHandler
			},
			expectedError:  nil,
			expectedStatus: core.CatalogReservation_RESERVATION_EXISTS,
		},
		{
			name: "Failed to get reservation - GetCatalogKey error",
			setup: func() (*nodeMocks.NodeExecutionContext, *futureCatalogMocks.FutureCatalogClient, *dynamicHandlerMock.TaskNodeHandler) {
				mockNodeExecutionContext, mockExecContext := setupMockExecutionContext(t)
				mockCatalog := &futureCatalogMocks.FutureCatalogClient{}
				mockCacheHandler := &dynamicHandlerMock.TaskNodeHandler{}

				mockParentInfo := &executorMocks.ImmutableParentInfo{}
				mockParentInfo.On("GetUniqueID").Return("test-unique-id")
				mockParentInfo.On("CurrentAttempt").Return(uint32(1))
				mockParentInfo.On("GetExecutionID").Return(&core.WorkflowExecutionIdentifier{
					Project: "test-project",
					Domain:  "test-domain",
					Name:    "test-name",
				})
				mockExecContext.OnGetParentInfo().Return(mockParentInfo)
				mockCacheHandler.On("GetCatalogKey", mock.Anything, mockNodeExecutionContext).Return(
					catalog.Key{}, errors.New("catalog key error"),
				)

				return mockNodeExecutionContext, mockCatalog, mockCacheHandler
			},
			expectedError:  errors.New("failed to initialize the future catalogKey: catalog key error"),
			expectedStatus: core.CatalogReservation_RESERVATION_DISABLED,
		},
		{
			name: "Failed to get reservation - GetOrExtendReservation error",
			setup: func() (*nodeMocks.NodeExecutionContext, *futureCatalogMocks.FutureCatalogClient, *dynamicHandlerMock.TaskNodeHandler) {
				mockNodeExecutionContext, mockExecContext := setupMockExecutionContext(t)
				mockCatalog := &futureCatalogMocks.FutureCatalogClient{}
				mockCacheHandler := &dynamicHandlerMock.TaskNodeHandler{}

				mockParentInfo := &executorMocks.ImmutableParentInfo{}
				mockParentInfo.On("GetUniqueID").Return("test-unique-id")
				mockParentInfo.On("CurrentAttempt").Return(uint32(1))
				mockParentInfo.On("GetExecutionID").Return(&core.WorkflowExecutionIdentifier{
					Project: "test-project",
					Domain:  "test-domain",
					Name:    "test-name",
				})

				mockExecContext.On("GetParentInfo").Return(mockParentInfo)
				mockNodeExecutionContext.On("ExecutionContext").Return(mockExecContext)

				mockCacheHandler.On("GetCatalogKey", mock.Anything, mockNodeExecutionContext).Return(testCatalogKey, nil)
				mockCatalog.On("GetOrExtendReservation",
					mock.Anything,
					testCatalogKey,
					mock.Anything,
					testHeartbeatInterval,
				).Return(nil, errors.New("reservation error"))

				return mockNodeExecutionContext, mockCatalog, mockCacheHandler
			},
			expectedError:  errors.New("reservation error"),
			expectedStatus: core.CatalogReservation_RESERVATION_FAILURE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNodeExecutionContext, mockCatalog, mockCacheHandler := tt.setup()

			d := &dynamicNodeTaskNodeHandler{
				catalog: mockCatalog,
				metrics: newMetrics(promutils.NewTestScope()),
			}

			reservation, err := d.GetOrExtendCatalogReservation(ctx, mockNodeExecutionContext, mockCacheHandler, testHeartbeatInterval)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reservation)
				assert.Equal(t, tt.expectedStatus, reservation.GetStatus())
			}
		})
	}
}

func TestDynamicNodeTaskNodeHandler_ReleaseCatalogReservation(t *testing.T) {
	ctx := context.Background()

	t.Run("Failed to get catalogKey", func(t *testing.T) {
		mockNodeExecContext, _ := setupMockExecutionContext(t)
		mockCatalog := &futureCatalogMocks.FutureCatalogClient{}
		mockCacheHandler := &dynamicHandlerMock.TaskNodeHandler{}

		h := &dynamicNodeTaskNodeHandler{
			catalog: mockCatalog,
			metrics: newMetrics(promutils.NewTestScope()),
		}

		expectedErr := errors.New("catalogKey error")
		mockCacheHandler.OnGetCatalogKeyMatch(ctx, mockNodeExecContext).Return(catalog.Key{}, expectedErr)

		entry, err := h.ReleaseCatalogReservation(ctx, mockNodeExecContext, mockCacheHandler)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to initialize the catalogKey")
		assert.Equal(t, core.CatalogReservation_RESERVATION_DISABLED, entry.GetStatus())
	})

	t.Run("Failed to release reservation", func(t *testing.T) {
		mockNodeExecContext, mockExecContext := setupMockExecutionContext(t)
		mockCatalog := &futureCatalogMocks.FutureCatalogClient{}
		mockCacheHandler := &dynamicHandlerMock.TaskNodeHandler{}

		mockParentInfo := &executorMocks.ImmutableParentInfo{}
		mockParentInfo.OnGetUniqueID().Return("test-unique-id")
		mockParentInfo.OnCurrentAttempt().Return(uint32(1))
		mockParentInfo.On("GetExecutionID").Return(&core.WorkflowExecutionIdentifier{
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "test-name",
		})
		mockExecContext.OnGetParentInfo().Return(mockParentInfo)
		mockNodeExecContext.OnExecutionContext().Return(mockExecContext)

		h := &dynamicNodeTaskNodeHandler{
			catalog: mockCatalog,
			metrics: newMetrics(promutils.NewTestScope()),
		}

		catalogKey := catalog.Key{}
		mockCacheHandler.OnGetCatalogKeyMatch(ctx, mockNodeExecContext).Return(catalogKey, nil)

		expectedErr := errors.New("release error")
		mockCatalog.EXPECT().ReleaseReservation(ctx, catalogKey, mock.Anything).Return(expectedErr)

		entry, err := h.ReleaseCatalogReservation(ctx, mockNodeExecContext, mockCacheHandler)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, core.CatalogReservation_RESERVATION_FAILURE, entry.GetStatus())
	})

	t.Run("Successfully released reservation", func(t *testing.T) {
		mockNodeExecContext, mockExecContext := setupMockExecutionContext(t)
		mockCatalog := &futureCatalogMocks.FutureCatalogClient{}
		mockCacheHandler := &dynamicHandlerMock.TaskNodeHandler{}

		mockParentInfo := &executorMocks.ImmutableParentInfo{}
		mockParentInfo.On("GetUniqueID").Return("test-unique-id")
		mockParentInfo.On("CurrentAttempt").Return(uint32(1))
		mockParentInfo.On("GetExecutionID").Return(&core.WorkflowExecutionIdentifier{
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "test-name",
		})

		mockExecContext.On("GetParentInfo").Return(mockParentInfo)
		mockNodeExecContext.On("ExecutionContext").Return(mockExecContext)

		h := &dynamicNodeTaskNodeHandler{
			catalog: mockCatalog,
			metrics: newMetrics(promutils.NewTestScope()),
		}

		catalogKey := catalog.Key{}
		mockCacheHandler.On("GetCatalogKey", mock.Anything, mockNodeExecContext).Return(catalogKey, nil)
		mockCatalog.On("ReleaseReservation", mock.Anything, catalogKey, mock.Anything).Return(nil)

		entry, err := h.ReleaseCatalogReservation(ctx, mockNodeExecContext, mockCacheHandler)

		assert.NoError(t, err)
		assert.Equal(t, core.CatalogReservation_RESERVATION_RELEASED, entry.GetStatus())
	})
}

func TestCheckFutureCache_WithOverwriteCache(t *testing.T) {
	d := &dynamicNodeTaskNodeHandler{
		metrics: newMetrics(promutils.NewTestScope()),
	}

	mockNodeExecContext, mockExecContext := setupMockExecutionContext(t)
	mockExecContext.OnGetExecutionConfig().Return(v1alpha1.ExecutionConfig{
		OverwriteCache: true,
	})

	status, err := d.CheckFutureCache(
		context.Background(),
		mockNodeExecContext,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, core.CatalogCacheStatus_CACHE_SKIPPED, status.GetCacheStatus())
}
