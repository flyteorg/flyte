// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	storage "github.com/flyteorg/flyte/flytestdlib/storage"
	mock "github.com/stretchr/testify/mock"
)

// RawOutputPaths is an autogenerated mock type for the RawOutputPaths type
type RawOutputPaths struct {
	mock.Mock
}

type RawOutputPaths_Expecter struct {
	mock *mock.Mock
}

func (_m *RawOutputPaths) EXPECT() *RawOutputPaths_Expecter {
	return &RawOutputPaths_Expecter{mock: &_m.Mock}
}

// GetRawOutputPrefix provides a mock function with no fields
func (_m *RawOutputPaths) GetRawOutputPrefix() storage.DataReference {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetRawOutputPrefix")
	}

	var r0 storage.DataReference
	if rf, ok := ret.Get(0).(func() storage.DataReference); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(storage.DataReference)
	}

	return r0
}

// RawOutputPaths_GetRawOutputPrefix_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetRawOutputPrefix'
type RawOutputPaths_GetRawOutputPrefix_Call struct {
	*mock.Call
}

// GetRawOutputPrefix is a helper method to define mock.On call
func (_e *RawOutputPaths_Expecter) GetRawOutputPrefix() *RawOutputPaths_GetRawOutputPrefix_Call {
	return &RawOutputPaths_GetRawOutputPrefix_Call{Call: _e.mock.On("GetRawOutputPrefix")}
}

func (_c *RawOutputPaths_GetRawOutputPrefix_Call) Run(run func()) *RawOutputPaths_GetRawOutputPrefix_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *RawOutputPaths_GetRawOutputPrefix_Call) Return(_a0 storage.DataReference) *RawOutputPaths_GetRawOutputPrefix_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *RawOutputPaths_GetRawOutputPrefix_Call) RunAndReturn(run func() storage.DataReference) *RawOutputPaths_GetRawOutputPrefix_Call {
	_c.Call.Return(run)
	return _c
}

// NewRawOutputPaths creates a new instance of RawOutputPaths. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRawOutputPaths(t interface {
	mock.TestingT
	Cleanup(func())
}) *RawOutputPaths {
	mock := &RawOutputPaths{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
