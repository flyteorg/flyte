package events

import (
	"context"
	"strings"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

//go:generate mockery -all -output=mocks -case=underscore

// Recorder for Task events
type TaskEventRecorder interface {
	// Records task execution events indicating the task has undergone a phase change and additional metadata.
	RecordTaskEvent(ctx context.Context, event *event.TaskExecutionEvent, eventConfig *config.EventConfig) error
}

type taskEventRecorder struct {
	eventRecorder EventRecorder
	store         *storage.DataStore
}

// In certain cases, a successful task execution event can be configured to include raw output data inline. However,
// for large outputs these events may exceed the event recipient's message size limit, so we fallback to passing
// the offloaded output URI instead.
func (r *taskEventRecorder) handleFailure(ctx context.Context, ev *event.TaskExecutionEvent, err error) error {
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.ResourceExhausted {
		// Error was not a status error
		return err
	}
	if !strings.HasPrefix(st.Message(), "message too large") {
		return err
	}

	// This time, we attempt to record the event with the output URI set.
	return r.eventRecorder.RecordTaskEvent(ctx, ev)
}

func (r *taskEventRecorder) RecordTaskEvent(ctx context.Context, ev *event.TaskExecutionEvent, eventConfig *config.EventConfig) error {
	var origEvent = ev
	var rawOutputPolicy = eventConfig.RawOutputPolicy
	if rawOutputPolicy == config.RawOutputPolicyInline && len(ev.GetOutputUri()) > 0 {
		outputs := &core.OutputData{}
		outputsLit := &core.LiteralMap{}
		msgIndex, err := r.store.ReadProtobufAny(ctx, storage.DataReference(ev.GetOutputUri()), outputs, outputsLit)
		if err != nil {
			// Fall back to forwarding along outputs by reference when we can't fetch them.
			logger.Warnf(ctx, "failed to fetch outputs by ref [%s] to send inline with err: %v", ev.GetOutputUri(), err)
			rawOutputPolicy = config.RawOutputPolicyReference
		} else if ev.EventVersion < int32(v1alpha1.EventVersion3) {
			origEvent = proto.Clone(ev).(*event.TaskExecutionEvent)
			if msgIndex == 0 { // OutputData
				outputsLit = outputs.Outputs
			}

			ev.OutputResult = &event.TaskExecutionEvent_DeprecatedOutputData{
				DeprecatedOutputData: outputsLit,
			}
		} else {
			origEvent = proto.Clone(ev).(*event.TaskExecutionEvent)
			if msgIndex == 1 { // LiteralMap
				outputs = &core.OutputData{
					Outputs: outputsLit,
				}
			}

			ev.OutputResult = &event.TaskExecutionEvent_OutputData{
				OutputData: outputs,
			}
		}
	}

	err := r.eventRecorder.RecordTaskEvent(ctx, ev)
	if err != nil {
		logger.Infof(ctx, "Failed to record task event [%+v] with err: %v", ev, err)
		// Only attempt to retry sending an event in the case we tried to send raw output data inline
		if eventConfig.FallbackToOutputReference && rawOutputPolicy == config.RawOutputPolicyInline {
			logger.Infof(ctx, "Falling back to sending task event outputs by reference for [%+v]", ev.TaskId)
			return r.handleFailure(ctx, origEvent, err)
		}
		return err
	}
	return nil
}

func NewTaskEventRecorder(eventSink EventSink, scope promutils.Scope, store *storage.DataStore) TaskEventRecorder {
	return &taskEventRecorder{
		eventRecorder: NewEventRecorder(eventSink, scope),
		store:         store,
	}
}
