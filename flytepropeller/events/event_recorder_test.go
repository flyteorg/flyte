package events

import (
	"context"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flytepropeller/events/mocks"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

var (
	workflowEventError = &event.WorkflowExecutionEvent{
		OutputResult: &event.WorkflowExecutionEvent_Error{
			Error: &core.ExecutionError{
				Message: "error",
			},
		},
	}

	nodeEventError = &event.NodeExecutionEvent{
		OutputResult: &event.NodeExecutionEvent_Error{
			Error: &core.ExecutionError{
				Message: "error",
			},
		},
	}

	taskEventError = &event.TaskExecutionEvent{
		OutputResult: &event.TaskExecutionEvent_Error{
			Error: &core.ExecutionError{
				Message: "error",
			},
		},
	}
)

var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func createRandomString(length int) string {
	b := make([]rune, length)
	for i := range b {
		randomIndex := rand.Intn(len(letter)) //nolint - ignore weak random
		b[i] = letter[randomIndex]
	}
	return string(b)
}

func TestRecordEvent(t *testing.T) {
	ctx := context.Background()
	scope := promutils.NewTestScope()
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey)

	eventSink := mocks.NewMockEventSink()
	eventRecorder := NewEventRecorder(eventSink, scope)

	wfErr := eventRecorder.RecordWorkflowEvent(ctx, wfEvent)
	assert.NoError(t, wfErr)

	nodeErr := eventRecorder.RecordNodeEvent(ctx, nodeEvent)
	assert.NoError(t, nodeErr)

	taskErr := eventRecorder.RecordTaskEvent(ctx, taskEvent)
	assert.NoError(t, taskErr)
}

func TestRecordErrorEvent(t *testing.T) {
	ctx := context.Background()
	scope := promutils.NewTestScope()
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey)

	eventSink := mocks.NewMockEventSink()
	eventRecorder := NewEventRecorder(eventSink, scope)

	wfErr := eventRecorder.RecordWorkflowEvent(ctx, workflowEventError)
	assert.NoError(t, wfErr)

	nodeErr := eventRecorder.RecordNodeEvent(ctx, nodeEventError)
	assert.NoError(t, nodeErr)

	taskErr := eventRecorder.RecordTaskEvent(ctx, taskEventError)
	assert.NoError(t, taskErr)
}

func TestTruncateErrorMessage(t *testing.T) {
	length := 100
	for i := 0; i <= length*2; i += 5 {
		executionError := core.ExecutionError{
			Message: createRandomString(i),
		}

		truncateErrorMessage(&executionError, length)
		assert.True(t, len(executionError.GetMessage()) <= length+len(truncationIndicator)+2)
	}
}
