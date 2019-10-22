package transformers

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	ptypesStruct "github.com/golang/protobuf/ptypes/struct"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/stretchr/testify/assert"
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

func TestAddTaskStartedState(t *testing.T) {
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
	err := addTaskTerminalState(&request, &taskExecutionModel, &closure)
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
	err := addTaskTerminalState(&request, &taskExecutionModel, closure)
	assert.Nil(t, err)

	duration, err := ptypes.Duration(closure.GetDuration())
	assert.Nil(t, err)
	assert.EqualValues(t, request.Event.OutputResult, closure.OutputResult)
	assert.EqualValues(t, outputURI, closure.GetOutputUri())
	assert.EqualValues(t, time.Minute, duration)

	assert.Equal(t, time.Minute, taskExecutionModel.Duration)
}

func TestCreateTaskExecutionModelQueued(t *testing.T) {
	taskExecutionModel, err := CreateTaskExecutionModel(CreateTaskExecutionModelInput{
		Request: &admin.TaskExecutionEventRequest{
			Event: &event.TaskExecutionEvent{
				TaskId:                sampleTaskID,
				ParentNodeExecutionId: sampleNodeExecID,
				Phase:                 core.TaskExecution_QUEUED,
				RetryAttempt:          1,
				InputUri:              "input uri",
				OccurredAt:            taskEventOccurredAtProto,
			},
		},
	})
	assert.Nil(t, err)

	expectedClosure := &admin.TaskExecutionClosure{
		Phase:     core.TaskExecution_QUEUED,
		StartedAt: nil,
		CreatedAt: taskEventOccurredAtProto,
		UpdatedAt: taskEventOccurredAtProto,
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
		InputURI:               "input uri",
		Closure:                expectedClosureBytes,
		StartedAt:              nil,
		TaskExecutionCreatedAt: &taskEventOccurredAt,
		TaskExecutionUpdatedAt: &taskEventOccurredAt,
	}, taskExecutionModel)
}

func TestCreateTaskExecutionModelRunning(t *testing.T) {
	taskExecutionModel, err := CreateTaskExecutionModel(CreateTaskExecutionModelInput{
		Request: &admin.TaskExecutionEventRequest{
			Event: &event.TaskExecutionEvent{
				TaskId:                sampleTaskID,
				ParentNodeExecutionId: sampleNodeExecID,
				Phase:                 core.TaskExecution_RUNNING,
				PhaseVersion:          uint32(2),
				RetryAttempt:          1,
				InputUri:              "input uri",
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
		InputURI:               "input uri",
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
			InputUri:              "input uri",
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
			CustomInfo: &customInfo,
		},
	}

	err = UpdateTaskExecutionModel(failedEventRequest, &existingTaskExecution)
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
				Uri: "uri_a",
			},
			{
				Uri: "uri_b",
			},
			{
				Uri: "uri_c",
			},
		},
		CustomInfo: &customInfo,
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
	})
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
	})
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
					Uri:  "uri_a",
					Name: "name_a",
				},
				{
					Uri:  "uri_b",
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
					Uri:  "uri_a",
					Name: "name_a",
				},
				{
					Uri:  "uri_b",
					Name: "name_b",
				},
				{
					Uri:  "uri_c",
					Name: "name_c",
				},
				{
					Uri:  "uri_d",
					Name: "name_d",
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
