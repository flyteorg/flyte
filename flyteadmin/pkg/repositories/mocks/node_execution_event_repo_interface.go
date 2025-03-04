// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	models "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
)

// NodeExecutionEventRepoInterface is an autogenerated mock type for the NodeExecutionEventRepoInterface type
type NodeExecutionEventRepoInterface struct {
	mock.Mock
}

type NodeExecutionEventRepoInterface_Expecter struct {
	mock *mock.Mock
}

func (_m *NodeExecutionEventRepoInterface) EXPECT() *NodeExecutionEventRepoInterface_Expecter {
	return &NodeExecutionEventRepoInterface_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: ctx, input
func (_m *NodeExecutionEventRepoInterface) Create(ctx context.Context, input models.NodeExecutionEvent) error {
	ret := _m.Called(ctx, input)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, models.NodeExecutionEvent) error); ok {
		r0 = rf(ctx, input)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NodeExecutionEventRepoInterface_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type NodeExecutionEventRepoInterface_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - input models.NodeExecutionEvent
func (_e *NodeExecutionEventRepoInterface_Expecter) Create(ctx interface{}, input interface{}) *NodeExecutionEventRepoInterface_Create_Call {
	return &NodeExecutionEventRepoInterface_Create_Call{Call: _e.mock.On("Create", ctx, input)}
}

func (_c *NodeExecutionEventRepoInterface_Create_Call) Run(run func(ctx context.Context, input models.NodeExecutionEvent)) *NodeExecutionEventRepoInterface_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(models.NodeExecutionEvent))
	})
	return _c
}

func (_c *NodeExecutionEventRepoInterface_Create_Call) Return(_a0 error) *NodeExecutionEventRepoInterface_Create_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionEventRepoInterface_Create_Call) RunAndReturn(run func(context.Context, models.NodeExecutionEvent) error) *NodeExecutionEventRepoInterface_Create_Call {
	_c.Call.Return(run)
	return _c
}

// NewNodeExecutionEventRepoInterface creates a new instance of NodeExecutionEventRepoInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewNodeExecutionEventRepoInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *NodeExecutionEventRepoInterface {
	mock := &NodeExecutionEventRepoInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
