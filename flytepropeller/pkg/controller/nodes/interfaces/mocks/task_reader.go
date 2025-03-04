// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	context "context"

	core "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"

	mock "github.com/stretchr/testify/mock"
)

// TaskReader is an autogenerated mock type for the TaskReader type
type TaskReader struct {
	mock.Mock
}

type TaskReader_Expecter struct {
	mock *mock.Mock
}

func (_m *TaskReader) EXPECT() *TaskReader_Expecter {
	return &TaskReader_Expecter{mock: &_m.Mock}
}

// GetTaskID provides a mock function with no fields
func (_m *TaskReader) GetTaskID() *core.Identifier {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetTaskID")
	}

	var r0 *core.Identifier
	if rf, ok := ret.Get(0).(func() *core.Identifier); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.Identifier)
		}
	}

	return r0
}

// TaskReader_GetTaskID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTaskID'
type TaskReader_GetTaskID_Call struct {
	*mock.Call
}

// GetTaskID is a helper method to define mock.On call
func (_e *TaskReader_Expecter) GetTaskID() *TaskReader_GetTaskID_Call {
	return &TaskReader_GetTaskID_Call{Call: _e.mock.On("GetTaskID")}
}

func (_c *TaskReader_GetTaskID_Call) Run(run func()) *TaskReader_GetTaskID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *TaskReader_GetTaskID_Call) Return(_a0 *core.Identifier) *TaskReader_GetTaskID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TaskReader_GetTaskID_Call) RunAndReturn(run func() *core.Identifier) *TaskReader_GetTaskID_Call {
	_c.Call.Return(run)
	return _c
}

// GetTaskType provides a mock function with no fields
func (_m *TaskReader) GetTaskType() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetTaskType")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// TaskReader_GetTaskType_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTaskType'
type TaskReader_GetTaskType_Call struct {
	*mock.Call
}

// GetTaskType is a helper method to define mock.On call
func (_e *TaskReader_Expecter) GetTaskType() *TaskReader_GetTaskType_Call {
	return &TaskReader_GetTaskType_Call{Call: _e.mock.On("GetTaskType")}
}

func (_c *TaskReader_GetTaskType_Call) Run(run func()) *TaskReader_GetTaskType_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *TaskReader_GetTaskType_Call) Return(_a0 string) *TaskReader_GetTaskType_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TaskReader_GetTaskType_Call) RunAndReturn(run func() string) *TaskReader_GetTaskType_Call {
	_c.Call.Return(run)
	return _c
}

// Read provides a mock function with given fields: ctx
func (_m *TaskReader) Read(ctx context.Context) (*core.TaskTemplate, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Read")
	}

	var r0 *core.TaskTemplate
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*core.TaskTemplate, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *core.TaskTemplate); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.TaskTemplate)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TaskReader_Read_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Read'
type TaskReader_Read_Call struct {
	*mock.Call
}

// Read is a helper method to define mock.On call
//   - ctx context.Context
func (_e *TaskReader_Expecter) Read(ctx interface{}) *TaskReader_Read_Call {
	return &TaskReader_Read_Call{Call: _e.mock.On("Read", ctx)}
}

func (_c *TaskReader_Read_Call) Run(run func(ctx context.Context)) *TaskReader_Read_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *TaskReader_Read_Call) Return(_a0 *core.TaskTemplate, _a1 error) *TaskReader_Read_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TaskReader_Read_Call) RunAndReturn(run func(context.Context) (*core.TaskTemplate, error)) *TaskReader_Read_Call {
	_c.Call.Return(run)
	return _c
}

// NewTaskReader creates a new instance of TaskReader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTaskReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *TaskReader {
	mock := &TaskReader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
