// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	context "context"

	admin "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

	core "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"

	mock "github.com/stretchr/testify/mock"
)

// Reader is an autogenerated mock type for the Reader type
type Reader struct {
	mock.Mock
}

type Reader_Expecter struct {
	mock *mock.Mock
}

func (_m *Reader) EXPECT() *Reader_Expecter {
	return &Reader_Expecter{mock: &_m.Mock}
}

// GetLaunchPlan provides a mock function with given fields: ctx, launchPlanRef
func (_m *Reader) GetLaunchPlan(ctx context.Context, launchPlanRef *core.Identifier) (*admin.LaunchPlan, error) {
	ret := _m.Called(ctx, launchPlanRef)

	if len(ret) == 0 {
		panic("no return value specified for GetLaunchPlan")
	}

	var r0 *admin.LaunchPlan
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *core.Identifier) (*admin.LaunchPlan, error)); ok {
		return rf(ctx, launchPlanRef)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *core.Identifier) *admin.LaunchPlan); ok {
		r0 = rf(ctx, launchPlanRef)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.LaunchPlan)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *core.Identifier) error); ok {
		r1 = rf(ctx, launchPlanRef)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Reader_GetLaunchPlan_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLaunchPlan'
type Reader_GetLaunchPlan_Call struct {
	*mock.Call
}

// GetLaunchPlan is a helper method to define mock.On call
//   - ctx context.Context
//   - launchPlanRef *core.Identifier
func (_e *Reader_Expecter) GetLaunchPlan(ctx interface{}, launchPlanRef interface{}) *Reader_GetLaunchPlan_Call {
	return &Reader_GetLaunchPlan_Call{Call: _e.mock.On("GetLaunchPlan", ctx, launchPlanRef)}
}

func (_c *Reader_GetLaunchPlan_Call) Run(run func(ctx context.Context, launchPlanRef *core.Identifier)) *Reader_GetLaunchPlan_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*core.Identifier))
	})
	return _c
}

func (_c *Reader_GetLaunchPlan_Call) Return(_a0 *admin.LaunchPlan, _a1 error) *Reader_GetLaunchPlan_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Reader_GetLaunchPlan_Call) RunAndReturn(run func(context.Context, *core.Identifier) (*admin.LaunchPlan, error)) *Reader_GetLaunchPlan_Call {
	_c.Call.Return(run)
	return _c
}

// NewReader creates a new instance of Reader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *Reader {
	mock := &Reader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
