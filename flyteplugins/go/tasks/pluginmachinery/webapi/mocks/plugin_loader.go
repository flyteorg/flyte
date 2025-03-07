// Code generated by mockery v2.40.3. DO NOT EDIT.

package mocks

import (
	context "context"

	webapi "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/webapi"
	mock "github.com/stretchr/testify/mock"
)

// PluginLoader is an autogenerated mock type for the PluginLoader type
type PluginLoader struct {
	mock.Mock
}

type PluginLoader_Expecter struct {
	mock *mock.Mock
}

func (_m *PluginLoader) EXPECT() *PluginLoader_Expecter {
	return &PluginLoader_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: ctx, iCtx
func (_m *PluginLoader) Execute(ctx context.Context, iCtx webapi.PluginSetupContext) (webapi.AsyncPlugin, error) {
	ret := _m.Called(ctx, iCtx)

	if len(ret) == 0 {
		panic("no return value specified for Execute")
	}

	var r0 webapi.AsyncPlugin
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, webapi.PluginSetupContext) (webapi.AsyncPlugin, error)); ok {
		return rf(ctx, iCtx)
	}
	if rf, ok := ret.Get(0).(func(context.Context, webapi.PluginSetupContext) webapi.AsyncPlugin); ok {
		r0 = rf(ctx, iCtx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(webapi.AsyncPlugin)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, webapi.PluginSetupContext) error); ok {
		r1 = rf(ctx, iCtx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PluginLoader_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type PluginLoader_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - ctx context.Context
//   - iCtx webapi.PluginSetupContext
func (_e *PluginLoader_Expecter) Execute(ctx interface{}, iCtx interface{}) *PluginLoader_Execute_Call {
	return &PluginLoader_Execute_Call{Call: _e.mock.On("Execute", ctx, iCtx)}
}

func (_c *PluginLoader_Execute_Call) Run(run func(ctx context.Context, iCtx webapi.PluginSetupContext)) *PluginLoader_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(webapi.PluginSetupContext))
	})
	return _c
}

func (_c *PluginLoader_Execute_Call) Return(_a0 webapi.AsyncPlugin, _a1 error) *PluginLoader_Execute_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *PluginLoader_Execute_Call) RunAndReturn(run func(context.Context, webapi.PluginSetupContext) (webapi.AsyncPlugin, error)) *PluginLoader_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// NewPluginLoader creates a new instance of PluginLoader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPluginLoader(t interface {
	mock.TestingT
	Cleanup(func())
}) *PluginLoader {
	mock := &PluginLoader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
