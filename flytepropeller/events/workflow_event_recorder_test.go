package events

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/runtime/protoiface"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flytepropeller/events/mocks"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	storageMocks "github.com/flyteorg/flyte/flytestdlib/storage/mocks"
)

func getReferenceWorkflowEv() *event.WorkflowExecutionEvent {
	return &event.WorkflowExecutionEvent{
		ExecutionId: workflowExecID,
		OutputResult: &event.WorkflowExecutionEvent_OutputUri{
			OutputUri: referenceURI,
		},
	}
}

func getRawOutputWorkflowEv() *event.WorkflowExecutionEvent {
	return &event.WorkflowExecutionEvent{
		ExecutionId: workflowExecID,
		OutputResult: &event.WorkflowExecutionEvent_OutputData{
			OutputData: outputData,
		},
	}
}

func TestRecordWorkflowEvent_Success_ReferenceOutputs(t *testing.T) {
	ctx := context.TODO()
	eventRecorder := mocks.EventRecorder{}
	eventRecorder.OnRecordWorkflowEventMatch(ctx, mock.MatchedBy(func(event *event.WorkflowExecutionEvent) bool {
		assert.True(t, proto.Equal(event, getReferenceWorkflowEv()))
		return true
	})).Return(nil)
	mockStore := &storage.DataStore{
		ComposedProtobufStore: &storageMocks.ComposedProtobufStore{},
		ReferenceConstructor:  &storageMocks.ReferenceConstructor{},
	}

	recorder := &workflowEventRecorder{
		eventRecorder: &eventRecorder,
		store:         mockStore,
	}
	err := recorder.RecordWorkflowEvent(ctx, getReferenceWorkflowEv(), referenceEventConfig)
	assert.NoError(t, err)
}

func TestRecordWorkflowEvent_Success_InlineOutputs(t *testing.T) {
	ctx := context.TODO()
	eventRecorder := mocks.EventRecorder{}
	eventRecorder.OnRecordWorkflowEventMatch(ctx, mock.MatchedBy(func(event *event.WorkflowExecutionEvent) bool {
		assert.True(t, proto.Equal(event, getRawOutputWorkflowEv()))
		return true
	})).Return(nil)
	pbStore := &storageMocks.ComposedProtobufStore{}
	pbStore.EXPECT().ReadProtobuf(mock.Anything, mock.MatchedBy(func(ref storage.DataReference) bool {
		return ref.String() == referenceURI
	}), mock.Anything).Return(nil).Run(func(ctx context.Context, reference storage.DataReference, msg protoiface.MessageV1) {
		arg := msg.(*core.LiteralMap)
		*arg = *outputData
	})
	mockStore := &storage.DataStore{
		ComposedProtobufStore: pbStore,
		ReferenceConstructor:  &storageMocks.ReferenceConstructor{},
	}

	recorder := &workflowEventRecorder{
		eventRecorder: &eventRecorder,
		store:         mockStore,
	}
	err := recorder.RecordWorkflowEvent(ctx, getReferenceWorkflowEv(), inlineEventConfig)
	assert.NoError(t, err)
}

func TestRecordWorkflowEvent_Failure_FetchInlineOutputs(t *testing.T) {
	ctx := context.TODO()
	eventRecorder := mocks.EventRecorder{}
	eventRecorder.OnRecordWorkflowEventMatch(ctx, mock.MatchedBy(func(event *event.WorkflowExecutionEvent) bool {
		assert.True(t, proto.Equal(event, getReferenceWorkflowEv()))
		return true
	})).Return(nil)
	pbStore := &storageMocks.ComposedProtobufStore{}
	pbStore.EXPECT().ReadProtobuf(mock.Anything, mock.MatchedBy(func(ref storage.DataReference) bool {
		return ref.String() == referenceURI
	}), mock.Anything).Return(errors.New("foo"))
	mockStore := &storage.DataStore{
		ComposedProtobufStore: pbStore,
		ReferenceConstructor:  &storageMocks.ReferenceConstructor{},
	}

	recorder := &workflowEventRecorder{
		eventRecorder: &eventRecorder,
		store:         mockStore,
	}
	err := recorder.RecordWorkflowEvent(ctx, getReferenceWorkflowEv(), inlineEventConfig)
	assert.NoError(t, err)
}

func TestRecordWorkflowEvent_Failure_FallbackReference_Retry(t *testing.T) {
	ctx := context.TODO()
	eventRecorder := mocks.EventRecorder{}
	eventRecorder.OnRecordWorkflowEventMatch(ctx, mock.MatchedBy(func(event *event.WorkflowExecutionEvent) bool {
		return event.GetOutputData() != nil
	})).Return(status.Error(codes.ResourceExhausted, "message too large"))
	eventRecorder.OnRecordWorkflowEventMatch(ctx, mock.MatchedBy(func(event *event.WorkflowExecutionEvent) bool {
		return event.GetOutputData() == nil && proto.Equal(event, getReferenceWorkflowEv())
	})).Return(nil)
	pbStore := &storageMocks.ComposedProtobufStore{}
	pbStore.EXPECT().ReadProtobuf(mock.Anything, mock.MatchedBy(func(ref storage.DataReference) bool {
		return ref.String() == referenceURI
	}), mock.Anything).Return(nil).Run(func(ctx context.Context, reference storage.DataReference, msg protoiface.MessageV1) {
		arg := msg.(*core.LiteralMap)
		*arg = *outputData
	})
	mockStore := &storage.DataStore{
		ComposedProtobufStore: pbStore,
		ReferenceConstructor:  &storageMocks.ReferenceConstructor{},
	}

	recorder := &workflowEventRecorder{
		eventRecorder: &eventRecorder,
		store:         mockStore,
	}
	err := recorder.RecordWorkflowEvent(ctx, getReferenceWorkflowEv(), inlineEventConfigFallback)
	assert.NoError(t, err)
}

func TestRecordWorkflowEvent_Failure_FallbackReference_Unretriable(t *testing.T) {
	ctx := context.TODO()
	eventRecorder := mocks.EventRecorder{}
	eventRecorder.OnRecordWorkflowEventMatch(ctx, mock.Anything).Return(errors.New("foo"))
	pbStore := &storageMocks.ComposedProtobufStore{}
	pbStore.EXPECT().ReadProtobuf(mock.Anything, mock.MatchedBy(func(ref storage.DataReference) bool {
		return ref.String() == referenceURI
	}), mock.Anything).Return(nil).Run(func(ctx context.Context, reference storage.DataReference, msg protoiface.MessageV1) {
		arg := msg.(*core.LiteralMap)
		*arg = *outputData
	})
	mockStore := &storage.DataStore{
		ComposedProtobufStore: pbStore,
		ReferenceConstructor:  &storageMocks.ReferenceConstructor{},
	}

	recorder := &workflowEventRecorder{
		eventRecorder: &eventRecorder,
		store:         mockStore,
	}
	err := recorder.RecordWorkflowEvent(context.TODO(), getReferenceWorkflowEv(), inlineEventConfigFallback)
	assert.EqualError(t, err, "foo")
}
