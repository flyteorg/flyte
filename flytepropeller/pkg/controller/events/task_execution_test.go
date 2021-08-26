package events

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flytestdlib/storage"
	storageMocks "github.com/flyteorg/flytestdlib/storage/mocks"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var taskID = &core.Identifier{
	Project: "p",
	Domain:  "d",
	Name:    "n",
	Version: "v",
}

func getReferenceTaskEv() *event.TaskExecutionEvent {
	return &event.TaskExecutionEvent{
		TaskId:                taskID,
		RetryAttempt:          1,
		ParentNodeExecutionId: nodeExecID,
		OutputResult: &event.TaskExecutionEvent_OutputUri{
			OutputUri: referenceURI,
		},
	}
}

func getRawOutputTaskEv() *event.TaskExecutionEvent {
	return &event.TaskExecutionEvent{
		TaskId:                taskID,
		RetryAttempt:          1,
		ParentNodeExecutionId: nodeExecID,
		OutputResult: &event.TaskExecutionEvent_OutputData{
			OutputData: outputData,
		},
	}
}

// TODO: move this mock definition to flyteidl/events
type mockTaskEventRecorder struct {
	RecordTaskEventCb func(ctx context.Context, event *event.TaskExecutionEvent) error
}

func (r *mockTaskEventRecorder) RecordTaskEvent(ctx context.Context, event *event.TaskExecutionEvent) error {
	if r.RecordTaskEventCb != nil {
		return r.RecordTaskEventCb(ctx, event)
	}
	return nil
}

func TestRecordTaskEvent_Success_ReferenceOutputs(t *testing.T) {
	eventRecorder := mockTaskEventRecorder{}
	eventRecorder.RecordTaskEventCb = func(ctx context.Context, event *event.TaskExecutionEvent) error {
		assert.True(t, proto.Equal(event, getReferenceTaskEv()))
		return nil
	}
	mockStore := &storage.DataStore{
		ComposedProtobufStore: &storageMocks.ComposedProtobufStore{},
		ReferenceConstructor:  &storageMocks.ReferenceConstructor{},
	}

	recorder := &taskEventRecorder{
		eventRecorder: &eventRecorder,
		store:         mockStore,
	}
	err := recorder.RecordTaskEvent(context.TODO(), getReferenceTaskEv(), referenceEventConfig)
	assert.NoError(t, err)
}

func TestRecordTaskEvent_Success_InlineOutputs(t *testing.T) {
	eventRecorder := mockTaskEventRecorder{}
	eventRecorder.RecordTaskEventCb = func(ctx context.Context, event *event.TaskExecutionEvent) error {
		assert.True(t, proto.Equal(event, getRawOutputTaskEv()))
		return nil
	}
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

	recorder := &taskEventRecorder{
		eventRecorder: &eventRecorder,
		store:         mockStore,
	}
	err := recorder.RecordTaskEvent(context.TODO(), getReferenceTaskEv(), inlineEventConfig)
	assert.NoError(t, err)
}

func TestRecordTaskEvent_Failure_FetchInlineOutputs(t *testing.T) {
	eventRecorder := mockTaskEventRecorder{}
	eventRecorder.RecordTaskEventCb = func(ctx context.Context, event *event.TaskExecutionEvent) error {
		assert.True(t, proto.Equal(event, getReferenceTaskEv()))
		return nil
	}
	pbStore := &storageMocks.ComposedProtobufStore{}
	pbStore.OnReadProtobufMatch(mock.Anything, mock.MatchedBy(func(ref storage.DataReference) bool {
		return ref.String() == referenceURI
	}), mock.Anything).Return(errors.New("foo"))
	mockStore := &storage.DataStore{
		ComposedProtobufStore: pbStore,
		ReferenceConstructor:  &storageMocks.ReferenceConstructor{},
	}

	recorder := &taskEventRecorder{
		eventRecorder: &eventRecorder,
		store:         mockStore,
	}
	err := recorder.RecordTaskEvent(context.TODO(), getReferenceTaskEv(), inlineEventConfig)
	assert.NoError(t, err)
}

func TestRecordTaskEvent_Failure_FallbackReference_Retry(t *testing.T) {
	eventRecorder := mockTaskEventRecorder{}
	eventRecorder.RecordTaskEventCb = func(ctx context.Context, event *event.TaskExecutionEvent) error {
		if event.GetOutputData() != nil {
			return status.Error(codes.ResourceExhausted, "message too large")
		}
		assert.True(t, proto.Equal(event, getReferenceTaskEv()))
		return nil
	}
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

	recorder := &taskEventRecorder{
		eventRecorder: &eventRecorder,
		store:         mockStore,
	}
	err := recorder.RecordTaskEvent(context.TODO(), getReferenceTaskEv(), inlineEventConfigFallback)
	assert.NoError(t, err)
}

func TestRecordTaskEvent_Failure_FallbackReference_Unretriable(t *testing.T) {
	eventRecorder := mockTaskEventRecorder{}
	eventRecorder.RecordTaskEventCb = func(ctx context.Context, event *event.TaskExecutionEvent) error {
		return errors.New("foo")
	}
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

	recorder := &taskEventRecorder{
		eventRecorder: &eventRecorder,
		store:         mockStore,
	}
	err := recorder.RecordTaskEvent(context.TODO(), getReferenceTaskEv(), inlineEventConfigFallback)
	assert.EqualError(t, err, "foo")
}
