package events

import (
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/lyft/flyteplugins/go/tasks/v1/errors"

	"github.com/lyft/flyteplugins/go/tasks/v1/types"
	"github.com/lyft/flyteplugins/go/tasks/v1/types/mocks"
	"github.com/lyft/flytestdlib/storage"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

const inputs = storage.DataReference("inputs.pb")
const outputs = storage.DataReference("outputs.pb")
const errorsFile = storage.DataReference("errorsFile.pb")

var testTaskExecutionIdentifier = core.TaskExecutionIdentifier{
	TaskId: &core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "proj",
		Domain:       "domain",
		Name:         "name",
	},
	NodeExecutionId: &core.NodeExecutionIdentifier{
		NodeId: "nodeId",
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "proj",
			Domain:  "domain",
			Name:    "name",
		},
	},
}

type mockTaskExecutionIdentifier struct{}

func (m mockTaskExecutionIdentifier) GetID() core.TaskExecutionIdentifier {
	return testTaskExecutionIdentifier
}

func (m mockTaskExecutionIdentifier) GetGeneratedName() string {
	return "task-exec-name"
}

func TestEventsPublisher_Queued(t *testing.T) {
	startedAt := time.Now()
	taskCtx := &mocks.TaskContext{}
	taskCtx.On("GetTaskExecutionID").Return(mockTaskExecutionIdentifier{})
	taskCtx.On("GetInputsFile").Return(inputs)

	e := CreateEvent(taskCtx, types.TaskStatusQueued, nil)
	assert.Equal(t, e.GetPhase(), core.TaskExecution_QUEUED)
	assert.Equal(t, e.GetTaskId(), testTaskExecutionIdentifier.TaskId)
	assert.Equal(t, e.GetParentNodeExecutionId(), testTaskExecutionIdentifier.NodeExecutionId)
	assert.Equal(t, e.GetInputUri(), string(inputs))
	assert.WithinDuration(t, startedAt, time.Now(), time.Millisecond*5)
}

func TestEventsPublisher_Running(t *testing.T) {
	startedAt := time.Now()

	taskCtx := &mocks.TaskContext{}
	taskCtx.On("GetTaskExecutionID").Return(mockTaskExecutionIdentifier{})
	taskCtx.On("GetInputsFile").Return(inputs)

	e := CreateEvent(taskCtx, types.TaskStatusRunning, nil)
	assert.Equal(t, e.GetPhase(), core.TaskExecution_RUNNING)
	assert.Equal(t, e.GetTaskId(), testTaskExecutionIdentifier.TaskId)
	assert.Equal(t, e.GetParentNodeExecutionId(), testTaskExecutionIdentifier.NodeExecutionId)
	assert.Equal(t, e.GetInputUri(), string(inputs))
	assert.WithinDuration(t, startedAt, time.Now(), time.Millisecond*5)
}

func TestEventsPublisher_Success(t *testing.T) {
	startedAt := time.Now()

	taskCtx := &mocks.TaskContext{}
	taskCtx.On("GetTaskExecutionID").Return(mockTaskExecutionIdentifier{})
	taskCtx.On("GetInputsFile").Return(inputs)
	taskCtx.On("GetOutputsFile").Return(outputs)

	e := CreateEvent(taskCtx, types.TaskStatusSucceeded, nil)
	assert.Equal(t, e.GetPhase(), core.TaskExecution_SUCCEEDED)
	assert.Equal(t, e.GetTaskId(), testTaskExecutionIdentifier.TaskId)
	assert.Equal(t, e.GetParentNodeExecutionId(), testTaskExecutionIdentifier.NodeExecutionId)
	assert.Equal(t, e.GetOutputUri(), string(outputs))
	assert.WithinDuration(t, startedAt, time.Now(), time.Millisecond*5)
}

func TestEventsPublisher_PermanentFailed(t *testing.T) {
	startedAt := time.Now()

	taskCtx := &mocks.TaskContext{}
	taskCtx.On("GetTaskExecutionID").Return(mockTaskExecutionIdentifier{})
	taskCtx.On("GetInputsFile").Return(inputs)
	taskCtx.On("GetOutputsFile").Return(outputs)
	taskCtx.On("GetErrorFile").Return(errorsFile)

	err := errors.Errorf("test", "failed")
	e := CreateEvent(taskCtx, types.TaskStatusPermanentFailure(err), nil)
	assert.Equal(t, e.GetPhase(), core.TaskExecution_FAILED)
	assert.Equal(t, e.GetTaskId(), testTaskExecutionIdentifier.TaskId)
	assert.Equal(t, e.GetParentNodeExecutionId(), testTaskExecutionIdentifier.NodeExecutionId)
	assert.NotNil(t, e.GetError())
	assert.Equal(t, e.GetError().Code, "test")
	assert.Equal(t, e.GetError().Message, "task failed, test: failed")
	assert.Equal(t, e.GetError().GetErrorUri(), string(errorsFile))
	assert.WithinDuration(t, startedAt, time.Now(), time.Millisecond*5)
}

func TestEventsPublisher_RetryableFailed(t *testing.T) {
	startedAt := time.Now()

	taskCtx := &mocks.TaskContext{}
	taskCtx.On("GetTaskExecutionID").Return(mockTaskExecutionIdentifier{})
	taskCtx.On("GetInputsFile").Return(inputs)
	taskCtx.On("GetOutputsFile").Return(outputs)
	taskCtx.On("GetErrorFile").Return(errorsFile)

	err := errors.Errorf("test", "failed")
	e := CreateEvent(taskCtx, types.TaskStatusRetryableFailure(err), nil)
	assert.Equal(t, e.GetPhase(), core.TaskExecution_FAILED)
	assert.Equal(t, e.GetTaskId(), testTaskExecutionIdentifier.TaskId)
	assert.Equal(t, e.GetParentNodeExecutionId(), testTaskExecutionIdentifier.NodeExecutionId)
	assert.NotNil(t, e.GetError())
	assert.Equal(t, e.GetError().Code, "test")
	assert.Equal(t, e.GetError().Message, "task failed, test: failed")
	assert.Equal(t, e.GetError().GetErrorUri(), string(errorsFile))
	assert.WithinDuration(t, startedAt, time.Now(), time.Millisecond*5)
}

func TestEventsPublisher_FailedNilError(t *testing.T) {
	startedAt := time.Now()

	taskCtx := &mocks.TaskContext{}
	taskCtx.On("GetTaskExecutionID").Return(mockTaskExecutionIdentifier{})
	taskCtx.On("GetInputsFile").Return(inputs)
	taskCtx.On("GetOutputsFile").Return(outputs)
	taskCtx.On("GetErrorFile").Return(errorsFile)

	e := CreateEvent(taskCtx, types.TaskStatusRetryableFailure(nil), nil)
	assert.Equal(t, e.GetPhase(), core.TaskExecution_FAILED)
	assert.Equal(t, e.GetTaskId(), testTaskExecutionIdentifier.TaskId)
	assert.Equal(t, e.GetParentNodeExecutionId(), testTaskExecutionIdentifier.NodeExecutionId)
	assert.NotNil(t, e.GetError())
	assert.Equal(t, e.GetError().Code, "UnknownTaskError")
	assert.Equal(t, e.GetError().Message, "unknown reason")
	assert.Equal(t, e.GetError().GetErrorUri(), string(errorsFile))
	assert.WithinDuration(t, startedAt, time.Now(), time.Millisecond*5)
}

func TestEventsPublisher_WithCustomInfo(t *testing.T) {
	startedAt := time.Now()
	taskCtx := &mocks.TaskContext{}
	taskCtx.On("GetTaskExecutionID").Return(mockTaskExecutionIdentifier{})
	taskCtx.On("GetInputsFile").Return(inputs)

	t.Run("emptyInfo", func(t *testing.T) {
		e := CreateEvent(taskCtx, types.TaskStatusQueued, &TaskEventInfo{})
		assert.Equal(t, e.GetPhase(), core.TaskExecution_QUEUED)
		assert.Equal(t, e.GetTaskId(), testTaskExecutionIdentifier.TaskId)
		assert.Equal(t, e.GetParentNodeExecutionId(), testTaskExecutionIdentifier.NodeExecutionId)
		assert.Equal(t, e.GetInputUri(), string(inputs))
		assert.WithinDuration(t, startedAt, time.Now(), time.Millisecond*5)
	})

	t.Run("withInfo", func(t *testing.T) {
		n := time.Now()
		s := structpb.Struct{}
		e := CreateEvent(taskCtx, types.TaskStatusQueued, &TaskEventInfo{
			Logs:       []*core.TaskLog{{Uri: "l1"}, {Uri: "l2"}},
			OccurredAt: &n,
			CustomInfo: &s,
		})
		assert.Equal(t, e.GetPhase(), core.TaskExecution_QUEUED)
		assert.Equal(t, e.GetTaskId(), testTaskExecutionIdentifier.TaskId)
		assert.Equal(t, e.GetParentNodeExecutionId(), testTaskExecutionIdentifier.NodeExecutionId)
		assert.Equal(t, e.GetInputUri(), string(inputs))
		o, err := ptypes.Timestamp(e.OccurredAt)
		assert.NoError(t, err)
		assert.Equal(t, n.Unix(), o.Unix())
		assert.Equal(t, len(e.Logs), 2)
		assert.Equal(t, e.Logs[0].Uri, "l1")
		assert.Equal(t, e.Logs[1].Uri, "l2")
		assert.Equal(t, e.CustomInfo, &s)

	})
}
