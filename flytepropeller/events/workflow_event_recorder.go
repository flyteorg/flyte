package events

import (
	"context"
	"strings"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate mockery -all -output=mocks -case=underscore

// Recorder for Workflow events
type WorkflowEventRecorder interface {
	// Records workflow execution events indicating the workflow has undergone a phase change and additional metadata.
	RecordWorkflowEvent(ctx context.Context, event *event.WorkflowExecutionEvent, eventConfig *config.EventConfig) error
}

type workflowEventRecorder struct {
	eventRecorder EventRecorder
	store         *storage.DataStore
}

// In certain cases, a successful workflow execution event can be configured to include raw output data inline. However,
// for large outputs these events may exceed the event recipient's message size limit, so we fallback to passing
// the offloaded output URI instead.
func (r *workflowEventRecorder) handleFailure(ctx context.Context, ev *event.WorkflowExecutionEvent, err error) error {
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.ResourceExhausted {
		// Error was not a status error
		return err
	}
	if !strings.HasPrefix(st.Message(), "message too large") {
		return err
	}

	// This time, we attempt to record the event with the output URI set.
	return r.eventRecorder.RecordWorkflowEvent(ctx, ev)
}

func (r *workflowEventRecorder) RecordWorkflowEvent(ctx context.Context, ev *event.WorkflowExecutionEvent, eventConfig *config.EventConfig) error {
	var origEvent = ev
	var rawOutputPolicy = eventConfig.RawOutputPolicy
	if rawOutputPolicy == config.RawOutputPolicyInline && len(ev.GetOutputUri()) > 0 {
		outputs := &core.LiteralMap{}
		err := r.store.ReadProtobuf(ctx, storage.DataReference(ev.GetOutputUri()), outputs)
		if err != nil {
			// Fall back to forwarding along outputs by reference when we can't fetch them.
			logger.Warnf(ctx, "failed to fetch outputs by ref [%s] to send inline with err: %v", ev.GetOutputUri(), err)
			rawOutputPolicy = config.RawOutputPolicyReference
		} else {
			origEvent = proto.Clone(ev).(*event.WorkflowExecutionEvent)
			ev.OutputResult = &event.WorkflowExecutionEvent_OutputData{
				OutputData: outputs,
			}
		}
	}

	err := r.eventRecorder.RecordWorkflowEvent(ctx, ev)
	if err != nil {
		logger.Infof(ctx, "Failed to record workflow event [%+v] with err: %v", ev, err)
		// Only attempt to retry sending an event in the case we tried to send raw output data inline
		if eventConfig.FallbackToOutputReference && rawOutputPolicy == config.RawOutputPolicyInline {
			logger.Infof(ctx, "Falling back to sending workflow event outputs by reference for [%+v]", ev.ExecutionId)
			return r.handleFailure(ctx, origEvent, err)
		}
		return err
	}
	return nil
}

func NewWorkflowEventRecorder(eventSink EventSink, scope promutils.Scope, store *storage.DataStore) WorkflowEventRecorder {
	return &workflowEventRecorder{
		eventRecorder: NewEventRecorder(eventSink, scope),
		store:         store,
	}
}
