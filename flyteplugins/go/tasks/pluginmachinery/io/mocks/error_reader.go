// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	context "context"

	io "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	mock "github.com/stretchr/testify/mock"
)

// ErrorReader is an autogenerated mock type for the ErrorReader type
type ErrorReader struct {
	mock.Mock
}

type ErrorReader_IsError struct {
	*mock.Call
}

func (_m ErrorReader_IsError) Return(_a0 bool, _a1 error) *ErrorReader_IsError {
	return &ErrorReader_IsError{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ErrorReader) OnIsError(ctx context.Context) *ErrorReader_IsError {
	c_call := _m.On("IsError", ctx)
	return &ErrorReader_IsError{Call: c_call}
}

func (_m *ErrorReader) OnIsErrorMatch(matchers ...interface{}) *ErrorReader_IsError {
	c_call := _m.On("IsError", matchers...)
	return &ErrorReader_IsError{Call: c_call}
}

// IsError provides a mock function with given fields: ctx
func (_m *ErrorReader) IsError(ctx context.Context) (bool, error) {
	ret := _m.Called(ctx)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context) bool); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ErrorReader_ReadError struct {
	*mock.Call
}

func (_m ErrorReader_ReadError) Return(_a0 io.ExecutionError, _a1 error) *ErrorReader_ReadError {
	return &ErrorReader_ReadError{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ErrorReader) OnReadError(ctx context.Context) *ErrorReader_ReadError {
	c_call := _m.On("ReadError", ctx)
	return &ErrorReader_ReadError{Call: c_call}
}

func (_m *ErrorReader) OnReadErrorMatch(matchers ...interface{}) *ErrorReader_ReadError {
	c_call := _m.On("ReadError", matchers...)
	return &ErrorReader_ReadError{Call: c_call}
}

// ReadError provides a mock function with given fields: ctx
func (_m *ErrorReader) ReadError(ctx context.Context) (io.ExecutionError, error) {
	ret := _m.Called(ctx)

	var r0 io.ExecutionError
	if rf, ok := ret.Get(0).(func(context.Context) io.ExecutionError); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(io.ExecutionError)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
