// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	v1alpha1 "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	mock "github.com/stretchr/testify/mock"
)

// TaskDetailsGetter is an autogenerated mock type for the TaskDetailsGetter type
type TaskDetailsGetter struct {
	mock.Mock
}

type TaskDetailsGetter_Expecter struct {
	mock *mock.Mock
}

func (_m *TaskDetailsGetter) EXPECT() *TaskDetailsGetter_Expecter {
	return &TaskDetailsGetter_Expecter{mock: &_m.Mock}
}

// GetTask provides a mock function with given fields: id
func (_m *TaskDetailsGetter) GetTask(id string) (v1alpha1.ExecutableTask, error) {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for GetTask")
	}

	var r0 v1alpha1.ExecutableTask
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (v1alpha1.ExecutableTask, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(string) v1alpha1.ExecutableTask); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1alpha1.ExecutableTask)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TaskDetailsGetter_GetTask_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTask'
type TaskDetailsGetter_GetTask_Call struct {
	*mock.Call
}

// GetTask is a helper method to define mock.On call
//   - id string
func (_e *TaskDetailsGetter_Expecter) GetTask(id interface{}) *TaskDetailsGetter_GetTask_Call {
	return &TaskDetailsGetter_GetTask_Call{Call: _e.mock.On("GetTask", id)}
}

func (_c *TaskDetailsGetter_GetTask_Call) Run(run func(id string)) *TaskDetailsGetter_GetTask_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *TaskDetailsGetter_GetTask_Call) Return(_a0 v1alpha1.ExecutableTask, _a1 error) *TaskDetailsGetter_GetTask_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TaskDetailsGetter_GetTask_Call) RunAndReturn(run func(string) (v1alpha1.ExecutableTask, error)) *TaskDetailsGetter_GetTask_Call {
	_c.Call.Return(run)
	return _c
}

// NewTaskDetailsGetter creates a new instance of TaskDetailsGetter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTaskDetailsGetter(t interface {
	mock.TestingT
	Cleanup(func())
}) *TaskDetailsGetter {
	mock := &TaskDetailsGetter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
