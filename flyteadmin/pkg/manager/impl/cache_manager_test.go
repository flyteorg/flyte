package impl

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	flyteAdminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	pluginCatalog "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog/mocks"
	"github.com/flyteorg/flytestdlib/catalog"
	"github.com/flyteorg/flytestdlib/promutils"
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

func setupCacheEvictionMockRepositories(t *testing.T, repository interfaces.Repository, executionModel *models.Execution,
	nodeExecutionModels []models.NodeExecution, taskExecutionModels map[string][]models.TaskExecution) map[string]int {
	if executionModel != nil {
		repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
			func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
				assert.Equal(t, executionIdentifier.Domain, input.Domain)
				assert.Equal(t, executionIdentifier.Name, input.Name)
				assert.Equal(t, executionIdentifier.Project, input.Project)
				return *executionModel, nil
			})
	}

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
			assert.Nil(t, md.TaskNodeMetadata.CatalogKey)

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

func TestEvictExecutionCache(t *testing.T) {
	t.Run("single task", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		artifactTags := []*core.CatalogArtifactTag{
			{
				ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
		}

		executionModel := models.Execution{
			BaseModel: models.BaseModel{
				ID: 1,
			},
			ExecutionKey: models.ExecutionKey{
				Project: executionIdentifier.Project,
				Domain:  executionIdentifier.Domain,
				Name:    executionIdentifier.Name,
			},
			LaunchPlanID: 6,
			WorkflowID:   4,
			TaskID:       0,
			Phase:        core.NodeExecution_SUCCEEDED.String(),
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
					NodeID: "start-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "start-node",
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 2,
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
			{
				BaseModel: models.BaseModel{
					ID: 3,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "end-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "end-node",
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

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, &executionModel, nodeExecutionModels,
			taskExecutionModels)
		deletedArtifactIDs := setupCacheEvictionCatalogClient(t, catalogClient, artifactTags, taskExecutionModels)

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictExecutionCacheRequest{
			WorkflowExecutionId: &executionIdentifier,
		}
		resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.GetErrors().GetErrors())

		for nodeID := range taskExecutionModels {
			assert.Contains(t, updatedNodeExecutions, nodeID)
		}
		assert.Len(t, updatedNodeExecutions, len(taskExecutionModels))

		for _, artifactID := range artifactTags {
			assert.Equal(t, 1, deletedArtifactIDs[artifactID.GetArtifactId()])
		}
		assert.Len(t, deletedArtifactIDs, len(artifactTags))
	})

	t.Run("multiple tasks", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		artifactTags := []*core.CatalogArtifactTag{
			{
				ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
			{
				ArtifactId: "4074be3e-7cee-4a7b-8c45-56577fa32f24",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYN",
			},
			{
				ArtifactId: "a8bd60d5-b2bb-4b06-a2ac-240d183a4ca8",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYO",
			},
		}

		executionModel := models.Execution{
			BaseModel: models.BaseModel{
				ID: 1,
			},
			ExecutionKey: models.ExecutionKey{
				Project: executionIdentifier.Project,
				Domain:  executionIdentifier.Domain,
				Name:    executionIdentifier.Name,
			},
			LaunchPlanID: 6,
			WorkflowID:   4,
			TaskID:       0,
			Phase:        core.NodeExecution_SUCCEEDED.String(),
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
					NodeID: "start-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "start-node",
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 2,
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
			{
				BaseModel: models.BaseModel{
					ID: 3,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "n1",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "n1",
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
								ArtifactTag: artifactTags[1],
							},
						},
					},
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 4,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "n2",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "n2",
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
								ArtifactTag: artifactTags[2],
							},
						},
					},
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 5,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "end-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "end-node",
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
			"n1": {
				{
					BaseModel: models.BaseModel{
						ID: 2,
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
							NodeID: "n1",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n1-0",
						},
					}),
				},
			},
			"n2": {
				{
					BaseModel: models.BaseModel{
						ID: 3,
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
							NodeID: "n2",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n2-0",
						},
					}),
				},
			},
		}

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, &executionModel, nodeExecutionModels,
			taskExecutionModels)
		deletedArtifactIDs := setupCacheEvictionCatalogClient(t, catalogClient, artifactTags, taskExecutionModels)

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictExecutionCacheRequest{
			WorkflowExecutionId: &executionIdentifier,
		}
		resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
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

	t.Run("single subtask", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		artifactTags := []*core.CatalogArtifactTag{
			{
				ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
			{
				ArtifactId: "4074be3e-7cee-4a7b-8c45-56577fa32f24",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYN",
			},
		}

		executionModel := models.Execution{
			BaseModel: models.BaseModel{
				ID: 1,
			},
			ExecutionKey: models.ExecutionKey{
				Project: executionIdentifier.Project,
				Domain:  executionIdentifier.Domain,
				Name:    executionIdentifier.Name,
			},
			LaunchPlanID: 6,
			WorkflowID:   4,
			TaskID:       0,
			Phase:        core.NodeExecution_SUCCEEDED.String(),
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
					NodeID: "start-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "start-node",
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 2,
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
					IsParentNode: true,
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
				ChildNodeExecutions: []models.NodeExecution{
					{
						BaseModel: models.BaseModel{
							ID: 3,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-start-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "start-node",
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 4,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-n0",
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
											Name:         "flyte_task-test.evict.execution_cache_sub",
											Version:      "version",
										},
										ArtifactTag: artifactTags[1],
										SourceExecution: &core.CatalogMetadata_SourceTaskExecution{
											SourceTaskExecution: &core.TaskExecutionIdentifier{
												TaskId: &core.Identifier{
													ResourceType: core.ResourceType_TASK,
													Project:      executionIdentifier.Project,
													Domain:       executionIdentifier.Domain,
													Name:         "flyte_task-test.evict.execution_cache",
													Version:      "version",
												},
												NodeExecutionId: &core.NodeExecutionIdentifier{
													NodeId:      "dn0",
													ExecutionId: &executionIdentifier,
												},
												RetryAttempt: 0,
											},
										},
									},
								},
							},
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 5,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-end-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "end-node",
						}),
					},
				},
			},
			{
				BaseModel: models.BaseModel{
					ID: 6,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "end-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "end-node",
				}),
			},
		}

		taskExecutionModels := map[string][]models.TaskExecution{
			"n0": {
				{
					BaseModel: models.BaseModel{
						ID: 2,
					},
					TaskExecutionKey: models.TaskExecutionKey{
						TaskKey: models.TaskKey{
							Project: executionIdentifier.Project,
							Domain:  executionIdentifier.Domain,
							Name:    "flyte_task-test.evict.execution_cache_single_task",
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
							GeneratedName: "name-n0-0-n0-0",
						},
					}),
				},
			},
			"n0-0-n0": {
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
							NodeID: "n0-0-n0",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0-n0-0",
						},
					}),
				},
			},
		}

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, &executionModel, nodeExecutionModels,
			taskExecutionModels)
		deletedArtifactIDs := setupCacheEvictionCatalogClient(t, catalogClient, artifactTags, taskExecutionModels)

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictExecutionCacheRequest{
			WorkflowExecutionId: &executionIdentifier,
		}
		resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.GetErrors().GetErrors())

		for nodeID := range taskExecutionModels {
			if len(taskExecutionModels[nodeID]) > 0 {
				assert.Contains(t, updatedNodeExecutions, nodeID)
			} else {
				assert.NotContains(t, updatedNodeExecutions, nodeID)
			}
		}
		assert.Len(t, updatedNodeExecutions, len(taskExecutionModels))

		for _, artifactTag := range artifactTags {
			assert.Equal(t, 1, deletedArtifactIDs[artifactTag.GetArtifactId()])
		}
		assert.Len(t, deletedArtifactIDs, len(artifactTags))
	})

	t.Run("multiple subtasks", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		artifactTags := []*core.CatalogArtifactTag{
			{
				ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
			{
				ArtifactId: "4074be3e-7cee-4a7b-8c45-56577fa32f24",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYN",
			},
			{
				ArtifactId: "a8bd60d5-b2bb-4b06-a2ac-240d183a4ca8",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYO",
			},
			{
				ArtifactId: "8a47c342-ff71-481e-9c7b-0e6ecb57e742",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYP",
			},
			{
				ArtifactId: "dafdef15-0aba-4f7c-a4aa-deba89568277",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYQ",
			},
			{
				ArtifactId: "55bbd39e-ff7d-42c8-a03b-bcc5f4019a0f",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYR",
			},
		}

		executionModel := models.Execution{
			BaseModel: models.BaseModel{
				ID: 1,
			},
			ExecutionKey: models.ExecutionKey{
				Project: executionIdentifier.Project,
				Domain:  executionIdentifier.Domain,
				Name:    executionIdentifier.Name,
			},
			LaunchPlanID: 6,
			WorkflowID:   4,
			TaskID:       0,
			Phase:        core.NodeExecution_SUCCEEDED.String(),
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
					NodeID: "start-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "start-node",
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 2,
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
					IsParentNode: true,
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
				ChildNodeExecutions: []models.NodeExecution{
					{
						BaseModel: models.BaseModel{
							ID: 3,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-start-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "start-node",
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 4,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-n0",
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
											Name:         "flyte_task-test.evict.execution_cache_sub",
											Version:      "version",
										},
										ArtifactTag: artifactTags[1],
										SourceExecution: &core.CatalogMetadata_SourceTaskExecution{
											SourceTaskExecution: &core.TaskExecutionIdentifier{
												TaskId: &core.Identifier{
													ResourceType: core.ResourceType_TASK,
													Project:      executionIdentifier.Project,
													Domain:       executionIdentifier.Domain,
													Name:         "flyte_task-test.evict.execution_cache",
													Version:      "version",
												},
												NodeExecutionId: &core.NodeExecutionIdentifier{
													NodeId:      "dn0",
													ExecutionId: &executionIdentifier,
												},
												RetryAttempt: 0,
											},
										},
									},
								},
							},
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 5,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-end-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "end-node",
						}),
					},
				},
			},
			{
				BaseModel: models.BaseModel{
					ID: 6,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "n1",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: true,
					IsDynamic:    false,
					SpecNodeId:   "n1",
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
								ArtifactTag: artifactTags[2],
							},
						},
					},
				}),
				ChildNodeExecutions: []models.NodeExecution{
					{
						BaseModel: models.BaseModel{
							ID: 7,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n1-0-start-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "start-node",
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 8,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n1-0-n0",
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
											Name:         "flyte_task-test.evict.execution_cache_sub",
											Version:      "version",
										},
										ArtifactTag: artifactTags[3],
										SourceExecution: &core.CatalogMetadata_SourceTaskExecution{
											SourceTaskExecution: &core.TaskExecutionIdentifier{
												TaskId: &core.Identifier{
													ResourceType: core.ResourceType_TASK,
													Project:      executionIdentifier.Project,
													Domain:       executionIdentifier.Domain,
													Name:         "flyte_task-test.evict.execution_cache",
													Version:      "version",
												},
												NodeExecutionId: &core.NodeExecutionIdentifier{
													NodeId:      "dn0",
													ExecutionId: &executionIdentifier,
												},
												RetryAttempt: 0,
											},
										},
									},
								},
							},
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 9,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n1-0-end-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "end-node",
						}),
					},
				},
			},
			{
				BaseModel: models.BaseModel{
					ID: 10,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "n2",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: true,
					IsDynamic:    false,
					SpecNodeId:   "n2",
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
								ArtifactTag: artifactTags[4],
							},
						},
					},
				}),
				ChildNodeExecutions: []models.NodeExecution{
					{
						BaseModel: models.BaseModel{
							ID: 11,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n2-0-start-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "start-node",
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 12,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n2-0-n0",
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
											Name:         "flyte_task-test.evict.execution_cache_sub",
											Version:      "version",
										},
										ArtifactTag: artifactTags[5],
										SourceExecution: &core.CatalogMetadata_SourceTaskExecution{
											SourceTaskExecution: &core.TaskExecutionIdentifier{
												TaskId: &core.Identifier{
													ResourceType: core.ResourceType_TASK,
													Project:      executionIdentifier.Project,
													Domain:       executionIdentifier.Domain,
													Name:         "flyte_task-test.evict.execution_cache",
													Version:      "version",
												},
												NodeExecutionId: &core.NodeExecutionIdentifier{
													NodeId:      "dn0",
													ExecutionId: &executionIdentifier,
												},
												RetryAttempt: 0,
											},
										},
									},
								},
							},
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 13,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n2-0-end-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "end-node",
						}),
					},
				},
			},
			{
				BaseModel: models.BaseModel{
					ID: 14,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "end-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "end-node",
				}),
			},
		}

		taskExecutionModels := map[string][]models.TaskExecution{
			"n0": {
				{
					BaseModel: models.BaseModel{
						ID: 4,
					},
					TaskExecutionKey: models.TaskExecutionKey{
						TaskKey: models.TaskKey{
							Project: executionIdentifier.Project,
							Domain:  executionIdentifier.Domain,
							Name:    "flyte_task-test.evict.execution_cache_single_task",
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
							GeneratedName: "name-n0-0-n0-0",
						},
					}),
				},
			},
			"n1": {
				{
					BaseModel: models.BaseModel{
						ID: 5,
					},
					TaskExecutionKey: models.TaskExecutionKey{
						TaskKey: models.TaskKey{
							Project: executionIdentifier.Project,
							Domain:  executionIdentifier.Domain,
							Name:    "flyte_task-test.evict.execution_cache_single_task",
							Version: "version",
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n1",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n1-0-n0-0",
						},
					}),
				},
			},
			"n2": {
				{
					BaseModel: models.BaseModel{
						ID: 6,
					},
					TaskExecutionKey: models.TaskExecutionKey{
						TaskKey: models.TaskKey{
							Project: executionIdentifier.Project,
							Domain:  executionIdentifier.Domain,
							Name:    "flyte_task-test.evict.execution_cache_single_task",
							Version: "version",
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n2",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n2-0-n0-0",
						},
					}),
				},
			},
			"n0-0-n0": {
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
							NodeID: "n0-0-n0",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0-n0-0",
						},
					}),
				},
			},
			"n1-0-n0": {
				{
					BaseModel: models.BaseModel{
						ID: 2,
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
							NodeID: "n1-0-n0",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n1-0-n0-0",
						},
					}),
				},
			},
			"n2-0-n0": {
				{
					BaseModel: models.BaseModel{
						ID: 3,
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
							NodeID: "n2-0-n0",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n2-0-n0-0",
						},
					}),
				},
			},
		}

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, &executionModel, nodeExecutionModels,
			taskExecutionModels)
		deletedArtifactIDs := setupCacheEvictionCatalogClient(t, catalogClient, artifactTags, taskExecutionModels)

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictExecutionCacheRequest{
			WorkflowExecutionId: &executionIdentifier,
		}
		resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.GetErrors().GetErrors())

		for nodeID := range taskExecutionModels {
			if len(taskExecutionModels[nodeID]) > 0 {
				assert.Contains(t, updatedNodeExecutions, nodeID)
			} else {
				assert.NotContains(t, updatedNodeExecutions, nodeID)
			}
		}
		assert.Len(t, updatedNodeExecutions, 6)

		for _, artifactTag := range artifactTags {
			assert.Equal(t, 1, deletedArtifactIDs[artifactTag.GetArtifactId()])
		}
		assert.Len(t, deletedArtifactIDs, len(artifactTags))
	})

	t.Run("single launch plan", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		artifactTags := []*core.CatalogArtifactTag{
			{
				ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
			{
				ArtifactId: "4074be3e-7cee-4a7b-8c45-56577fa32f24",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYN",
			},
		}

		childExecutionIdentifier := core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "childname",
		}

		executionModels := map[string]models.Execution{
			executionIdentifier.Name: {
				BaseModel: models.BaseModel{
					ID: 1,
				},
				ExecutionKey: models.ExecutionKey{
					Project: executionIdentifier.Project,
					Domain:  executionIdentifier.Domain,
					Name:    executionIdentifier.Name,
				},
				LaunchPlanID: 6,
				WorkflowID:   4,
				TaskID:       0,
				Phase:        core.NodeExecution_SUCCEEDED.String(),
			},
			childExecutionIdentifier.Name: {
				BaseModel: models.BaseModel{
					ID: 2,
				},
				ExecutionKey: models.ExecutionKey{
					Project: childExecutionIdentifier.Project,
					Domain:  childExecutionIdentifier.Domain,
					Name:    childExecutionIdentifier.Name,
				},
				LaunchPlanID: 7,
				WorkflowID:   5,
				TaskID:       0,
				Phase:        core.NodeExecution_SUCCEEDED.String(),
			},
		}

		nodeExecutionModels := map[string][]models.NodeExecution{
			executionIdentifier.Name: {
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
						NodeID: "start-node",
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
						IsParentNode: false,
						IsDynamic:    false,
						SpecNodeId:   "start-node",
					}),
				},
				{
					BaseModel: models.BaseModel{
						ID: 2,
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
						IsParentNode: true,
						IsDynamic:    true,
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
					ChildNodeExecutions: []models.NodeExecution{
						{
							BaseModel: models.BaseModel{
								ID: 3,
							},
							NodeExecutionKey: models.NodeExecutionKey{
								ExecutionKey: models.ExecutionKey{
									Project: executionIdentifier.Project,
									Domain:  executionIdentifier.Domain,
									Name:    executionIdentifier.Name,
								},
								NodeID: "n0-0-start-node",
							},
							Phase: core.NodeExecution_SUCCEEDED.String(),
							NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
								IsParentNode: false,
								IsDynamic:    false,
								SpecNodeId:   "start-node",
							}),
						},
						{
							BaseModel: models.BaseModel{
								ID: 4,
							},
							NodeExecutionKey: models.NodeExecutionKey{
								ExecutionKey: models.ExecutionKey{
									Project: executionIdentifier.Project,
									Domain:  executionIdentifier.Domain,
									Name:    executionIdentifier.Name,
								},
								NodeID: "n0-0-dn0",
							},
							Phase: core.NodeExecution_SUCCEEDED.String(),
							NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
								IsParentNode: false,
								IsDynamic:    false,
								SpecNodeId:   "dn0",
							}),
							Closure: serializeNodeExecutionClosure(t, &admin.NodeExecutionClosure{
								TargetMetadata: &admin.NodeExecutionClosure_WorkflowNodeMetadata{
									WorkflowNodeMetadata: &admin.WorkflowNodeMetadata{
										ExecutionId: &childExecutionIdentifier,
									},
								},
							}),
						},
						{
							BaseModel: models.BaseModel{
								ID: 5,
							},
							NodeExecutionKey: models.NodeExecutionKey{
								ExecutionKey: models.ExecutionKey{
									Project: executionIdentifier.Project,
									Domain:  executionIdentifier.Domain,
									Name:    executionIdentifier.Name,
								},
								NodeID: "n0-0-end-node",
							},
							Phase: core.NodeExecution_SUCCEEDED.String(),
							NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
								IsParentNode: false,
								IsDynamic:    false,
								SpecNodeId:   "end-node",
							}),
						},
					},
				},
				{
					BaseModel: models.BaseModel{
						ID: 6,
					},
					NodeExecutionKey: models.NodeExecutionKey{
						ExecutionKey: models.ExecutionKey{
							Project: executionIdentifier.Project,
							Domain:  executionIdentifier.Domain,
							Name:    executionIdentifier.Name,
						},
						NodeID: "end-node",
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
						IsParentNode: false,
						IsDynamic:    false,
						SpecNodeId:   "end-node",
					}),
				},
			},
			childExecutionIdentifier.Name: {
				{
					BaseModel: models.BaseModel{
						ID: 7,
					},
					NodeExecutionKey: models.NodeExecutionKey{
						ExecutionKey: models.ExecutionKey{
							Project: childExecutionIdentifier.Project,
							Domain:  childExecutionIdentifier.Domain,
							Name:    childExecutionIdentifier.Name,
						},
						NodeID: "start-node",
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
						IsParentNode: false,
						IsDynamic:    false,
						SpecNodeId:   "start-node",
					}),
				},
				{
					BaseModel: models.BaseModel{
						ID: 8,
					},
					NodeExecutionKey: models.NodeExecutionKey{
						ExecutionKey: models.ExecutionKey{
							Project: childExecutionIdentifier.Project,
							Domain:  childExecutionIdentifier.Domain,
							Name:    childExecutionIdentifier.Name,
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
										Project:      childExecutionIdentifier.Project,
										Domain:       childExecutionIdentifier.Domain,
										Name:         "flyte_task-test.evict.execution_cache_single_task",
										Version:      "version",
									},
									ArtifactTag: artifactTags[1],
								},
							},
						},
					}),
				},
				{
					BaseModel: models.BaseModel{
						ID: 9,
					},
					NodeExecutionKey: models.NodeExecutionKey{
						ExecutionKey: models.ExecutionKey{
							Project: childExecutionIdentifier.Project,
							Domain:  childExecutionIdentifier.Domain,
							Name:    childExecutionIdentifier.Name,
						},
						NodeID: "end-node",
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
						IsParentNode: false,
						IsDynamic:    false,
						SpecNodeId:   "end-node",
					}),
				},
			},
		}

		taskExecutionModels := map[string]map[string][]models.TaskExecution{
			executionIdentifier.Name: {
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
				"n0-0-dn0": {},
			},
			childExecutionIdentifier.Name: {
				"n0": {
					{
						BaseModel: models.BaseModel{
							ID: 2,
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
								GeneratedName: "childname-n0-0",
							},
						}),
					},
				},
			},
		}

		repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
			func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
				execution, ok := executionModels[input.Name]
				require.True(t, ok)
				return execution, nil
			})

		repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(
			func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
				nodeExecutions, ok := nodeExecutionModels[input.NodeExecutionIdentifier.ExecutionId.Name]
				require.True(t, ok)
				for _, nodeExecution := range nodeExecutions {
					if nodeExecution.NodeID == input.NodeExecutionIdentifier.NodeId {
						return nodeExecution, nil
					}
				}
				return models.NodeExecution{}, errors.New("not found")
			})

		repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetListCallback(
			func(ctx context.Context, input interfaces.ListResourceInput) (interfaces.NodeExecutionCollectionOutput, error) {
				var executionName string
				var parentID uint
				for _, filter := range input.InlineFilters {
					if filter.GetField() == "execution_name" {
						query, err := filter.GetGormJoinTableQueryExpr("")
						require.NoError(t, err)
						var ok bool
						executionName, ok = query.Args.(string)
						require.True(t, ok)
					} else if filter.GetField() == "parent_id" {
						query, err := filter.GetGormJoinTableQueryExpr("")
						require.NoError(t, err)
						var ok bool
						parentID, ok = query.Args.(uint)
						require.True(t, ok)
					}
				}

				nodeExecutions, ok := nodeExecutionModels[executionName]
				require.True(t, ok)

				if parentID == 0 {
					return interfaces.NodeExecutionCollectionOutput{NodeExecutions: nodeExecutions}, nil
				}
				for _, nodeExecution := range nodeExecutions {
					if nodeExecution.ID == parentID {
						return interfaces.NodeExecutionCollectionOutput{NodeExecutions: nodeExecution.ChildNodeExecutions}, nil
					}
				}
				return interfaces.NodeExecutionCollectionOutput{}, errors.New("not found")
			})

		updatedNodeExecutions := make(map[string]map[string]struct{})
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
				assert.Nil(t, md.TaskNodeMetadata.CatalogKey)

				if _, ok := updatedNodeExecutions[nodeExecution.Name]; !ok {
					updatedNodeExecutions[nodeExecution.Name] = make(map[string]struct{})
				}
				updatedNodeExecutions[nodeExecution.Name][nodeExecution.NodeID] = struct{}{}
				return nil
			})

		repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetListCallback(
			func(ctx context.Context, input interfaces.ListResourceInput) (interfaces.TaskExecutionCollectionOutput, error) {
				var executionName string
				var nodeID string
				for _, filter := range input.InlineFilters {
					if filter.GetField() == "execution_name" {
						query, err := filter.GetGormJoinTableQueryExpr("")
						require.NoError(t, err)
						var ok bool
						executionName, ok = query.Args.(string)
						require.True(t, ok)
					} else if filter.GetField() == "node_id" {
						query, err := filter.GetGormJoinTableQueryExpr("")
						require.NoError(t, err)
						var ok bool
						nodeID, ok = query.Args.(string)
						require.True(t, ok)
					}
				}
				require.NotEmpty(t, nodeID)
				taskExecutions, ok := taskExecutionModels[executionName]
				require.True(t, ok)
				executions, ok := taskExecutions[nodeID]
				require.True(t, ok)
				return interfaces.TaskExecutionCollectionOutput{TaskExecutions: executions}, nil
			})

		repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetGetCallback(
			func(ctx context.Context, input interfaces.GetTaskExecutionInput) (models.TaskExecution, error) {
				taskExecutions, ok := taskExecutionModels[input.TaskExecutionID.NodeExecutionId.ExecutionId.Name]
				require.True(t, ok)
				executions, ok := taskExecutions[input.TaskExecutionID.NodeExecutionId.NodeId]
				require.True(t, ok)
				return executions[0], nil
			})

		for executionName := range taskExecutionModels {
			for nodeID := range taskExecutionModels[executionName] {
				for _, taskExecution := range taskExecutionModels[executionName][nodeID] {
					require.NotNil(t, taskExecution.RetryAttempt)
					ownerID := fmt.Sprintf("%s-%s-%d", executionName, nodeID, *taskExecution.RetryAttempt)
					for _, artifactTag := range artifactTags {
						catalogClient.On("GetOrExtendReservationByArtifactTag", mock.Anything, mock.Anything,
							artifactTag.GetName(), ownerID, mock.Anything).Return(&datacatalog.Reservation{OwnerId: ownerID}, nil)
						catalogClient.On("ReleaseReservationByArtifactTag", mock.Anything, mock.Anything,
							artifactTag.GetName(), ownerID).Return(nil)
					}
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

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictExecutionCacheRequest{
			WorkflowExecutionId: &executionIdentifier,
		}
		resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.GetErrors().GetErrors())

		for executionName := range taskExecutionModels {
			require.Contains(t, updatedNodeExecutions, executionName)
			for nodeID := range taskExecutionModels[executionName] {
				if len(taskExecutionModels[executionName][nodeID]) > 0 {
					assert.Contains(t, updatedNodeExecutions[executionName], nodeID)
				} else {
					assert.NotContains(t, updatedNodeExecutions[executionName], nodeID)
				}
			}
			assert.Len(t, updatedNodeExecutions[executionName], 1)
		}

		for _, artifactTag := range artifactTags {
			assert.Equal(t, 1, deletedArtifactIDs[artifactTag.GetArtifactId()])
		}
		assert.Len(t, deletedArtifactIDs, len(artifactTags))
	})

	t.Run("single dynamic task", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		artifactTags := []*core.CatalogArtifactTag{
			{
				ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
			{
				ArtifactId: "4074be3e-7cee-4a7b-8c45-56577fa32f24",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYN",
			},
			{
				ArtifactId: "a8bd60d5-b2bb-4b06-a2ac-240d183a4ca8",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYO",
			},
			{
				ArtifactId: "8a47c342-ff71-481e-9c7b-0e6ecb57e742",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYP",
			},
			{
				ArtifactId: "dafdef15-0aba-4f7c-a4aa-deba89568277",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYQ",
			},
		}

		executionModel := models.Execution{
			BaseModel: models.BaseModel{
				ID: 1,
			},
			ExecutionKey: models.ExecutionKey{
				Project: executionIdentifier.Project,
				Domain:  executionIdentifier.Domain,
				Name:    executionIdentifier.Name,
			},
			LaunchPlanID: 6,
			WorkflowID:   4,
			TaskID:       0,
			Phase:        core.NodeExecution_SUCCEEDED.String(),
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
					NodeID: "start-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "start-node",
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 2,
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
					IsParentNode: true,
					IsDynamic:    true,
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
									Name:         "flyte_task-test.evict.execution_cache_dynamic",
									Version:      "version",
								},
								ArtifactTag: artifactTags[0],
							},
						},
					},
				}),
				ChildNodeExecutions: []models.NodeExecution{
					{
						BaseModel: models.BaseModel{
							ID: 3,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-start-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "start-node",
						}),
						ParentID: ptr[uint](2),
					},
					{
						BaseModel: models.BaseModel{
							ID: 4,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-dn0",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "dn0",
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
											Name:         "flyte_task-test.evict.execution_cache",
											Version:      "version",
										},
										ArtifactTag: artifactTags[1],
									},
								},
							},
						}),
						ParentID:    ptr[uint](2),
						CacheStatus: ptr[string](core.CatalogCacheStatus_CACHE_POPULATED.String()),
					},
					{
						BaseModel: models.BaseModel{
							ID: 5,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-dn1",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "dn1",
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
											Name:         "flyte_task-test.evict.execution_cache",
											Version:      "version",
										},
										ArtifactTag: artifactTags[2],
									},
								},
							},
						}),
						ParentID:    ptr[uint](2),
						CacheStatus: ptr[string](core.CatalogCacheStatus_CACHE_POPULATED.String()),
					},
					{
						BaseModel: models.BaseModel{
							ID: 6,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-dn2",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "dn2",
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
											Name:         "flyte_task-test.evict.execution_cache",
											Version:      "version",
										},
										ArtifactTag: artifactTags[3],
									},
								},
							},
						}),
						ParentID:    ptr[uint](2),
						CacheStatus: ptr[string](core.CatalogCacheStatus_CACHE_POPULATED.String()),
					},
					{
						BaseModel: models.BaseModel{
							ID: 7,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-dn3",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "dn3",
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
											Name:         "flyte_task-test.evict.execution_cache",
											Version:      "version",
										},
										ArtifactTag: artifactTags[4],
									},
								},
							},
						}),
						ParentID:    ptr[uint](2),
						CacheStatus: ptr[string](core.CatalogCacheStatus_CACHE_POPULATED.String()),
					},
					{
						BaseModel: models.BaseModel{
							ID: 8,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-end-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "end-node",
						}),
						ParentID: ptr[uint](2),
					},
				},
			},
			{
				BaseModel: models.BaseModel{
					ID: 9,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "end-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "end-node",
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
			"n0-0-dn0": {
				{
					BaseModel: models.BaseModel{
						ID: 2,
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
							NodeID: "n0-0-dn0",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0-dn0-0",
						},
					}),
				},
			},
			"n0-0-dn1": {
				{
					BaseModel: models.BaseModel{
						ID: 3,
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
							NodeID: "n0-0-dn1",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0-dn1-0",
						},
					}),
				},
			},
			"n0-0-dn2": {
				{
					BaseModel: models.BaseModel{
						ID: 4,
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
							NodeID: "n0-0-dn2",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0-dn2-0",
						},
					}),
				},
			},
			"n0-0-dn3": {
				{
					BaseModel: models.BaseModel{
						ID: 5,
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
							NodeID: "n0-0-dn3",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0-dn3-0",
						},
					}),
				},
			},
		}

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, &executionModel, nodeExecutionModels,
			taskExecutionModels)
		deletedArtifactIDs := setupCacheEvictionCatalogClient(t, catalogClient, artifactTags, taskExecutionModels)

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictExecutionCacheRequest{
			WorkflowExecutionId: &executionIdentifier,
		}
		resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
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

		artifactTags := []string{"flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM"}

		executionModel := models.Execution{
			BaseModel: models.BaseModel{
				ID: 1,
			},
			ExecutionKey: models.ExecutionKey{
				Project: executionIdentifier.Project,
				Domain:  executionIdentifier.Domain,
				Name:    executionIdentifier.Name,
			},
			LaunchPlanID: 6,
			WorkflowID:   4,
			TaskID:       0,
			Phase:        core.NodeExecution_SUCCEEDED.String(),
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
					NodeID: "start-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "start-node",
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 2,
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
								ArtifactTag: &core.CatalogArtifactTag{
									ArtifactId: "c7cc868e-767f-4896-a97d-9f4887826cdd",
									Name:       artifactTags[0],
								},
							},
						},
					},
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 3,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "end-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "end-node",
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

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, &executionModel, nodeExecutionModels,
			taskExecutionModels)

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictExecutionCacheRequest{
			WorkflowExecutionId: &executionIdentifier,
		}
		resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.GetErrors().GetErrors())

		assert.Empty(t, updatedNodeExecutions)
	})

	t.Run("unknown execution", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		mockConfig := getMockExecutionsConfigProvider()

		repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
			func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
				return models.Execution{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "entry not found")
			})

		catalogClient := &mocks.Client{}
		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictExecutionCacheRequest{
			WorkflowExecutionId: &executionIdentifier,
		}
		resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
		assert.Error(t, err)
		assert.True(t, pluginCatalog.IsNotFound(err))
		assert.Nil(t, resp)
	})

	t.Run("single task without cached results", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		executionModel := models.Execution{
			BaseModel: models.BaseModel{
				ID: 1,
			},
			ExecutionKey: models.ExecutionKey{
				Project: executionIdentifier.Project,
				Domain:  executionIdentifier.Domain,
				Name:    executionIdentifier.Name,
			},
			LaunchPlanID: 6,
			WorkflowID:   4,
			TaskID:       0,
			Phase:        core.NodeExecution_SUCCEEDED.String(),
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
					NodeID: "start-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "start-node",
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 2,
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
			{
				BaseModel: models.BaseModel{
					ID: 3,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "end-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "end-node",
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

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, &executionModel, nodeExecutionModels,
			taskExecutionModels)

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictExecutionCacheRequest{
			WorkflowExecutionId: &executionIdentifier,
		}
		resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.GetErrors().GetErrors())

		assert.Empty(t, updatedNodeExecutions)
	})

	t.Run("multiple tasks with partially cached results", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		artifactTags := []*core.CatalogArtifactTag{
			{
				ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
			{
				ArtifactId: "4074be3e-7cee-4a7b-8c45-56577fa32f24",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYN",
			},
		}

		executionModel := models.Execution{
			BaseModel: models.BaseModel{
				ID: 1,
			},
			ExecutionKey: models.ExecutionKey{
				Project: executionIdentifier.Project,
				Domain:  executionIdentifier.Domain,
				Name:    executionIdentifier.Name,
			},
			LaunchPlanID: 6,
			WorkflowID:   4,
			TaskID:       0,
			Phase:        core.NodeExecution_SUCCEEDED.String(),
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
					NodeID: "start-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "start-node",
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 2,
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
			{
				BaseModel: models.BaseModel{
					ID: 3,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "n1",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "n1",
				}),
				Closure: serializeNodeExecutionClosure(t, &admin.NodeExecutionClosure{
					TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
						TaskNodeMetadata: &admin.TaskNodeMetadata{
							CacheStatus: core.CatalogCacheStatus_CACHE_DISABLED,
						},
					},
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 4,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "n2",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "n2",
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
								ArtifactTag: artifactTags[1],
							},
						},
					},
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 5,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "end-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "end-node",
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
			"n1": {
				{
					BaseModel: models.BaseModel{
						ID: 2,
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
							NodeID: "n1",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n1-0",
						},
					}),
				},
			},
			"n2": {
				{
					BaseModel: models.BaseModel{
						ID: 3,
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
							NodeID: "n2",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n2-0",
						},
					}),
				},
			},
		}

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, &executionModel, nodeExecutionModels,
			taskExecutionModels)
		deletedArtifactIDs := setupCacheEvictionCatalogClient(t, catalogClient, artifactTags, taskExecutionModels)

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictExecutionCacheRequest{
			WorkflowExecutionId: &executionIdentifier,
		}
		resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.GetErrors().GetErrors())

		assert.Contains(t, updatedNodeExecutions, "n0")
		assert.Contains(t, updatedNodeExecutions, "n2")
		assert.Len(t, updatedNodeExecutions, 2)

		for _, artifactTag := range artifactTags {
			assert.Equal(t, 1, deletedArtifactIDs[artifactTag.GetArtifactId()])
		}
		assert.Len(t, deletedArtifactIDs, len(artifactTags))
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

		executionModel := models.Execution{
			BaseModel: models.BaseModel{
				ID: 1,
			},
			ExecutionKey: models.ExecutionKey{
				Project: executionIdentifier.Project,
				Domain:  executionIdentifier.Domain,
				Name:    executionIdentifier.Name,
			},
			LaunchPlanID: 6,
			WorkflowID:   4,
			TaskID:       0,
			Phase:        core.NodeExecution_SUCCEEDED.String(),
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
					NodeID: "start-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "start-node",
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 2,
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
			{
				BaseModel: models.BaseModel{
					ID: 3,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "end-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "end-node",
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

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, &executionModel, nodeExecutionModels,
			taskExecutionModels)
		deletedArtifactIDs := setupCacheEvictionCatalogClient(t, catalogClient, artifactTags, taskExecutionModels)

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictExecutionCacheRequest{
			WorkflowExecutionId: &executionIdentifier,
		}
		resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.GetErrors().GetErrors())

		resp, err = cacheManager.EvictExecutionCache(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.GetErrors().GetErrors())

		resp, err = cacheManager.EvictExecutionCache(context.Background(), request)
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

		executionModel := models.Execution{
			BaseModel: models.BaseModel{
				ID: 1,
			},
			ExecutionKey: models.ExecutionKey{
				Project: executionIdentifier.Project,
				Domain:  executionIdentifier.Domain,
				Name:    executionIdentifier.Name,
			},
			LaunchPlanID: 6,
			WorkflowID:   4,
			TaskID:       0,
			Phase:        core.NodeExecution_SUCCEEDED.String(),
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
					NodeID: "start-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "start-node",
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 2,
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
			{
				BaseModel: models.BaseModel{
					ID: 3,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "end-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "end-node",
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

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, &executionModel, nodeExecutionModels,
			taskExecutionModels)

		catalogClient.On("GetOrExtendReservationByArtifactTag", mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(&datacatalog.Reservation{OwnerId: "differentOwnerID"}, nil)

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictExecutionCacheRequest{
			WorkflowExecutionId: &executionIdentifier,
		}
		resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
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

		executionModel := models.Execution{
			BaseModel: models.BaseModel{
				ID: 1,
			},
			ExecutionKey: models.ExecutionKey{
				Project: executionIdentifier.Project,
				Domain:  executionIdentifier.Domain,
				Name:    executionIdentifier.Name,
			},
			LaunchPlanID: 6,
			WorkflowID:   4,
			TaskID:       0,
			Phase:        core.NodeExecution_SUCCEEDED.String(),
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
					NodeID: "start-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "start-node",
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 2,
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
			{
				BaseModel: models.BaseModel{
					ID: 3,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "end-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "end-node",
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

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, &executionModel, nodeExecutionModels,
			taskExecutionModels)

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
		request := service.EvictExecutionCacheRequest{
			WorkflowExecutionId: &executionIdentifier,
		}
		resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
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

			executionModel := models.Execution{
				BaseModel: models.BaseModel{
					ID: 1,
				},
				ExecutionKey: models.ExecutionKey{
					Project: executionIdentifier.Project,
					Domain:  executionIdentifier.Domain,
					Name:    executionIdentifier.Name,
				},
				LaunchPlanID: 6,
				WorkflowID:   4,
				TaskID:       0,
				Phase:        core.NodeExecution_SUCCEEDED.String(),
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
						NodeID: "start-node",
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
						IsParentNode: false,
						IsDynamic:    false,
						SpecNodeId:   "start-node",
					}),
				},
				{
					BaseModel: models.BaseModel{
						ID: 2,
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
				{
					BaseModel: models.BaseModel{
						ID: 3,
					},
					NodeExecutionKey: models.NodeExecutionKey{
						ExecutionKey: models.ExecutionKey{
							Project: executionIdentifier.Project,
							Domain:  executionIdentifier.Domain,
							Name:    executionIdentifier.Name,
						},
						NodeID: "end-node",
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
						IsParentNode: false,
						IsDynamic:    false,
						SpecNodeId:   "end-node",
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

			updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, &executionModel, nodeExecutionModels,
				taskExecutionModels)

			catalogClient.On("GetOrExtendReservationByArtifactTag", mock.Anything, mock.Anything,
				mock.Anything, mock.Anything, mock.Anything).Return(nil, status.Error(codes.Internal, "error"))

			cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
			request := service.EvictExecutionCacheRequest{
				WorkflowExecutionId: &executionIdentifier,
			}
			resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
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

			executionModel := models.Execution{
				BaseModel: models.BaseModel{
					ID: 1,
				},
				ExecutionKey: models.ExecutionKey{
					Project: executionIdentifier.Project,
					Domain:  executionIdentifier.Domain,
					Name:    executionIdentifier.Name,
				},
				LaunchPlanID: 6,
				WorkflowID:   4,
				TaskID:       0,
				Phase:        core.NodeExecution_SUCCEEDED.String(),
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
						NodeID: "start-node",
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
						IsParentNode: false,
						IsDynamic:    false,
						SpecNodeId:   "start-node",
					}),
				},
				{
					BaseModel: models.BaseModel{
						ID: 2,
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
				{
					BaseModel: models.BaseModel{
						ID: 3,
					},
					NodeExecutionKey: models.NodeExecutionKey{
						ExecutionKey: models.ExecutionKey{
							Project: executionIdentifier.Project,
							Domain:  executionIdentifier.Domain,
							Name:    executionIdentifier.Name,
						},
						NodeID: "end-node",
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
						IsParentNode: false,
						IsDynamic:    false,
						SpecNodeId:   "end-node",
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

			updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, &executionModel, nodeExecutionModels,
				taskExecutionModels)

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

			cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
			request := service.EvictExecutionCacheRequest{
				WorkflowExecutionId: &executionIdentifier,
			}
			resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Len(t, resp.GetErrors().GetErrors(), len(artifactTags))
			eErr := resp.GetErrors().GetErrors()[0]
			assert.Equal(t, core.CacheEvictionError_ARTIFACT_DELETE_FAILED, eErr.Code)
			assert.Equal(t, "n0", eErr.NodeExecutionId.NodeId)
			assert.True(t, proto.Equal(&executionIdentifier, eErr.NodeExecutionId.ExecutionId))
			require.NotNil(t, eErr.Source)
			assert.IsType(t, &core.CacheEvictionError_TaskExecutionId{}, eErr.Source)

			for nodeID := range taskExecutionModels {
				assert.Contains(t, updatedNodeExecutions, nodeID)
			}
			assert.Len(t, updatedNodeExecutions, len(taskExecutionModels))
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

			executionModel := models.Execution{
				BaseModel: models.BaseModel{
					ID: 1,
				},
				ExecutionKey: models.ExecutionKey{
					Project: executionIdentifier.Project,
					Domain:  executionIdentifier.Domain,
					Name:    executionIdentifier.Name,
				},
				LaunchPlanID: 6,
				WorkflowID:   4,
				TaskID:       0,
				Phase:        core.NodeExecution_SUCCEEDED.String(),
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
						NodeID: "start-node",
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
						IsParentNode: false,
						IsDynamic:    false,
						SpecNodeId:   "start-node",
					}),
				},
				{
					BaseModel: models.BaseModel{
						ID: 2,
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
				{
					BaseModel: models.BaseModel{
						ID: 3,
					},
					NodeExecutionKey: models.NodeExecutionKey{
						ExecutionKey: models.ExecutionKey{
							Project: executionIdentifier.Project,
							Domain:  executionIdentifier.Domain,
							Name:    executionIdentifier.Name,
						},
						NodeID: "end-node",
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
						IsParentNode: false,
						IsDynamic:    false,
						SpecNodeId:   "end-node",
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

			updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, &executionModel, nodeExecutionModels,
				taskExecutionModels)

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
			request := service.EvictExecutionCacheRequest{
				WorkflowExecutionId: &executionIdentifier,
			}
			resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
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
		t.Run("ExecutionRepo", func(t *testing.T) {
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

				executionModel := models.Execution{
					BaseModel: models.BaseModel{
						ID: 1,
					},
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					LaunchPlanID: 6,
					WorkflowID:   4,
					TaskID:       0,
					Phase:        core.NodeExecution_SUCCEEDED.String(),
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
							NodeID: "start-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "start-node",
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 2,
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
					{
						BaseModel: models.BaseModel{
							ID: 3,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "end-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "end-node",
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

				updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, &executionModel, nodeExecutionModels,
					taskExecutionModels)

				repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
					func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
						return models.Execution{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.Internal, "error")
					})

				cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
				request := service.EvictExecutionCacheRequest{
					WorkflowExecutionId: &executionIdentifier,
				}
				resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
				require.Error(t, err)
				s, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, codes.Internal, s.Code())
				assert.Nil(t, resp)

				assert.Empty(t, updatedNodeExecutions)
			})
		})

		t.Run("NodeExecutionRepo", func(t *testing.T) {
			t.Run("List", func(t *testing.T) {
				repository := repositoryMocks.NewMockRepository()
				catalogClient := &mocks.Client{}
				mockConfig := getMockExecutionsConfigProvider()

				artifactTags := []*core.CatalogArtifactTag{
					{
						ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
						Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
					},
				}

				executionModel := models.Execution{
					BaseModel: models.BaseModel{
						ID: 1,
					},
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					LaunchPlanID: 6,
					WorkflowID:   4,
					TaskID:       0,
					Phase:        core.NodeExecution_SUCCEEDED.String(),
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
							NodeID: "start-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "start-node",
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 2,
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
					{
						BaseModel: models.BaseModel{
							ID: 3,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "end-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "end-node",
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

				updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, &executionModel, nodeExecutionModels,
					taskExecutionModels)

				repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetListCallback(
					func(ctx context.Context, input interfaces.ListResourceInput) (interfaces.NodeExecutionCollectionOutput, error) {
						return interfaces.NodeExecutionCollectionOutput{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.Internal, "error")
					})

				cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
				request := service.EvictExecutionCacheRequest{
					WorkflowExecutionId: &executionIdentifier,
				}
				resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
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

				executionModel := models.Execution{
					BaseModel: models.BaseModel{
						ID: 1,
					},
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					LaunchPlanID: 6,
					WorkflowID:   4,
					TaskID:       0,
					Phase:        core.NodeExecution_SUCCEEDED.String(),
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
							NodeID: "start-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "start-node",
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 2,
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
					{
						BaseModel: models.BaseModel{
							ID: 3,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "end-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "end-node",
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

				updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, &executionModel, nodeExecutionModels,
					taskExecutionModels)

				repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetUpdateSelectedCallback(
					func(ctx context.Context, nodeExecution *models.NodeExecution, selectedFields []string) error {
						return flyteAdminErrors.NewFlyteAdminError(codes.Internal, "error")
					})

				deletedArtifactIDs := setupCacheEvictionCatalogClient(t, catalogClient, artifactTags, taskExecutionModels)

				cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
				request := service.EvictExecutionCacheRequest{
					WorkflowExecutionId: &executionIdentifier,
				}
				resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
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
				assert.Empty(t, deletedArtifactIDs)
			})
		})

		t.Run("TaskExecutionRepo", func(t *testing.T) {
			t.Run("List", func(t *testing.T) {
				repository := repositoryMocks.NewMockRepository()
				catalogClient := &mocks.Client{}
				mockConfig := getMockExecutionsConfigProvider()

				artifactTags := []*core.CatalogArtifactTag{
					{
						ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
						Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
					},
				}

				executionModel := models.Execution{
					BaseModel: models.BaseModel{
						ID: 1,
					},
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					LaunchPlanID: 6,
					WorkflowID:   4,
					TaskID:       0,
					Phase:        core.NodeExecution_SUCCEEDED.String(),
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
							NodeID: "start-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "start-node",
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 2,
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
					{
						BaseModel: models.BaseModel{
							ID: 3,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "end-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "end-node",
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

				updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, &executionModel, nodeExecutionModels,
					taskExecutionModels)

				repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetListCallback(
					func(ctx context.Context, input interfaces.ListResourceInput) (interfaces.TaskExecutionCollectionOutput, error) {
						return interfaces.TaskExecutionCollectionOutput{}, flyteAdminErrors.NewFlyteAdminError(codes.Internal, "error")
					})

				deletedArtifactIDs := setupCacheEvictionCatalogClient(t, catalogClient, artifactTags, taskExecutionModels)

				cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
				request := service.EvictExecutionCacheRequest{
					WorkflowExecutionId: &executionIdentifier,
				}
				resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Len(t, resp.GetErrors().GetErrors(), len(artifactTags))
				eErr := resp.GetErrors().GetErrors()[0]
				assert.Equal(t, core.CacheEvictionError_INTERNAL, eErr.Code)
				assert.Equal(t, "n0", eErr.NodeExecutionId.NodeId)
				assert.True(t, proto.Equal(&executionIdentifier, eErr.NodeExecutionId.ExecutionId))
				require.NotNil(t, eErr.Source)
				assert.IsType(t, &core.CacheEvictionError_WorkflowExecutionId{}, eErr.Source)

				assert.Empty(t, updatedNodeExecutions)
				assert.Empty(t, deletedArtifactIDs)
			})
		})
	})

	t.Run("multiple tasks with identical artifact tags", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		artifactTags := []*core.CatalogArtifactTag{
			{
				ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
			{
				ArtifactId: "4074be3e-7cee-4a7b-8c45-56577fa32f24",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
			{
				ArtifactId: "a8bd60d5-b2bb-4b06-a2ac-240d183a4ca8",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
			{
				ArtifactId: "8a47c342-ff71-481e-9c7b-0e6ecb57e742",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
		}

		executionModel := models.Execution{
			BaseModel: models.BaseModel{
				ID: 1,
			},
			ExecutionKey: models.ExecutionKey{
				Project: executionIdentifier.Project,
				Domain:  executionIdentifier.Domain,
				Name:    executionIdentifier.Name,
			},
			LaunchPlanID: 6,
			WorkflowID:   4,
			TaskID:       0,
			Phase:        core.NodeExecution_SUCCEEDED.String(),
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
					NodeID: "start-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "start-node",
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 2,
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
			{
				BaseModel: models.BaseModel{
					ID: 3,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "n1",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "n1",
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
								ArtifactTag: artifactTags[1],
							},
						},
					},
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 4,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "n2",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "n1",
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
								ArtifactTag: artifactTags[2],
							},
						},
					},
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 5,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "n3",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "n3",
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
								ArtifactTag: artifactTags[3],
							},
						},
					},
				}),
			},
			{
				BaseModel: models.BaseModel{
					ID: 6,
				},
				NodeExecutionKey: models.NodeExecutionKey{
					ExecutionKey: models.ExecutionKey{
						Project: executionIdentifier.Project,
						Domain:  executionIdentifier.Domain,
						Name:    executionIdentifier.Name,
					},
					NodeID: "end-node",
				},
				Phase: core.NodeExecution_SUCCEEDED.String(),
				NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
					IsParentNode: false,
					IsDynamic:    false,
					SpecNodeId:   "end-node",
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
			"n1": {
				{
					BaseModel: models.BaseModel{
						ID: 2,
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
							NodeID: "n1",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n1-0",
						},
					}),
				},
			},
			"n2": {
				{
					BaseModel: models.BaseModel{
						ID: 3,
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
							NodeID: "n2",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n2-0",
						},
					}),
				},
			},
			"n3": {
				{
					BaseModel: models.BaseModel{
						ID: 4,
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
							NodeID: "n3",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n3-0",
						},
					}),
				},
			},
		}

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, &executionModel, nodeExecutionModels,
			taskExecutionModels)
		deletedArtifactIDs := setupCacheEvictionCatalogClient(t, catalogClient, artifactTags, taskExecutionModels)

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictExecutionCacheRequest{
			WorkflowExecutionId: &executionIdentifier,
		}
		resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
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

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nil, nodeExecutionModels,
			taskExecutionModels)
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

	t.Run("subtask", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		artifactTags := []*core.CatalogArtifactTag{
			{
				ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
			{
				ArtifactId: "4074be3e-7cee-4a7b-8c45-56577fa32f24",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYN",
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
					IsParentNode: true,
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
				ChildNodeExecutions: []models.NodeExecution{
					{
						BaseModel: models.BaseModel{
							ID: 2,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-start-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "start-node",
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 3,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-n0",
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
											Name:         "flyte_task-test.evict.execution_cache_sub",
											Version:      "version",
										},
										ArtifactTag: artifactTags[1],
										SourceExecution: &core.CatalogMetadata_SourceTaskExecution{
											SourceTaskExecution: &core.TaskExecutionIdentifier{
												TaskId: &core.Identifier{
													ResourceType: core.ResourceType_TASK,
													Project:      executionIdentifier.Project,
													Domain:       executionIdentifier.Domain,
													Name:         "flyte_task-test.evict.execution_cache",
													Version:      "version",
												},
												NodeExecutionId: &core.NodeExecutionIdentifier{
													NodeId:      "dn0",
													ExecutionId: &executionIdentifier,
												},
												RetryAttempt: 0,
											},
										},
									},
								},
							},
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 4,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-end-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "end-node",
						}),
					},
				},
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
							Name:    "flyte_task-test.evict.execution_cache_single_task",
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
							GeneratedName: "name-n0-0-n0-0",
						},
					}),
				},
			},
			"n0-0-n0": {
				{
					BaseModel: models.BaseModel{
						ID: 2,
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
							NodeID: "n0-0-n0",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0-n0-0",
						},
					}),
				},
			},
		}

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nil, nodeExecutionModels,
			taskExecutionModels)
		deletedArtifactIDs := setupCacheEvictionCatalogClient(t, catalogClient, artifactTags, taskExecutionModels)

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictExecutionCacheRequest{
			WorkflowExecutionId: &executionIdentifier,
		}
		resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.GetErrors().GetErrors())

		for nodeID := range taskExecutionModels {
			if len(taskExecutionModels[nodeID]) > 0 {
				assert.Contains(t, updatedNodeExecutions, nodeID)
			} else {
				assert.NotContains(t, updatedNodeExecutions, nodeID)
			}
		}
		assert.Len(t, updatedNodeExecutions, len(taskExecutionModels))

		for _, artifactTag := range artifactTags {
			assert.Equal(t, 1, deletedArtifactIDs[artifactTag.GetArtifactId()])
		}
		assert.Len(t, deletedArtifactIDs, len(artifactTags))
	})

	t.Run("dynamic task", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		artifactTags := []*core.CatalogArtifactTag{
			{
				ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
			{
				ArtifactId: "4074be3e-7cee-4a7b-8c45-56577fa32f24",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYN",
			},
			{
				ArtifactId: "a8bd60d5-b2bb-4b06-a2ac-240d183a4ca8",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYO",
			},
			{
				ArtifactId: "8a47c342-ff71-481e-9c7b-0e6ecb57e742",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYP",
			},
			{
				ArtifactId: "dafdef15-0aba-4f7c-a4aa-deba89568277",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYQ",
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
					IsParentNode: true,
					IsDynamic:    true,
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
									Name:         "flyte_task-test.evict.execution_cache_dynamic",
									Version:      "version",
								},
								ArtifactTag: artifactTags[0],
							},
						},
					},
				}),
				ChildNodeExecutions: []models.NodeExecution{
					{
						BaseModel: models.BaseModel{
							ID: 2,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-start-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "start-node",
						}),
						ParentID: ptr[uint](2),
					},
					{
						BaseModel: models.BaseModel{
							ID: 3,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-dn0",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "dn0",
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
											Name:         "flyte_task-test.evict.execution_cache",
											Version:      "version",
										},
										ArtifactTag: artifactTags[1],
									},
								},
							},
						}),
						ParentID:    ptr[uint](2),
						CacheStatus: ptr[string](core.CatalogCacheStatus_CACHE_POPULATED.String()),
					},
					{
						BaseModel: models.BaseModel{
							ID: 4,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-dn1",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "dn1",
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
											Name:         "flyte_task-test.evict.execution_cache",
											Version:      "version",
										},
										ArtifactTag: artifactTags[2],
									},
								},
							},
						}),
						ParentID:    ptr[uint](2),
						CacheStatus: ptr[string](core.CatalogCacheStatus_CACHE_POPULATED.String()),
					},
					{
						BaseModel: models.BaseModel{
							ID: 5,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-dn2",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "dn2",
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
											Name:         "flyte_task-test.evict.execution_cache",
											Version:      "version",
										},
										ArtifactTag: artifactTags[3],
									},
								},
							},
						}),
						ParentID:    ptr[uint](2),
						CacheStatus: ptr[string](core.CatalogCacheStatus_CACHE_POPULATED.String()),
					},
					{
						BaseModel: models.BaseModel{
							ID: 6,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-dn3",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "dn3",
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
											Name:         "flyte_task-test.evict.execution_cache",
											Version:      "version",
										},
										ArtifactTag: artifactTags[4],
									},
								},
							},
						}),
						ParentID:    ptr[uint](2),
						CacheStatus: ptr[string](core.CatalogCacheStatus_CACHE_POPULATED.String()),
					},
					{
						BaseModel: models.BaseModel{
							ID: 7,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-end-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "end-node",
						}),
						ParentID: ptr[uint](2),
					},
				},
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
			"n0-0-dn0": {
				{
					BaseModel: models.BaseModel{
						ID: 2,
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
							NodeID: "n0-0-dn0",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0-dn0-0",
						},
					}),
				},
			},
			"n0-0-dn1": {
				{
					BaseModel: models.BaseModel{
						ID: 3,
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
							NodeID: "n0-0-dn1",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0-dn1-0",
						},
					}),
				},
			},
			"n0-0-dn2": {
				{
					BaseModel: models.BaseModel{
						ID: 4,
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
							NodeID: "n0-0-dn2",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0-dn2-0",
						},
					}),
				},
			},
			"n0-0-dn3": {
				{
					BaseModel: models.BaseModel{
						ID: 5,
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
							NodeID: "n0-0-dn3",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0-dn3-0",
						},
					}),
				},
			},
		}

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nil, nodeExecutionModels,
			taskExecutionModels)
		deletedArtifactIDs := setupCacheEvictionCatalogClient(t, catalogClient, artifactTags, taskExecutionModels)

		cacheManager := NewCacheManager(repository, mockConfig, catalogClient, promutils.NewTestScope())
		request := service.EvictExecutionCacheRequest{
			WorkflowExecutionId: &executionIdentifier,
		}
		resp, err := cacheManager.EvictExecutionCache(context.Background(), request)
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

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nil, nodeExecutionModels,
			taskExecutionModels)

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

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nil, nodeExecutionModels,
			taskExecutionModels)

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

	t.Run("subtask with partially cached results", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		artifactTags := []*core.CatalogArtifactTag{
			{
				ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
			{
				ArtifactId: "4074be3e-7cee-4a7b-8c45-56577fa32f24",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYN",
			},
			{
				ArtifactId: "a8bd60d5-b2bb-4b06-a2ac-240d183a4ca8",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYO",
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
					IsParentNode: true,
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
				ChildNodeExecutions: []models.NodeExecution{
					{
						BaseModel: models.BaseModel{
							ID: 2,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-start-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "start-node",
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 3,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-n0",
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
											Name:         "flyte_task-test.evict.execution_cache_sub",
											Version:      "version",
										},
										ArtifactTag: artifactTags[1],
										SourceExecution: &core.CatalogMetadata_SourceTaskExecution{
											SourceTaskExecution: &core.TaskExecutionIdentifier{
												TaskId: &core.Identifier{
													ResourceType: core.ResourceType_TASK,
													Project:      executionIdentifier.Project,
													Domain:       executionIdentifier.Domain,
													Name:         "flyte_task-test.evict.execution_cache",
													Version:      "version",
												},
												NodeExecutionId: &core.NodeExecutionIdentifier{
													NodeId:      "dn0",
													ExecutionId: &executionIdentifier,
												},
												RetryAttempt: 0,
											},
										},
									},
								},
							},
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 4,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-n1",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "n1",
						}),
						Closure: serializeNodeExecutionClosure(t, &admin.NodeExecutionClosure{
							TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
								TaskNodeMetadata: &admin.TaskNodeMetadata{
									CacheStatus: core.CatalogCacheStatus_CACHE_DISABLED,
								},
							},
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 5,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-n2",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "n2",
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
											Name:         "flyte_task-test.evict.execution_cache_sub",
											Version:      "version",
										},
										ArtifactTag: artifactTags[2],
										SourceExecution: &core.CatalogMetadata_SourceTaskExecution{
											SourceTaskExecution: &core.TaskExecutionIdentifier{
												TaskId: &core.Identifier{
													ResourceType: core.ResourceType_TASK,
													Project:      executionIdentifier.Project,
													Domain:       executionIdentifier.Domain,
													Name:         "flyte_task-test.evict.execution_cache",
													Version:      "version",
												},
												NodeExecutionId: &core.NodeExecutionIdentifier{
													NodeId:      "dn2",
													ExecutionId: &executionIdentifier,
												},
												RetryAttempt: 0,
											},
										},
									},
								},
							},
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 6,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-end-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "end-node",
						}),
					},
				},
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
							Name:    "flyte_task-test.evict.execution_cache_single_task",
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
							GeneratedName: "name-n0-0-n0-0",
						},
					}),
				},
			},
			"n0-0-n0": {
				{
					BaseModel: models.BaseModel{
						ID: 2,
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
							NodeID: "n0-0-n0",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0-n0-0",
						},
					}),
				},
			},
			"n0-0-n1": {
				{
					BaseModel: models.BaseModel{
						ID: 2,
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
							NodeID: "n0-0-n1",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0-n1-0",
						},
					}),
				},
			},
			"n0-0-n2": {
				{
					BaseModel: models.BaseModel{
						ID: 2,
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
							NodeID: "n0-0-n2",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0-n2-0",
						},
					}),
				},
			},
		}

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nil, nodeExecutionModels,
			taskExecutionModels)
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

		assert.Contains(t, updatedNodeExecutions, "n0")
		assert.Contains(t, updatedNodeExecutions, "n0-0-n0")
		assert.Contains(t, updatedNodeExecutions, "n0-0-n2")
		assert.Len(t, updatedNodeExecutions, 3)

		for _, artifactTag := range artifactTags {
			assert.Equal(t, 1, deletedArtifactIDs[artifactTag.GetArtifactId()])
		}
		assert.Len(t, deletedArtifactIDs, len(artifactTags))
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

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nil, nodeExecutionModels,
			taskExecutionModels)
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

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nil, nodeExecutionModels,
			taskExecutionModels)

		catalogClient.On("GetOrExtendReservationByArtifactTag", mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(&datacatalog.Reservation{OwnerId: "otherOwnerID"}, nil)

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

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nil, nodeExecutionModels,
			taskExecutionModels)

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

			updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nil, nodeExecutionModels,
				taskExecutionModels)

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

			updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nil, nodeExecutionModels,
				taskExecutionModels)

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

			for nodeID := range taskExecutionModels {
				assert.Contains(t, updatedNodeExecutions, nodeID)
			}
			assert.Len(t, updatedNodeExecutions, len(taskExecutionModels))
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

			updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nil, nodeExecutionModels,
				taskExecutionModels)

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

				updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nil, nodeExecutionModels,
					taskExecutionModels)

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

				updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nil, nodeExecutionModels,
					taskExecutionModels)

				repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetUpdateSelectedCallback(
					func(ctx context.Context, nodeExecution *models.NodeExecution, selectedFields []string) error {
						return flyteAdminErrors.NewFlyteAdminError(codes.Internal, "error")
					})

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
				assert.Len(t, resp.GetErrors().GetErrors(), len(artifactTags))
				eErr := resp.GetErrors().GetErrors()[0]
				assert.Equal(t, core.CacheEvictionError_DATABASE_UPDATE_FAILED, eErr.Code)
				assert.Equal(t, "n0", eErr.NodeExecutionId.NodeId)
				assert.True(t, proto.Equal(&executionIdentifier, eErr.NodeExecutionId.ExecutionId))
				require.NotNil(t, eErr.Source)
				assert.IsType(t, &core.CacheEvictionError_TaskExecutionId{}, eErr.Source)

				assert.Empty(t, updatedNodeExecutions)
				assert.Empty(t, deletedArtifactIDs)
			})
		})

		t.Run("TaskExecutionRepo", func(t *testing.T) {
			t.Run("List", func(t *testing.T) {
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

				updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nil, nodeExecutionModels,
					taskExecutionModels)

				repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetListCallback(
					func(ctx context.Context, input interfaces.ListResourceInput) (interfaces.TaskExecutionCollectionOutput, error) {
						return interfaces.TaskExecutionCollectionOutput{}, flyteAdminErrors.NewFlyteAdminError(codes.Internal, "error")
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
				assert.Equal(t, core.CacheEvictionError_INTERNAL, eErr.Code)
				assert.Equal(t, "n0", eErr.NodeExecutionId.NodeId)
				assert.True(t, proto.Equal(&executionIdentifier, eErr.NodeExecutionId.ExecutionId))
				require.NotNil(t, eErr.Source)
				assert.IsType(t, &core.CacheEvictionError_WorkflowExecutionId{}, eErr.Source)

				assert.Empty(t, updatedNodeExecutions)
			})

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

				updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nil, nodeExecutionModels,
					taskExecutionModels)

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

	t.Run("subtask with identical artifact tags", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		catalogClient := &mocks.Client{}
		mockConfig := getMockExecutionsConfigProvider()

		artifactTags := []*core.CatalogArtifactTag{
			{
				ArtifactId: "0285ddb9-ddfb-4835-bc22-80e1bdf7f560",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYM",
			},
			{
				ArtifactId: "4074be3e-7cee-4a7b-8c45-56577fa32f24",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYN",
			},
			{
				ArtifactId: "a8bd60d5-b2bb-4b06-a2ac-240d183a4ca8",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYN",
			},
			{
				ArtifactId: "8a47c342-ff71-481e-9c7b-0e6ecb57e742",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYN",
			},
			{
				ArtifactId: "dafdef15-0aba-4f7c-a4aa-deba89568277",
				Name:       "flyte_cached-G3ACyxqY0U3sEf99tLMta5vuLCOk7j9O7MStxubzxYN",
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
					IsParentNode: true,
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
				ChildNodeExecutions: []models.NodeExecution{
					{
						BaseModel: models.BaseModel{
							ID: 2,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-start-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "start-node",
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 3,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-n0",
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
											Name:         "flyte_task-test.evict.execution_cache_sub",
											Version:      "version",
										},
										ArtifactTag: artifactTags[1],
										SourceExecution: &core.CatalogMetadata_SourceTaskExecution{
											SourceTaskExecution: &core.TaskExecutionIdentifier{
												TaskId: &core.Identifier{
													ResourceType: core.ResourceType_TASK,
													Project:      executionIdentifier.Project,
													Domain:       executionIdentifier.Domain,
													Name:         "flyte_task-test.evict.execution_cache",
													Version:      "version",
												},
												NodeExecutionId: &core.NodeExecutionIdentifier{
													NodeId:      "dn0",
													ExecutionId: &executionIdentifier,
												},
												RetryAttempt: 0,
											},
										},
									},
								},
							},
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 4,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-n1",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "n1",
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
											Name:         "flyte_task-test.evict.execution_cache_sub",
											Version:      "version",
										},
										ArtifactTag: artifactTags[2],
										SourceExecution: &core.CatalogMetadata_SourceTaskExecution{
											SourceTaskExecution: &core.TaskExecutionIdentifier{
												TaskId: &core.Identifier{
													ResourceType: core.ResourceType_TASK,
													Project:      executionIdentifier.Project,
													Domain:       executionIdentifier.Domain,
													Name:         "flyte_task-test.evict.execution_cache",
													Version:      "version",
												},
												NodeExecutionId: &core.NodeExecutionIdentifier{
													NodeId:      "dn0",
													ExecutionId: &executionIdentifier,
												},
												RetryAttempt: 0,
											},
										},
									},
								},
							},
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 5,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-n2",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "n2",
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
											Name:         "flyte_task-test.evict.execution_cache_sub",
											Version:      "version",
										},
										ArtifactTag: artifactTags[3],
										SourceExecution: &core.CatalogMetadata_SourceTaskExecution{
											SourceTaskExecution: &core.TaskExecutionIdentifier{
												TaskId: &core.Identifier{
													ResourceType: core.ResourceType_TASK,
													Project:      executionIdentifier.Project,
													Domain:       executionIdentifier.Domain,
													Name:         "flyte_task-test.evict.execution_cache",
													Version:      "version",
												},
												NodeExecutionId: &core.NodeExecutionIdentifier{
													NodeId:      "dn2",
													ExecutionId: &executionIdentifier,
												},
												RetryAttempt: 0,
											},
										},
									},
								},
							},
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 6,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-n3",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "n3",
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
											Name:         "flyte_task-test.evict.execution_cache_sub",
											Version:      "version",
										},
										ArtifactTag: artifactTags[4],
										SourceExecution: &core.CatalogMetadata_SourceTaskExecution{
											SourceTaskExecution: &core.TaskExecutionIdentifier{
												TaskId: &core.Identifier{
													ResourceType: core.ResourceType_TASK,
													Project:      executionIdentifier.Project,
													Domain:       executionIdentifier.Domain,
													Name:         "flyte_task-test.evict.execution_cache",
													Version:      "version",
												},
												NodeExecutionId: &core.NodeExecutionIdentifier{
													NodeId:      "dn3",
													ExecutionId: &executionIdentifier,
												},
												RetryAttempt: 0,
											},
										},
									},
								},
							},
						}),
					},
					{
						BaseModel: models.BaseModel{
							ID: 7,
						},
						NodeExecutionKey: models.NodeExecutionKey{
							ExecutionKey: models.ExecutionKey{
								Project: executionIdentifier.Project,
								Domain:  executionIdentifier.Domain,
								Name:    executionIdentifier.Name,
							},
							NodeID: "n0-0-end-node",
						},
						Phase: core.NodeExecution_SUCCEEDED.String(),
						NodeExecutionMetadata: serializeNodeExecutionMetadata(t, &admin.NodeExecutionMetaData{
							IsParentNode: false,
							IsDynamic:    false,
							SpecNodeId:   "end-node",
						}),
					},
				},
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
							Name:    "flyte_task-test.evict.execution_cache_single_task",
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
							GeneratedName: "name-n0-0-n0-0",
						},
					}),
				},
			},
			"n0-0-n0": {
				{
					BaseModel: models.BaseModel{
						ID: 2,
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
							NodeID: "n0-0-n0",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0-n0-0",
						},
					}),
				},
			},
			"n0-0-n1": {
				{
					BaseModel: models.BaseModel{
						ID: 2,
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
							NodeID: "n0-0-n1",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0-n1-0",
						},
					}),
				},
			},
			"n0-0-n2": {
				{
					BaseModel: models.BaseModel{
						ID: 3,
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
							NodeID: "n0-0-n2",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0-n2-0",
						},
					}),
				},
			},
			"n0-0-n3": {
				{
					BaseModel: models.BaseModel{
						ID: 4,
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
							NodeID: "n0-0-n3",
						},
						RetryAttempt: ptr[uint32](0),
					},
					Phase: core.NodeExecution_SUCCEEDED.String(),
					Closure: serializeTaskExecutionClosure(t, &admin.TaskExecutionClosure{
						Metadata: &event.TaskExecutionMetadata{
							GeneratedName: "name-n0-0-n3-0",
						},
					}),
				},
			},
		}

		updatedNodeExecutions := setupCacheEvictionMockRepositories(t, repository, nil, nodeExecutionModels,
			taskExecutionModels)
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
}
