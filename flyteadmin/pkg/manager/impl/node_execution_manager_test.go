package impl

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/util"

	genModel "github.com/flyteorg/flyteadmin/pkg/repositories/gen/models"

	eventWriterMocks "github.com/flyteorg/flyteadmin/pkg/async/events/mocks"

	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyteadmin/pkg/common"
	commonMocks "github.com/flyteorg/flyteadmin/pkg/common/mocks"
	dataMocks "github.com/flyteorg/flyteadmin/pkg/data/mocks"
	flyteAdminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

var occurredAt = time.Now().UTC()
var occurredAtProto, _ = ptypes.TimestampProto(occurredAt)

var dynamicWorkflowClosure = core.CompiledWorkflowClosure{
	Primary: &core.CompiledWorkflow{
		Template: &core.WorkflowTemplate{
			Id: &core.Identifier{
				ResourceType: core.ResourceType_WORKFLOW,
				Project:      "proj",
				Domain:       "domain",
				Name:         "dynamic_wf",
				Version:      "abc123",
			},
		},
	},
}

var request = admin.NodeExecutionEventRequest{
	RequestId: "request id",
	Event: &event.NodeExecutionEvent{
		ProducerId: "propeller",
		Id: &core.NodeExecutionIdentifier{
			NodeId: "node id",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
		},
		OccurredAt: occurredAtProto,
		Phase:      core.NodeExecution_RUNNING,
		InputValue: &event.NodeExecutionEvent_InputUri{
			InputUri: "input uri",
		},
		TargetMetadata: &event.NodeExecutionEvent_TaskNodeMetadata{
			TaskNodeMetadata: &event.TaskNodeMetadata{
				DynamicWorkflow: &event.DynamicWorkflowNodeMetadata{
					Id:               dynamicWorkflowClosure.Primary.Template.Id,
					CompiledWorkflow: &dynamicWorkflowClosure,
				},
			},
		},
		EventVersion: 2,
	},
}
var internalData = genModel.NodeExecutionInternalData{
	EventVersion: 2,
}
var internalDataBytes, _ = proto.Marshal(&internalData)
var nodeExecutionIdentifier = core.NodeExecutionIdentifier{
	NodeId: "node id",
	ExecutionId: &core.WorkflowExecutionIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	},
}
var workflowExecutionIdentifier = core.WorkflowExecutionIdentifier{
	Project: "project",
	Domain:  "domain",
	Name:    "name",
}

var mockNodeExecutionRemoteURL = dataMocks.NewMockRemoteURL()

func addGetExecutionCallback(t *testing.T, repository interfaces.Repository) {
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
			assert.Equal(t, "project", input.Project)
			assert.Equal(t, "domain", input.Domain)
			assert.Equal(t, "name", input.Name)
			return models.Execution{
				BaseModel: models.BaseModel{
					ID: uint(8),
				},
				ExecutionKey: models.ExecutionKey{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
				Cluster: "propeller",
			}, nil
		})
}

func TestCreateNodeEvent(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	addGetExecutionCallback(t, repository)
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
			assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
				NodeId:      "node id",
				ExecutionId: &workflowExecutionIdentifier,
			}, &input.NodeExecutionIdentifier))
			return models.NodeExecution{}, flyteAdminErrors.NewFlyteAdminError(codes.NotFound, "foo")
		})
	expectedClosure := admin.NodeExecutionClosure{
		Phase:     request.Event.Phase,
		StartedAt: occurredAtProto,
		CreatedAt: occurredAtProto,
		UpdatedAt: occurredAtProto,
		TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
			TaskNodeMetadata: &admin.TaskNodeMetadata{},
		},
	}
	closureBytes, _ := proto.Marshal(&expectedClosure)
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetCreateCallback(
		func(ctx context.Context, input *models.NodeExecution) error {
			assert.Equal(t, models.NodeExecution{
				NodeExecutionKey: models.NodeExecutionKey{
					NodeID: "node id",
					ExecutionKey: models.ExecutionKey{
						Project: "project",
						Domain:  "domain",
						Name:    "name",
					},
				},
				Phase:                                 core.NodeExecution_RUNNING.String(),
				InputURI:                              "input uri",
				StartedAt:                             &occurredAt,
				Closure:                               closureBytes,
				NodeExecutionMetadata:                 []byte{},
				NodeExecutionCreatedAt:                &occurredAt,
				NodeExecutionUpdatedAt:                &occurredAt,
				DynamicWorkflowRemoteClosureReference: "s3://bucket/admin/metadata/project/domain/name/node id/proj_domain_dynamic_wf_abc123",
				InternalData:                          internalDataBytes,
			}, *input)
			return nil
		})

	mockDbEventWriter := &eventWriterMocks.NodeExecutionEventWriter{}
	mockDbEventWriter.On("Write", request)
	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(),
		[]string{"admin", "metadata"}, getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL,
		&mockPublisher, &mockPublisher, mockDbEventWriter)
	resp, err := nodeExecManager.CreateNodeEvent(context.Background(), request)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestCreateNodeEvent_Update(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	addGetExecutionCallback(t, repository)
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
			assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
				NodeId:      "node id",
				ExecutionId: &workflowExecutionIdentifier,
			}, &input.NodeExecutionIdentifier))
			return models.NodeExecution{
				NodeExecutionKey: models.NodeExecutionKey{
					NodeID: "node id",
					ExecutionKey: models.ExecutionKey{
						Project: "project",
						Domain:  "domain",
						Name:    "name",
					},
				},
				Phase:     core.NodeExecution_UNDEFINED.String(),
				InputURI:  "input uri",
				StartedAt: &occurredAt,
			}, nil
		})
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetUpdateCallback(
		func(ctx context.Context, nodeExecution *models.NodeExecution) error {
			expectedClosure := admin.NodeExecutionClosure{
				StartedAt: occurredAtProto,
				Phase:     core.NodeExecution_RUNNING,
				UpdatedAt: occurredAtProto,
				TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
					TaskNodeMetadata: &admin.TaskNodeMetadata{},
				},
			}
			expectedClosureBytes, _ := proto.Marshal(&expectedClosure)
			actualClosure := admin.NodeExecutionClosure{}
			_ = proto.Unmarshal(nodeExecution.Closure, &actualClosure)
			assert.Equal(t, models.NodeExecution{
				NodeExecutionKey: models.NodeExecutionKey{
					NodeID: "node id",
					ExecutionKey: models.ExecutionKey{
						Project: "project",
						Domain:  "domain",
						Name:    "name",
					},
				},
				Phase:                                 core.NodeExecution_RUNNING.String(),
				InputURI:                              "input uri",
				StartedAt:                             &occurredAt,
				Closure:                               expectedClosureBytes,
				NodeExecutionUpdatedAt:                &occurredAt,
				DynamicWorkflowRemoteClosureReference: "s3://bucket/admin/metadata/project/domain/name/node id/proj_domain_dynamic_wf_abc123",
			}, *nodeExecution)

			return nil
		})

	mockDbEventWriter := &eventWriterMocks.NodeExecutionEventWriter{}
	mockDbEventWriter.On("Write", request)
	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(),
		[]string{"admin", "metadata"}, getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL, &mockPublisher, &mockPublisher, mockDbEventWriter)
	resp, err := nodeExecManager.CreateNodeEvent(context.Background(), request)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestCreateNodeEvent_MissingExecution(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	expectedErr := flyteAdminErrors.NewFlyteAdminErrorf(codes.Internal, "expected error")
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
			return models.NodeExecution{}, expectedErr
		})
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
			return models.Execution{}, expectedErr
		})

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
		return models.Execution{}, expectedErr
	})
	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(), make([]string, 0), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL, &mockPublisher, &mockPublisher, &eventWriterMocks.NodeExecutionEventWriter{})
	resp, err := nodeExecManager.CreateNodeEvent(context.Background(), request)
	assert.EqualError(t, err, "Failed to get existing execution id: [project:\"project\" domain:\"domain\" name:\"name\" ] with err: expected error")
	assert.Nil(t, resp)
}

func TestCreateNodeEvent_CreateDatabaseError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	addGetExecutionCallback(t, repository)
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
			return models.NodeExecution{}, flyteAdminErrors.NewFlyteAdminError(codes.NotFound, "foo")
		})

	expectedErr := errors.New("expected error")
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetCreateCallback(
		func(ctx context.Context, input *models.NodeExecution) error {
			return expectedErr
		})
	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(), make([]string, 0), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL, nil, nil, &eventWriterMocks.NodeExecutionEventWriter{})
	resp, err := nodeExecManager.CreateNodeEvent(context.Background(), request)
	assert.EqualError(t, err, expectedErr.Error())
	assert.Nil(t, resp)
}

func TestCreateNodeEvent_UpdateDatabaseError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	addGetExecutionCallback(t, repository)
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
			assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
				NodeId:      "node id",
				ExecutionId: &workflowExecutionIdentifier,
			}, &input.NodeExecutionIdentifier))
			return models.NodeExecution{
				NodeExecutionKey: models.NodeExecutionKey{
					NodeID: "node id",
					ExecutionKey: models.ExecutionKey{
						Project: "project",
						Domain:  "domain",
						Name:    "name",
					},
				},
				Phase:     core.NodeExecution_UNDEFINED.String(),
				InputURI:  "input uri",
				StartedAt: &occurredAt,
			}, nil
		})

	expectedErr := errors.New("expected error")
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetUpdateCallback(
		func(ctx context.Context, nodeExecution *models.NodeExecution) error {
			return expectedErr
		})
	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(), make([]string, 0), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL, nil, nil, &eventWriterMocks.NodeExecutionEventWriter{})
	resp, err := nodeExecManager.CreateNodeEvent(context.Background(), request)
	assert.EqualError(t, err, expectedErr.Error())
	assert.Nil(t, resp)
}

func TestCreateNodeEvent_UpdateTerminalEventError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	addGetExecutionCallback(t, repository)
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
			assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
				NodeId:      "node id",
				ExecutionId: &workflowExecutionIdentifier,
			}, &input.NodeExecutionIdentifier))
			return models.NodeExecution{
				NodeExecutionKey: models.NodeExecutionKey{
					NodeID: "node id",
					ExecutionKey: models.ExecutionKey{
						Project: "project",
						Domain:  "domain",
						Name:    "name",
					},
				},
				Phase:     core.NodeExecution_SUCCEEDED.String(),
				InputURI:  "input uri",
				StartedAt: &occurredAt,
			}, nil
		})
	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(), make([]string, 0), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL, nil, nil, &eventWriterMocks.NodeExecutionEventWriter{})
	resp, err := nodeExecManager.CreateNodeEvent(context.Background(), request)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	adminError := err.(flyteAdminErrors.FlyteAdminError)
	assert.Equal(t, codes.FailedPrecondition, adminError.GRPCStatus().Code())
	details, ok := adminError.GRPCStatus().Details()[0].(*admin.EventFailureReason)
	assert.True(t, ok)
	_, ok = details.GetReason().(*admin.EventFailureReason_AlreadyInTerminalState)
	assert.True(t, ok)
}

func TestCreateNodeEvent_UpdateDuplicateEventError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	addGetExecutionCallback(t, repository)
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
			assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
				NodeId:      "node id",
				ExecutionId: &workflowExecutionIdentifier,
			}, &input.NodeExecutionIdentifier))
			return models.NodeExecution{
				NodeExecutionKey: models.NodeExecutionKey{
					NodeID: "node id",
					ExecutionKey: models.ExecutionKey{
						Project: "project",
						Domain:  "domain",
						Name:    "name",
					},
				},
				Phase:     core.NodeExecution_RUNNING.String(),
				InputURI:  "input uri",
				StartedAt: &occurredAt,
			}, nil
		})
	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(), make([]string, 0), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL, nil, nil, &eventWriterMocks.NodeExecutionEventWriter{})
	resp, err := nodeExecManager.CreateNodeEvent(context.Background(), request)
	assert.Equal(t, codes.AlreadyExists, err.(flyteAdminErrors.FlyteAdminError).Code())
	assert.Nil(t, resp)
}

func TestCreateNodeEvent_FirstEventIsTerminal(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	addGetExecutionCallback(t, repository)
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
			return models.NodeExecution{}, flyteAdminErrors.NewFlyteAdminError(codes.NotFound, "foo")
		})

	succeededRequest := admin.NodeExecutionEventRequest{
		RequestId: "request id",
		Event: &event.NodeExecutionEvent{
			ProducerId: "propeller",
			Id: &core.NodeExecutionIdentifier{
				NodeId: "node id",
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
			},
			OccurredAt: occurredAtProto,
			Phase:      core.NodeExecution_SUCCEEDED,
			InputValue: &event.NodeExecutionEvent_InputUri{
				InputUri: "input uri",
			},
		},
	}
	mockDbEventWriter := &eventWriterMocks.NodeExecutionEventWriter{}
	mockDbEventWriter.On("Write", succeededRequest)
	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(), make([]string, 0), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL, &mockPublisher, &mockPublisher, mockDbEventWriter)
	resp, err := nodeExecManager.CreateNodeEvent(context.Background(), succeededRequest)
	assert.NotNil(t, resp)
	assert.Nil(t, err)
}

func TestTransformNodeExecutionModel(t *testing.T) {
	ctx := context.TODO()
	repository := repositoryMocks.NewMockRepository()
	nodeExecID := &core.NodeExecutionIdentifier{
		NodeId:      "node id",
		ExecutionId: &workflowExecutionIdentifier,
	}
	t.Run("event version 0", func(t *testing.T) {
		repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetWithChildrenCallback(
			func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
				assert.True(t, proto.Equal(nodeExecID, &input.NodeExecutionIdentifier))
				return models.NodeExecution{
					NodeExecutionKey: models.NodeExecutionKey{
						NodeID: "node id",
						ExecutionKey: models.ExecutionKey{
							Project: "project",
							Domain:  "domain",
							Name:    "name",
						},
					},
					Phase:     core.NodeExecution_SUCCEEDED.String(),
					InputURI:  "input uri",
					StartedAt: &occurredAt,
					Closure:   closureBytes,
					ChildNodeExecutions: []models.NodeExecution{
						{
							Phase: "unknown",
						},
					},
				}, nil
			})

		manager := NodeExecutionManager{
			db: repository,
		}
		nodeExecution, err := manager.transformNodeExecutionModel(ctx, models.NodeExecution{}, nodeExecID, transformers.DefaultExecutionTransformerOptions)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(nodeExecID, nodeExecution.Id))
		assert.True(t, nodeExecution.Metadata.IsParentNode)
	})
	t.Run("event version > 0", func(t *testing.T) {
		manager := NodeExecutionManager{
			db: repository,
		}
		nodeExecutionMetadata := &admin.NodeExecutionMetaData{
			IsParentNode: true,
			IsDynamic:    true,
			SpecNodeId:   "spec",
		}
		nodeExecutionMetadataBytes, _ := proto.Marshal(nodeExecutionMetadata)
		nodeExecution, err := manager.transformNodeExecutionModel(ctx, models.NodeExecution{
			NodeExecutionKey: models.NodeExecutionKey{
				NodeID: "node id",
				ExecutionKey: models.ExecutionKey{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
			},
			Phase:                 core.NodeExecution_SUCCEEDED.String(),
			InputURI:              "input uri",
			StartedAt:             &occurredAt,
			Closure:               closureBytes,
			NodeExecutionMetadata: nodeExecutionMetadataBytes,
			InternalData:          internalDataBytes,
		}, nodeExecID, transformers.DefaultExecutionTransformerOptions)
		assert.NoError(t, err)
		assert.True(t, nodeExecution.Metadata.IsParentNode)
		assert.True(t, nodeExecution.Metadata.IsDynamic)
	})
	t.Run("transform internal data err", func(t *testing.T) {
		manager := NodeExecutionManager{
			db: repository,
		}
		_, err := manager.transformNodeExecutionModel(ctx, models.NodeExecution{
			InternalData: []byte("i'm invalid"),
		}, nodeExecID, transformers.DefaultExecutionTransformerOptions)
		assert.NotNil(t, err)
		assert.Equal(t, err.(flyteAdminErrors.FlyteAdminError).Code(), codes.Internal)
	})
	t.Run("get with children err", func(t *testing.T) {
		expectedErr := flyteAdminErrors.NewFlyteAdminError(codes.Internal, "foo")
		repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetWithChildrenCallback(
			func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
				assert.True(t, proto.Equal(nodeExecID, &input.NodeExecutionIdentifier))
				return models.NodeExecution{}, expectedErr
			})

		manager := NodeExecutionManager{
			db: repository,
		}
		_, err := manager.transformNodeExecutionModel(ctx, models.NodeExecution{}, nodeExecID, transformers.DefaultExecutionTransformerOptions)
		assert.Equal(t, err, expectedErr)
	})
}

func TestTransformNodeExecutionModelList(t *testing.T) {
	ctx := context.TODO()
	repository := repositoryMocks.NewMockRepository()
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetWithChildrenCallback(
		func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
			return models.NodeExecution{
				NodeExecutionKey: models.NodeExecutionKey{
					NodeID: "node id",
					ExecutionKey: models.ExecutionKey{
						Project: "project",
						Domain:  "domain",
						Name:    "name",
					},
				},
				Phase:     core.NodeExecution_SUCCEEDED.String(),
				InputURI:  "input uri",
				StartedAt: &occurredAt,
				Closure:   closureBytes,
				ChildNodeExecutions: []models.NodeExecution{
					{
						Phase: "unknown",
					},
				},
			}, nil
		})

	manager := NodeExecutionManager{
		db: repository,
	}
	nodeExecutions, err := manager.transformNodeExecutionModelList(ctx, []models.NodeExecution{
		{
			Phase: "unknown",
		},
	})
	assert.NoError(t, err)
	assert.Len(t, nodeExecutions, 1)

}

func TestGetNodeExecution(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	expectedClosure := admin.NodeExecutionClosure{
		Phase: core.NodeExecution_SUCCEEDED,
		TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
			TaskNodeMetadata: &admin.TaskNodeMetadata{
				CheckpointUri: "last checkpoint uri",
			},
		},
	}
	expectedMetadata := admin.NodeExecutionMetaData{
		SpecNodeId: "spec_node_id",
		RetryGroup: "retry_group",
	}
	metadataBytes, _ := proto.Marshal(&expectedMetadata)
	closureBytes, _ := proto.Marshal(&expectedClosure)
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
			workflowExecutionIdentifier := core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			}
			assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
				NodeId:      "node id",
				ExecutionId: &workflowExecutionIdentifier,
			}, &input.NodeExecutionIdentifier))
			return models.NodeExecution{
				NodeExecutionKey: models.NodeExecutionKey{
					NodeID: "node id",
					ExecutionKey: models.ExecutionKey{
						Project: "project",
						Domain:  "domain",
						Name:    "name",
					},
				},
				Phase:                 core.NodeExecution_SUCCEEDED.String(),
				InputURI:              "input uri",
				StartedAt:             &occurredAt,
				Closure:               closureBytes,
				NodeExecutionMetadata: metadataBytes,
				InternalData:          internalDataBytes,
			}, nil
		})
	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(), make([]string, 0), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL, nil, nil, &eventWriterMocks.NodeExecutionEventWriter{})
	nodeExecution, err := nodeExecManager.GetNodeExecution(context.Background(), admin.NodeExecutionGetRequest{
		Id: &nodeExecutionIdentifier,
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.NodeExecution{
		Id:       &nodeExecutionIdentifier,
		InputUri: "input uri",
		Closure:  &expectedClosure,
		Metadata: &expectedMetadata,
	}, nodeExecution))
}

func TestGetNodeExecutionParentNode(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	expectedClosure := admin.NodeExecutionClosure{
		Phase: core.NodeExecution_SUCCEEDED,
	}
	expectedMetadata := admin.NodeExecutionMetaData{
		SpecNodeId: "spec_node_id",
		RetryGroup: "retry_group",
	}
	metadataBytes, _ := proto.Marshal(&expectedMetadata)
	closureBytes, _ := proto.Marshal(&expectedClosure)
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetWithChildrenCallback(
		func(
			ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
			workflowExecutionIdentifier := core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			}
			assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
				NodeId:      "node id",
				ExecutionId: &workflowExecutionIdentifier,
			}, &input.NodeExecutionIdentifier))
			return models.NodeExecution{
				NodeExecutionKey: models.NodeExecutionKey{
					NodeID: "node id",
					ExecutionKey: models.ExecutionKey{
						Project: "project",
						Domain:  "domain",
						Name:    "name",
					},
				},
				Phase:                 core.NodeExecution_SUCCEEDED.String(),
				InputURI:              "input uri",
				StartedAt:             &occurredAt,
				Closure:               closureBytes,
				NodeExecutionMetadata: metadataBytes,
				ChildNodeExecutions: []models.NodeExecution{
					{
						NodeExecutionKey: models.NodeExecutionKey{
							NodeID: "node-child",
							ExecutionKey: models.ExecutionKey{
								Project: "project",
								Domain:  "domain",
								Name:    "name",
							},
						},
					},
				},
			}, nil
		})
	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(), make([]string, 0), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL, nil, nil, &eventWriterMocks.NodeExecutionEventWriter{})
	nodeExecution, err := nodeExecManager.GetNodeExecution(context.Background(), admin.NodeExecutionGetRequest{
		Id: &nodeExecutionIdentifier,
	})
	assert.Nil(t, err)
	expectedMetadata.IsParentNode = true
	assert.True(t, proto.Equal(&admin.NodeExecution{
		Id:       &nodeExecutionIdentifier,
		InputUri: "input uri",
		Closure:  &expectedClosure,
		Metadata: &expectedMetadata,
	}, nodeExecution))
}

func TestGetNodeExecutionEventVersion0(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	expectedClosure := admin.NodeExecutionClosure{
		Phase: core.NodeExecution_SUCCEEDED,
	}
	expectedMetadata := admin.NodeExecutionMetaData{
		SpecNodeId: "spec_node_id",
		RetryGroup: "retry_group",
	}
	metadataBytes, _ := proto.Marshal(&expectedMetadata)
	closureBytes, _ := proto.Marshal(&expectedClosure)
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetWithChildrenCallback(
		func(
			ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
			workflowExecutionIdentifier := core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			}
			assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
				NodeId:      "node id",
				ExecutionId: &workflowExecutionIdentifier,
			}, &input.NodeExecutionIdentifier))
			return models.NodeExecution{
				NodeExecutionKey: models.NodeExecutionKey{
					NodeID: "node id",
					ExecutionKey: models.ExecutionKey{
						Project: "project",
						Domain:  "domain",
						Name:    "name",
					},
				},
				Phase:                 core.NodeExecution_SUCCEEDED.String(),
				InputURI:              "input uri",
				StartedAt:             &occurredAt,
				Closure:               closureBytes,
				NodeExecutionMetadata: metadataBytes,
			}, nil
		})

	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(), make([]string, 0), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL, nil, nil, &eventWriterMocks.NodeExecutionEventWriter{})
	nodeExecution, err := nodeExecManager.GetNodeExecution(context.Background(), admin.NodeExecutionGetRequest{
		Id: &nodeExecutionIdentifier,
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.NodeExecution{
		Id:       &nodeExecutionIdentifier,
		InputUri: "input uri",
		Closure:  &expectedClosure,
		Metadata: &expectedMetadata,
	}, nodeExecution))
}

func TestGetNodeExecution_DatabaseError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	expectedErr := errors.New("expected error")
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
			return models.NodeExecution{}, expectedErr
		})
	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(), make([]string, 0), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL, nil, nil, &eventWriterMocks.NodeExecutionEventWriter{})
	nodeExecution, err := nodeExecManager.GetNodeExecution(context.Background(), admin.NodeExecutionGetRequest{
		Id: &nodeExecutionIdentifier,
	})
	assert.Nil(t, nodeExecution)
	assert.EqualError(t, err, expectedErr.Error())
}

func TestGetNodeExecution_TransformerError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
			return models.NodeExecution{
				NodeExecutionKey: models.NodeExecutionKey{
					NodeID: "node id",
					ExecutionKey: models.ExecutionKey{
						Project: "project",
						Domain:  "domain",
						Name:    "name",
					},
				},
				Phase:        core.NodeExecution_SUCCEEDED.String(),
				InputURI:     "input uri",
				StartedAt:    &occurredAt,
				Closure:      []byte("i'm invalid"),
				InternalData: internalDataBytes,
			}, nil
		})
	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(), make([]string, 0), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL, nil, nil, &eventWriterMocks.NodeExecutionEventWriter{})
	nodeExecution, err := nodeExecManager.GetNodeExecution(context.Background(), admin.NodeExecutionGetRequest{
		Id: &nodeExecutionIdentifier,
	})
	assert.Nil(t, nodeExecution)
	assert.Equal(t, err.(flyteAdminErrors.FlyteAdminError).Code(), codes.Internal)
}

func TestListNodeExecutionsLevelZero(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	expectedClosure := admin.NodeExecutionClosure{
		Phase: core.NodeExecution_SUCCEEDED,
	}
	expectedMetadata := admin.NodeExecutionMetaData{
		SpecNodeId: "spec_node_id",
		RetryGroup: "retry_group",
	}
	metadataBytes, _ := proto.Marshal(&expectedMetadata)
	closureBytes, _ := proto.Marshal(&expectedClosure)

	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetListCallback(
		func(ctx context.Context, input interfaces.ListResourceInput) (
			interfaces.NodeExecutionCollectionOutput, error) {
			assert.Equal(t, 1, input.Limit)
			assert.Equal(t, 2, input.Offset)
			assert.Len(t, input.InlineFilters, 3)
			assert.Equal(t, common.Execution, input.InlineFilters[0].GetEntity())
			queryExpr, _ := input.InlineFilters[0].GetGormQueryExpr()
			assert.Equal(t, "project", queryExpr.Args)
			assert.Equal(t, "execution_project = ?", queryExpr.Query)

			assert.Equal(t, common.Execution, input.InlineFilters[1].GetEntity())
			queryExpr, _ = input.InlineFilters[1].GetGormQueryExpr()
			assert.Equal(t, "domain", queryExpr.Args)
			assert.Equal(t, "execution_domain = ?", queryExpr.Query)

			assert.Equal(t, common.Execution, input.InlineFilters[2].GetEntity())
			queryExpr, _ = input.InlineFilters[2].GetGormQueryExpr()
			assert.Equal(t, "name", queryExpr.Args)
			assert.Equal(t, "execution_name = ?", queryExpr.Query)

			assert.Len(t, input.MapFilters, 1)
			filter := input.MapFilters[0].GetFilter()
			assert.Equal(t, map[string]interface{}{
				"parent_id":                nil,
				"parent_task_execution_id": nil,
			}, filter)

			assert.Equal(t, "execution_domain asc", input.SortParameter.GetGormOrderExpr())
			return interfaces.NodeExecutionCollectionOutput{
				NodeExecutions: []models.NodeExecution{
					{
						NodeExecutionKey: models.NodeExecutionKey{
							NodeID: "node id",
							ExecutionKey: models.ExecutionKey{
								Project: "project",
								Domain:  "domain",
								Name:    "name",
							},
						},
						Phase:                 core.NodeExecution_SUCCEEDED.String(),
						InputURI:              "input uri",
						StartedAt:             &occurredAt,
						Closure:               closureBytes,
						NodeExecutionMetadata: metadataBytes,
					},
				},
			}, nil
		})
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetWithChildrenCallback(
		func(
			ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
			return models.NodeExecution{
				NodeExecutionKey: models.NodeExecutionKey{
					NodeID: "node id",
					ExecutionKey: models.ExecutionKey{
						Project: "project",
						Domain:  "domain",
						Name:    "name",
					},
				},
				Phase:                 core.NodeExecution_SUCCEEDED.String(),
				InputURI:              "input uri",
				StartedAt:             &occurredAt,
				Closure:               closureBytes,
				NodeExecutionMetadata: metadataBytes,
			}, nil
		})
	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(), make([]string, 0), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL, nil, nil, &eventWriterMocks.NodeExecutionEventWriter{})
	nodeExecutions, err := nodeExecManager.ListNodeExecutions(context.Background(), admin.NodeExecutionListRequest{
		WorkflowExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Limit: 1,
		Token: "2",
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       "execution_domain",
		},
	})
	assert.NoError(t, err)
	assert.Len(t, nodeExecutions.NodeExecutions, 1)
	assert.True(t, proto.Equal(&admin.NodeExecution{
		Id: &core.NodeExecutionIdentifier{
			NodeId: "node id",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
		},
		InputUri: "input uri",
		Closure:  &expectedClosure,
		Metadata: &expectedMetadata,
	}, nodeExecutions.NodeExecutions[0]))
	assert.Equal(t, "3", nodeExecutions.Token)
}

func TestListNodeExecutionsWithParent(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	expectedClosure := admin.NodeExecutionClosure{
		Phase: core.NodeExecution_SUCCEEDED,
	}
	expectedMetadata := admin.NodeExecutionMetaData{
		SpecNodeId: "spec_node_id",
		RetryGroup: "retry_group",
	}
	metadataBytes, _ := proto.Marshal(&expectedMetadata)
	closureBytes, _ := proto.Marshal(&expectedClosure)
	parentID := uint(12)
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(func(ctx context.Context, input interfaces.NodeExecutionResource) (execution models.NodeExecution, e error) {
		assert.Equal(t, "parent_1", input.NodeExecutionIdentifier.NodeId)
		return models.NodeExecution{
			BaseModel: models.BaseModel{
				ID: parentID,
			},
		}, nil
	})
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetListCallback(
		func(ctx context.Context, input interfaces.ListResourceInput) (
			interfaces.NodeExecutionCollectionOutput, error) {
			assert.Equal(t, 1, input.Limit)
			assert.Equal(t, 2, input.Offset)
			assert.Len(t, input.InlineFilters, 4)
			assert.Equal(t, common.Execution, input.InlineFilters[0].GetEntity())
			queryExpr, _ := input.InlineFilters[0].GetGormQueryExpr()
			assert.Equal(t, "project", queryExpr.Args)
			assert.Equal(t, "execution_project = ?", queryExpr.Query)

			assert.Equal(t, common.Execution, input.InlineFilters[1].GetEntity())
			queryExpr, _ = input.InlineFilters[1].GetGormQueryExpr()
			assert.Equal(t, "domain", queryExpr.Args)
			assert.Equal(t, "execution_domain = ?", queryExpr.Query)

			assert.Equal(t, common.Execution, input.InlineFilters[2].GetEntity())
			queryExpr, _ = input.InlineFilters[2].GetGormQueryExpr()
			assert.Equal(t, "name", queryExpr.Args)
			assert.Equal(t, "execution_name = ?", queryExpr.Query)

			assert.Equal(t, common.NodeExecution, input.InlineFilters[3].GetEntity())
			queryExpr, _ = input.InlineFilters[3].GetGormQueryExpr()
			assert.Equal(t, parentID, queryExpr.Args)
			assert.Equal(t, "parent_id = ?", queryExpr.Query)

			assert.Equal(t, "execution_domain asc", input.SortParameter.GetGormOrderExpr())
			return interfaces.NodeExecutionCollectionOutput{
				NodeExecutions: []models.NodeExecution{
					{
						NodeExecutionKey: models.NodeExecutionKey{
							NodeID: "node id",
							ExecutionKey: models.ExecutionKey{
								Project: "project",
								Domain:  "domain",
								Name:    "name",
							},
						},
						Phase:                 core.NodeExecution_SUCCEEDED.String(),
						InputURI:              "input uri",
						StartedAt:             &occurredAt,
						Closure:               closureBytes,
						NodeExecutionMetadata: metadataBytes,
						InternalData:          internalDataBytes,
					},
				},
			}, nil
		})
	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(), make([]string, 0), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL, nil, nil, &eventWriterMocks.NodeExecutionEventWriter{})
	nodeExecutions, err := nodeExecManager.ListNodeExecutions(context.Background(), admin.NodeExecutionListRequest{
		WorkflowExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Limit: 1,
		Token: "2",
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       "execution_domain",
		},
		UniqueParentId: "parent_1",
	})
	assert.Nil(t, err)
	assert.Len(t, nodeExecutions.NodeExecutions, 1)
	assert.True(t, proto.Equal(&admin.NodeExecution{
		Id: &core.NodeExecutionIdentifier{
			NodeId: "node id",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
		},
		InputUri: "input uri",
		Closure:  &expectedClosure,
		Metadata: &expectedMetadata,
	}, nodeExecutions.NodeExecutions[0]))
	assert.Equal(t, "3", nodeExecutions.Token)
}

func TestListNodeExecutions_InvalidParams(t *testing.T) {
	nodeExecManager := NewNodeExecutionManager(nil, getMockExecutionsConfigProvider(), make([]string, 0), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL, nil, nil, &eventWriterMocks.NodeExecutionEventWriter{})
	_, err := nodeExecManager.ListNodeExecutions(context.Background(), admin.NodeExecutionListRequest{
		Filters: "eq(execution.project, project)",
	})
	assert.Equal(t, codes.InvalidArgument, err.(flyteAdminErrors.FlyteAdminError).Code())

	_, err = nodeExecManager.ListNodeExecutions(context.Background(), admin.NodeExecutionListRequest{
		Limit: 1,
	})
	assert.Equal(t, codes.InvalidArgument, err.(flyteAdminErrors.FlyteAdminError).Code())

	_, err = nodeExecManager.ListNodeExecutions(context.Background(), admin.NodeExecutionListRequest{
		Limit:   1,
		Filters: "foo",
	})
	assert.Equal(t, codes.InvalidArgument, err.(flyteAdminErrors.FlyteAdminError).Code())
}

func TestListNodeExecutions_DatabaseError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	expectedErr := errors.New("expected error")
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetListCallback(
		func(ctx context.Context, input interfaces.ListResourceInput) (
			interfaces.NodeExecutionCollectionOutput, error) {
			return interfaces.NodeExecutionCollectionOutput{}, expectedErr
		})
	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(), make([]string, 0), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL, nil, nil, &eventWriterMocks.NodeExecutionEventWriter{})
	nodeExecutions, err := nodeExecManager.ListNodeExecutions(context.Background(), admin.NodeExecutionListRequest{
		WorkflowExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Limit: 1,
		Token: "2",
	})
	assert.Nil(t, nodeExecutions)
	assert.EqualError(t, err, expectedErr.Error())
}

func TestListNodeExecutions_TransformerError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetListCallback(
		func(ctx context.Context, input interfaces.ListResourceInput) (
			interfaces.NodeExecutionCollectionOutput, error) {
			return interfaces.NodeExecutionCollectionOutput{
				NodeExecutions: []models.NodeExecution{
					{
						NodeExecutionKey: models.NodeExecutionKey{
							NodeID: "node id",
							ExecutionKey: models.ExecutionKey{
								Project: "project",
								Domain:  "domain",
								Name:    "name",
							},
						},
						Phase:        core.NodeExecution_SUCCEEDED.String(),
						InputURI:     "input uri",
						StartedAt:    &occurredAt,
						Closure:      []byte("i'm invalid"),
						InternalData: internalDataBytes,
					},
				},
			}, nil
		})
	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(), make([]string, 0), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL, nil, nil, &eventWriterMocks.NodeExecutionEventWriter{})
	nodeExecutions, err := nodeExecManager.ListNodeExecutions(context.Background(), admin.NodeExecutionListRequest{
		WorkflowExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Limit: 1,
		Token: "2",
	})
	assert.Nil(t, nodeExecutions)
	assert.Equal(t, err.(flyteAdminErrors.FlyteAdminError).Code(), codes.Internal)
}

func TestListNodeExecutions_NothingToReturn(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetListCallback(
		func(ctx context.Context, input interfaces.ListResourceInput) (
			interfaces.NodeExecutionCollectionOutput, error) {
			return interfaces.NodeExecutionCollectionOutput{}, nil
		})
	var listExecutionsCalled bool
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetListCallback(
		func(ctx context.Context, input interfaces.ListResourceInput) (
			interfaces.ExecutionCollectionOutput, error) {
			listExecutionsCalled = true
			return interfaces.ExecutionCollectionOutput{}, nil
		})
	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(), make([]string, 0), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL, nil, nil, &eventWriterMocks.NodeExecutionEventWriter{})

	_, err := nodeExecManager.ListNodeExecutions(context.Background(), admin.NodeExecutionListRequest{
		WorkflowExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Limit: 1,
		Token: "2",
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       "execution_domain",
		},
	})

	assert.NoError(t, err)
	assert.False(t, listExecutionsCalled)
}

func TestListNodeExecutionsForTask(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	expectedClosure := admin.NodeExecutionClosure{
		Phase: core.NodeExecution_SUCCEEDED,
	}
	execMetadata := admin.NodeExecutionMetaData{
		SpecNodeId:   "spec-n1",
		IsParentNode: true,
	}

	closureBytes, _ := proto.Marshal(&expectedClosure)
	execMetadataBytes, _ := proto.Marshal(&execMetadata)

	repository.TaskExecutionRepo().(*repositoryMocks.MockTaskExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.GetTaskExecutionInput) (models.TaskExecution, error) {
			return models.TaskExecution{
				BaseModel: models.BaseModel{
					ID: uint(8),
				},
			}, nil
		})
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetListCallback(
		func(ctx context.Context, input interfaces.ListResourceInput) (
			interfaces.NodeExecutionCollectionOutput, error) {
			assert.Equal(t, 1, input.Limit)
			assert.Equal(t, 2, input.Offset)
			assert.Len(t, input.InlineFilters, 4)
			assert.Equal(t, common.Execution, input.InlineFilters[0].GetEntity())
			queryExpr, _ := input.InlineFilters[0].GetGormQueryExpr()
			assert.Equal(t, "project", queryExpr.Args)
			assert.Equal(t, "execution_project = ?", queryExpr.Query)

			assert.Equal(t, common.Execution, input.InlineFilters[1].GetEntity())
			queryExpr, _ = input.InlineFilters[1].GetGormQueryExpr()
			assert.Equal(t, "domain", queryExpr.Args)
			assert.Equal(t, "execution_domain = ?", queryExpr.Query)

			assert.Equal(t, common.Execution, input.InlineFilters[2].GetEntity())
			queryExpr, _ = input.InlineFilters[2].GetGormQueryExpr()
			assert.Equal(t, "name", queryExpr.Args)
			assert.Equal(t, "execution_name = ?", queryExpr.Query)

			assert.Equal(t, common.NodeExecution, input.InlineFilters[3].GetEntity())
			queryExpr, _ = input.InlineFilters[3].GetGormQueryExpr()
			assert.Equal(t, uint(8), queryExpr.Args)
			assert.Equal(t, "parent_task_execution_id = ?", queryExpr.Query)

			assert.Equal(t, "execution_domain asc", input.SortParameter.GetGormOrderExpr())
			return interfaces.NodeExecutionCollectionOutput{
				NodeExecutions: []models.NodeExecution{
					{
						NodeExecutionKey: models.NodeExecutionKey{
							NodeID: "node id",
							ExecutionKey: models.ExecutionKey{
								Project: "project",
								Domain:  "domain",
								Name:    "name",
							},
						},
						Phase:                 core.NodeExecution_SUCCEEDED.String(),
						InputURI:              "input uri",
						StartedAt:             &occurredAt,
						Closure:               closureBytes,
						NodeExecutionMetadata: execMetadataBytes,
						InternalData:          internalDataBytes,
					},
				},
			}, nil
		})
	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(), make([]string, 0), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockNodeExecutionRemoteURL, nil, nil, &eventWriterMocks.NodeExecutionEventWriter{})
	nodeExecutions, err := nodeExecManager.ListNodeExecutionsForTask(context.Background(), admin.NodeExecutionForTaskListRequest{
		TaskExecutionId: &core.TaskExecutionIdentifier{
			NodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
				NodeId: "node_id",
			},
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "project",
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
		},
		Limit: 1,
		Token: "2",
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       "execution_domain",
		},
	})
	assert.Nil(t, err)
	assert.Len(t, nodeExecutions.NodeExecutions, 1)
	expectedMetadata := admin.NodeExecutionMetaData{
		SpecNodeId:   "spec-n1",
		IsParentNode: true,
	}
	assert.True(t, proto.Equal(&admin.NodeExecution{
		Id: &core.NodeExecutionIdentifier{
			NodeId: "node id",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
		},
		InputUri: "input uri",
		Closure:  &expectedClosure,
		Metadata: &expectedMetadata,
	}, nodeExecutions.NodeExecutions[0]))
	assert.Equal(t, "3", nodeExecutions.Token)
}

func TestGetNodeExecutionData(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	expectedClosure := admin.NodeExecutionClosure{
		Phase: core.NodeExecution_SUCCEEDED,
		OutputResult: &admin.NodeExecutionClosure_OutputUri{
			OutputUri: util.OutputsFile,
		},
		DeckUri: util.DeckFile,
	}
	dynamicWorkflowClosureRef := "s3://my-s3-bucket/foo/bar/dynamic.pb"

	closureBytes, _ := proto.Marshal(&expectedClosure)
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
			workflowExecutionIdentifier := core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			}
			assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
				NodeId:      "node id",
				ExecutionId: &workflowExecutionIdentifier,
			}, &input.NodeExecutionIdentifier))
			return models.NodeExecution{
				NodeExecutionKey: models.NodeExecutionKey{
					NodeID: "node id",
					ExecutionKey: models.ExecutionKey{
						Project: "project",
						Domain:  "domain",
						Name:    "name",
					},
				},
				Phase:                                 core.NodeExecution_SUCCEEDED.String(),
				InputURI:                              "input uri",
				StartedAt:                             &occurredAt,
				Closure:                               closureBytes,
				DynamicWorkflowRemoteClosureReference: dynamicWorkflowClosureRef,
			}, nil
		})

	mockNodeExecutionRemoteURL := dataMocks.NewMockRemoteURL()
	mockNodeExecutionRemoteURL.(*dataMocks.MockRemoteURL).GetCallback = func(ctx context.Context, uri string) (admin.UrlBlob, error) {
		if uri == "input uri" {
			return admin.UrlBlob{
				Url:   "inputs",
				Bytes: 100,
			}, nil
		} else if uri == util.OutputsFile {
			return admin.UrlBlob{
				Url:   "outputs",
				Bytes: 200,
			}, nil
		}

		return admin.UrlBlob{}, errors.New("unexpected input")
	}
	mockStorage := commonMocks.GetMockStorageClient()
	fullInputs := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo": testutils.MakeStringLiteral("foo-value-1"),
		},
	}
	fullOutputs := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"bar": testutils.MakeStringLiteral("bar-value-1"),
		},
	}

	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).ReadProtobufCb = func(
		ctx context.Context, reference storage.DataReference, msg proto.Message) error {
		if reference.String() == "input uri" {
			marshalled, _ := proto.Marshal(fullInputs)
			_ = proto.Unmarshal(marshalled, msg)
			return nil
		} else if reference.String() == util.OutputsFile {
			marshalled, _ := proto.Marshal(fullOutputs)
			_ = proto.Unmarshal(marshalled, msg)
			return nil
		} else if reference.String() == dynamicWorkflowClosureRef {
			marshalled, _ := proto.Marshal(&dynamicWorkflowClosure)
			_ = proto.Unmarshal(marshalled, msg)
			return nil
		}
		return fmt.Errorf("unexpected call to find value in storage [%v]", reference.String())
	}
	nodeExecManager := NewNodeExecutionManager(repository, getMockExecutionsConfigProvider(), make([]string, 0), mockStorage, mockScope.NewTestScope(), mockNodeExecutionRemoteURL, nil, nil, &eventWriterMocks.NodeExecutionEventWriter{})
	dataResponse, err := nodeExecManager.GetNodeExecutionData(context.Background(), admin.NodeExecutionGetDataRequest{
		Id: &nodeExecutionIdentifier,
	})
	assert.NoError(t, err)
	assert.True(t, proto.Equal(&admin.NodeExecutionGetDataResponse{
		Inputs: &admin.UrlBlob{
			Url:   "inputs",
			Bytes: 100,
		},
		Outputs: &admin.UrlBlob{
			Url:   "outputs",
			Bytes: 200,
		},
		FullInputs:  fullInputs,
		FullOutputs: fullOutputs,
		DynamicWorkflow: &admin.DynamicWorkflowNodeMetadata{
			Id:               dynamicWorkflowClosure.Primary.Template.Id,
			CompiledWorkflow: &dynamicWorkflowClosure,
		},
		FlyteUrls: &admin.FlyteURLs{
			Inputs:  "flyte://v1/project/domain/name/node id/i",
			Outputs: "flyte://v1/project/domain/name/node id/o",
			Deck:    "flyte://v1/project/domain/name/node id/d",
		},
	}, dataResponse))
}
