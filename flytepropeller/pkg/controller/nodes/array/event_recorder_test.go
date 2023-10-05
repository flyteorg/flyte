package array

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
)

type bufferedEventRecorder struct {
	taskExecutionEvents []*event.TaskExecutionEvent
	nodeExecutionEvents []*event.NodeExecutionEvent
}

func (b *bufferedEventRecorder) RecordTaskEvent(ctx context.Context, taskExecutionEvent *event.TaskExecutionEvent, eventConfig *config.EventConfig) error {
	b.taskExecutionEvents = append(b.taskExecutionEvents, taskExecutionEvent)
	return nil
}

func (b *bufferedEventRecorder) RecordNodeEvent(ctx context.Context, nodeExecutionEvent *event.NodeExecutionEvent, eventConfig *config.EventConfig) error {
	b.nodeExecutionEvents = append(b.nodeExecutionEvents, nodeExecutionEvent)
	return nil
}

func newBufferedEventRecorder() *bufferedEventRecorder {
	return &bufferedEventRecorder{}
}
