package events

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flytepropeller/events/mocks"
	"github.com/flyteorg/flytestdlib/storage"
	storageMocks "github.com/flyteorg/flytestdlib/storage/mocks"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	eventRecorder.OnRecordNodeEventMatch(ctx, mock.MatchedBy(func(event *event.NodeExecutionEvent) bool {
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
	eventRecorder.OnRecordNodeEventMatch(ctx, mock.MatchedBy(func(event *event.NodeExecutionEvent) bool {
		assert.True(t, proto.Equal(event, getRawOutputNodeEv()))
		return true
	})).Return(nil)
	pbStore := &storageMocks.ComposedProtobufStore{}
	pbStore.OnReadProtobufMatch(mock.Anything, mock.MatchedBy(func(ref storage.DataReference) bool {
		return ref.String() == referenceURI
	}), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*core.LiteralMap)
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
	assert.Equal(t, deckURI, nodeEvent.DeckUri)
	assert.NoError(t, err)
}

func TestRecordNodeEvent_Failure_FetchInlineOutputs(t *testing.T) {
	ctx := context.TODO()
	eventRecorder := mocks.EventRecorder{}
	eventRecorder.OnRecordNodeEventMatch(ctx, mock.MatchedBy(func(event *event.NodeExecutionEvent) bool {
		assert.True(t, proto.Equal(event, getReferenceNodeEv()))
		return true
	})).Return(nil)
	pbStore := &storageMocks.ComposedProtobufStore{}
	pbStore.OnReadProtobufMatch(mock.Anything, mock.MatchedBy(func(ref storage.DataReference) bool {
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
	eventRecorder.OnRecordNodeEventMatch(ctx, mock.MatchedBy(func(event *event.NodeExecutionEvent) bool {
		return event.GetOutputData() != nil
	})).Return(status.Error(codes.ResourceExhausted, "message too large"))
	eventRecorder.OnRecordNodeEventMatch(ctx, mock.MatchedBy(func(event *event.NodeExecutionEvent) bool {
		return event.GetOutputData() == nil && proto.Equal(event, getReferenceNodeEv())
	})).Return(nil)
	pbStore := &storageMocks.ComposedProtobufStore{}
	pbStore.OnReadProtobufMatch(mock.Anything, mock.MatchedBy(func(ref storage.DataReference) bool {
		return ref.String() == referenceURI
	}), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*core.LiteralMap)
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
	eventRecorder.OnRecordNodeEventMatch(ctx, mock.Anything).Return(errors.New("foo"))
	pbStore := &storageMocks.ComposedProtobufStore{}
	pbStore.OnReadProtobufMatch(mock.Anything, mock.MatchedBy(func(ref storage.DataReference) bool {
		return ref.String() == referenceURI
	}), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*core.LiteralMap)
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
