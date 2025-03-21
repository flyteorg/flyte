// Code generated by mockery v2.40.3. DO NOT EDIT.

package mocks

import (
	context "context"

	core "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"

	mock "github.com/stretchr/testify/mock"

	webapi "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/webapi"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

type Client_Expecter struct {
	mock *mock.Mock
}

func (_m *Client) EXPECT() *Client_Expecter {
	return &Client_Expecter{mock: &_m.Mock}
}

// Get provides a mock function with given fields: ctx, tCtx
func (_m *Client) Get(ctx context.Context, tCtx webapi.GetContext) (interface{}, error) {
	ret := _m.Called(ctx, tCtx)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 interface{}
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, webapi.GetContext) (interface{}, error)); ok {
		return rf(ctx, tCtx)
	}
	if rf, ok := ret.Get(0).(func(context.Context, webapi.GetContext) interface{}); ok {
		r0 = rf(ctx, tCtx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, webapi.GetContext) error); ok {
		r1 = rf(ctx, tCtx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Client_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type Client_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - tCtx webapi.GetContext
func (_e *Client_Expecter) Get(ctx interface{}, tCtx interface{}) *Client_Get_Call {
	return &Client_Get_Call{Call: _e.mock.On("Get", ctx, tCtx)}
}

func (_c *Client_Get_Call) Run(run func(ctx context.Context, tCtx webapi.GetContext)) *Client_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(webapi.GetContext))
	})
	return _c
}

func (_c *Client_Get_Call) Return(latest interface{}, err error) *Client_Get_Call {
	_c.Call.Return(latest, err)
	return _c
}

func (_c *Client_Get_Call) RunAndReturn(run func(context.Context, webapi.GetContext) (interface{}, error)) *Client_Get_Call {
	_c.Call.Return(run)
	return _c
}

// Status provides a mock function with given fields: ctx, tCtx
func (_m *Client) Status(ctx context.Context, tCtx webapi.StatusContext) (core.PhaseInfo, error) {
	ret := _m.Called(ctx, tCtx)

	if len(ret) == 0 {
		panic("no return value specified for Status")
	}

	var r0 core.PhaseInfo
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, webapi.StatusContext) (core.PhaseInfo, error)); ok {
		return rf(ctx, tCtx)
	}
	if rf, ok := ret.Get(0).(func(context.Context, webapi.StatusContext) core.PhaseInfo); ok {
		r0 = rf(ctx, tCtx)
	} else {
		r0 = ret.Get(0).(core.PhaseInfo)
	}

	if rf, ok := ret.Get(1).(func(context.Context, webapi.StatusContext) error); ok {
		r1 = rf(ctx, tCtx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Client_Status_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Status'
type Client_Status_Call struct {
	*mock.Call
}

// Status is a helper method to define mock.On call
//   - ctx context.Context
//   - tCtx webapi.StatusContext
func (_e *Client_Expecter) Status(ctx interface{}, tCtx interface{}) *Client_Status_Call {
	return &Client_Status_Call{Call: _e.mock.On("Status", ctx, tCtx)}
}

func (_c *Client_Status_Call) Run(run func(ctx context.Context, tCtx webapi.StatusContext)) *Client_Status_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(webapi.StatusContext))
	})
	return _c
}

func (_c *Client_Status_Call) Return(phase core.PhaseInfo, err error) *Client_Status_Call {
	_c.Call.Return(phase, err)
	return _c
}

func (_c *Client_Status_Call) RunAndReturn(run func(context.Context, webapi.StatusContext) (core.PhaseInfo, error)) *Client_Status_Call {
	_c.Call.Return(run)
	return _c
}

// NewClient creates a new instance of Client. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *Client {
	mock := &Client{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
