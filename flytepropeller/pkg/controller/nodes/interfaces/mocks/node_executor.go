// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	context "context"

	executors "github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	interfaces "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"

	mock "github.com/stretchr/testify/mock"
)

// NodeExecutor is an autogenerated mock type for the NodeExecutor type
type NodeExecutor struct {
	mock.Mock
}

type NodeExecutor_Expecter struct {
	mock *mock.Mock
}

func (_m *NodeExecutor) EXPECT() *NodeExecutor_Expecter {
	return &NodeExecutor_Expecter{mock: &_m.Mock}
}

// Abort provides a mock function with given fields: ctx, h, nCtx, reason, finalTransition
func (_m *NodeExecutor) Abort(ctx context.Context, h interfaces.NodeHandler, nCtx interfaces.NodeExecutionContext, reason string, finalTransition bool) error {
	ret := _m.Called(ctx, h, nCtx, reason, finalTransition)

	if len(ret) == 0 {
		panic("no return value specified for Abort")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, interfaces.NodeHandler, interfaces.NodeExecutionContext, string, bool) error); ok {
		r0 = rf(ctx, h, nCtx, reason, finalTransition)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NodeExecutor_Abort_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Abort'
type NodeExecutor_Abort_Call struct {
	*mock.Call
}

// Abort is a helper method to define mock.On call
//   - ctx context.Context
//   - h interfaces.NodeHandler
//   - nCtx interfaces.NodeExecutionContext
//   - reason string
//   - finalTransition bool
func (_e *NodeExecutor_Expecter) Abort(ctx interface{}, h interface{}, nCtx interface{}, reason interface{}, finalTransition interface{}) *NodeExecutor_Abort_Call {
	return &NodeExecutor_Abort_Call{Call: _e.mock.On("Abort", ctx, h, nCtx, reason, finalTransition)}
}

func (_c *NodeExecutor_Abort_Call) Run(run func(ctx context.Context, h interfaces.NodeHandler, nCtx interfaces.NodeExecutionContext, reason string, finalTransition bool)) *NodeExecutor_Abort_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(interfaces.NodeHandler), args[2].(interfaces.NodeExecutionContext), args[3].(string), args[4].(bool))
	})
	return _c
}

func (_c *NodeExecutor_Abort_Call) Return(_a0 error) *NodeExecutor_Abort_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutor_Abort_Call) RunAndReturn(run func(context.Context, interfaces.NodeHandler, interfaces.NodeExecutionContext, string, bool) error) *NodeExecutor_Abort_Call {
	_c.Call.Return(run)
	return _c
}

// Finalize provides a mock function with given fields: ctx, h, nCtx
func (_m *NodeExecutor) Finalize(ctx context.Context, h interfaces.NodeHandler, nCtx interfaces.NodeExecutionContext) error {
	ret := _m.Called(ctx, h, nCtx)

	if len(ret) == 0 {
		panic("no return value specified for Finalize")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, interfaces.NodeHandler, interfaces.NodeExecutionContext) error); ok {
		r0 = rf(ctx, h, nCtx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NodeExecutor_Finalize_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Finalize'
type NodeExecutor_Finalize_Call struct {
	*mock.Call
}

// Finalize is a helper method to define mock.On call
//   - ctx context.Context
//   - h interfaces.NodeHandler
//   - nCtx interfaces.NodeExecutionContext
func (_e *NodeExecutor_Expecter) Finalize(ctx interface{}, h interface{}, nCtx interface{}) *NodeExecutor_Finalize_Call {
	return &NodeExecutor_Finalize_Call{Call: _e.mock.On("Finalize", ctx, h, nCtx)}
}

func (_c *NodeExecutor_Finalize_Call) Run(run func(ctx context.Context, h interfaces.NodeHandler, nCtx interfaces.NodeExecutionContext)) *NodeExecutor_Finalize_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(interfaces.NodeHandler), args[2].(interfaces.NodeExecutionContext))
	})
	return _c
}

func (_c *NodeExecutor_Finalize_Call) Return(_a0 error) *NodeExecutor_Finalize_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutor_Finalize_Call) RunAndReturn(run func(context.Context, interfaces.NodeHandler, interfaces.NodeExecutionContext) error) *NodeExecutor_Finalize_Call {
	_c.Call.Return(run)
	return _c
}

// HandleNode provides a mock function with given fields: ctx, dag, nCtx, h
func (_m *NodeExecutor) HandleNode(ctx context.Context, dag executors.DAGStructure, nCtx interfaces.NodeExecutionContext, h interfaces.NodeHandler) (interfaces.NodeStatus, error) {
	ret := _m.Called(ctx, dag, nCtx, h)

	if len(ret) == 0 {
		panic("no return value specified for HandleNode")
	}

	var r0 interfaces.NodeStatus
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, executors.DAGStructure, interfaces.NodeExecutionContext, interfaces.NodeHandler) (interfaces.NodeStatus, error)); ok {
		return rf(ctx, dag, nCtx, h)
	}
	if rf, ok := ret.Get(0).(func(context.Context, executors.DAGStructure, interfaces.NodeExecutionContext, interfaces.NodeHandler) interfaces.NodeStatus); ok {
		r0 = rf(ctx, dag, nCtx, h)
	} else {
		r0 = ret.Get(0).(interfaces.NodeStatus)
	}

	if rf, ok := ret.Get(1).(func(context.Context, executors.DAGStructure, interfaces.NodeExecutionContext, interfaces.NodeHandler) error); ok {
		r1 = rf(ctx, dag, nCtx, h)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NodeExecutor_HandleNode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HandleNode'
type NodeExecutor_HandleNode_Call struct {
	*mock.Call
}

// HandleNode is a helper method to define mock.On call
//   - ctx context.Context
//   - dag executors.DAGStructure
//   - nCtx interfaces.NodeExecutionContext
//   - h interfaces.NodeHandler
func (_e *NodeExecutor_Expecter) HandleNode(ctx interface{}, dag interface{}, nCtx interface{}, h interface{}) *NodeExecutor_HandleNode_Call {
	return &NodeExecutor_HandleNode_Call{Call: _e.mock.On("HandleNode", ctx, dag, nCtx, h)}
}

func (_c *NodeExecutor_HandleNode_Call) Run(run func(ctx context.Context, dag executors.DAGStructure, nCtx interfaces.NodeExecutionContext, h interfaces.NodeHandler)) *NodeExecutor_HandleNode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(executors.DAGStructure), args[2].(interfaces.NodeExecutionContext), args[3].(interfaces.NodeHandler))
	})
	return _c
}

func (_c *NodeExecutor_HandleNode_Call) Return(_a0 interfaces.NodeStatus, _a1 error) *NodeExecutor_HandleNode_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NodeExecutor_HandleNode_Call) RunAndReturn(run func(context.Context, executors.DAGStructure, interfaces.NodeExecutionContext, interfaces.NodeHandler) (interfaces.NodeStatus, error)) *NodeExecutor_HandleNode_Call {
	_c.Call.Return(run)
	return _c
}

// NewNodeExecutor creates a new instance of NodeExecutor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewNodeExecutor(t interface {
	mock.TestingT
	Cleanup(func())
}) *NodeExecutor {
	mock := &NodeExecutor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
