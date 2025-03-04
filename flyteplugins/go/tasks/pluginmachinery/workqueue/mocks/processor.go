// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	context "context"

	workqueue "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/workqueue"
	mock "github.com/stretchr/testify/mock"
)

// Processor is an autogenerated mock type for the Processor type
type Processor struct {
	mock.Mock
}

type Processor_Expecter struct {
	mock *mock.Mock
}

func (_m *Processor) EXPECT() *Processor_Expecter {
	return &Processor_Expecter{mock: &_m.Mock}
}

// Process provides a mock function with given fields: ctx, workItem
func (_m *Processor) Process(ctx context.Context, workItem workqueue.WorkItem) (workqueue.WorkStatus, error) {
	ret := _m.Called(ctx, workItem)

	if len(ret) == 0 {
		panic("no return value specified for Process")
	}

	var r0 workqueue.WorkStatus
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, workqueue.WorkItem) (workqueue.WorkStatus, error)); ok {
		return rf(ctx, workItem)
	}
	if rf, ok := ret.Get(0).(func(context.Context, workqueue.WorkItem) workqueue.WorkStatus); ok {
		r0 = rf(ctx, workItem)
	} else {
		r0 = ret.Get(0).(workqueue.WorkStatus)
	}

	if rf, ok := ret.Get(1).(func(context.Context, workqueue.WorkItem) error); ok {
		r1 = rf(ctx, workItem)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Processor_Process_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Process'
type Processor_Process_Call struct {
	*mock.Call
}

// Process is a helper method to define mock.On call
//   - ctx context.Context
//   - workItem workqueue.WorkItem
func (_e *Processor_Expecter) Process(ctx interface{}, workItem interface{}) *Processor_Process_Call {
	return &Processor_Process_Call{Call: _e.mock.On("Process", ctx, workItem)}
}

func (_c *Processor_Process_Call) Run(run func(ctx context.Context, workItem workqueue.WorkItem)) *Processor_Process_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(workqueue.WorkItem))
	})
	return _c
}

func (_c *Processor_Process_Call) Return(_a0 workqueue.WorkStatus, _a1 error) *Processor_Process_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Processor_Process_Call) RunAndReturn(run func(context.Context, workqueue.WorkItem) (workqueue.WorkStatus, error)) *Processor_Process_Call {
	_c.Call.Return(run)
	return _c
}

// NewProcessor creates a new instance of Processor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewProcessor(t interface {
	mock.TestingT
	Cleanup(func())
}) *Processor {
	mock := &Processor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
