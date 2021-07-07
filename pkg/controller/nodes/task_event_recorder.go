package nodes

import (
	"context"

	"github.com/flyteorg/flyteidl/clients/go/events"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/pkg/errors"

	eventsErr "github.com/flyteorg/flyteidl/clients/go/events/errors"
)

type taskEventRecorder struct {
	events.TaskEventRecorder
}

func (t taskEventRecorder) RecordTaskEvent(ctx context.Context, ev *event.TaskExecutionEvent) error {
	if err := t.TaskEventRecorder.RecordTaskEvent(ctx, ev); err != nil {
		if eventsErr.IsAlreadyExists(err) {
			logger.Warningf(ctx, "Failed to record taskEvent, error [%s]. Trying to record state: %s. Ignoring this error!", err.Error(), ev.Phase)
			return nil
		} else if eventsErr.IsEventAlreadyInTerminalStateError(err) {
			if IsTerminalTaskPhase(ev.Phase) {
				// Event is terminal and the stored value in flyteadmin is already terminal. This implies aborted case. So ignoring
				logger.Warningf(ctx, "Failed to record taskEvent, error [%s]. Trying to record state: %s. Ignoring this error!", err.Error(), ev.Phase)
				return nil
			}
			logger.Warningf(ctx, "Failed to record taskEvent in state: %s, error: %s", ev.Phase, err)
			return errors.Wrapf(err, "failed to record task event, as it already exists in terminal state. Event state: %s", ev.Phase)
		}
		return err
	}
	return nil
}
