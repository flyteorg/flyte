package utils

import (
	"context"

	"github.com/lyft/flyteidl/clients/go/events"
	eventsErr "github.com/lyft/flyteidl/clients/go/events/errors"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/lyft/flytestdlib/logger"
)

// Construct task event recorder to pass down to plugin. This is a just a wrapper around the normal
// taskEventRecorder that can encapsulate logic to validate and handle errors.
func NewPluginTaskEventRecorder(taskEventRecorder events.TaskEventRecorder) events.TaskEventRecorder {
	return &pluginTaskEventRecorder{
		taskEventRecorder: taskEventRecorder,
	}
}

type pluginTaskEventRecorder struct {
	taskEventRecorder events.TaskEventRecorder
}

func (r pluginTaskEventRecorder) RecordTaskEvent(ctx context.Context, event *event.TaskExecutionEvent) error {
	err := r.taskEventRecorder.RecordTaskEvent(ctx, event)
	if err != nil && eventsErr.IsAlreadyExists(err) {
		logger.Infof(ctx, "Task event phase: %s, taskId %s, retry attempt %d - already exists",
			event.Phase.String(), event.GetTaskId(), event.RetryAttempt)
		return nil
	}
	return err
}
