package events

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
)

type MockRecorder struct {
	RecordNodeEventCb     func(ctx context.Context, event *event.NodeExecutionEvent) error
	RecordWorkflowEventCb func(ctx context.Context, event *event.WorkflowExecutionEvent) error
}

func (m *MockRecorder) RecordNodeEvent(ctx context.Context, event *event.NodeExecutionEvent) error {
	if m.RecordNodeEventCb != nil {
		return m.RecordNodeEventCb(ctx, event)
	}

	return nil
}

func (m *MockRecorder) RecordWorkflowEvent(ctx context.Context, event *event.WorkflowExecutionEvent) error {
	if m.RecordWorkflowEventCb != nil {
		return m.RecordWorkflowEventCb(ctx, event)
	}

	return nil
}

func NewMock() NodeEventRecorder {
	return &MockRecorder{
		RecordNodeEventCb: func(ctx context.Context, event *event.NodeExecutionEvent) error {
			return nil
		},
	}
}

func NewMockWorkflowRecorder() WorkflowEventRecorder {
	return &MockRecorder{
		RecordWorkflowEventCb: func(ctx context.Context, event *event.WorkflowExecutionEvent) error {
			return nil
		},
	}
}
