// Code generated by mockery v2.40.3. DO NOT EDIT.

package mocks

import (
	storage "github.com/flyteorg/flyte/flytestdlib/storage"
	mock "github.com/stretchr/testify/mock"
)

// InputFilePaths is an autogenerated mock type for the InputFilePaths type
type InputFilePaths struct {
	mock.Mock
}

type InputFilePaths_Expecter struct {
	mock *mock.Mock
}

func (_m *InputFilePaths) EXPECT() *InputFilePaths_Expecter {
	return &InputFilePaths_Expecter{mock: &_m.Mock}
}

// GetInputPath provides a mock function with given fields:
func (_m *InputFilePaths) GetInputPath() storage.DataReference {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetInputPath")
	}

	var r0 storage.DataReference
	if rf, ok := ret.Get(0).(func() storage.DataReference); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(storage.DataReference)
	}

	return r0
}

// InputFilePaths_GetInputPath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetInputPath'
type InputFilePaths_GetInputPath_Call struct {
	*mock.Call
}

// GetInputPath is a helper method to define mock.On call
func (_e *InputFilePaths_Expecter) GetInputPath() *InputFilePaths_GetInputPath_Call {
	return &InputFilePaths_GetInputPath_Call{Call: _e.mock.On("GetInputPath")}
}

func (_c *InputFilePaths_GetInputPath_Call) Run(run func()) *InputFilePaths_GetInputPath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *InputFilePaths_GetInputPath_Call) Return(_a0 storage.DataReference) *InputFilePaths_GetInputPath_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *InputFilePaths_GetInputPath_Call) RunAndReturn(run func() storage.DataReference) *InputFilePaths_GetInputPath_Call {
	_c.Call.Return(run)
	return _c
}

// GetInputPrefixPath provides a mock function with given fields:
func (_m *InputFilePaths) GetInputPrefixPath() storage.DataReference {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetInputPrefixPath")
	}

	var r0 storage.DataReference
	if rf, ok := ret.Get(0).(func() storage.DataReference); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(storage.DataReference)
	}

	return r0
}

// InputFilePaths_GetInputPrefixPath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetInputPrefixPath'
type InputFilePaths_GetInputPrefixPath_Call struct {
	*mock.Call
}

// GetInputPrefixPath is a helper method to define mock.On call
func (_e *InputFilePaths_Expecter) GetInputPrefixPath() *InputFilePaths_GetInputPrefixPath_Call {
	return &InputFilePaths_GetInputPrefixPath_Call{Call: _e.mock.On("GetInputPrefixPath")}
}

func (_c *InputFilePaths_GetInputPrefixPath_Call) Run(run func()) *InputFilePaths_GetInputPrefixPath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *InputFilePaths_GetInputPrefixPath_Call) Return(_a0 storage.DataReference) *InputFilePaths_GetInputPrefixPath_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *InputFilePaths_GetInputPrefixPath_Call) RunAndReturn(run func() storage.DataReference) *InputFilePaths_GetInputPrefixPath_Call {
	_c.Call.Return(run)
	return _c
}

// NewInputFilePaths creates a new instance of InputFilePaths. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewInputFilePaths(t interface {
	mock.TestingT
	Cleanup(func())
}) *InputFilePaths {
	mock := &InputFilePaths{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
