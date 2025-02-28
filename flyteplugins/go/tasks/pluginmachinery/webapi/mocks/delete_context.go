// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// DeleteContext is an autogenerated mock type for the DeleteContext type
type DeleteContext struct {
	mock.Mock
}

type DeleteContext_Expecter struct {
	mock *mock.Mock
}

func (_m *DeleteContext) EXPECT() *DeleteContext_Expecter {
	return &DeleteContext_Expecter{mock: &_m.Mock}
}

// Reason provides a mock function with no fields
func (_m *DeleteContext) Reason() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Reason")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// DeleteContext_Reason_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Reason'
type DeleteContext_Reason_Call struct {
	*mock.Call
}

// Reason is a helper method to define mock.On call
func (_e *DeleteContext_Expecter) Reason() *DeleteContext_Reason_Call {
	return &DeleteContext_Reason_Call{Call: _e.mock.On("Reason")}
}

func (_c *DeleteContext_Reason_Call) Run(run func()) *DeleteContext_Reason_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DeleteContext_Reason_Call) Return(_a0 string) *DeleteContext_Reason_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DeleteContext_Reason_Call) RunAndReturn(run func() string) *DeleteContext_Reason_Call {
	_c.Call.Return(run)
	return _c
}

// ResourceMeta provides a mock function with no fields
func (_m *DeleteContext) ResourceMeta() interface{} {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ResourceMeta")
	}

	var r0 interface{}
	if rf, ok := ret.Get(0).(func() interface{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	return r0
}

// DeleteContext_ResourceMeta_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResourceMeta'
type DeleteContext_ResourceMeta_Call struct {
	*mock.Call
}

// ResourceMeta is a helper method to define mock.On call
func (_e *DeleteContext_Expecter) ResourceMeta() *DeleteContext_ResourceMeta_Call {
	return &DeleteContext_ResourceMeta_Call{Call: _e.mock.On("ResourceMeta")}
}

func (_c *DeleteContext_ResourceMeta_Call) Run(run func()) *DeleteContext_ResourceMeta_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DeleteContext_ResourceMeta_Call) Return(_a0 interface{}) *DeleteContext_ResourceMeta_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DeleteContext_ResourceMeta_Call) RunAndReturn(run func() interface{}) *DeleteContext_ResourceMeta_Call {
	_c.Call.Return(run)
	return _c
}

// NewDeleteContext creates a new instance of DeleteContext. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDeleteContext(t interface {
	mock.TestingT
	Cleanup(func())
}) *DeleteContext {
	mock := &DeleteContext{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
