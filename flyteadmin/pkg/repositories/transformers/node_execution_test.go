package transformers

import (
	"context"
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flytestdlib/promutils"

	flyteAdminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"google.golang.org/grpc/codes"

	commonMocks "github.com/flyteorg/flyteadmin/pkg/common/mocks"
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/golang/protobuf/ptypes"

	genModel "github.com/flyteorg/flyteadmin/pkg/repositories/gen/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

var occurredAt = time.Now().UTC()
var occurredAtProto, _ = ptypes.TimestampProto(occurredAt)
var duration = time.Hour
var closure = &admin.NodeExecutionClosure{
	Phase: core.NodeExecution_SUCCEEDED,
	OutputResult: &admin.NodeExecutionClosure_OutputUri{
		OutputUri: "output uri",
	},
	StartedAt: occurredAtProto,
	Duration:  ptypes.DurationProto(duration),
}
var closureBytes, _ = proto.Marshal(closure)
var nodeExecutionMetadata = admin.NodeExecutionMetaData{
	IsParentNode: false,
	RetryGroup:   "r",
	SpecNodeId:   "sp",
}
var nodeExecutionMetadataBytes, _ = proto.Marshal(&nodeExecutionMetadata)

var childExecutionID = &core.WorkflowExecutionIdentifier{
	Project: "p",
	Domain:  "d",
	Name:    "name",
}

const dynamicWorkflowClosureRef = "s3://bucket/admin/metadata/workflow"

const testInputURI = "fake://bucket/inputs.pb"

var testInputs = &core.LiteralMap{
	Literals: map[string]*core.Literal{
		"foo": coreutils.MustMakeLiteral("bar"),
	},
}

func TestAddRunningState(t *testing.T) {
	var startedAt = time.Now().UTC()
	var startedAtProto, _ = ptypes.TimestampProto(startedAt)
	request := admin.NodeExecutionEventRequest{
		Event: &event.NodeExecutionEvent{
			Phase:      core.NodeExecution_RUNNING,
			OccurredAt: startedAtProto,
		},
	}
	nodeExecutionModel := models.NodeExecution{}
	closure := admin.NodeExecutionClosure{}
	err := addNodeRunningState(&request, &nodeExecutionModel, &closure)
	assert.Nil(t, err)
	assert.Equal(t, startedAt, *nodeExecutionModel.StartedAt)
	assert.True(t, proto.Equal(startedAtProto, closure.StartedAt))
}

func TestAddTerminalState_OutputURI(t *testing.T) {
	outputURI := "output uri"
	request := admin.NodeExecutionEventRequest{
		Event: &event.NodeExecutionEvent{
			Phase: core.NodeExecution_SUCCEEDED,
			OutputResult: &event.NodeExecutionEvent_OutputUri{
				OutputUri: outputURI,
			},
			OccurredAt: occurredAtProto,
		},
	}
	startedAt := occurredAt.Add(-time.Minute)
	startedAtProto, _ := ptypes.TimestampProto(startedAt)
	nodeExecutionModel := models.NodeExecution{
		StartedAt: &startedAt,
	}
	closure := admin.NodeExecutionClosure{
		StartedAt: startedAtProto,
	}
	err := addTerminalState(context.TODO(), &request, &nodeExecutionModel, &closure,
		interfaces.InlineEventDataPolicyStoreInline, commonMocks.GetMockStorageClient())
	assert.Nil(t, err)
	assert.EqualValues(t, outputURI, closure.GetOutputUri())
	assert.Equal(t, time.Minute, nodeExecutionModel.Duration)
}

func TestAddTerminalState_OutputData(t *testing.T) {
	outputData := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo": {
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{
							Primitive: &core.Primitive{
								Value: &core.Primitive_Integer{
									Integer: 4,
								},
							},
						},
					},
				},
			},
		},
	}
	request := admin.NodeExecutionEventRequest{
		Event: &event.NodeExecutionEvent{
			Id: &core.NodeExecutionIdentifier{
				NodeId: "node id",
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
			},
			Phase: core.NodeExecution_SUCCEEDED,
			OutputResult: &event.NodeExecutionEvent_OutputData{
				OutputData: outputData,
			},
			OccurredAt: occurredAtProto,
		},
	}
	startedAt := occurredAt.Add(-time.Minute)
	startedAtProto, _ := ptypes.TimestampProto(startedAt)
	nodeExecutionModel := models.NodeExecution{
		StartedAt: &startedAt,
	}
	closure := admin.NodeExecutionClosure{
		StartedAt: startedAtProto,
	}
	t.Run("output data stored inline", func(t *testing.T) {
		err := addTerminalState(context.TODO(), &request, &nodeExecutionModel, &closure,
			interfaces.InlineEventDataPolicyStoreInline, commonMocks.GetMockStorageClient())
		assert.Nil(t, err)
		assert.EqualValues(t, outputData, closure.GetOutputData())
		assert.Equal(t, time.Minute, nodeExecutionModel.Duration)
	})
	t.Run("output data stored offloaded", func(t *testing.T) {
		mockStorage := commonMocks.GetMockStorageClient()
		mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).WriteProtobufCb = func(ctx context.Context, reference storage.DataReference, opts storage.Options, msg proto.Message) error {
			assert.Equal(t, reference.String(), "s3://bucket/metadata/project/domain/name/node id/offloaded_outputs")
			return nil
		}

		err := addTerminalState(context.TODO(), &request, &nodeExecutionModel, &closure,
			interfaces.InlineEventDataPolicyOffload, mockStorage)
		assert.Nil(t, err)
		assert.Equal(t, "s3://bucket/metadata/project/domain/name/node id/offloaded_outputs", closure.GetOutputUri())
	})
}

func TestAddTerminalState_Error(t *testing.T) {
	error := &core.ExecutionError{
		Code: "foo",
	}
	request := admin.NodeExecutionEventRequest{
		Event: &event.NodeExecutionEvent{
			Phase: core.NodeExecution_FAILED,
			OutputResult: &event.NodeExecutionEvent_Error{
				Error: error,
			},
			OccurredAt: occurredAtProto,
		},
	}
	startedAt := occurredAt.Add(-time.Minute)
	startedAtProto, _ := ptypes.TimestampProto(startedAt)
	nodeExecutionModel := models.NodeExecution{
		StartedAt: &startedAt,
	}
	closure := admin.NodeExecutionClosure{
		StartedAt: startedAtProto,
	}
	err := addTerminalState(context.TODO(), &request, &nodeExecutionModel, &closure,
		interfaces.InlineEventDataPolicyStoreInline, commonMocks.GetMockStorageClient())
	assert.Nil(t, err)
	assert.True(t, proto.Equal(error, closure.GetError()))
	assert.Equal(t, time.Minute, nodeExecutionModel.Duration)
}

func TestCreateNodeExecutionModel(t *testing.T) {
	parentTaskExecID := uint(8)
	request := &admin.NodeExecutionEventRequest{
		Event: &event.NodeExecutionEvent{
			Id: &core.NodeExecutionIdentifier{
				NodeId: "node id",
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
			},
			Phase: core.NodeExecution_RUNNING,
			InputValue: &event.NodeExecutionEvent_InputUri{
				InputUri: testInputURI,
			},
			OutputResult: &event.NodeExecutionEvent_OutputUri{
				OutputUri: "output uri",
			},
			OccurredAt: occurredAtProto,
			TargetMetadata: &event.NodeExecutionEvent_TaskNodeMetadata{
				TaskNodeMetadata: &event.TaskNodeMetadata{
					CacheStatus: core.CatalogCacheStatus_CACHE_POPULATED,
					CatalogKey: &core.CatalogMetadata{
						DatasetId: &core.Identifier{
							ResourceType: core.ResourceType_DATASET,
							Name:         "x",
							Project:      "proj",
							Domain:       "domain",
						},
					},
					CheckpointUri: "last checkpoint uri",
				},
			},
			ParentTaskMetadata: &event.ParentTaskExecutionMetadata{
				Id: &core.TaskExecutionIdentifier{
					RetryAttempt: 1,
				},
			},
			IsParent:     true,
			IsDynamic:    true,
			EventVersion: 2,
		},
	}

	nodeExecutionModel, err := CreateNodeExecutionModel(context.TODO(), ToNodeExecutionModelInput{
		Request:               request,
		ParentTaskExecutionID: &parentTaskExecID,
	})
	assert.Nil(t, err)

	var closure = &admin.NodeExecutionClosure{
		Phase:     core.NodeExecution_RUNNING,
		StartedAt: occurredAtProto,
		CreatedAt: occurredAtProto,
		UpdatedAt: occurredAtProto,
		TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
			TaskNodeMetadata: &admin.TaskNodeMetadata{
				CacheStatus:   request.Event.GetTaskNodeMetadata().CacheStatus,
				CatalogKey:    request.Event.GetTaskNodeMetadata().CatalogKey,
				CheckpointUri: request.Event.GetTaskNodeMetadata().CheckpointUri,
			},
		},
	}
	var closureBytes, _ = proto.Marshal(closure)
	var nodeExecutionMetadata, _ = proto.Marshal(&admin.NodeExecutionMetaData{
		IsParentNode: true,
		IsDynamic:    true,
	})
	internalData := &genModel.NodeExecutionInternalData{
		EventVersion: 2,
	}
	internalDataBytes, _ := proto.Marshal(internalData)
	cacheStatus := request.Event.GetTaskNodeMetadata().CacheStatus.String()
	assert.Equal(t, &models.NodeExecution{
		NodeExecutionKey: models.NodeExecutionKey{
			NodeID: "node id",
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
		},
		Phase:                  "RUNNING",
		Closure:                closureBytes,
		InputURI:               testInputURI,
		StartedAt:              &occurredAt,
		NodeExecutionCreatedAt: &occurredAt,
		NodeExecutionUpdatedAt: &occurredAt,
		NodeExecutionMetadata:  nodeExecutionMetadata,
		ParentTaskExecutionID:  &parentTaskExecID,
		CacheStatus:            &cacheStatus,
		InternalData:           internalDataBytes,
	}, nodeExecutionModel)
}

func TestUpdateNodeExecutionModel(t *testing.T) {
	t.Run("child-workflow", func(t *testing.T) {
		request := admin.NodeExecutionEventRequest{
			Event: &event.NodeExecutionEvent{
				Phase:      core.NodeExecution_RUNNING,
				OccurredAt: occurredAtProto,
				TargetMetadata: &event.NodeExecutionEvent_WorkflowNodeMetadata{
					WorkflowNodeMetadata: &event.WorkflowNodeMetadata{
						ExecutionId: childExecutionID,
					},
				},
				InputValue: &event.NodeExecutionEvent_InputUri{
					InputUri: testInputURI,
				},
			},
		}
		nodeExecutionModel := models.NodeExecution{
			Phase: core.NodeExecution_UNDEFINED.String(),
		}
		mockStore := commonMocks.GetMockStorageClient()
		err := UpdateNodeExecutionModel(context.TODO(), &request, &nodeExecutionModel, childExecutionID, dynamicWorkflowClosureRef,
			interfaces.InlineEventDataPolicyStoreInline, mockStore)
		assert.Nil(t, err)
		assert.Equal(t, core.NodeExecution_RUNNING.String(), nodeExecutionModel.Phase)
		assert.Equal(t, occurredAt, *nodeExecutionModel.StartedAt)
		assert.EqualValues(t, occurredAt, *nodeExecutionModel.NodeExecutionUpdatedAt)
		assert.Nil(t, nodeExecutionModel.CacheStatus)
		assert.Equal(t, nodeExecutionModel.DynamicWorkflowRemoteClosureReference, dynamicWorkflowClosureRef)
		assert.Equal(t, nodeExecutionModel.InputURI, testInputURI)

		var closure = &admin.NodeExecutionClosure{
			Phase:     core.NodeExecution_RUNNING,
			StartedAt: occurredAtProto,
			UpdatedAt: occurredAtProto,
			TargetMetadata: &admin.NodeExecutionClosure_WorkflowNodeMetadata{
				WorkflowNodeMetadata: &admin.WorkflowNodeMetadata{
					ExecutionId: childExecutionID,
				},
			},
		}
		var closureBytes, _ = proto.Marshal(closure)
		assert.Equal(t, nodeExecutionModel.Closure, closureBytes)
	})

	t.Run("task-node-metadata", func(t *testing.T) {
		request := admin.NodeExecutionEventRequest{
			Event: &event.NodeExecutionEvent{
				Phase:      core.NodeExecution_RUNNING,
				OccurredAt: occurredAtProto,
				TargetMetadata: &event.NodeExecutionEvent_TaskNodeMetadata{
					TaskNodeMetadata: &event.TaskNodeMetadata{
						CacheStatus: core.CatalogCacheStatus_CACHE_POPULATED,
						CatalogKey: &core.CatalogMetadata{
							DatasetId: &core.Identifier{
								ResourceType: core.ResourceType_DATASET,
								Name:         "x",
								Project:      "proj",
								Domain:       "domain",
							},
						},
						DynamicWorkflow: &event.DynamicWorkflowNodeMetadata{
							Id: &core.Identifier{
								ResourceType: core.ResourceType_WORKFLOW,
								Name:         "n",
								Project:      "proj",
								Domain:       "domain",
								Version:      "v",
							},
							CompiledWorkflow: &core.CompiledWorkflowClosure{
								Primary: &core.CompiledWorkflow{
									Template: &core.WorkflowTemplate{
										Metadata: &core.WorkflowMetadata{
											OnFailure: core.WorkflowMetadata_FAIL_AFTER_EXECUTABLE_NODES_COMPLETE,
										},
									},
								},
							},
							DynamicJobSpecUri: "/foo/bar",
						},
						CheckpointUri: "last checkpoint uri",
					},
				},
			},
		}
		nodeExecutionModel := models.NodeExecution{
			Phase: core.NodeExecution_UNDEFINED.String(),
		}
		err := UpdateNodeExecutionModel(context.TODO(), &request, &nodeExecutionModel, childExecutionID, dynamicWorkflowClosureRef,
			interfaces.InlineEventDataPolicyStoreInline, commonMocks.GetMockStorageClient())
		assert.Nil(t, err)
		assert.Equal(t, core.NodeExecution_RUNNING.String(), nodeExecutionModel.Phase)
		assert.Equal(t, occurredAt, *nodeExecutionModel.StartedAt)
		assert.EqualValues(t, occurredAt, *nodeExecutionModel.NodeExecutionUpdatedAt)
		assert.NotNil(t, nodeExecutionModel.CacheStatus)
		assert.Equal(t, *nodeExecutionModel.CacheStatus, request.Event.GetTaskNodeMetadata().CacheStatus.String())
		assert.Equal(t, nodeExecutionModel.DynamicWorkflowRemoteClosureReference, dynamicWorkflowClosureRef)

		var closure = &admin.NodeExecutionClosure{
			Phase:     core.NodeExecution_RUNNING,
			StartedAt: occurredAtProto,
			UpdatedAt: occurredAtProto,
			TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
				TaskNodeMetadata: &admin.TaskNodeMetadata{
					CacheStatus:   request.Event.GetTaskNodeMetadata().CacheStatus,
					CatalogKey:    request.Event.GetTaskNodeMetadata().CatalogKey,
					CheckpointUri: request.Event.GetTaskNodeMetadata().CheckpointUri,
				},
			},
			DynamicJobSpecUri: request.Event.GetTaskNodeMetadata().DynamicWorkflow.DynamicJobSpecUri,
		}
		var closureBytes, _ = proto.Marshal(closure)
		assert.Equal(t, nodeExecutionModel.Closure, closureBytes)
	})
	t.Run("is parent & is dynamic", func(t *testing.T) {
		request := admin.NodeExecutionEventRequest{
			Event: &event.NodeExecutionEvent{
				Phase:      core.NodeExecution_RUNNING,
				OccurredAt: occurredAtProto,
				TargetMetadata: &event.NodeExecutionEvent_WorkflowNodeMetadata{
					WorkflowNodeMetadata: &event.WorkflowNodeMetadata{
						ExecutionId: childExecutionID,
					},
				},
				IsParent:  true,
				IsDynamic: true,
			},
		}
		nodeExecMetadata := admin.NodeExecutionMetaData{
			SpecNodeId: "foo",
		}
		nodeExecMetadataSerialized, _ := proto.Marshal(&nodeExecMetadata)
		nodeExecutionModel := models.NodeExecution{
			Phase:                 core.NodeExecution_UNDEFINED.String(),
			NodeExecutionMetadata: nodeExecMetadataSerialized,
		}
		err := UpdateNodeExecutionModel(context.TODO(), &request, &nodeExecutionModel, childExecutionID, dynamicWorkflowClosureRef,
			interfaces.InlineEventDataPolicyStoreInline, commonMocks.GetMockStorageClient())
		assert.Nil(t, err)
		assert.Equal(t, core.NodeExecution_RUNNING.String(), nodeExecutionModel.Phase)
		assert.Equal(t, occurredAt, *nodeExecutionModel.StartedAt)
		assert.EqualValues(t, occurredAt, *nodeExecutionModel.NodeExecutionUpdatedAt)
		assert.Nil(t, nodeExecutionModel.CacheStatus)
		assert.Equal(t, nodeExecutionModel.DynamicWorkflowRemoteClosureReference, dynamicWorkflowClosureRef)

		nodeExecMetadata.IsParentNode = true
		nodeExecMetadata.IsDynamic = true
		nodeExecMetadataExpected, _ := proto.Marshal(&nodeExecMetadata)
		assert.Equal(t, nodeExecutionModel.NodeExecutionMetadata, nodeExecMetadataExpected)
	})
	t.Run("inline input data", func(t *testing.T) {
		request := admin.NodeExecutionEventRequest{
			Event: &event.NodeExecutionEvent{
				Id:         sampleNodeExecID,
				Phase:      core.NodeExecution_RUNNING,
				OccurredAt: occurredAtProto,
				InputValue: &event.NodeExecutionEvent_InputData{
					InputData: testInputs,
				},
			},
		}
		nodeExecMetadata := admin.NodeExecutionMetaData{
			SpecNodeId: "foo",
		}
		nodeExecMetadataSerialized, _ := proto.Marshal(&nodeExecMetadata)
		nodeExecutionModel := models.NodeExecution{
			Phase:                 core.NodeExecution_UNDEFINED.String(),
			NodeExecutionMetadata: nodeExecMetadataSerialized,
		}
		ds, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)
		err = UpdateNodeExecutionModel(context.TODO(), &request, &nodeExecutionModel, childExecutionID, dynamicWorkflowClosureRef,
			interfaces.InlineEventDataPolicyStoreInline, ds)
		assert.Nil(t, err)
		assert.Equal(t, nodeExecutionModel.InputURI, "/metadata/project/domain/name/node-id/offloaded_inputs")
	})
	t.Run("input data URI", func(t *testing.T) {
		request := admin.NodeExecutionEventRequest{
			Event: &event.NodeExecutionEvent{
				Id:         sampleNodeExecID,
				Phase:      core.NodeExecution_RUNNING,
				OccurredAt: occurredAtProto,
				InputValue: &event.NodeExecutionEvent_InputUri{
					InputUri: testInputURI,
				},
			},
		}
		nodeExecMetadata := admin.NodeExecutionMetaData{
			SpecNodeId: "foo",
		}
		nodeExecMetadataSerialized, _ := proto.Marshal(&nodeExecMetadata)
		nodeExecutionModel := models.NodeExecution{
			Phase:                 core.NodeExecution_UNDEFINED.String(),
			NodeExecutionMetadata: nodeExecMetadataSerialized,
		}
		err := UpdateNodeExecutionModel(context.TODO(), &request, &nodeExecutionModel, childExecutionID, dynamicWorkflowClosureRef,
			interfaces.InlineEventDataPolicyStoreInline, commonMocks.GetMockStorageClient())
		assert.Nil(t, err)
		assert.Equal(t, nodeExecutionModel.InputURI, testInputURI)
	})
}

func TestFromNodeExecutionModel(t *testing.T) {
	nodeExecutionIdentifier := core.NodeExecutionIdentifier{
		NodeId: "nodey",
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
	}
	nodeExecution, err := FromNodeExecutionModel(models.NodeExecution{
		NodeExecutionKey: models.NodeExecutionKey{
			NodeID: "nodey",
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
		},
		Phase:                 "NodeExecutionPhase_NODE_PHASE_RUNNING",
		Closure:               closureBytes,
		NodeExecutionMetadata: nodeExecutionMetadataBytes,
		InputURI:              "input uri",
		Duration:              duration,
	}, DefaultExecutionTransformerOptions)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.NodeExecution{
		Id:       &nodeExecutionIdentifier,
		InputUri: "input uri",
		Closure:  closure,
		Metadata: &nodeExecutionMetadata,
	}, nodeExecution))
}

func TestFromNodeExecutionModel_Error(t *testing.T) {
	extraLongErrMsg := string(make([]byte, 2*trimmedErrMessageLen))
	execErr := &core.ExecutionError{
		Code:    "CODE",
		Message: extraLongErrMsg,
		Kind:    core.ExecutionError_USER,
	}
	executionClosureBytes, _ := proto.Marshal(&admin.ExecutionClosure{
		Phase:        core.WorkflowExecution_FAILED,
		OutputResult: &admin.ExecutionClosure_Error{Error: execErr},
	})
	nodeExecution, err := FromNodeExecutionModel(models.NodeExecution{
		NodeExecutionKey: models.NodeExecutionKey{
			NodeID: "nodey",
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
		},
		Closure:               executionClosureBytes,
		NodeExecutionMetadata: nodeExecutionMetadataBytes,
		InputURI:              "input uri",
		Duration:              duration,
	}, &ExecutionTransformerOptions{TrimErrorMessage: true})
	assert.Nil(t, err)

	expectedExecErr := execErr
	expectedExecErr.Message = string(make([]byte, trimmedErrMessageLen))
	assert.Nil(t, err)
	assert.True(t, proto.Equal(expectedExecErr, nodeExecution.Closure.GetError()))
}

func TestFromNodeExecutionModelWithChildren(t *testing.T) {
	nodeExecutionIdentifier := core.NodeExecutionIdentifier{
		NodeId: "nodey",
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
	}
	nodeExecModel := models.NodeExecution{
		NodeExecutionKey: models.NodeExecutionKey{
			NodeID: "nodey",
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
		},
		Phase:                 "NodeExecutionPhase_NODE_PHASE_RUNNING",
		Closure:               closureBytes,
		NodeExecutionMetadata: nodeExecutionMetadataBytes,
		ChildNodeExecutions: []models.NodeExecution{
			{NodeExecutionKey: models.NodeExecutionKey{
				NodeID: "nodec1",
				ExecutionKey: models.ExecutionKey{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
			}},
		},
		InputURI: "input uri",
		Duration: duration,
	}
	t.Run("dynamic workflow", func(t *testing.T) {
		nodeExecModel.DynamicWorkflowRemoteClosureReference = "dummy_dynamic_worklfow_ref"
		nodeExecution, err := FromNodeExecutionModel(nodeExecModel, DefaultExecutionTransformerOptions)
		assert.Nil(t, err)
		assert.True(t, proto.Equal(&admin.NodeExecution{
			Id:       &nodeExecutionIdentifier,
			InputUri: "input uri",
			Closure:  closure,
			Metadata: &admin.NodeExecutionMetaData{
				IsParentNode: true,
				RetryGroup:   "r",
				SpecNodeId:   "sp",
				IsDynamic:    true,
			},
		}, nodeExecution))
	})
	t.Run("non dynamic workflow", func(t *testing.T) {
		nodeExecModel.DynamicWorkflowRemoteClosureReference = ""
		nodeExecution, err := FromNodeExecutionModel(nodeExecModel, DefaultExecutionTransformerOptions)
		assert.Nil(t, err)
		assert.True(t, proto.Equal(&admin.NodeExecution{
			Id:       &nodeExecutionIdentifier,
			InputUri: "input uri",
			Closure:  closure,
			Metadata: &admin.NodeExecutionMetaData{
				IsParentNode: true,
				RetryGroup:   "r",
				SpecNodeId:   "sp",
				IsDynamic:    false,
			},
		}, nodeExecution))
	})
}

func TestGetNodeExecutionInternalData(t *testing.T) {
	t.Run("unset", func(t *testing.T) {
		data, err := GetNodeExecutionInternalData(nil)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(data, &genModel.NodeExecutionInternalData{}))
	})
	t.Run("set", func(t *testing.T) {
		internalData := &genModel.NodeExecutionInternalData{
			EventVersion: 2,
		}
		serializedData, _ := proto.Marshal(internalData)
		actualData, err := GetNodeExecutionInternalData(serializedData)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(internalData, actualData))
	})
	t.Run("invalid internal", func(t *testing.T) {
		_, err := GetNodeExecutionInternalData([]byte("i'm invalid"))
		assert.Equal(t, err.(flyteAdminErrors.FlyteAdminError).Code(), codes.Internal)
	})
}

func TestHandleNodeExecutionInputs(t *testing.T) {
	ctx := context.TODO()
	t.Run("no need to update", func(t *testing.T) {
		nodeExecutionModel := models.NodeExecution{
			InputURI: testInputURI,
		}
		err := handleNodeExecutionInputs(ctx, &nodeExecutionModel, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, nodeExecutionModel.InputURI, testInputURI)
	})
	t.Run("read event input data", func(t *testing.T) {
		nodeExecutionModel := models.NodeExecution{}
		ds, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)
		err = handleNodeExecutionInputs(ctx, &nodeExecutionModel, &admin.NodeExecutionEventRequest{
			Event: &event.NodeExecutionEvent{
				Id: sampleNodeExecID,
				InputValue: &event.NodeExecutionEvent_InputData{
					InputData: testInputs,
				},
			},
		}, ds)
		assert.NoError(t, err)
		expectedOffloadedInputsLocation := "/metadata/project/domain/name/node-id/offloaded_inputs"
		assert.Equal(t, nodeExecutionModel.InputURI, expectedOffloadedInputsLocation)
		actualInputs := &core.LiteralMap{}
		err = ds.ReadProtobuf(ctx, storage.DataReference(expectedOffloadedInputsLocation), actualInputs)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(actualInputs, testInputs))
	})
	t.Run("read event input uri", func(t *testing.T) {
		nodeExecutionModel := models.NodeExecution{}
		err := handleNodeExecutionInputs(ctx, &nodeExecutionModel, &admin.NodeExecutionEventRequest{
			Event: &event.NodeExecutionEvent{
				Id: sampleNodeExecID,
				InputValue: &event.NodeExecutionEvent_InputUri{
					InputUri: testInputURI,
				},
			},
		}, nil)
		assert.NoError(t, err)
		assert.Equal(t, nodeExecutionModel.InputURI, testInputURI)
	})
	t.Run("request contained no input data", func(t *testing.T) {
		nodeExecutionModel := models.NodeExecution{
			InputURI: testInputURI,
		}
		err := handleNodeExecutionInputs(ctx, &nodeExecutionModel, &admin.NodeExecutionEventRequest{
			Event: &event.NodeExecutionEvent{
				Id: sampleNodeExecID,
			},
		}, nil)
		assert.NoError(t, err)
		assert.Equal(t, nodeExecutionModel.InputURI, testInputURI)
	})
}
