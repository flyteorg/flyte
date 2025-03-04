// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	v1alpha1 "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

// Workflow is an autogenerated mock type for the Workflow type
type Workflow struct {
	mock.Mock
}

type Workflow_Expecter struct {
	mock *mock.Mock
}

func (_m *Workflow) EXPECT() *Workflow_Expecter {
	return &Workflow_Expecter{mock: &_m.Mock}
}

// HandleAbortedWorkflow provides a mock function with given fields: ctx, w, maxRetries
func (_m *Workflow) HandleAbortedWorkflow(ctx context.Context, w *v1alpha1.FlyteWorkflow, maxRetries uint32) error {
	ret := _m.Called(ctx, w, maxRetries)

	if len(ret) == 0 {
		panic("no return value specified for HandleAbortedWorkflow")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1alpha1.FlyteWorkflow, uint32) error); ok {
		r0 = rf(ctx, w, maxRetries)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Workflow_HandleAbortedWorkflow_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HandleAbortedWorkflow'
type Workflow_HandleAbortedWorkflow_Call struct {
	*mock.Call
}

// HandleAbortedWorkflow is a helper method to define mock.On call
//   - ctx context.Context
//   - w *v1alpha1.FlyteWorkflow
//   - maxRetries uint32
func (_e *Workflow_Expecter) HandleAbortedWorkflow(ctx interface{}, w interface{}, maxRetries interface{}) *Workflow_HandleAbortedWorkflow_Call {
	return &Workflow_HandleAbortedWorkflow_Call{Call: _e.mock.On("HandleAbortedWorkflow", ctx, w, maxRetries)}
}

func (_c *Workflow_HandleAbortedWorkflow_Call) Run(run func(ctx context.Context, w *v1alpha1.FlyteWorkflow, maxRetries uint32)) *Workflow_HandleAbortedWorkflow_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v1alpha1.FlyteWorkflow), args[2].(uint32))
	})
	return _c
}

func (_c *Workflow_HandleAbortedWorkflow_Call) Return(_a0 error) *Workflow_HandleAbortedWorkflow_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Workflow_HandleAbortedWorkflow_Call) RunAndReturn(run func(context.Context, *v1alpha1.FlyteWorkflow, uint32) error) *Workflow_HandleAbortedWorkflow_Call {
	_c.Call.Return(run)
	return _c
}

// HandleFlyteWorkflow provides a mock function with given fields: ctx, w
func (_m *Workflow) HandleFlyteWorkflow(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
	ret := _m.Called(ctx, w)

	if len(ret) == 0 {
		panic("no return value specified for HandleFlyteWorkflow")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1alpha1.FlyteWorkflow) error); ok {
		r0 = rf(ctx, w)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Workflow_HandleFlyteWorkflow_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HandleFlyteWorkflow'
type Workflow_HandleFlyteWorkflow_Call struct {
	*mock.Call
}

// HandleFlyteWorkflow is a helper method to define mock.On call
//   - ctx context.Context
//   - w *v1alpha1.FlyteWorkflow
func (_e *Workflow_Expecter) HandleFlyteWorkflow(ctx interface{}, w interface{}) *Workflow_HandleFlyteWorkflow_Call {
	return &Workflow_HandleFlyteWorkflow_Call{Call: _e.mock.On("HandleFlyteWorkflow", ctx, w)}
}

func (_c *Workflow_HandleFlyteWorkflow_Call) Run(run func(ctx context.Context, w *v1alpha1.FlyteWorkflow)) *Workflow_HandleFlyteWorkflow_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v1alpha1.FlyteWorkflow))
	})
	return _c
}

func (_c *Workflow_HandleFlyteWorkflow_Call) Return(_a0 error) *Workflow_HandleFlyteWorkflow_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Workflow_HandleFlyteWorkflow_Call) RunAndReturn(run func(context.Context, *v1alpha1.FlyteWorkflow) error) *Workflow_HandleFlyteWorkflow_Call {
	_c.Call.Return(run)
	return _c
}

// Initialize provides a mock function with given fields: ctx
func (_m *Workflow) Initialize(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Initialize")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Workflow_Initialize_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Initialize'
type Workflow_Initialize_Call struct {
	*mock.Call
}

// Initialize is a helper method to define mock.On call
//   - ctx context.Context
func (_e *Workflow_Expecter) Initialize(ctx interface{}) *Workflow_Initialize_Call {
	return &Workflow_Initialize_Call{Call: _e.mock.On("Initialize", ctx)}
}

func (_c *Workflow_Initialize_Call) Run(run func(ctx context.Context)) *Workflow_Initialize_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Workflow_Initialize_Call) Return(_a0 error) *Workflow_Initialize_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Workflow_Initialize_Call) RunAndReturn(run func(context.Context) error) *Workflow_Initialize_Call {
	_c.Call.Return(run)
	return _c
}

// NewWorkflow creates a new instance of Workflow. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewWorkflow(t interface {
	mock.TestingT
	Cleanup(func())
}) *Workflow {
	mock := &Workflow{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
