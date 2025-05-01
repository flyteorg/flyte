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

func getReferenceNodeEv() *event.NodeExecutionEvent {
	return &event.NodeExecutionEvent{
		Id: nodeExecID,
		OutputResult: &event.NodeExecutionEvent_OutputUri{
			OutputUri: referenceURI,
		},
		DeckUri: deckURI,
	}
}

func getRawOutputNodeEv() *event.NodeExecutionEvent {
	return &event.NodeExecutionEvent{
		Id: nodeExecID,
		OutputResult: &event.NodeExecutionEvent_OutputData{
			OutputData: outputData,
		},
		DeckUri: deckURI,
	}
}

func TestRecordNodeEvent_Success_ReferenceOutputs(t *testing.T) {
	ctx := context.TODO()
	eventRecorder := mocks.EventRecorder{}
	eventRecorder.EXPECT().RecordNodeEvent(ctx, mock.MatchedBy(func(event *event.NodeExecutionEvent) bool {
		assert.True(t, proto.Equal(event, getReferenceNodeEv()))
		return true
	})).Return(nil)
	mockStore := &storage.DataStore{
		ComposedProtobufStore: &storageMocks.ComposedProtobufStore{},
		ReferenceConstructor:  &storageMocks.ReferenceConstructor{},
	}

	recorder := &nodeEventRecorder{
		eventRecorder: &eventRecorder,
		store:         mockStore,
	}
	err := recorder.RecordNodeEvent(ctx, getReferenceNodeEv(), referenceEventConfig)
	assert.NoError(t, err)
}

func TestRecordNodeEvent_Success_InlineOutputs(t *testing.T) {
	ctx := context.TODO()
	eventRecorder := mocks.EventRecorder{}
	eventRecorder.EXPECT().RecordNodeEvent(ctx, mock.MatchedBy(func(event *event.NodeExecutionEvent) bool {
		assert.True(t, proto.Equal(event, getRawOutputNodeEv()))
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

	recorder := &nodeEventRecorder{
		eventRecorder: &eventRecorder,
		store:         mockStore,
	}
	err := recorder.RecordNodeEvent(ctx, getReferenceNodeEv(), inlineEventConfig)
	assert.Equal(t, deckURI, nodeEvent.GetDeckUri())
	assert.NoError(t, err)
}

func TestRecordNodeEvent_Failure_FetchInlineOutputs(t *testing.T) {
	ctx := context.TODO()
	eventRecorder := mocks.EventRecorder{}
	eventRecorder.EXPECT().RecordNodeEvent(ctx, mock.MatchedBy(func(event *event.NodeExecutionEvent) bool {
		assert.True(t, proto.Equal(event, getReferenceNodeEv()))
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

	recorder := &nodeEventRecorder{
		eventRecorder: &eventRecorder,
		store:         mockStore,
	}
	err := recorder.RecordNodeEvent(ctx, getReferenceNodeEv(), inlineEventConfig)
	assert.NoError(t, err)
}

func TestRecordNodeEvent_Failure_FallbackReference_Retry(t *testing.T) {
	ctx := context.TODO()
	eventRecorder := mocks.EventRecorder{}
	eventRecorder.EXPECT().RecordNodeEvent(ctx, mock.MatchedBy(func(event *event.NodeExecutionEvent) bool {
		return event.GetOutputData() != nil
	})).Return(status.Error(codes.ResourceExhausted, "message too large"))
	eventRecorder.EXPECT().RecordNodeEvent(ctx, mock.MatchedBy(func(event *event.NodeExecutionEvent) bool {
		return event.GetOutputData() == nil && proto.Equal(event, getReferenceNodeEv())
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

	recorder := &nodeEventRecorder{
		eventRecorder: &eventRecorder,
		store:         mockStore,
	}
	err := recorder.RecordNodeEvent(ctx, getReferenceNodeEv(), inlineEventConfigFallback)
	assert.NoError(t, err)
}

func TestRecordNodeEvent_Failure_FallbackReference_Unretriable(t *testing.T) {
	ctx := context.TODO()
	eventRecorder := mocks.EventRecorder{}
	eventRecorder.EXPECT().RecordNodeEvent(ctx, mock.Anything).Return(errors.New("foo"))
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

	recorder := &nodeEventRecorder{
		eventRecorder: &eventRecorder,
		store:         mockStore,
	}
	err := recorder.RecordNodeEvent(ctx, getReferenceNodeEv(), inlineEventConfigFallback)
	assert.EqualError(t, err, "foo")
}
