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

// NodeEventRecorder records Node events
type NodeEventRecorder interface {
	// RecordNodeEvent records execution events indicating the node has undergone a phase change and additional metadata.
	RecordNodeEvent(ctx context.Context, event *event.NodeExecutionEvent, eventConfig *config.EventConfig) error
}

type nodeEventRecorder struct {
	eventRecorder EventRecorder
	store         *storage.DataStore
}

// In certain cases, a successful node execution event can be configured to include raw output data inline. However,
// for large outputs these events may exceed the event recipient's message size limit, so we fallback to passing
// the offloaded output URI instead.
func (r *nodeEventRecorder) handleFailure(ctx context.Context, ev *event.NodeExecutionEvent, err error) error {
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.ResourceExhausted {
		// Error was not a status error
		return err
	}
	if !strings.HasPrefix(st.Message(), "message too large") {
		return err
	}

	// This time, we attempt to record the event with the output URI set.
	return r.eventRecorder.RecordNodeEvent(ctx, ev)
}

func (r *nodeEventRecorder) RecordNodeEvent(ctx context.Context, ev *event.NodeExecutionEvent, eventConfig *config.EventConfig) error {
	var origEvent = ev
	var rawOutputPolicy = eventConfig.RawOutputPolicy
	if rawOutputPolicy == config.RawOutputPolicyInline && len(ev.GetOutputUri()) > 0 {
		outputs := &core.OutputData{}
		outputLit := &core.LiteralMap{}
		msgIndex, err := r.store.ReadProtobufAny(ctx, storage.DataReference(ev.GetOutputUri()), outputs, outputLit)
		if err != nil {
			// Fall back to forwarding along outputs by reference when we can't fetch them.
			logger.Warnf(ctx, "failed to fetch outputs by ref [%s] to send inline with err: %v", ev.GetOutputUri(), err)
			rawOutputPolicy = config.RawOutputPolicyReference
		} else if ev.GetEventVersion() < int32(v1alpha1.EventVersion3) {
			origEvent = proto.Clone(ev).(*event.NodeExecutionEvent)
			// Admin is not updated yet, send the old format. Set literal maps.

			if msgIndex == 0 { // OutputData
				outputLit = outputs.Outputs
			}

			ev.OutputResult = &event.NodeExecutionEvent_DeprecatedOutputData{
				DeprecatedOutputData: outputLit,
			}
		} else {
			origEvent = proto.Clone(ev).(*event.NodeExecutionEvent)
			// Use OutputData
			if msgIndex == 1 { // LiteralMap
				outputs = &core.OutputData{
					Outputs: outputLit,
				}
			}

			ev.OutputResult = &event.NodeExecutionEvent_OutputData{
				OutputData: outputs,
			}
		}
	}

	err := r.eventRecorder.RecordNodeEvent(ctx, ev)
	if err != nil {
		logger.Infof(ctx, "Failed to record node event [%+v] with err: %v", ev, err)
		// Only attempt to retry sending an event in the case we tried to send raw output data inline
		if eventConfig.FallbackToOutputReference && rawOutputPolicy == config.RawOutputPolicyInline {
			logger.Infof(ctx, "Falling back to sending node event outputs by reference for [%+v]", ev.Id)
			return r.handleFailure(ctx, origEvent, err)
		}
		return err
	}
	return nil
}

func NewNodeEventRecorder(eventSink EventSink, scope promutils.Scope, store *storage.DataStore) NodeEventRecorder {
	return &nodeEventRecorder{
		eventRecorder: NewEventRecorder(eventSink, scope),
		store:         store,
	}
}
