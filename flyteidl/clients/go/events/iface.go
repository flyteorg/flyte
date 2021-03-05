package events

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
)

// Recorder for Workflow events
type WorkflowEventRecorder interface {
	RecordWorkflowEvent(ctx context.Context, event *event.WorkflowExecutionEvent) error
}

// Recorder for Node events
type NodeEventRecorder interface {
	RecordNodeEvent(ctx context.Context, event *event.NodeExecutionEvent) error
}

// Recorder for Task events
type TaskEventRecorder interface {
	RecordTaskEvent(ctx context.Context, event *event.TaskExecutionEvent) error
}
