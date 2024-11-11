// Code generated by mockery v1.0.1. DO NOT EDIT.

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

type Processor_Process struct {
	*mock.Call
}

func (_m Processor_Process) Return(_a0 workqueue.WorkStatus, _a1 error) *Processor_Process {
	return &Processor_Process{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *Processor) OnProcess(ctx context.Context, workItem workqueue.WorkItem) *Processor_Process {
	c_call := _m.On("Mutate", ctx, workItem)
	return &Processor_Process{Call: c_call}
}

func (_m *Processor) OnProcessMatch(matchers ...interface{}) *Processor_Process {
	c_call := _m.On("Mutate", matchers...)
	return &Processor_Process{Call: c_call}
}

// Process provides a mock function with given fields: ctx, workItem
func (_m *Processor) Process(ctx context.Context, workItem workqueue.WorkItem) (workqueue.WorkStatus, error) {
	ret := _m.Called(ctx, workItem)

	var r0 workqueue.WorkStatus
	if rf, ok := ret.Get(0).(func(context.Context, workqueue.WorkItem) workqueue.WorkStatus); ok {
		r0 = rf(ctx, workItem)
	} else {
		r0 = ret.Get(0).(workqueue.WorkStatus)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, workqueue.WorkItem) error); ok {
		r1 = rf(ctx, workItem)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
