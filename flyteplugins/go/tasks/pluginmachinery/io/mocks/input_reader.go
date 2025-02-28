// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	context "context"

	core "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"

	mock "github.com/stretchr/testify/mock"

	storage "github.com/flyteorg/flyte/flytestdlib/storage"
)

// InputReader is an autogenerated mock type for the InputReader type
type InputReader struct {
	mock.Mock
}

type InputReader_Expecter struct {
	mock *mock.Mock
}

func (_m *InputReader) EXPECT() *InputReader_Expecter {
	return &InputReader_Expecter{mock: &_m.Mock}
}

// Get provides a mock function with given fields: ctx
func (_m *InputReader) Get(ctx context.Context) (*core.LiteralMap, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *core.LiteralMap
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*core.LiteralMap, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *core.LiteralMap); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.LiteralMap)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// InputReader_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type InputReader_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
func (_e *InputReader_Expecter) Get(ctx interface{}) *InputReader_Get_Call {
	return &InputReader_Get_Call{Call: _e.mock.On("Get", ctx)}
}

func (_c *InputReader_Get_Call) Run(run func(ctx context.Context)) *InputReader_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *InputReader_Get_Call) Return(_a0 *core.LiteralMap, _a1 error) *InputReader_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *InputReader_Get_Call) RunAndReturn(run func(context.Context) (*core.LiteralMap, error)) *InputReader_Get_Call {
	_c.Call.Return(run)
	return _c
}

// GetInputPath provides a mock function with no fields
func (_m *InputReader) GetInputPath() storage.DataReference {
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

// InputReader_GetInputPath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetInputPath'
type InputReader_GetInputPath_Call struct {
	*mock.Call
}

// GetInputPath is a helper method to define mock.On call
func (_e *InputReader_Expecter) GetInputPath() *InputReader_GetInputPath_Call {
	return &InputReader_GetInputPath_Call{Call: _e.mock.On("GetInputPath")}
}

func (_c *InputReader_GetInputPath_Call) Run(run func()) *InputReader_GetInputPath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *InputReader_GetInputPath_Call) Return(_a0 storage.DataReference) *InputReader_GetInputPath_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *InputReader_GetInputPath_Call) RunAndReturn(run func() storage.DataReference) *InputReader_GetInputPath_Call {
	_c.Call.Return(run)
	return _c
}

// GetInputPrefixPath provides a mock function with no fields
func (_m *InputReader) GetInputPrefixPath() storage.DataReference {
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

// InputReader_GetInputPrefixPath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetInputPrefixPath'
type InputReader_GetInputPrefixPath_Call struct {
	*mock.Call
}

// GetInputPrefixPath is a helper method to define mock.On call
func (_e *InputReader_Expecter) GetInputPrefixPath() *InputReader_GetInputPrefixPath_Call {
	return &InputReader_GetInputPrefixPath_Call{Call: _e.mock.On("GetInputPrefixPath")}
}

func (_c *InputReader_GetInputPrefixPath_Call) Run(run func()) *InputReader_GetInputPrefixPath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *InputReader_GetInputPrefixPath_Call) Return(_a0 storage.DataReference) *InputReader_GetInputPrefixPath_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *InputReader_GetInputPrefixPath_Call) RunAndReturn(run func() storage.DataReference) *InputReader_GetInputPrefixPath_Call {
	_c.Call.Return(run)
	return _c
}

// NewInputReader creates a new instance of InputReader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewInputReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *InputReader {
	mock := &InputReader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
