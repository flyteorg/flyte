// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// RateLimiter is an autogenerated mock type for the RateLimiter type
type RateLimiter struct {
	mock.Mock
}

type RateLimiter_Expecter struct {
	mock *mock.Mock
}

func (_m *RateLimiter) EXPECT() *RateLimiter_Expecter {
	return &RateLimiter_Expecter{mock: &_m.Mock}
}

// Wait provides a mock function with given fields: ctx
func (_m *RateLimiter) Wait(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Wait")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RateLimiter_Wait_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Wait'
type RateLimiter_Wait_Call struct {
	*mock.Call
}

// Wait is a helper method to define mock.On call
//   - ctx context.Context
func (_e *RateLimiter_Expecter) Wait(ctx interface{}) *RateLimiter_Wait_Call {
	return &RateLimiter_Wait_Call{Call: _e.mock.On("Wait", ctx)}
}

func (_c *RateLimiter_Wait_Call) Run(run func(ctx context.Context)) *RateLimiter_Wait_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *RateLimiter_Wait_Call) Return(_a0 error) *RateLimiter_Wait_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *RateLimiter_Wait_Call) RunAndReturn(run func(context.Context) error) *RateLimiter_Wait_Call {
	_c.Call.Return(run)
	return _c
}

// NewRateLimiter creates a new instance of RateLimiter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRateLimiter(t interface {
	mock.TestingT
	Cleanup(func())
}) *RateLimiter {
	mock := &RateLimiter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
