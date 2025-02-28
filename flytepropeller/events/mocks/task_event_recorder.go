// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	context "context"

	config "github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"

	event "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"

	mock "github.com/stretchr/testify/mock"
)

// TaskEventRecorder is an autogenerated mock type for the TaskEventRecorder type
type TaskEventRecorder struct {
	mock.Mock
}

type TaskEventRecorder_Expecter struct {
	mock *mock.Mock
}

func (_m *TaskEventRecorder) EXPECT() *TaskEventRecorder_Expecter {
	return &TaskEventRecorder_Expecter{mock: &_m.Mock}
}

// RecordTaskEvent provides a mock function with given fields: ctx, _a1, eventConfig
func (_m *TaskEventRecorder) RecordTaskEvent(ctx context.Context, _a1 *event.TaskExecutionEvent, eventConfig *config.EventConfig) error {
	ret := _m.Called(ctx, _a1, eventConfig)

	if len(ret) == 0 {
		panic("no return value specified for RecordTaskEvent")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *event.TaskExecutionEvent, *config.EventConfig) error); ok {
		r0 = rf(ctx, _a1, eventConfig)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TaskEventRecorder_RecordTaskEvent_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RecordTaskEvent'
type TaskEventRecorder_RecordTaskEvent_Call struct {
	*mock.Call
}

// RecordTaskEvent is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 *event.TaskExecutionEvent
//   - eventConfig *config.EventConfig
func (_e *TaskEventRecorder_Expecter) RecordTaskEvent(ctx interface{}, _a1 interface{}, eventConfig interface{}) *TaskEventRecorder_RecordTaskEvent_Call {
	return &TaskEventRecorder_RecordTaskEvent_Call{Call: _e.mock.On("RecordTaskEvent", ctx, _a1, eventConfig)}
}

func (_c *TaskEventRecorder_RecordTaskEvent_Call) Run(run func(ctx context.Context, _a1 *event.TaskExecutionEvent, eventConfig *config.EventConfig)) *TaskEventRecorder_RecordTaskEvent_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*event.TaskExecutionEvent), args[2].(*config.EventConfig))
	})
	return _c
}

func (_c *TaskEventRecorder_RecordTaskEvent_Call) Return(_a0 error) *TaskEventRecorder_RecordTaskEvent_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TaskEventRecorder_RecordTaskEvent_Call) RunAndReturn(run func(context.Context, *event.TaskExecutionEvent, *config.EventConfig) error) *TaskEventRecorder_RecordTaskEvent_Call {
	_c.Call.Return(run)
	return _c
}

// NewTaskEventRecorder creates a new instance of TaskEventRecorder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTaskEventRecorder(t interface {
	mock.TestingT
	Cleanup(func())
}) *TaskEventRecorder {
	mock := &TaskEventRecorder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
