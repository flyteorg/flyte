package nodes

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flytepropeller/events"
	eventsErr "github.com/flyteorg/flytepropeller/events/errors"
	"github.com/flyteorg/flytepropeller/pkg/controller/config"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
)

type fakeTaskEventsRecorder struct {
	err error
}

func (f fakeTaskEventsRecorder) RecordTaskEvent(ctx context.Context, event *event.TaskExecutionEvent, eventConfig *config.EventConfig) error {
	if f.err != nil {
		return f.err
	}
	return nil
}

func Test_taskEventRecorder_RecordTaskEvent(t1 *testing.T) {
	noErrRecorder := fakeTaskEventsRecorder{}
	alreadyExistsError := fakeTaskEventsRecorder{&eventsErr.EventError{Code: eventsErr.AlreadyExists, Cause: fmt.Errorf("err")}}
	inTerminalError := fakeTaskEventsRecorder{&eventsErr.EventError{Code: eventsErr.EventAlreadyInTerminalStateError, Cause: fmt.Errorf("err")}}
	otherError := fakeTaskEventsRecorder{&eventsErr.EventError{Code: eventsErr.ResourceExhausted, Cause: fmt.Errorf("err")}}

	tests := []struct {
		name    string
		rec     events.TaskEventRecorder
		p       core.TaskExecution_Phase
		wantErr bool
	}{
		{"aborted-success", noErrRecorder, core.TaskExecution_ABORTED, false},
		{"aborted-failure", otherError, core.TaskExecution_ABORTED, true},
		{"aborted-already", alreadyExistsError, core.TaskExecution_ABORTED, false},
		{"aborted-terminal", inTerminalError, core.TaskExecution_ABORTED, false},
		{"running-terminal", inTerminalError, core.TaskExecution_RUNNING, true},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := taskEventRecorder{
				TaskEventRecorder: tt.rec,
			}
			ev := &event.TaskExecutionEvent{
				Phase: tt.p,
			}
			if err := t.RecordTaskEvent(context.TODO(), ev, &config.EventConfig{
				RawOutputPolicy: config.RawOutputPolicyReference,
			}); (err != nil) != tt.wantErr {
				t1.Errorf("RecordTaskEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
