// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	context "context"

	core "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	io "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"

	mock "github.com/stretchr/testify/mock"
)

// OutputReader is an autogenerated mock type for the OutputReader type
type OutputReader struct {
	mock.Mock
}

type OutputReader_Expecter struct {
	mock *mock.Mock
}

func (_m *OutputReader) EXPECT() *OutputReader_Expecter {
	return &OutputReader_Expecter{mock: &_m.Mock}
}

// DeckExists provides a mock function with given fields: ctx
func (_m *OutputReader) DeckExists(ctx context.Context) (bool, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for DeckExists")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (bool, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) bool); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OutputReader_DeckExists_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeckExists'
type OutputReader_DeckExists_Call struct {
	*mock.Call
}

// DeckExists is a helper method to define mock.On call
//   - ctx context.Context
func (_e *OutputReader_Expecter) DeckExists(ctx interface{}) *OutputReader_DeckExists_Call {
	return &OutputReader_DeckExists_Call{Call: _e.mock.On("DeckExists", ctx)}
}

func (_c *OutputReader_DeckExists_Call) Run(run func(ctx context.Context)) *OutputReader_DeckExists_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *OutputReader_DeckExists_Call) Return(_a0 bool, _a1 error) *OutputReader_DeckExists_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *OutputReader_DeckExists_Call) RunAndReturn(run func(context.Context) (bool, error)) *OutputReader_DeckExists_Call {
	_c.Call.Return(run)
	return _c
}

// Exists provides a mock function with given fields: ctx
func (_m *OutputReader) Exists(ctx context.Context) (bool, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Exists")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (bool, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) bool); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OutputReader_Exists_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Exists'
type OutputReader_Exists_Call struct {
	*mock.Call
}

// Exists is a helper method to define mock.On call
//   - ctx context.Context
func (_e *OutputReader_Expecter) Exists(ctx interface{}) *OutputReader_Exists_Call {
	return &OutputReader_Exists_Call{Call: _e.mock.On("Exists", ctx)}
}

func (_c *OutputReader_Exists_Call) Run(run func(ctx context.Context)) *OutputReader_Exists_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *OutputReader_Exists_Call) Return(_a0 bool, _a1 error) *OutputReader_Exists_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *OutputReader_Exists_Call) RunAndReturn(run func(context.Context) (bool, error)) *OutputReader_Exists_Call {
	_c.Call.Return(run)
	return _c
}

// IsError provides a mock function with given fields: ctx
func (_m *OutputReader) IsError(ctx context.Context) (bool, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for IsError")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (bool, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) bool); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OutputReader_IsError_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsError'
type OutputReader_IsError_Call struct {
	*mock.Call
}

// IsError is a helper method to define mock.On call
//   - ctx context.Context
func (_e *OutputReader_Expecter) IsError(ctx interface{}) *OutputReader_IsError_Call {
	return &OutputReader_IsError_Call{Call: _e.mock.On("IsError", ctx)}
}

func (_c *OutputReader_IsError_Call) Run(run func(ctx context.Context)) *OutputReader_IsError_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *OutputReader_IsError_Call) Return(_a0 bool, _a1 error) *OutputReader_IsError_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *OutputReader_IsError_Call) RunAndReturn(run func(context.Context) (bool, error)) *OutputReader_IsError_Call {
	_c.Call.Return(run)
	return _c
}

// IsFile provides a mock function with given fields: ctx
func (_m *OutputReader) IsFile(ctx context.Context) bool {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for IsFile")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context) bool); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// OutputReader_IsFile_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsFile'
type OutputReader_IsFile_Call struct {
	*mock.Call
}

// IsFile is a helper method to define mock.On call
//   - ctx context.Context
func (_e *OutputReader_Expecter) IsFile(ctx interface{}) *OutputReader_IsFile_Call {
	return &OutputReader_IsFile_Call{Call: _e.mock.On("IsFile", ctx)}
}

func (_c *OutputReader_IsFile_Call) Run(run func(ctx context.Context)) *OutputReader_IsFile_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *OutputReader_IsFile_Call) Return(_a0 bool) *OutputReader_IsFile_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *OutputReader_IsFile_Call) RunAndReturn(run func(context.Context) bool) *OutputReader_IsFile_Call {
	_c.Call.Return(run)
	return _c
}

// Read provides a mock function with given fields: ctx
func (_m *OutputReader) Read(ctx context.Context) (*core.LiteralMap, *io.ExecutionError, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Read")
	}

	var r0 *core.LiteralMap
	var r1 *io.ExecutionError
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context) (*core.LiteralMap, *io.ExecutionError, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *core.LiteralMap); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.LiteralMap)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) *io.ExecutionError); ok {
		r1 = rf(ctx)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*io.ExecutionError)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context) error); ok {
		r2 = rf(ctx)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// OutputReader_Read_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Read'
type OutputReader_Read_Call struct {
	*mock.Call
}

// Read is a helper method to define mock.On call
//   - ctx context.Context
func (_e *OutputReader_Expecter) Read(ctx interface{}) *OutputReader_Read_Call {
	return &OutputReader_Read_Call{Call: _e.mock.On("Read", ctx)}
}

func (_c *OutputReader_Read_Call) Run(run func(ctx context.Context)) *OutputReader_Read_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *OutputReader_Read_Call) Return(_a0 *core.LiteralMap, _a1 *io.ExecutionError, _a2 error) *OutputReader_Read_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *OutputReader_Read_Call) RunAndReturn(run func(context.Context) (*core.LiteralMap, *io.ExecutionError, error)) *OutputReader_Read_Call {
	_c.Call.Return(run)
	return _c
}

// ReadError provides a mock function with given fields: ctx
func (_m *OutputReader) ReadError(ctx context.Context) (io.ExecutionError, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for ReadError")
	}

	var r0 io.ExecutionError
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (io.ExecutionError, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) io.ExecutionError); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(io.ExecutionError)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OutputReader_ReadError_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ReadError'
type OutputReader_ReadError_Call struct {
	*mock.Call
}

// ReadError is a helper method to define mock.On call
//   - ctx context.Context
func (_e *OutputReader_Expecter) ReadError(ctx interface{}) *OutputReader_ReadError_Call {
	return &OutputReader_ReadError_Call{Call: _e.mock.On("ReadError", ctx)}
}

func (_c *OutputReader_ReadError_Call) Run(run func(ctx context.Context)) *OutputReader_ReadError_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *OutputReader_ReadError_Call) Return(_a0 io.ExecutionError, _a1 error) *OutputReader_ReadError_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *OutputReader_ReadError_Call) RunAndReturn(run func(context.Context) (io.ExecutionError, error)) *OutputReader_ReadError_Call {
	_c.Call.Return(run)
	return _c
}

// NewOutputReader creates a new instance of OutputReader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewOutputReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *OutputReader {
	mock := &OutputReader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
