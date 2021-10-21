package mocks

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
)

type MockRecorder struct {
	RecordNodeEventCb     func(ctx context.Context, event *event.NodeExecutionEvent) error
	RecordTaskEventCb     func(ctx context.Context, event *event.TaskExecutionEvent) error
	RecordWorkflowEventCb func(ctx context.Context, event *event.WorkflowExecutionEvent) error
}

func (m *MockRecorder) RecordNodeEvent(ctx context.Context, event *event.NodeExecutionEvent) error {
	if m.RecordNodeEventCb != nil {
		return m.RecordNodeEventCb(ctx, event)
	}

	return nil
}

func (m *MockRecorder) RecordTaskEvent(ctx context.Context, event *event.TaskExecutionEvent) error {
	if m.RecordTaskEventCb != nil {
		return m.RecordTaskEventCb(ctx, event)
	}

	return nil
}

func (m *MockRecorder) RecordWorkflowEvent(ctx context.Context, event *event.WorkflowExecutionEvent) error {
	if m.RecordWorkflowEventCb != nil {
		return m.RecordWorkflowEventCb(ctx, event)
	}

	return nil
}

func NewMock() *MockRecorder {
	return &MockRecorder{
		RecordNodeEventCb: func(ctx context.Context, event *event.NodeExecutionEvent) error {
			return nil
		},
		RecordTaskEventCb: func(ctx context.Context, event *event.TaskExecutionEvent) error {
			return nil
		},
		RecordWorkflowEventCb: func(ctx context.Context, event *event.WorkflowExecutionEvent) error {
			return nil
		},
	}
}
