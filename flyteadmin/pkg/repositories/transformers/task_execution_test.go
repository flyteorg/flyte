package transformers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/flyteorg/flytestdlib/promutils"

	commonMocks "github.com/flyteorg/flyteadmin/pkg/common/mocks"
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	ptypesStruct "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
)

var taskEventOccurredAt = time.Now().UTC()
var taskEventOccurredAtProto, _ = ptypes.TimestampProto(taskEventOccurredAt)

var sampleTaskID = &core.Identifier{
	ResourceType: core.ResourceType_TASK,
	Project:      "project",
	Domain:       "domain",
	Name:         "task-id",
	Version:      "task-v",
}

var sampleNodeExecID = &core.NodeExecutionIdentifier{
	NodeId: "node-id",
	ExecutionId: &core.WorkflowExecutionIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	},
}

var retryAttemptValue = uint32(1)

var customInfo = ptypesStruct.Struct{
	Fields: map[string]*ptypesStruct.Value{
		"phase": {
			Kind: &ptypesStruct.Value_StringValue{
				StringValue: "value",
			},
		},
	},
}

func transformMapToStructPB(t *testing.T, thing map[string]string) *structpb.Struct {
	b, err := json.Marshal(thing)
	if err != nil {
		t.Fatal(t, err)
	}

	thingAsCustom := &structpb.Struct{}
	if err := jsonpb.UnmarshalString(string(b), thingAsCustom); err != nil {
		t.Fatal(t, err)
	}
	return thingAsCustom
}

func TestAddTaskStartedState(t *testing.T) {
	t.Run("model with unset started At ", func(t *testing.T) {
		var startedAt = time.Now().UTC()
		var startedAtProto, _ = ptypes.TimestampProto(startedAt)
		request := admin.TaskExecutionEventRequest{
			Event: &event.TaskExecutionEvent{
				Phase:      core.TaskExecution_RUNNING,
				OccurredAt: startedAtProto,
			},
		}
		taskExecutionModel := models.TaskExecution{}
		closure := &admin.TaskExecutionClosure{}
		err := addTaskStartedState(&request, &taskExecutionModel, closure)
		assert.Nil(t, err)

		timestamp, err := ptypes.Timestamp(closure.StartedAt)
		assert.Nil(t, err)
		assert.Equal(t, startedAt, timestamp)
		assert.Equal(t, &startedAt, taskExecutionModel.StartedAt)
	})
	t.Run("model with set started At ", func(t *testing.T) {
		var oldStartedAt = time.Now().UTC()
		var newStartedAt = time.Now().UTC().Add(time.Minute * -10)
		var startedAtProto, _ = ptypes.TimestampProto(newStartedAt)
		request := admin.TaskExecutionEventRequest{
			Event: &event.TaskExecutionEvent{
				Phase:      core.TaskExecution_RUNNING,
				OccurredAt: startedAtProto,
			},
		}
		taskExecutionModel := models.TaskExecution{
			StartedAt: &oldStartedAt,
		}
		closure := &admin.TaskExecutionClosure{
			StartedAt: startedAtProto,
		}
		err := addTaskStartedState(&request, &taskExecutionModel, closure)
		assert.Nil(t, err)

		timestamp, err := ptypes.Timestamp(closure.StartedAt)
		assert.Nil(t, err)
		assert.NotEqual(t, oldStartedAt, timestamp)
		assert.Equal(t, &oldStartedAt, taskExecutionModel.StartedAt)
	})

}

func TestAddTaskTerminalState_Error(t *testing.T) {
	expectedErr := &core.ExecutionError{
		Code: "foo",
	}
	request := admin.TaskExecutionEventRequest{
		Event: &event.TaskExecutionEvent{
			Phase: core.TaskExecution_FAILED,
			OutputResult: &event.TaskExecutionEvent_Error{
				Error: expectedErr,
			},
			OccurredAt: occurredAtProto,
		},
	}
	startedAt := occurredAt.Add(-time.Minute)

	startedAtProto, _ := ptypes.TimestampProto(startedAt)
	taskExecutionModel := models.TaskExecution{
		StartedAt: &startedAt,
	}
	closure := admin.TaskExecutionClosure{
		StartedAt: startedAtProto,
	}
	err := addTaskTerminalState(context.TODO(), &request, &taskExecutionModel, &closure,
		interfaces.InlineEventDataPolicyStoreInline, commonMocks.GetMockStorageClient())
	assert.Nil(t, err)
	assert.True(t, proto.Equal(expectedErr, closure.GetError()))
	assert.Equal(t, time.Minute, taskExecutionModel.Duration)
}

func TestAddTaskTerminalState_OutputURI(t *testing.T) {
	outputURI := "output uri"
	request := admin.TaskExecutionEventRequest{
		Event: &event.TaskExecutionEvent{
			Phase: core.TaskExecution_SUCCEEDED,
			OutputResult: &event.TaskExecutionEvent_OutputUri{
				OutputUri: outputURI,
			},
			OccurredAt: taskEventOccurredAtProto,
		},
	}
	startedAt := taskEventOccurredAt.Add(-time.Minute)
	taskExecutionModel := models.TaskExecution{
		StartedAt: &startedAt,
	}

	closure := &admin.TaskExecutionClosure{}
	err := addTaskTerminalState(context.TODO(), &request, &taskExecutionModel, closure,
		interfaces.InlineEventDataPolicyStoreInline, commonMocks.GetMockStorageClient())
	assert.Nil(t, err)

	duration, err := ptypes.Duration(closure.GetDuration())
	assert.Nil(t, err)
	assert.EqualValues(t, request.Event.OutputResult, closure.OutputResult)
	assert.EqualValues(t, outputURI, closure.GetOutputUri())
	assert.EqualValues(t, time.Minute, duration)

	assert.Equal(t, time.Minute, taskExecutionModel.Duration)
}

func TestAddTaskTerminalState_OutputData(t *testing.T) {
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
	request := admin.TaskExecutionEventRequest{
		Event: &event.TaskExecutionEvent{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "project",
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
			RetryAttempt: 1,
			ParentNodeExecutionId: &core.NodeExecutionIdentifier{
				NodeId: "node id",
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "ex project",
					Domain:  "ex domain",
					Name:    "ex name",
				},
			},
			Phase: core.TaskExecution_SUCCEEDED,
			OutputResult: &event.TaskExecutionEvent_OutputData{
				OutputData: outputData,
			},
			OccurredAt: taskEventOccurredAtProto,
		},
	}
	startedAt := taskEventOccurredAt.Add(-time.Minute)
	taskExecutionModel := models.TaskExecution{
		StartedAt: &startedAt,
	}

	t.Run("output data stored inline", func(t *testing.T) {
		closure := &admin.TaskExecutionClosure{}
		err := addTaskTerminalState(context.TODO(), &request, &taskExecutionModel, closure,
			interfaces.InlineEventDataPolicyStoreInline, commonMocks.GetMockStorageClient())
		assert.Nil(t, err)

		duration, err := ptypes.Duration(closure.GetDuration())
		assert.Nil(t, err)
		assert.EqualValues(t, request.Event.OutputResult, closure.OutputResult)
		assert.True(t, proto.Equal(outputData, closure.GetOutputData()))
		assert.EqualValues(t, time.Minute, duration)
	})
	t.Run("output data stored offloaded", func(t *testing.T) {
		mockStorage := commonMocks.GetMockStorageClient()
		mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).WriteProtobufCb = func(ctx context.Context, reference storage.DataReference, opts storage.Options, msg proto.Message) error {
			assert.Equal(t, reference.String(), "s3://bucket/metadata/ex project/ex domain/ex name/node id/project/domain/name/version/1/offloaded_outputs")
			return nil
		}

		closure := &admin.TaskExecutionClosure{}
		err := addTaskTerminalState(context.TODO(), &request, &taskExecutionModel, closure,
			interfaces.InlineEventDataPolicyOffload, mockStorage)
		assert.Nil(t, err)
		assert.Equal(t, "s3://bucket/metadata/ex project/ex domain/ex name/node id/project/domain/name/version/1/offloaded_outputs", closure.GetOutputUri())
	})

	assert.Equal(t, time.Minute, taskExecutionModel.Duration)
}

func TestCreateTaskExecutionModelQueued(t *testing.T) {
	ds, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)
	taskExecutionModel, err := CreateTaskExecutionModel(context.TODO(), CreateTaskExecutionModelInput{
		Request: &admin.TaskExecutionEventRequest{
			Event: &event.TaskExecutionEvent{
				TaskId:                sampleTaskID,
				ParentNodeExecutionId: sampleNodeExecID,
				Phase:                 core.TaskExecution_QUEUED,
				RetryAttempt:          1,
				InputValue: &event.TaskExecutionEvent_InputData{
					InputData: testInputs,
				},
				OccurredAt: taskEventOccurredAtProto,
				Reason:     "Task was scheduled",
				TaskType:   "sidecar",
			},
		},
		StorageClient: ds,
	})
	assert.Nil(t, err)

	expectedClosure := &admin.TaskExecutionClosure{
		Phase:     core.TaskExecution_QUEUED,
		StartedAt: nil,
		CreatedAt: taskEventOccurredAtProto,
		UpdatedAt: taskEventOccurredAtProto,
		Reason:    "Task was scheduled",
		Reasons: []*admin.Reason{
			&admin.Reason{
				OccurredAt: taskEventOccurredAtProto,
				Message:    "Task was scheduled",
			},
		},
		TaskType: "sidecar",
	}

	expectedClosureBytes, err := proto.Marshal(expectedClosure)
	assert.Nil(t, err)

	assert.Equal(t, &models.TaskExecution{
		TaskExecutionKey: models.TaskExecutionKey{
			TaskKey: models.TaskKey{
				Project: sampleTaskID.Project,
				Domain:  sampleTaskID.Domain,
				Name:    sampleTaskID.Name,
				Version: sampleTaskID.Version,
			},
			NodeExecutionKey: models.NodeExecutionKey{
				NodeID: sampleNodeExecID.NodeId,
				ExecutionKey: models.ExecutionKey{
					Project: sampleNodeExecID.ExecutionId.Project,
					Domain:  sampleNodeExecID.ExecutionId.Domain,
					Name:    sampleNodeExecID.ExecutionId.Name,
				},
			},
			RetryAttempt: &retryAttemptValue,
		},
		Phase:                  "QUEUED",
		InputURI:               "/metadata/project/domain/name/node-id/project/domain/task-id/task-v/1/offloaded_inputs",
		Closure:                expectedClosureBytes,
		StartedAt:              nil,
		TaskExecutionCreatedAt: &taskEventOccurredAt,
		TaskExecutionUpdatedAt: &taskEventOccurredAt,
	}, taskExecutionModel)
}

func TestCreateTaskExecutionModelRunning(t *testing.T) {
	taskExecutionModel, err := CreateTaskExecutionModel(context.TODO(), CreateTaskExecutionModelInput{
		Request: &admin.TaskExecutionEventRequest{
			Event: &event.TaskExecutionEvent{
				TaskId:                sampleTaskID,
				ParentNodeExecutionId: sampleNodeExecID,
				Phase:                 core.TaskExecution_RUNNING,
				PhaseVersion:          uint32(2),
				RetryAttempt:          1,
				InputValue: &event.TaskExecutionEvent_InputUri{
					InputUri: testInputURI,
				},
				OutputResult: &event.TaskExecutionEvent_OutputUri{
					OutputUri: "output uri",
				},
				OccurredAt: taskEventOccurredAtProto,
				Logs: []*core.TaskLog{
					{
						Name: "some_log",
						Uri:  "some_uri",
					},
					{
						Name: "some_log2",
						Uri:  "some_uri2",
					},
				},
				CustomInfo: &customInfo,
			},
		},
	})
	assert.Nil(t, err)

	expectedClosure := &admin.TaskExecutionClosure{
		Phase:     core.TaskExecution_RUNNING,
		StartedAt: taskEventOccurredAtProto,
		CreatedAt: taskEventOccurredAtProto,
		UpdatedAt: taskEventOccurredAtProto,
		Logs: []*core.TaskLog{
			{
				Name: "some_log",
				Uri:  "some_uri",
			},
			{
				Name: "some_log2",
				Uri:  "some_uri2",
			},
		},
		CustomInfo: &customInfo,
	}

	t.Logf("expected %+v %+v\n", expectedClosure.Reason, expectedClosure.Reasons)

	expectedClosureBytes, err := proto.Marshal(expectedClosure)
	assert.Nil(t, err)

	assert.Equal(t, &models.TaskExecution{
		TaskExecutionKey: models.TaskExecutionKey{
			TaskKey: models.TaskKey{
				Project: sampleTaskID.Project,
				Domain:  sampleTaskID.Domain,
				Name:    sampleTaskID.Name,
				Version: sampleTaskID.Version,
			},
			NodeExecutionKey: models.NodeExecutionKey{
				NodeID: sampleNodeExecID.NodeId,
				ExecutionKey: models.ExecutionKey{
					Project: sampleNodeExecID.ExecutionId.Project,
					Domain:  sampleNodeExecID.ExecutionId.Domain,
					Name:    sampleNodeExecID.ExecutionId.Name,
				},
			},
			RetryAttempt: &retryAttemptValue,
		},
		Phase:                  "RUNNING",
		PhaseVersion:           uint32(2),
		InputURI:               testInputURI,
		Closure:                expectedClosureBytes,
		StartedAt:              &taskEventOccurredAt,
		TaskExecutionCreatedAt: &taskEventOccurredAt,
		TaskExecutionUpdatedAt: &taskEventOccurredAt,
	}, taskExecutionModel)
}

func TestUpdateTaskExecutionModelRunningToFailed(t *testing.T) {
	existingClosure := &admin.TaskExecutionClosure{
		Phase:     core.TaskExecution_RUNNING,
		StartedAt: taskEventOccurredAtProto,
		CreatedAt: taskEventOccurredAtProto,
		UpdatedAt: taskEventOccurredAtProto,
		Logs: []*core.TaskLog{
			{
				Uri: "uri_a",
			},
			{
				Uri: "uri_b",
			},
		},
		CustomInfo: transformMapToStructPB(t, map[string]string{
			"key1": "value1",
		}),
		Reason: "Task was scheduled",
		Reasons: []*admin.Reason{
			&admin.Reason{
				OccurredAt: taskEventOccurredAtProto,
				Message:    "Task was scheduled",
			},
		},
	}

	closureBytes, err := proto.Marshal(existingClosure)
	assert.Nil(t, err)

	existingTaskExecution := models.TaskExecution{
		TaskExecutionKey: models.TaskExecutionKey{
			TaskKey: models.TaskKey{
				Project: sampleTaskID.Project,
				Domain:  sampleTaskID.Domain,
				Name:    sampleTaskID.Name,
				Version: sampleTaskID.Version,
			},
			NodeExecutionKey: models.NodeExecutionKey{
				NodeID: sampleNodeExecID.NodeId,
				ExecutionKey: models.ExecutionKey{
					Project: sampleNodeExecID.ExecutionId.Project,
					Domain:  sampleNodeExecID.ExecutionId.Domain,
					Name:    sampleNodeExecID.ExecutionId.Name,
				},
			},
			RetryAttempt: &retryAttemptValue,
		},
		Phase:                  "TaskExecutionPhase_TASK_PHASE_RUNNING",
		InputURI:               "input uri",
		Closure:                closureBytes,
		StartedAt:              &taskEventOccurredAt,
		TaskExecutionCreatedAt: &taskEventOccurredAt,
		TaskExecutionUpdatedAt: &taskEventOccurredAt,
	}

	occuredAt := taskEventOccurredAt.Add(time.Minute)
	occuredAtProto, err := ptypes.TimestampProto(occuredAt)
	assert.Nil(t, err)

	outputError := &core.ExecutionError{
		ErrorUri: "error.pb",
	}

	failedEventRequest := &admin.TaskExecutionEventRequest{
		Event: &event.TaskExecutionEvent{
			TaskId:                sampleTaskID,
			ParentNodeExecutionId: sampleNodeExecID,
			Phase:                 core.TaskExecution_FAILED,
			RetryAttempt:          1,
			InputValue: &event.TaskExecutionEvent_InputUri{
				InputUri: testInputURI,
			},
			OutputResult: &event.TaskExecutionEvent_Error{
				Error: outputError,
			},
			OccurredAt: occuredAtProto,
			Logs: []*core.TaskLog{
				{
					Uri: "uri_b",
				},
				{
					Uri: "uri_c",
				},
			},
			CustomInfo: transformMapToStructPB(t, map[string]string{
				"key1": "value1 updated",
			}),
			Reason: "task failed",
		},
	}

	err = UpdateTaskExecutionModel(context.TODO(), failedEventRequest, &existingTaskExecution,
		interfaces.InlineEventDataPolicyStoreInline, commonMocks.GetMockStorageClient())
	assert.Nil(t, err)

	expectedClosure := &admin.TaskExecutionClosure{
		Phase:     core.TaskExecution_FAILED,
		StartedAt: taskEventOccurredAtProto,
		UpdatedAt: occuredAtProto,
		CreatedAt: taskEventOccurredAtProto,
		Duration:  ptypes.DurationProto(time.Minute),
		OutputResult: &admin.TaskExecutionClosure_Error{
			Error: outputError,
		},
		Logs: []*core.TaskLog{
			{
				Uri: "uri_b",
			},
			{
				Uri: "uri_c",
			},
			{
				Uri: "uri_a",
			},
		},
		CustomInfo: transformMapToStructPB(t, map[string]string{
			"key1": "value1 updated",
		}),
		Reason: "task failed",
		Reasons: []*admin.Reason{
			&admin.Reason{
				OccurredAt: taskEventOccurredAtProto,
				Message:    "Task was scheduled",
			},
			&admin.Reason{
				OccurredAt: occuredAtProto,
				Message:    "task failed",
			},
		},
	}

	expectedClosureBytes, err := proto.Marshal(expectedClosure)
	assert.Nil(t, err)

	assert.EqualValues(t, models.TaskExecution{
		TaskExecutionKey: models.TaskExecutionKey{
			TaskKey: models.TaskKey{
				Project: sampleTaskID.Project,
				Domain:  sampleTaskID.Domain,
				Name:    sampleTaskID.Name,
				Version: sampleTaskID.Version,
			},
			NodeExecutionKey: models.NodeExecutionKey{
				NodeID: sampleNodeExecID.NodeId,
				ExecutionKey: models.ExecutionKey{
					Project: sampleNodeExecID.ExecutionId.Project,
					Domain:  sampleNodeExecID.ExecutionId.Domain,
					Name:    sampleNodeExecID.ExecutionId.Name,
				},
			},
			RetryAttempt: &retryAttemptValue,
		},
		Phase:                  "FAILED",
		InputURI:               "input uri",
		Closure:                expectedClosureBytes,
		StartedAt:              &taskEventOccurredAt,
		TaskExecutionUpdatedAt: &occuredAt,
		TaskExecutionCreatedAt: &taskEventOccurredAt,
		Duration:               time.Minute,
	}, existingTaskExecution)

}

func TestFromTaskExecutionModel(t *testing.T) {
	taskClosure := &admin.TaskExecutionClosure{
		Phase: core.TaskExecution_RUNNING,
		OutputResult: &admin.TaskExecutionClosure_OutputUri{
			OutputUri: "out.pb",
		},
		Duration:  ptypes.DurationProto(time.Minute),
		StartedAt: taskEventOccurredAtProto,
	}
	closureBytes, err := proto.Marshal(taskClosure)
	assert.Nil(t, err)
	taskExecution, err := FromTaskExecutionModel(models.TaskExecution{
		TaskExecutionKey: models.TaskExecutionKey{
			TaskKey: models.TaskKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
				Version: "version",
			},
			NodeExecutionKey: models.NodeExecutionKey{
				NodeID: "node id",
				ExecutionKey: models.ExecutionKey{
					Project: "ex project",
					Domain:  "ex domain",
					Name:    "ex name",
				},
			},
			RetryAttempt: &retryAttemptValue,
		},
		Phase:    "TaskExecutionPhase_TASK_PHASE_RUNNING",
		InputURI: "input uri",
		Duration: duration,
		Closure:  closureBytes,
	}, DefaultExecutionTransformerOptions)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.TaskExecution{
		Id: &core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "project",
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
			RetryAttempt: 1,
			NodeExecutionId: &core.NodeExecutionIdentifier{
				NodeId: "node id",
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "ex project",
					Domain:  "ex domain",
					Name:    "ex name",
				},
			},
		},
		InputUri: "input uri",
		Closure:  taskClosure,
	}, taskExecution))
}

func TestFromTaskExecutionModel_Error(t *testing.T) {
	extraLongErrMsg := string(make([]byte, 2*trimmedErrMessageLen))
	execErr := &core.ExecutionError{
		Code:    "CODE",
		Message: extraLongErrMsg,
		Kind:    core.ExecutionError_USER,
	}
	closureBytes, _ := proto.Marshal(&admin.ExecutionClosure{
		Phase:        core.WorkflowExecution_FAILED,
		OutputResult: &admin.ExecutionClosure_Error{Error: execErr},
	})
	taskExecutionModel := models.TaskExecution{
		TaskExecutionKey: models.TaskExecutionKey{
			TaskKey: models.TaskKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
				Version: "version",
			},
			NodeExecutionKey: models.NodeExecutionKey{
				NodeID: "node id",
				ExecutionKey: models.ExecutionKey{
					Project: "ex project",
					Domain:  "ex domain",
					Name:    "ex name",
				},
			},
			RetryAttempt: &retryAttemptValue,
		},
		InputURI: "input uri",
		Duration: duration,
		Closure:  closureBytes,
	}
	taskExecution, err := FromTaskExecutionModel(taskExecutionModel, &ExecutionTransformerOptions{
		TrimErrorMessage: true,
	})

	expectedExecErr := execErr
	expectedExecErr.Message = string(make([]byte, trimmedErrMessageLen))
	assert.Nil(t, err)
	assert.True(t, proto.Equal(expectedExecErr, taskExecution.Closure.GetError()))

	extraShortErrMsg := string(make([]byte, 10))
	execErr = &core.ExecutionError{
		Code:    "CODE",
		Message: extraShortErrMsg,
		Kind:    core.ExecutionError_USER,
	}
	closureBytes, _ = proto.Marshal(&admin.ExecutionClosure{
		Phase:        core.WorkflowExecution_FAILED,
		OutputResult: &admin.ExecutionClosure_Error{Error: execErr},
	})
	taskExecutionModel.Closure = closureBytes
	taskExecution, err = FromTaskExecutionModel(taskExecutionModel, &ExecutionTransformerOptions{
		TrimErrorMessage: true,
	})
	expectedExecErr = execErr
	expectedExecErr.Message = string(make([]byte, 10))
	assert.Nil(t, err)
	assert.True(t, proto.Equal(expectedExecErr, taskExecution.Closure.GetError()))
}

func TestFromTaskExecutionModels(t *testing.T) {
	taskClosure := &admin.TaskExecutionClosure{
		Phase: core.TaskExecution_RUNNING,
		OutputResult: &admin.TaskExecutionClosure_OutputUri{
			OutputUri: "out.pb",
		},
		Duration:  ptypes.DurationProto(time.Minute),
		StartedAt: taskEventOccurredAtProto,
	}
	closureBytes, err := proto.Marshal(taskClosure)
	assert.Nil(t, err)
	taskExecutions, err := FromTaskExecutionModels([]models.TaskExecution{
		{
			TaskExecutionKey: models.TaskExecutionKey{
				TaskKey: models.TaskKey{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
					Version: "version",
				},
				NodeExecutionKey: models.NodeExecutionKey{
					NodeID: "nodey",
					ExecutionKey: models.ExecutionKey{
						Project: "ex project",
						Domain:  "ex domain",
						Name:    "ex name",
					},
				},
				RetryAttempt: &retryAttemptValue,
			},
			Phase:    "TaskExecutionPhase_TASK_PHASE_RUNNING",
			InputURI: "input uri",
			Duration: duration,
			Closure:  closureBytes,
		},
	}, DefaultExecutionTransformerOptions)
	assert.Nil(t, err)
	assert.Len(t, taskExecutions, 1)
	assert.True(t, proto.Equal(&admin.TaskExecution{
		Id: &core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "project",
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
			RetryAttempt: 1,
			NodeExecutionId: &core.NodeExecutionIdentifier{
				NodeId: "nodey",
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "ex project",
					Domain:  "ex domain",
					Name:    "ex name",
				},
			},
		},
		InputUri: "input uri",
		Closure:  taskClosure,
	}, taskExecutions[0]))
}

func TestMergeLogs(t *testing.T) {
	type testCase struct {
		existing []*core.TaskLog
		latest   []*core.TaskLog
		expected []*core.TaskLog
		name     string
	}

	testCases := []testCase{
		{
			existing: []*core.TaskLog{
				{
					Uri: "uri_a",
				},
				{
					Uri: "uri_b",
				},
			},
			latest: []*core.TaskLog{
				{
					Uri: "uri_b",
				},
				{
					Uri: "uri_c",
				},
			},
			expected: []*core.TaskLog{
				{
					Uri: "uri_b",
				},
				{
					Uri: "uri_c",
				},
				{
					Uri: "uri_a",
				},
			},
			name: "Merge logs with empty names",
		},
		{
			existing: []*core.TaskLog{
				{
					Uri:  "uri_a",
					Name: "name_a",
				},
				{
					Uri:  "uri_b_old",
					Name: "name_b",
				},
				{
					Uri:  "uri_c",
					Name: "name_c",
				},
			},
			latest: []*core.TaskLog{
				{
					Uri:  "uri_b",
					Name: "name_b",
				},
				{
					Uri:  "uri_d",
					Name: "name_d",
				},
			},
			expected: []*core.TaskLog{
				{
					Uri:  "uri_b",
					Name: "name_b",
				},
				{
					Uri:  "uri_d",
					Name: "name_d",
				},
				{
					Uri:  "uri_a",
					Name: "name_a",
				},
				{
					Uri:  "uri_c",
					Name: "name_c",
				},
			},
			name: "Merge unique logs by name",
		},
		{
			existing: []*core.TaskLog{
				{
					Uri:  "uri_a",
					Name: "name_a",
				},
				{
					Uri:  "uri_b",
					Name: "ignored_name_b",
				},
				{
					Uri:  "uri_c",
					Name: "name_c",
				},
			},
			latest: []*core.TaskLog{
				{
					Uri:  "uri_b",
					Name: "name_b",
				},
				{
					Uri:  "uri_d",
					Name: "name_d",
				},
			},
			expected: []*core.TaskLog{
				{
					Uri:  "uri_b",
					Name: "name_b",
				},
				{
					Uri:  "uri_d",
					Name: "name_d",
				},
				{
					Uri:  "uri_a",
					Name: "name_a",
				},
				{
					Uri:  "uri_c",
					Name: "name_c",
				},
			},
			name: "Merge unique logs",
		},
		{
			latest: []*core.TaskLog{
				{
					Uri:  "uri_b",
					Name: "name_b",
				},
			},
			expected: []*core.TaskLog{
				{
					Uri:  "uri_b",
					Name: "name_b",
				},
			},
			name: "Empty existing logs",
		},
		{
			existing: []*core.TaskLog{
				{
					Uri:  "uri_b",
					Name: "name_b",
				},
			},
			expected: []*core.TaskLog{
				{
					Uri:  "uri_b",
					Name: "name_b",
				},
			},
			name: "Empty latest logs",
		},
		{
			name: "Nothing to do",
		},
	}
	for _, mergeTestCase := range testCases {
		actual := mergeLogs(mergeTestCase.existing, mergeTestCase.latest)
		assert.Equal(t, len(mergeTestCase.expected), len(actual), fmt.Sprintf("%s failed", mergeTestCase.name))
		for idx, expectedLog := range mergeTestCase.expected {
			assert.True(t, proto.Equal(expectedLog, actual[idx]), fmt.Sprintf("%s failed", mergeTestCase.name))
		}
	}
}

func TestMergeCustoms(t *testing.T) {
	t.Run("nothing to do", func(t *testing.T) {
		custom, err := mergeCustom(nil, nil)
		assert.NoError(t, err)
		assert.Nil(t, custom)
	})

	// Turn JSON into a protobuf struct
	existingCustom := transformMapToStructPB(t, map[string]string{
		"foo": "bar",
		"1":   "value1",
	})
	latestCustom := transformMapToStructPB(t, map[string]string{
		"foo": "something different",
		"2":   "value2",
	})

	t.Run("use existing", func(t *testing.T) {
		mergedCustom, err := mergeCustom(existingCustom, nil)
		assert.Nil(t, err)
		assert.True(t, proto.Equal(mergedCustom, existingCustom))
	})
	t.Run("use latest", func(t *testing.T) {
		mergedCustom, err := mergeCustom(nil, latestCustom)
		assert.Nil(t, err)
		assert.True(t, proto.Equal(mergedCustom, latestCustom))
	})
	t.Run("merge", func(t *testing.T) {
		mergedCustom, err := mergeCustom(existingCustom, latestCustom)
		assert.Nil(t, err)

		var marshaler jsonpb.Marshaler
		mergedJSON, err := marshaler.MarshalToString(mergedCustom)
		assert.Nil(t, err)

		var mergedMap map[string]string
		err = json.Unmarshal([]byte(mergedJSON), &mergedMap)
		assert.Nil(t, err)
		assert.EqualValues(t, map[string]string{
			"1":   "value1",
			"foo": "something different",
			"2":   "value2",
		}, mergedMap)

	})
}

func TestMergeExternalResource(t *testing.T) {
	type testCase struct {
		existing *event.ExternalResourceInfo
		latest   *event.ExternalResourceInfo
		expected *event.ExternalResourceInfo
		name     string
	}

	testCases := []testCase{
		{
			existing: nil,
			latest:   nil,
			expected: nil,
			name:     "do nothing",
		},
		{
			existing: &event.ExternalResourceInfo{
				ExternalId: "foo",
			},
			latest: nil,
			expected: &event.ExternalResourceInfo{
				ExternalId: "foo",
			},
			name: "use existing",
		},
		{
			existing: nil,
			latest: &event.ExternalResourceInfo{
				ExternalId: "foo",
			},
			expected: &event.ExternalResourceInfo{
				ExternalId: "foo",
			},
			name: "use latest",
		},
		{
			existing: &event.ExternalResourceInfo{
				ExternalId:   "foo",
				CacheStatus:  core.CatalogCacheStatus_CACHE_HIT,
				RetryAttempt: 1,
				Phase:        core.TaskExecution_RUNNING,
			},
			latest: &event.ExternalResourceInfo{
				Phase: core.TaskExecution_SUCCEEDED,
			},
			expected: &event.ExternalResourceInfo{
				ExternalId:   "foo",
				CacheStatus:  core.CatalogCacheStatus_CACHE_HIT,
				RetryAttempt: 1,
				Phase:        core.TaskExecution_SUCCEEDED,
			},
			name: "update phase",
		},
		{
			existing: &event.ExternalResourceInfo{},
			latest: &event.ExternalResourceInfo{
				ExternalId:  "foo",
				CacheStatus: core.CatalogCacheStatus_CACHE_HIT,
				Phase:       core.TaskExecution_RUNNING,
			},
			expected: &event.ExternalResourceInfo{
				ExternalId:  "foo",
				CacheStatus: core.CatalogCacheStatus_CACHE_HIT,
				Phase:       core.TaskExecution_RUNNING,
			},
			name: "update everything",
		},
	}

	for _, mergeTestCase := range testCases {
		t.Run(mergeTestCase.name, func(t *testing.T) {
			actual := mergeExternalResource(mergeTestCase.existing, mergeTestCase.latest)
			assert.True(t, proto.Equal(mergeTestCase.expected, actual))
		})
	}
}

func TestMergeExternalResources(t *testing.T) {
	type testCase struct {
		existing []*event.ExternalResourceInfo
		latest   []*event.ExternalResourceInfo
		expected []*event.ExternalResourceInfo
		name     string
	}

	testCases := []testCase{
		{
			existing: nil,
			latest:   nil,
			expected: nil,
			name:     "do nothing",
		},
		{
			existing: []*event.ExternalResourceInfo{
				&event.ExternalResourceInfo{
					ExternalId: "foo",
					Index:      1,
				},
			},
			latest: nil,
			expected: []*event.ExternalResourceInfo{
				&event.ExternalResourceInfo{
					ExternalId: "foo",
					Index:      1,
				},
			},
			name: "use existing",
		},
		{
			existing: nil,
			latest: []*event.ExternalResourceInfo{
				&event.ExternalResourceInfo{
					ExternalId: "foo",
					Index:      1,
				},
			},
			expected: []*event.ExternalResourceInfo{
				&event.ExternalResourceInfo{
					ExternalId: "foo",
					Index:      1,
				},
			},
			name: "use latest",
		},
		{
			existing: []*event.ExternalResourceInfo{
				&event.ExternalResourceInfo{
					ExternalId: "baz",
					Index:      1,
				},
			},
			latest: []*event.ExternalResourceInfo{
				&event.ExternalResourceInfo{
					ExternalId: "bar",
					Index:      0,
				},
			},
			expected: []*event.ExternalResourceInfo{
				&event.ExternalResourceInfo{
					ExternalId: "bar",
					Index:      0,
				},
				&event.ExternalResourceInfo{
					ExternalId: "baz",
					Index:      1,
				},
			},
			name: "add subtask before",
		},
		{
			existing: []*event.ExternalResourceInfo{
				&event.ExternalResourceInfo{
					ExternalId: "baz",
					Index:      1,
				},
			},
			latest: []*event.ExternalResourceInfo{
				&event.ExternalResourceInfo{
					ExternalId: "foo",
					Index:      2,
				},
			},
			expected: []*event.ExternalResourceInfo{
				&event.ExternalResourceInfo{
					ExternalId: "baz",
					Index:      1,
				},
				&event.ExternalResourceInfo{
					ExternalId: "foo",
					Index:      2,
				},
			},
			name: "add subtask after",
		},
		{
			existing: []*event.ExternalResourceInfo{
				&event.ExternalResourceInfo{
					ExternalId: "baz",
					Index:      1,
				},
			},
			latest: []*event.ExternalResourceInfo{
				&event.ExternalResourceInfo{
					ExternalId:   "baz",
					Index:        1,
					RetryAttempt: 1,
				},
			},
			expected: []*event.ExternalResourceInfo{
				&event.ExternalResourceInfo{
					ExternalId: "baz",
					Index:      1,
				},
				&event.ExternalResourceInfo{
					ExternalId:   "baz",
					Index:        1,
					RetryAttempt: 1,
				},
			},
			name: "add subtask retry",
		},
		{
			existing: []*event.ExternalResourceInfo{
				&event.ExternalResourceInfo{
					ExternalId:   "foo",
					Index:        0,
					RetryAttempt: 0,
					Phase:        core.TaskExecution_UNDEFINED,
				},
				&event.ExternalResourceInfo{
					ExternalId:   "bar",
					Index:        1,
					RetryAttempt: 0,
					Phase:        core.TaskExecution_UNDEFINED,
				},
				&event.ExternalResourceInfo{
					ExternalId:   "baz",
					Index:        2,
					RetryAttempt: 0,
					Phase:        core.TaskExecution_UNDEFINED,
				},
			},
			latest: []*event.ExternalResourceInfo{
				&event.ExternalResourceInfo{
					ExternalId:   "baz",
					Index:        2,
					RetryAttempt: 0,
					Phase:        core.TaskExecution_RUNNING,
				},
			},
			expected: []*event.ExternalResourceInfo{
				&event.ExternalResourceInfo{
					ExternalId:   "foo",
					Index:        0,
					RetryAttempt: 0,
					Phase:        core.TaskExecution_UNDEFINED,
				},
				&event.ExternalResourceInfo{
					ExternalId:   "bar",
					Index:        1,
					RetryAttempt: 0,
					Phase:        core.TaskExecution_UNDEFINED,
				},
				&event.ExternalResourceInfo{
					ExternalId:   "baz",
					Index:        2,
					RetryAttempt: 0,
					Phase:        core.TaskExecution_RUNNING,
				},
			},
			name: "update subtask",
		},
	}

	for _, mergeTestCase := range testCases {
		t.Run(mergeTestCase.name, func(t *testing.T) {
			actual := mergeExternalResources(mergeTestCase.existing, mergeTestCase.latest)
			assert.Equal(t, len(mergeTestCase.expected), len(actual))
			for idx, expectedExternalResource := range mergeTestCase.expected {
				assert.True(t, proto.Equal(expectedExternalResource, actual[idx]))
			}
		})
	}
}

func TestMergeMetadata(t *testing.T) {
	type testCase struct {
		existing *event.TaskExecutionMetadata
		latest   *event.TaskExecutionMetadata
		expected *event.TaskExecutionMetadata
		name     string
	}

	testCases := []testCase{
		{
			existing: nil,
			latest:   nil,
			expected: nil,
			name:     "do nothing",
		},
		{
			existing: &event.TaskExecutionMetadata{},
			latest:   nil,
			expected: &event.TaskExecutionMetadata{},
			name:     "use existing",
		},
		{
			existing: nil,
			latest:   &event.TaskExecutionMetadata{},
			expected: &event.TaskExecutionMetadata{},
			name:     "use latest",
		},
		{
			existing: &event.TaskExecutionMetadata{
				GeneratedName:    "foo",
				ResourcePoolInfo: []*event.ResourcePoolInfo{},
				PluginIdentifier: "bar",
				InstanceClass:    1,
			},
			latest: &event.TaskExecutionMetadata{},
			expected: &event.TaskExecutionMetadata{
				GeneratedName:    "foo",
				ResourcePoolInfo: []*event.ResourcePoolInfo{},
				PluginIdentifier: "bar",
				InstanceClass:    1,
			},
			name: "no updates",
		},
		{
			existing: &event.TaskExecutionMetadata{
				GeneratedName:    "foo",
				ResourcePoolInfo: []*event.ResourcePoolInfo{},
				PluginIdentifier: "bar",
				InstanceClass:    0,
			},
			latest: &event.TaskExecutionMetadata{
				GeneratedName:    "bar",
				ResourcePoolInfo: []*event.ResourcePoolInfo{},
				PluginIdentifier: "foo",
				InstanceClass:    1,
			},
			expected: &event.TaskExecutionMetadata{
				GeneratedName:    "bar",
				ResourcePoolInfo: []*event.ResourcePoolInfo{},
				PluginIdentifier: "foo",
				InstanceClass:    1,
			},
			name: "all updates",
		},
	}

	for _, mergeTestCase := range testCases {
		t.Run(mergeTestCase.name, func(t *testing.T) {
			metadata := mergeMetadata(mergeTestCase.existing, mergeTestCase.latest)
			assert.True(t, proto.Equal(mergeTestCase.expected, metadata))
		})
	}
}

func TestHandleTaskExecutionInputs(t *testing.T) {
	ctx := context.TODO()
	t.Run("no need to update", func(t *testing.T) {
		taskExecutionModel := models.TaskExecution{
			InputURI: testInputURI,
		}
		err := handleTaskExecutionInputs(ctx, &taskExecutionModel, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, taskExecutionModel.InputURI, testInputURI)
	})
	t.Run("read event input data", func(t *testing.T) {
		taskExecutionModel := models.TaskExecution{}
		ds, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)
		err = handleTaskExecutionInputs(ctx, &taskExecutionModel, &admin.TaskExecutionEventRequest{
			Event: &event.TaskExecutionEvent{
				TaskId:                sampleTaskID,
				ParentNodeExecutionId: sampleNodeExecID,
				RetryAttempt:          retryAttemptValue,
				InputValue: &event.TaskExecutionEvent_InputData{
					InputData: testInputs,
				},
			},
		}, ds)
		assert.NoError(t, err)
		expectedOffloadedInputsLocation := "/metadata/project/domain/name/node-id/project/domain/task-id/task-v/1/offloaded_inputs"
		assert.Equal(t, taskExecutionModel.InputURI, expectedOffloadedInputsLocation)
		actualInputs := &core.LiteralMap{}
		err = ds.ReadProtobuf(ctx, storage.DataReference(expectedOffloadedInputsLocation), actualInputs)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(actualInputs, testInputs))
	})
	t.Run("read event input uri", func(t *testing.T) {
		taskExecutionModel := models.TaskExecution{}
		err := handleTaskExecutionInputs(ctx, &taskExecutionModel, &admin.TaskExecutionEventRequest{
			Event: &event.TaskExecutionEvent{
				TaskId:                sampleTaskID,
				ParentNodeExecutionId: sampleNodeExecID,
				RetryAttempt:          retryAttemptValue,
				InputValue: &event.TaskExecutionEvent_InputUri{
					InputUri: testInputURI,
				},
			},
		}, nil)
		assert.NoError(t, err)
		assert.Equal(t, taskExecutionModel.InputURI, testInputURI)
	})
	t.Run("request contained no input data", func(t *testing.T) {
		taskExecutionModel := models.TaskExecution{
			InputURI: testInputURI,
		}
		err := handleTaskExecutionInputs(ctx, &taskExecutionModel, &admin.TaskExecutionEventRequest{
			Event: &event.TaskExecutionEvent{
				TaskId:                sampleTaskID,
				ParentNodeExecutionId: sampleNodeExecID,
				RetryAttempt:          retryAttemptValue,
			},
		}, nil)
		assert.NoError(t, err)
		assert.Equal(t, taskExecutionModel.InputURI, testInputURI)
	})
}
