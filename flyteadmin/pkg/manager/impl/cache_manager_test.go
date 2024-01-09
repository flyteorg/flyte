package impl

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	flyteAdminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	pluginCatalog "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog/mocks"
	"github.com/flyteorg/flyte/flytestdlib/catalog"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

func serializeNodeExecutionMetadata(t *testing.T, nodeExecutionMetadata *admin.NodeExecutionMetaData) []byte {
	t.Helper()
	marshalled, err := proto.Marshal(nodeExecutionMetadata)
	require.NoError(t, err)
	return marshalled
}

func serializeNodeExecutionClosure(t *testing.T, nodeExecutionClosure *admin.NodeExecutionClosure) []byte {
	t.Helper()
	marshalled, err := proto.Marshal(nodeExecutionClosure)
	require.NoError(t, err)
	return marshalled
}

func serializeTaskExecutionClosure(t *testing.T, taskExecutionClosure *admin.TaskExecutionClosure) []byte {
	t.Helper()
	marshalled, err := proto.Marshal(taskExecutionClosure)
	require.NoError(t, err)
	return marshalled
}

func ptr[T any](val T) *T {
	return &val
}

func setupCacheEvictionMockRepositories(t *testing.T, repository interfaces.Repository, nodeExecutionModels []models.NodeExecution,
	taskExecutionModels map[string][]models.TaskExecution) map[string]int {

	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
			for _, nodeExecution := range nodeExecutionModels {
				if nodeExecution.NodeID == input.NodeExecutionIdentifier.NodeId {
					return nodeExecution, nil
				}
			}
			return models.NodeExecution{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "entry not found")
		})

	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetListCallback(
		func(ctx context.Context, input interfaces.ListResourceInput) (interfaces.NodeExecutionCollectionOutput, error) {
			var parentID uint
			for _, filter := range input.InlineFilters {
				if filter.GetField() == "parent_id" {
					query, err := filter.GetGormJoinTableQueryExpr("")
					require.NoError(t, err)
					var ok bool
					parentID, ok = query.Args.(uint)
					require.True(t, ok)
				}
			}
			if parentID == 0 {
				return interfaces.NodeExecutionCollectionOutput{NodeExecutions: nodeExecutionModels}, nil
			}
			for _, nodeExecution := range nodeExecutionModels {
				if nodeExecution.ID == parentID {
					return interfaces.NodeExecutionCollectionOutput{NodeExecutions: nodeExecution.ChildNodeExecutions}, nil
				}
			}
			return interfaces.NodeExecutionCollectionOutput{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "entry not found")
		})

	updatedNodeExecutions := make(map[string]int)
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetUpdateSelectedCallback(
		func(ctx context.Context, nodeExecution *models.NodeExecution, selectedFields []string) error {
			assert.Nil(t, nodeExecution.CacheStatus)

			var closure admin.NodeExecutionClosure
			err := proto.Unmarshal(nodeExecution.Closure, &closure)
			require.NoError(t, err)
			md, ok := closure.TargetMetadata.(*admin.NodeExecutionClosure_TaskNodeMetadata)
			require.True(t, ok)
			require.NotNil(t, md.TaskNodeMetadata)
			assert.Equal(t, core.CatalogCacheStatus_CACHE_DISABLED, md.TaskNodeMetadata.CacheStatus)

			updatedNodeExecutions[nodeExecution.NodeID]++

			for i := range nodeExecutionModels {
				if nodeExecutionModels[i].ID == nodeExecution.ID {
					nodeExecutionModels[i] = *nodeExecution
					break
				}
			}
			return nil
		})

	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetListCallback(
		func(ctx context.Context, input interfaces.ListResourceInput) (interfaces.TaskExecutionCollectionOutput, error) {
			var nodeID string
			for _, filter := range input.InlineFilters {
				if filter.GetField() == "node_id" {
					query, err := filter.GetGormJoinTableQueryExpr("")
					require.NoError(t, err)
					var ok bool
					nodeID, ok = query.Args.(string)
					require.True(t, ok)
				}
			}
			require.NotEmpty(t, nodeID)
			taskExecutions, ok := taskExecutionModels[nodeID]
			require.True(t, ok)
			return interfaces.TaskExecutionCollectionOutput{TaskExecutions: taskExecutions}, nil
		})

	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.GetTaskExecutionInput) (models.TaskExecution, error) {
			for nodeID := range taskExecutionModels {
				if nodeID == input.TaskExecutionID.NodeExecutionId.NodeId {
					return taskExecutionModels[nodeID][0], nil
				}
			}
			return models.TaskExecution{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "entry not found")
		})

	return updatedNodeExecutions
}

func setupCacheEvictionCatalogClient(t *testing.T, catalogClient *mocks.Client, artifactTags []*core.CatalogArtifactTag,
	taskExecutionModels map[string][]models.TaskExecution) map[string]int {
	for nodeID := range taskExecutionModels {
		for _, taskExecution := range taskExecutionModels[nodeID] {
			require.NotNil(t, taskExecution.RetryAttempt)
			ownerID := fmt.Sprintf("%s-%s-%d", executionIdentifier.Name, nodeID, *taskExecution.RetryAttempt)
			for _, artifactTag := range artifactTags {
				catalogClient.On("GetOrExtendReservationByArtifactTag", mock.Anything, mock.Anything,
					artifactTag.GetName(), ownerID, mock.Anything).Return(&datacatalog.Reservation{OwnerId: ownerID}, nil)
				catalogClient.On("ReleaseReservationByArtifactTag", mock.Anything, mock.Anything,
					artifactTag.GetName(), ownerID).Return(nil)
			}
		}
	}

	deletedArtifactIDs := make(map[string]int)
	for _, artifactTag := range artifactTags {
		catalogClient.On("DeleteByArtifactID", mock.Anything, mock.Anything, artifactTag.GetArtifactId()).
			Run(func(args mock.Arguments) {
				deletedArtifactIDs[args.Get(2).(string)] = deletedArtifactIDs[args.Get(2).(string)] + 1
			}).
			Return(nil)
	}

	return deletedArtifactIDs
}

func TestEvictTaskExecutionCache(t *testing.T) {
	t.Run("task", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		artifactTags := []*core.CatalogArtifactTag{
			{
				ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
		}

		nodeExecutionModels := []models.NodeExecution{
			{
				BaseModel: models.BaseModel{
					ID: 1,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "n0",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "n0",
				}),
				Closure: serializeNodeExecutionClosure(t, &admin.NodeExecutionClosure{
					TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
						TaskNodeMetadata: &admin.TaskNodeMetadata{
							CacheStatus: core.CatalogCacheStatus_CACHE_POPULATED,
							CatalogKey: &core.CatalogMetadata{
								DatasetId: &core.Identifier{
									ResourceType: core.ResourceType_DATASET,
									Project:      executionIdentifier.Project,
									Domain:       executionIdentifier.Domain,
									Name:         "flyte_task-test.evict.execution_cache_single_task",
									Version:      "version",
								},
								ArtifactTag: artifactTags[0],
							},
						},
					},
				}),
			},
		}

		taskExecutionModels := map[string][]models.TaskExecution{
			"n0": {
				{
					BaseModel: models.BaseModel{
						ID: 1,
					},
					TaskExecutionKey: models.TaskExecutionKey{
						TaskKey: models.TaskKey{
							Project: executionIdentifier.Project,
							Domain:  executionIdentifier.Domain,
							Name:    "flyte_task-test.evict.execution_cache",
							Version: "version",
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0",
						},
					}),
				},
			},
		}

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nodeExecutionModels, taskExecutionModels)
		deletedArtifactIDs := setupCacheEvictionCatalogClient(t, catalogClient, artifactTags, taskExecutionModels)

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictTaskExecutionCacheRequest{
			TaskExecutionId: &core.TaskExecutionIdentifier{
				TaskId: &core.Identifier{
					ResourceType: core.ResourceType_TASK,
					Project:      executionIdentifier.Project,
					Domain:       executionIdentifier.Domain,
					Name:         executionIdentifier.Name,
					Version:      version,
				},
				NodeExecutionId: &core.NodeExecutionIdentifier{
					ExecutionId: &executionIdentifier,
					NodeId:      "n0",
				},
				RetryAttempt: uint32(0),
			},
		}
		resp, err := cacheManager.EvictTaskExecutionCache(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.GetErrors().GetErrors())

		for nodeID := range taskExecutionModels {
			assert.Contains(t, updatedNodeExecutions, nodeID)
		}
		assert.Len(t, updatedNodeExecutions, len(taskExecutionModels))

		for _, artifactTag := range artifactTags {
			assert.Equal(t, 1, deletedArtifactIDs[artifactTag.GetArtifactId()])
		}
		assert.Len(t, deletedArtifactIDs, len(artifactTags))
	})

	t.Run("noop catalog", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := catalog.NOOPCatalog{}
		mockConfig := getMockExecutionsConfigProvider()

		artifactTags := []*core.CatalogArtifactTag{
			{
				ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
		}

		nodeExecutionModels := []models.NodeExecution{
			{
				BaseModel: models.BaseModel{
					ID: 1,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "n0",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "n0",
				}),
				Closure: serializeNodeExecutionClosure(t, &admin.NodeExecutionClosure{
					TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
						TaskNodeMetadata: &admin.TaskNodeMetadata{
							CacheStatus: core.CatalogCacheStatus_CACHE_POPULATED,
							CatalogKey: &core.CatalogMetadata{
								DatasetId: &core.Identifier{
									ResourceType: core.ResourceType_DATASET,
									Project:      executionIdentifier.Project,
									Domain:       executionIdentifier.Domain,
									Name:         "flyte_task-test.evict.execution_cache_single_task",
									Version:      "version",
								},
								ArtifactTag: artifactTags[0],
							},
						},
					},
				}),
			},
		}

		taskExecutionModels := map[string][]models.TaskExecution{
			"n0": {
				{
					BaseModel: models.BaseModel{
						ID: 1,
					},
					TaskExecutionKey: models.TaskExecutionKey{
						TaskKey: models.TaskKey{
							Project: executionIdentifier.Project,
							Domain:  executionIdentifier.Domain,
							Name:    "flyte_task-test.evict.execution_cache",
							Version: "version",
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0",
						},
					}),
				},
			},
		}

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nodeExecutionModels, taskExecutionModels)

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictTaskExecutionCacheRequest{
			TaskExecutionId: &core.TaskExecutionIdentifier{
				TaskId: &core.Identifier{
					ResourceType: core.ResourceType_TASK,
					Project:      executionIdentifier.Project,
					Domain:       executionIdentifier.Domain,
					Name:         executionIdentifier.Name,
					Version:      version,
				},
				NodeExecutionId: &core.NodeExecutionIdentifier{
					ExecutionId: &executionIdentifier,
					NodeId:      "n0",
				},
				RetryAttempt: uint32(0),
			},
		}
		resp, err := cacheManager.EvictTaskExecutionCache(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.GetErrors().GetErrors())

		assert.Empty(t, updatedNodeExecutions)
	})

	t.Run("unknown task execution", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetGetCallback(
			func(ctx context.Context, input interfaces.GetTaskExecutionInput) (models.TaskExecution, error) {
				return models.TaskExecution{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "entry not found")
			})

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictTaskExecutionCacheRequest{
			TaskExecutionId: &core.TaskExecutionIdentifier{
				TaskId: &core.Identifier{
					ResourceType: core.ResourceType_TASK,
					Project:      executionIdentifier.Project,
					Domain:       executionIdentifier.Domain,
					Name:         executionIdentifier.Name,
					Version:      version,
				},
				NodeExecutionId: &core.NodeExecutionIdentifier{
					ExecutionId: &executionIdentifier,
					NodeId:      "n0",
				},
				RetryAttempt: uint32(0),
			},
		}
		resp, err := cacheManager.EvictTaskExecutionCache(context.Background(), request)
		assert.Error(t, err)
		assert.True(t, pluginCatalog.IsNotFound(err))
		assert.Nil(t, resp)
	})

	t.Run("unknown node execution", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		taskExecutionModel := models.TaskExecution{
			BaseModel: models.BaseModel{
				ID: 1,
			},
			TaskExecutionKey: models.TaskExecutionKey{
				TaskKey: models.TaskKey{
					Project: executionIdentifier.Project,
					Domain:  executionIdentifier.Domain,
					Name:    "flyte_task-test.evict.execution_cache",
					Version: "version",
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "n0",
				},
				RetryAttempt: ptr[uint32](0),
			},
			Phase: core.NodeExecution_SUCCEEDED.String(),
			Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
				Metadata: &event.TaskExecutionMetadata{
					GeneratedName: "name-n0-0",
				},
			}),
		}

		repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetGetCallback(
			func(ctx context.Context, input interfaces.GetTaskExecutionInput) (models.TaskExecution, error) {
				assert.EqualValues(t, executionIdentifier, *input.TaskExecutionID.NodeExecutionId.ExecutionId)
				assert.Equal(t, taskExecutionModel.TaskExecutionKey.NodeID, input.TaskExecutionID.NodeExecutionId.NodeId)
				return taskExecutionModel, nil
			})

		repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(
			func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
				return models.NodeExecution{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "entry not found")
			})

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictTaskExecutionCacheRequest{
			TaskExecutionId: &core.TaskExecutionIdentifier{
				TaskId: &core.Identifier{
					ResourceType: core.ResourceType_TASK,
					Project:      executionIdentifier.Project,
					Domain:       executionIdentifier.Domain,
					Name:         executionIdentifier.Name,
					Version:      version,
				},
				NodeExecutionId: &core.NodeExecutionIdentifier{
					ExecutionId: &executionIdentifier,
					NodeId:      "n0",
				},
				RetryAttempt: uint32(0),
			},
		}
		resp, err := cacheManager.EvictTaskExecutionCache(context.Background(), request)
		assert.Error(t, err)
		assert.True(t, pluginCatalog.IsNotFound(err))
		assert.Nil(t, resp)
	})

	t.Run("task without cached results", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		nodeExecutionModels := []models.NodeExecution{
			{
				BaseModel: models.BaseModel{
					ID: 1,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "n0",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "n0",
				}),
				Closure: serializeNodeExecutionClosure(t, &admin.NodeExecutionClosure{
					TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
						TaskNodeMetadata: &admin.TaskNodeMetadata{
							CacheStatus: core.CatalogCacheStatus_CACHE_DISABLED,
						},
					},
				}),
			},
		}

		taskExecutionModels := map[string][]models.TaskExecution{
			"n0": {
				{
					BaseModel: models.BaseModel{
						ID: 1,
					},
					TaskExecutionKey: models.TaskExecutionKey{
						TaskKey: models.TaskKey{
							Project: executionIdentifier.Project,
							Domain:  executionIdentifier.Domain,
							Name:    "flyte_task-test.evict.execution_cache",
							Version: "version",
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0",
						},
					}),
				},
			},
		}

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nodeExecutionModels, taskExecutionModels)

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictTaskExecutionCacheRequest{
			TaskExecutionId: &core.TaskExecutionIdentifier{
				TaskId: &core.Identifier{
					ResourceType: core.ResourceType_TASK,
					Project:      executionIdentifier.Project,
					Domain:       executionIdentifier.Domain,
					Name:         executionIdentifier.Name,
					Version:      version,
				},
				NodeExecutionId: &core.NodeExecutionIdentifier{
					ExecutionId: &executionIdentifier,
					NodeId:      "n0",
				},
				RetryAttempt: uint32(0),
			},
		}
		resp, err := cacheManager.EvictTaskExecutionCache(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.GetErrors().GetErrors())

		assert.Empty(t, updatedNodeExecutions)
	})

	t.Run("idempotency", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		artifactTags := []*core.CatalogArtifactTag{
			{
				ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
		}

		nodeExecutionModels := []models.NodeExecution{
			{
				BaseModel: models.BaseModel{
					ID: 1,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "n0",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "n0",
				}),
				Closure: serializeNodeExecutionClosure(t, &admin.NodeExecutionClosure{
					TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
						TaskNodeMetadata: &admin.TaskNodeMetadata{
							CacheStatus: core.CatalogCacheStatus_CACHE_POPULATED,
							CatalogKey: &core.CatalogMetadata{
								DatasetId: &core.Identifier{
									ResourceType: core.ResourceType_DATASET,
									Project:      executionIdentifier.Project,
									Domain:       executionIdentifier.Domain,
									Name:         "flyte_task-test.evict.execution_cache_single_task",
									Version:      "version",
								},
								ArtifactTag: artifactTags[0],
							},
						},
					},
				}),
			},
		}

		taskExecutionModels := map[string][]models.TaskExecution{
			"n0": {
				{
					BaseModel: models.BaseModel{
						ID: 1,
					},
					TaskExecutionKey: models.TaskExecutionKey{
						TaskKey: models.TaskKey{
							Project: executionIdentifier.Project,
							Domain:  executionIdentifier.Domain,
							Name:    "flyte_task-test.evict.execution_cache",
							Version: "version",
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0",
						},
					}),
				},
			},
		}

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nodeExecutionModels, taskExecutionModels)
		deletedArtifactIDs := setupCacheEvictionCatalogClient(t, catalogClient, artifactTags, taskExecutionModels)

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictTaskExecutionCacheRequest{
			TaskExecutionId: &core.TaskExecutionIdentifier{
				TaskId: &core.Identifier{
					ResourceType: core.ResourceType_TASK,
					Project:      executionIdentifier.Project,
					Domain:       executionIdentifier.Domain,
					Name:         executionIdentifier.Name,
					Version:      version,
				},
				NodeExecutionId: &core.NodeExecutionIdentifier{
					ExecutionId: &executionIdentifier,
					NodeId:      "n0",
				},
				RetryAttempt: uint32(0),
			},
		}
		resp, err := cacheManager.EvictTaskExecutionCache(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.GetErrors().GetErrors())

		for nodeID := range taskExecutionModels {
			assert.Contains(t, updatedNodeExecutions, nodeID)
		}
		assert.Len(t, updatedNodeExecutions, len(taskExecutionModels))

		for _, artifactTag := range artifactTags {
			assert.Equal(t, 1, deletedArtifactIDs[artifactTag.GetArtifactId()])
		}
		assert.Len(t, deletedArtifactIDs, len(artifactTags))
	})

	t.Run("reserved artifact", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		artifactTags := []*core.CatalogArtifactTag{
			{
				ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
		}

		nodeExecutionModels := []models.NodeExecution{
			{
				BaseModel: models.BaseModel{
					ID: 1,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "n0",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "n0",
				}),
				Closure: serializeNodeExecutionClosure(t, &admin.NodeExecutionClosure{
					TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
						TaskNodeMetadata: &admin.TaskNodeMetadata{
							CacheStatus: core.CatalogCacheStatus_CACHE_POPULATED,
							CatalogKey: &core.CatalogMetadata{
								DatasetId: &core.Identifier{
									ResourceType: core.ResourceType_DATASET,
									Project:      executionIdentifier.Project,
									Domain:       executionIdentifier.Domain,
									Name:         "flyte_task-test.evict.execution_cache_single_task",
									Version:      "version",
								},
								ArtifactTag: artifactTags[0],
							},
						},
					},
				}),
			},
		}

		taskExecutionModels := map[string][]models.TaskExecution{
			"n0": {
				{
					BaseModel: models.BaseModel{
						ID: 1,
					},
					TaskExecutionKey: models.TaskExecutionKey{
						TaskKey: models.TaskKey{
							Project: executionIdentifier.Project,
							Domain:  executionIdentifier.Domain,
							Name:    "flyte_task-test.evict.execution_cache",
							Version: "version",
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0",
						},
					}),
				},
			},
		}

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nodeExecutionModels, taskExecutionModels)

		catalogClient.On("GetOrExtendReservationByArtifactTag", mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(&datacatalog.Reservation{OwnerId: "otherOwnerID"}, nil)
		catalogClient.On("ReleaseReservationByArtifactTag", mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictTaskExecutionCacheRequest{
			TaskExecutionId: &core.TaskExecutionIdentifier{
				TaskId: &core.Identifier{
					ResourceType: core.ResourceType_TASK,
					Project:      executionIdentifier.Project,
					Domain:       executionIdentifier.Domain,
					Name:         executionIdentifier.Name,
					Version:      version,
				},
				NodeExecutionId: &core.NodeExecutionIdentifier{
					ExecutionId: &executionIdentifier,
					NodeId:      "n0",
				},
				RetryAttempt: uint32(0),
			},
		}
		resp, err := cacheManager.EvictTaskExecutionCache(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.GetErrors().GetErrors(), len(artifactTags))
		eErr := resp.GetErrors().GetErrors()[0]
		assert.Equal(t, core.CacheEvictionError_RESERVATION_NOT_ACQUIRED, eErr.Code)
		assert.Equal(t, "n0", eErr.NodeExecutionId.NodeId)
		assert.True(t, proto.Equal(&executionIdentifier, eErr.NodeExecutionId.ExecutionId))
		require.NotNil(t, eErr.Source)
		assert.IsType(t, &core.CacheEvictionError_TaskExecutionId{}, eErr.Source)

		assert.Empty(t, updatedNodeExecutions)
	})

	t.Run("unknown artifact", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		artifactTags := []*core.CatalogArtifactTag{
			{
				ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
		}

		nodeExecutionModels := []models.NodeExecution{
			{
				BaseModel: models.BaseModel{
					ID: 1,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "n0",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "n0",
				}),
				Closure: serializeNodeExecutionClosure(t, &admin.NodeExecutionClosure{
					TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
						TaskNodeMetadata: &admin.TaskNodeMetadata{
							CacheStatus: core.CatalogCacheStatus_CACHE_POPULATED,
							CatalogKey: &core.CatalogMetadata{
								DatasetId: &core.Identifier{
									ResourceType: core.ResourceType_DATASET,
									Project:      executionIdentifier.Project,
									Domain:       executionIdentifier.Domain,
									Name:         "flyte_task-test.evict.execution_cache_single_task",
									Version:      "version",
								},
								ArtifactTag: artifactTags[0],
							},
						},
					},
				}),
			},
		}

		taskExecutionModels := map[string][]models.TaskExecution{
			"n0": {
				{
					BaseModel: models.BaseModel{
						ID: 1,
					},
					TaskExecutionKey: models.TaskExecutionKey{
						TaskKey: models.TaskKey{
							Project: executionIdentifier.Project,
							Domain:  executionIdentifier.Domain,
							Name:    "flyte_task-test.evict.execution_cache",
							Version: "version",
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0",
						},
					}),
				},
			},
		}

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nodeExecutionModels, taskExecutionModels)

		for nodeID := range taskExecutionModels {
			for _, taskExecution := range taskExecutionModels[nodeID] {
				require.NotNil(t, taskExecution.RetryAttempt)
				ownerID := fmt.Sprintf("%s-%s-%d", executionIdentifier.Name, nodeID, *taskExecution.RetryAttempt)
				for _, artifactTag := range artifactTags {
					catalogClient.On("GetOrExtendReservationByArtifactTag", mock.Anything, mock.Anything,
						artifactTag.GetName(), ownerID, mock.Anything).Return(&datacatalog.Reservation{OwnerId: ownerID}, nil)
					catalogClient.On("ReleaseReservationByArtifactTag", mock.Anything, mock.Anything,
						artifactTag.GetName(), ownerID).Return(nil)
				}
			}
		}

		catalogClient.On("DeleteByArtifactID", mock.Anything, mock.Anything, mock.Anything).
			Return(status.Error(codes.NotFound, "not found"))

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictTaskExecutionCacheRequest{
			TaskExecutionId: &core.TaskExecutionIdentifier{
				TaskId: &core.Identifier{
					ResourceType: core.ResourceType_TASK,
					Project:      executionIdentifier.Project,
					Domain:       executionIdentifier.Domain,
					Name:         executionIdentifier.Name,
					Version:      version,
				},
				NodeExecutionId: &core.NodeExecutionIdentifier{
					ExecutionId: &executionIdentifier,
					NodeId:      "n0",
				},
				RetryAttempt: uint32(0),
			},
		}
		resp, err := cacheManager.EvictTaskExecutionCache(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.GetErrors().GetErrors())

		for nodeID := range taskExecutionModels {
			assert.Contains(t, updatedNodeExecutions, nodeID)
		}
		assert.Len(t, updatedNodeExecutions, len(taskExecutionModels))
	})

	t.Run("datacatalog error", func(t *testing.T) {
		t.Run("GetOrExtendReservationByArtifactTag", func(t *testing.T) {
			repository := repositoryMocks.NewMockRepository()
			catalogClient := &mocks.Client{}
			mockConfig := getMockExecutionsConfigProvider()

			artifactTags := []*core.CatalogArtifactTag{
				{
					ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
					Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
				},
			}

			nodeExecutionModels := []models.NodeExecution{
				{
					BaseModel: models.BaseModel{
						ID: 1,
					},
					NodeExecutionKey: models.NodeExecutionKey{
						ExecutionKey: models.ExecutionKey{
							Project: executionIdentifier.Project,
							Domain:  executionIdentifier.Domain,
							Name:    executionIdentifier.Name,
						},
						NodeID: "n0",
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
						IsParentNode: false,
						IsDynamic:    false,
						SpecNodeId:   "n0",
					}),
					Closure: serializeNodeExecutionClosure(t, &admin.NodeExecutionClosure{
						TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
							TaskNodeMetadata: &admin.TaskNodeMetadata{
								CacheStatus: core.CatalogCacheStatus_CACHE_POPULATED,
								CatalogKey: &core.CatalogMetadata{
									DatasetId: &core.Identifier{
										ResourceType: core.ResourceType_DATASET,
										Project:      executionIdentifier.Project,
										Domain:       executionIdentifier.Domain,
										Name:         "flyte_task-test.evict.execution_cache_single_task",
										Version:      "version",
									},
									ArtifactTag: artifactTags[0],
								},
							},
						},
					}),
				},
			}

			taskExecutionModels := map[string][]models.TaskExecution{
				"n0": {
					{
						BaseModel: models.BaseModel{
							ID: 1,
						},
						TaskExecutionKey: models.TaskExecutionKey{
							TaskKey: models.TaskKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    "flyte_task-test.evict.execution_cache",
								Version: "version",
							},
							NodeExecutionKey: models.NodeExecutionKey{
								ExecutionKey: models.ExecutionKey{
									Project: executionIdentifier.Project,
									Domain:  executionIdentifier.Domain,
									Name:    executionIdentifier.Name,
								},
								NodeID: "n0",
							},
							RetryAttempt: ptr[uint32](0),
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
							Metadata: &event.TaskExecutionMetadata{
								GeneratedName: "name-n0-0",
							},
						}),
					},
				},
			}

			updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nodeExecutionModels, taskExecutionModels)

			catalogClient.On("GetOrExtendReservationByArtifactTag", mock.Anything, mock.Anything,
				mock.Anything, mock.Anything, mock.Anything).Return(nil, status.Error(codes.Internal, "error"))

			cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
			request := service.EvictTaskExecutionCacheRequest{
				TaskExecutionId: &core.TaskExecutionIdentifier{
					TaskId: &core.Identifier{
						ResourceType: core.ResourceType_TASK,
						Project:      executionIdentifier.Project,
						Domain:       executionIdentifier.Domain,
						Name:         executionIdentifier.Name,
						Version:      version,
					},
					NodeExecutionId: &core.NodeExecutionIdentifier{
						ExecutionId: &executionIdentifier,
						NodeId:      "n0",
					},
					RetryAttempt: uint32(0),
				},
			}
			resp, err := cacheManager.EvictTaskExecutionCache(context.Background(), request)
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Len(t, resp.GetErrors().GetErrors(), len(artifactTags))
			eErr := resp.GetErrors().GetErrors()[0]
			assert.Equal(t, core.CacheEvictionError_RESERVATION_NOT_ACQUIRED, eErr.Code)
			assert.Equal(t, "n0", eErr.NodeExecutionId.NodeId)
			assert.True(t, proto.Equal(&executionIdentifier, eErr.NodeExecutionId.ExecutionId))
			require.NotNil(t, eErr.Source)
			assert.IsType(t, &core.CacheEvictionError_TaskExecutionId{}, eErr.Source)

			assert.Empty(t, updatedNodeExecutions)
		})

		t.Run("DeleteByArtifactID", func(t *testing.T) {
			repository := repositoryMocks.NewMockRepository()
			catalogClient := &mocks.Client{}
			mockConfig := getMockExecutionsConfigProvider()

			artifactTags := []*core.CatalogArtifactTag{
				{
					ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
					Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
				},
			}

			nodeExecutionModels := []models.NodeExecution{
				{
					BaseModel: models.BaseModel{
						ID: 1,
					},
					NodeExecutionKey: models.NodeExecutionKey{
						ExecutionKey: models.ExecutionKey{
							Project: executionIdentifier.Project,
							Domain:  executionIdentifier.Domain,
							Name:    executionIdentifier.Name,
						},
						NodeID: "n0",
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
						IsParentNode: false,
						IsDynamic:    false,
						SpecNodeId:   "n0",
					}),
					Closure: serializeNodeExecutionClosure(t, &admin.NodeExecutionClosure{
						TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
							TaskNodeMetadata: &admin.TaskNodeMetadata{
								CacheStatus: core.CatalogCacheStatus_CACHE_POPULATED,
								CatalogKey: &core.CatalogMetadata{
									DatasetId: &core.Identifier{
										ResourceType: core.ResourceType_DATASET,
										Project:      executionIdentifier.Project,
										Domain:       executionIdentifier.Domain,
										Name:         "flyte_task-test.evict.execution_cache_single_task",
										Version:      "version",
									},
									ArtifactTag: artifactTags[0],
								},
							},
						},
					}),
				},
			}

			taskExecutionModels := map[string][]models.TaskExecution{
				"n0": {
					{
						BaseModel: models.BaseModel{
							ID: 1,
						},
						TaskExecutionKey: models.TaskExecutionKey{
							TaskKey: models.TaskKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    "flyte_task-test.evict.execution_cache",
								Version: "version",
							},
							NodeExecutionKey: models.NodeExecutionKey{
								ExecutionKey: models.ExecutionKey{
									Project: executionIdentifier.Project,
									Domain:  executionIdentifier.Domain,
									Name:    executionIdentifier.Name,
								},
								NodeID: "n0",
							},
							RetryAttempt: ptr[uint32](0),
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
							Metadata: &event.TaskExecutionMetadata{
								GeneratedName: "name-n0-0",
							},
						}),
					},
				},
			}

			updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nodeExecutionModels, taskExecutionModels)

			for nodeID := range taskExecutionModels {
				for _, taskExecution := range taskExecutionModels[nodeID] {
					require.NotNil(t, taskExecution.RetryAttempt)
					ownerID := fmt.Sprintf("%s-%s-%d", executionIdentifier.Name, nodeID, *taskExecution.RetryAttempt)
					for _, artifactTag := range artifactTags {
						catalogClient.On("GetOrExtendReservationByArtifactTag", mock.Anything, mock.Anything,
							artifactTag.GetName(), ownerID, mock.Anything).Return(&datacatalog.Reservation{OwnerId: ownerID}, nil)
					}
				}
			}

			catalogClient.On("DeleteByArtifactID", mock.Anything, mock.Anything, mock.Anything).
				Return(status.Error(codes.Internal, "error"))
			catalogClient.On("ReleaseReservationByArtifactTag", mock.Anything, mock.Anything,
				mock.Anything, mock.Anything).Return(nil)

			cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
			request := service.EvictTaskExecutionCacheRequest{
				TaskExecutionId: &core.TaskExecutionIdentifier{
					TaskId: &core.Identifier{
						ResourceType: core.ResourceType_TASK,
						Project:      executionIdentifier.Project,
						Domain:       executionIdentifier.Domain,
						Name:         executionIdentifier.Name,
						Version:      version,
					},
					NodeExecutionId: &core.NodeExecutionIdentifier{
						ExecutionId: &executionIdentifier,
						NodeId:      "n0",
					},
					RetryAttempt: uint32(0),
				},
			}
			resp, err := cacheManager.EvictTaskExecutionCache(context.Background(), request)
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Len(t, resp.GetErrors().GetErrors(), len(artifactTags))
			eErr := resp.GetErrors().GetErrors()[0]
			assert.Equal(t, core.CacheEvictionError_ARTIFACT_DELETE_FAILED, eErr.Code)
			assert.Equal(t, "n0", eErr.NodeExecutionId.NodeId)
			assert.True(t, proto.Equal(&executionIdentifier, eErr.NodeExecutionId.ExecutionId))
			require.NotNil(t, eErr.Source)
			assert.IsType(t, &core.CacheEvictionError_TaskExecutionId{}, eErr.Source)

			assert.Len(t, updatedNodeExecutions, 0)
		})

		t.Run("ReleaseReservationByArtifactTag", func(t *testing.T) {
			repository := repositoryMocks.NewMockRepository()
			catalogClient := &mocks.Client{}
			mockConfig := getMockExecutionsConfigProvider()

			artifactTags := []*core.CatalogArtifactTag{
				{
					ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
					Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
				},
			}

			nodeExecutionModels := []models.NodeExecution{
				{
					BaseModel: models.BaseModel{
						ID: 1,
					},
					NodeExecutionKey: models.NodeExecutionKey{
						ExecutionKey: models.ExecutionKey{
							Project: executionIdentifier.Project,
							Domain:  executionIdentifier.Domain,
							Name:    executionIdentifier.Name,
						},
						NodeID: "n0",
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
						IsParentNode: false,
						IsDynamic:    false,
						SpecNodeId:   "n0",
					}),
					Closure: serializeNodeExecutionClosure(t, &admin.NodeExecutionClosure{
						TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
							TaskNodeMetadata: &admin.TaskNodeMetadata{
								CacheStatus: core.CatalogCacheStatus_CACHE_POPULATED,
								CatalogKey: &core.CatalogMetadata{
									DatasetId: &core.Identifier{
										ResourceType: core.ResourceType_DATASET,
										Project:      executionIdentifier.Project,
										Domain:       executionIdentifier.Domain,
										Name:         "flyte_task-test.evict.execution_cache_single_task",
										Version:      "version",
									},
									ArtifactTag: artifactTags[0],
								},
							},
						},
					}),
				},
			}

			taskExecutionModels := map[string][]models.TaskExecution{
				"n0": {
					{
						BaseModel: models.BaseModel{
							ID: 1,
						},
						TaskExecutionKey: models.TaskExecutionKey{
							TaskKey: models.TaskKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    "flyte_task-test.evict.execution_cache",
								Version: "version",
							},
							NodeExecutionKey: models.NodeExecutionKey{
								ExecutionKey: models.ExecutionKey{
									Project: executionIdentifier.Project,
									Domain:  executionIdentifier.Domain,
									Name:    executionIdentifier.Name,
								},
								NodeID: "n0",
							},
							RetryAttempt: ptr[uint32](0),
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
							Metadata: &event.TaskExecutionMetadata{
								GeneratedName: "name-n0-0",
							},
						}),
					},
				},
			}

			updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nodeExecutionModels, taskExecutionModels)

			for nodeID := range taskExecutionModels {
				for _, taskExecution := range taskExecutionModels[nodeID] {
					require.NotNil(t, taskExecution.RetryAttempt)
					ownerID := fmt.Sprintf("%s-%s-%d", executionIdentifier.Name, nodeID, *taskExecution.RetryAttempt)
					for _, artifactTag := range artifactTags {
						catalogClient.On("GetOrExtendReservationByArtifactTag", mock.Anything, mock.Anything,
							artifactTag.GetName(), ownerID, mock.Anything).Return(&datacatalog.Reservation{OwnerId: ownerID}, nil)
					}
				}
			}

			deletedArtifactIDs := make(map[string]int)
			for _, artifactTag := range artifactTags {
				catalogClient.On("DeleteByArtifactID", mock.Anything, mock.Anything, artifactTag.GetArtifactId()).
					Run(func(args mock.Arguments) {
						deletedArtifactIDs[args.Get(2).(string)] = deletedArtifactIDs[args.Get(2).(string)] + 1
					}).
					Return(nil)
			}

			catalogClient.On("ReleaseReservationByArtifactTag", mock.Anything, mock.Anything,
				mock.Anything, mock.Anything).Return(status.Error(codes.Internal, "error"))

			cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
			request := service.EvictTaskExecutionCacheRequest{
				TaskExecutionId: &core.TaskExecutionIdentifier{
					TaskId: &core.Identifier{
						ResourceType: core.ResourceType_TASK,
						Project:      executionIdentifier.Project,
						Domain:       executionIdentifier.Domain,
						Name:         executionIdentifier.Name,
						Version:      version,
					},
					NodeExecutionId: &core.NodeExecutionIdentifier{
						ExecutionId: &executionIdentifier,
						NodeId:      "n0",
					},
					RetryAttempt: uint32(0),
				},
			}
			resp, err := cacheManager.EvictTaskExecutionCache(context.Background(), request)
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Len(t, resp.GetErrors().GetErrors(), len(artifactTags))
			eErr := resp.GetErrors().GetErrors()[0]
			assert.Equal(t, core.CacheEvictionError_RESERVATION_NOT_RELEASED, eErr.Code)
			assert.Equal(t, "n0", eErr.NodeExecutionId.NodeId)
			assert.True(t, proto.Equal(&executionIdentifier, eErr.NodeExecutionId.ExecutionId))
			require.NotNil(t, eErr.Source)
			assert.IsType(t, &core.CacheEvictionError_TaskExecutionId{}, eErr.Source)

			for nodeID := range taskExecutionModels {
				assert.Contains(t, updatedNodeExecutions, nodeID)
			}
			assert.Len(t, updatedNodeExecutions, len(taskExecutionModels))
		})
	})

	t.Run("repository error", func(t *testing.T) {
		t.Run("NodeExecutionRepo", func(t *testing.T) {
			t.Run("Get", func(t *testing.T) {
				repository := repositoryMocks.NewMockRepository()
				catalogClient := &mocks.Client{}
				mockConfig := getMockExecutionsConfigProvider()

				artifactTags := []*core.CatalogArtifactTag{
					{
						ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
						Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
					},
				}

				nodeExecutionModels := []models.NodeExecution{
					{
						BaseModel: models.BaseModel{
							ID: 1,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "n0",
						}),
						Closure: serializeNodeExecutionClosure(t, &admin.NodeExecutionClosure{
							TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
								TaskNodeMetadata: &admin.TaskNodeMetadata{
									CacheStatus: core.CatalogCacheStatus_CACHE_POPULATED,
									CatalogKey: &core.CatalogMetadata{
										DatasetId: &core.Identifier{
											ResourceType: core.ResourceType_DATASET,
											Project:      executionIdentifier.Project,
											Domain:       executionIdentifier.Domain,
											Name:         "flyte_task-test.evict.execution_cache_single_task",
											Version:      "version",
										},
										ArtifactTag: artifactTags[0],
									},
								},
							},
						}),
					},
				}

				taskExecutionModels := map[string][]models.TaskExecution{
					"n0": {
						{
							BaseModel: models.BaseModel{
								ID: 1,
							},
							TaskExecutionKey: models.TaskExecutionKey{
								TaskKey: models.TaskKey{
									Project: executionIdentifier.Project,
									Domain:  executionIdentifier.Domain,
									Name:    "flyte_task-test.evict.execution_cache",
									Version: "version",
								},
								NodeExecutionKey: models.NodeExecutionKey{
									ExecutionKey: models.ExecutionKey{
										Project: executionIdentifier.Project,
										Domain:  executionIdentifier.Domain,
										Name:    executionIdentifier.Name,
									},
									NodeID: "n0",
								},
								RetryAttempt: ptr[uint32](0),
							},
							Phase: core.NodeExecution_SUCCEEDED.String(),
							Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
								Metadata: &event.TaskExecutionMetadata{
									GeneratedName: "name-n0-0",
								},
							}),
						},
					},
				}

				updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nodeExecutionModels, taskExecutionModels)

				repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(
					func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
						return models.NodeExecution{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.Internal, "error")
					})

				cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
				request := service.EvictTaskExecutionCacheRequest{
					TaskExecutionId: &core.TaskExecutionIdentifier{
						TaskId: &core.Identifier{
							ResourceType: core.ResourceType_TASK,
							Project:      executionIdentifier.Project,
							Domain:       executionIdentifier.Domain,
							Name:         executionIdentifier.Name,
							Version:      version,
						},
						NodeExecutionId: &core.NodeExecutionIdentifier{
							ExecutionId: &executionIdentifier,
							NodeId:      "n0",
						},
						RetryAttempt: uint32(0),
					},
				}
				resp, err := cacheManager.EvictTaskExecutionCache(context.Background(), request)
				require.Error(t, err)
				s, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, codes.Internal, s.Code())
				assert.Nil(t, resp)

				assert.Empty(t, updatedNodeExecutions)
			})

			t.Run("UpdateSelected", func(t *testing.T) {
				repository := repositoryMocks.NewMockRepository()
				catalogClient := &mocks.Client{}
				mockConfig := getMockExecutionsConfigProvider()

				artifactTags := []*core.CatalogArtifactTag{
					{
						ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
						Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
					},
				}

				nodeExecutionModels := []models.NodeExecution{
					{
						BaseModel: models.BaseModel{
							ID: 1,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "n0",
						}),
						Closure: serializeNodeExecutionClosure(t, &admin.NodeExecutionClosure{
							TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
								TaskNodeMetadata: &admin.TaskNodeMetadata{
									CacheStatus: core.CatalogCacheStatus_CACHE_POPULATED,
									CatalogKey: &core.CatalogMetadata{
										DatasetId: &core.Identifier{
											ResourceType: core.ResourceType_DATASET,
											Project:      executionIdentifier.Project,
											Domain:       executionIdentifier.Domain,
											Name:         "flyte_task-test.evict.execution_cache_single_task",
											Version:      "version",
										},
										ArtifactTag: artifactTags[0],
									},
								},
							},
						}),
					},
				}

				taskExecutionModels := map[string][]models.TaskExecution{
					"n0": {
						{
							BaseModel: models.BaseModel{
								ID: 1,
							},
							TaskExecutionKey: models.TaskExecutionKey{
								TaskKey: models.TaskKey{
									Project: executionIdentifier.Project,
									Domain:  executionIdentifier.Domain,
									Name:    "flyte_task-test.evict.execution_cache",
									Version: "version",
								},
								NodeExecutionKey: models.NodeExecutionKey{
									ExecutionKey: models.ExecutionKey{
										Project: executionIdentifier.Project,
										Domain:  executionIdentifier.Domain,
										Name:    executionIdentifier.Name,
									},
									NodeID: "n0",
								},
								RetryAttempt: ptr[uint32](0),
							},
							Phase: core.NodeExecution_SUCCEEDED.String(),
							Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
								Metadata: &event.TaskExecutionMetadata{
									GeneratedName: "name-n0-0",
								},
							}),
						},
					},
				}

				updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nodeExecutionModels, taskExecutionModels)

				for nodeID := range taskExecutionModels {
					for _, taskExecution := range taskExecutionModels[nodeID] {
						require.NotNil(t, taskExecution.RetryAttempt)
						ownerID := fmt.Sprintf("%s-%s-%d", executionIdentifier.Name, nodeID, *taskExecution.RetryAttempt)
						for _, artifactTag := range artifactTags {
							catalogClient.On("GetOrExtendReservationByArtifactTag", mock.Anything, mock.Anything,
								artifactTag.GetName(), ownerID, mock.Anything).Return(&datacatalog.Reservation{OwnerId: ownerID}, nil)
							catalogClient.On("ReleaseReservationByArtifactTag", mock.Anything, mock.Anything,
								artifactTag.GetName(), ownerID).Return(nil)
							catalogClient.On("DeleteByArtifactID", mock.Anything, mock.Anything, artifactTag.GetArtifactId()).
								Return(nil)
						}
					}
				}

				repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetUpdateSelectedCallback(
					func(ctx context.Context, nodeExecution *models.NodeExecution, selectedFields []string) error {
						return flyteAdminErrors.NewFlyteAdminError(codes.Internal, "error")
					})

				cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
				request := service.EvictTaskExecutionCacheRequest{
					TaskExecutionId: &core.TaskExecutionIdentifier{
						TaskId: &core.Identifier{
							ResourceType: core.ResourceType_TASK,
							Project:      executionIdentifier.Project,
							Domain:       executionIdentifier.Domain,
							Name:         executionIdentifier.Name,
							Version:      version,
						},
						NodeExecutionId: &core.NodeExecutionIdentifier{
							ExecutionId: &executionIdentifier,
							NodeId:      "n0",
						},
						RetryAttempt: uint32(0),
					},
				}
				resp, err := cacheManager.EvictTaskExecutionCache(context.Background(), request)
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Len(t, resp.GetErrors().GetErrors(), len(artifactTags))
				eErr := resp.GetErrors().GetErrors()[0]
				assert.Equal(t, core.CacheEvictionError_DATABASE_UPDATE_FAILED, eErr.Code)
				assert.Equal(t, "n0", eErr.NodeExecutionId.NodeId)
				assert.True(t, proto.Equal(&executionIdentifier, eErr.NodeExecutionId.ExecutionId))
				require.NotNil(t, eErr.Source)
				assert.IsType(t, &core.CacheEvictionError_TaskExecutionId{}, eErr.Source)

				assert.Empty(t, updatedNodeExecutions)
			})
		})

		t.Run("TaskExecutionRepo", func(t *testing.T) {
			t.Run("Get", func(t *testing.T) {
				repository := repositoryMocks.NewMockRepository()
				catalogClient := &mocks.Client{}
				mockConfig := getMockExecutionsConfigProvider()

				artifactTags := []*core.CatalogArtifactTag{
					{
						ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
						Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
					},
				}

				nodeExecutionModels := []models.NodeExecution{
					{
						BaseModel: models.BaseModel{
							ID: 1,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "n0",
						}),
						Closure: serializeNodeExecutionClosure(t, &admin.NodeExecutionClosure{
							TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
								TaskNodeMetadata: &admin.TaskNodeMetadata{
									CacheStatus: core.CatalogCacheStatus_CACHE_POPULATED,
									CatalogKey: &core.CatalogMetadata{
										DatasetId: &core.Identifier{
											ResourceType: core.ResourceType_DATASET,
											Project:      executionIdentifier.Project,
											Domain:       executionIdentifier.Domain,
											Name:         "flyte_task-test.evict.execution_cache_single_task",
											Version:      "version",
										},
										ArtifactTag: artifactTags[0],
									},
								},
							},
						}),
					},
				}

				taskExecutionModels := map[string][]models.TaskExecution{
					"n0": {
						{
							BaseModel: models.BaseModel{
								ID: 1,
							},
							TaskExecutionKey: models.TaskExecutionKey{
								TaskKey: models.TaskKey{
									Project: executionIdentifier.Project,
									Domain:  executionIdentifier.Domain,
									Name:    "flyte_task-test.evict.execution_cache",
									Version: "version",
								},
								NodeExecutionKey: models.NodeExecutionKey{
									ExecutionKey: models.ExecutionKey{
										Project: executionIdentifier.Project,
										Domain:  executionIdentifier.Domain,
										Name:    executionIdentifier.Name,
									},
									NodeID: "n0",
								},
								RetryAttempt: ptr[uint32](0),
							},
							Phase: core.NodeExecution_SUCCEEDED.String(),
							Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
								Metadata: &event.TaskExecutionMetadata{
									GeneratedName: "name-n0-0",
								},
							}),
						},
					},
				}

				updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nodeExecutionModels, taskExecutionModels)

				repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetGetCallback(
					func(ctx context.Context, input interfaces.GetTaskExecutionInput) (models.TaskExecution, error) {
						return models.TaskExecution{}, flyteAdminErrors.NewFlyteAdminError(codes.Internal, "error")
					})

				cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
				request := service.EvictTaskExecutionCacheRequest{
					TaskExecutionId: &core.TaskExecutionIdentifier{
						TaskId: &core.Identifier{
							ResourceType: core.ResourceType_TASK,
							Project:      executionIdentifier.Project,
							Domain:       executionIdentifier.Domain,
							Name:         executionIdentifier.Name,
							Version:      version,
						},
						NodeExecutionId: &core.NodeExecutionIdentifier{
							ExecutionId: &executionIdentifier,
							NodeId:      "n0",
						},
						RetryAttempt: uint32(0),
					},
				}
				resp, err := cacheManager.EvictTaskExecutionCache(context.Background(), request)
				require.Error(t, err)
				s, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, codes.Internal, s.Code())
				assert.Nil(t, resp)

				assert.Empty(t, updatedNodeExecutions)
			})
		})
	})
}
