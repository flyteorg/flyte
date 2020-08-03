package transformers

import (
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"

	"github.com/golang/protobuf/proto"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/stretchr/testify/assert"

	"github.com/lyft/flyteadmin/pkg/repositories/models"
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
	err := addTerminalState(&request, &nodeExecutionModel, &closure)
	assert.Nil(t, err)
	assert.EqualValues(t, outputURI, closure.GetOutputUri())
	assert.Equal(t, time.Minute, nodeExecutionModel.Duration)
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
	err := addTerminalState(&request, &nodeExecutionModel, &closure)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(error, closure.GetError()))
	assert.Equal(t, time.Minute, nodeExecutionModel.Duration)
}

func TestCreateNodeExecutionModel(t *testing.T) {
	nodeExecutionModel, err := CreateNodeExecutionModel(ToNodeExecutionModelInput{
		Request: &admin.NodeExecutionEventRequest{
			Event: &event.NodeExecutionEvent{
				Id: &core.NodeExecutionIdentifier{
					NodeId: "node id",
					ExecutionId: &core.WorkflowExecutionIdentifier{
						Project: "project",
						Domain:  "domain",
						Name:    "name",
					},
				},
				Phase:    core.NodeExecution_RUNNING,
				InputUri: "input uri",
				OutputResult: &event.NodeExecutionEvent_OutputUri{
					OutputUri: "output uri",
				},
				OccurredAt: occurredAtProto,
				ParentTaskMetadata: &event.ParentTaskExecutionMetadata{
					Id: &core.TaskExecutionIdentifier{
						RetryAttempt: 1,
					},
				},
			},
		},
		ParentTaskExecutionID: 8,
	})
	assert.Nil(t, err)

	var closure = &admin.NodeExecutionClosure{
		Phase:     core.NodeExecution_RUNNING,
		StartedAt: occurredAtProto,
		CreatedAt: occurredAtProto,
		UpdatedAt: occurredAtProto,
	}
	var closureBytes, _ = proto.Marshal(closure)
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
		InputURI:               "input uri",
		StartedAt:              &occurredAt,
		NodeExecutionCreatedAt: &occurredAt,
		NodeExecutionUpdatedAt: &occurredAt,
		NodeExecutionMetadata:  []byte{},
		ParentTaskExecutionID:  8,
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
			},
		}
		nodeExecutionModel := models.NodeExecution{
			Phase: core.NodeExecution_UNDEFINED.String(),
		}
		err := UpdateNodeExecutionModel(&request, &nodeExecutionModel, childExecutionID)
		assert.Nil(t, err)
		assert.Equal(t, core.NodeExecution_RUNNING.String(), nodeExecutionModel.Phase)
		assert.Equal(t, occurredAt, *nodeExecutionModel.StartedAt)
		assert.EqualValues(t, occurredAt, *nodeExecutionModel.NodeExecutionUpdatedAt)
		assert.Nil(t, nodeExecutionModel.CacheStatus)

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
					},
				},
			},
		}
		nodeExecutionModel := models.NodeExecution{
			Phase: core.NodeExecution_UNDEFINED.String(),
		}
		err := UpdateNodeExecutionModel(&request, &nodeExecutionModel, childExecutionID)
		assert.Nil(t, err)
		assert.Equal(t, core.NodeExecution_RUNNING.String(), nodeExecutionModel.Phase)
		assert.Equal(t, occurredAt, *nodeExecutionModel.StartedAt)
		assert.EqualValues(t, occurredAt, *nodeExecutionModel.NodeExecutionUpdatedAt)
		assert.NotNil(t, nodeExecutionModel.CacheStatus)
		assert.Equal(t, *nodeExecutionModel.CacheStatus, request.Event.GetTaskNodeMetadata().CacheStatus.String())

		var closure = &admin.NodeExecutionClosure{
			Phase:     core.NodeExecution_RUNNING,
			StartedAt: occurredAtProto,
			UpdatedAt: occurredAtProto,
			TargetMetadata: &admin.NodeExecutionClosure_TaskNodeMetadata{
				TaskNodeMetadata: &admin.TaskNodeMetadata{
					CacheStatus: request.Event.GetTaskNodeMetadata().CacheStatus,
					CatalogKey:  request.Event.GetTaskNodeMetadata().CatalogKey,
				},
			},
		}
		var closureBytes, _ = proto.Marshal(closure)
		assert.Equal(t, nodeExecutionModel.Closure, closureBytes)
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
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.NodeExecution{
		Id:       &nodeExecutionIdentifier,
		InputUri: "input uri",
		Closure:  closure,
		Metadata: &nodeExecutionMetadata,
	}, nodeExecution))
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
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.NodeExecution{
		Id:       &nodeExecutionIdentifier,
		InputUri: "input uri",
		Closure:  closure,
		Metadata: &admin.NodeExecutionMetaData{
			IsParentNode: true,
			RetryGroup:   "r",
			SpecNodeId:   "sp",
		},
	}, nodeExecution))
}

func TestFromNodeExecutionModels(t *testing.T) {
	nodeExecutions, err := FromNodeExecutionModels([]models.NodeExecution{
		{
			NodeExecutionKey: models.NodeExecutionKey{
				NodeID: "node id",
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
		},
	})
	assert.Nil(t, err)
	assert.Len(t, nodeExecutions, 1)
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
		Closure:  closure,
		Metadata: &nodeExecutionMetadata,
	}, nodeExecutions[0]))
}
