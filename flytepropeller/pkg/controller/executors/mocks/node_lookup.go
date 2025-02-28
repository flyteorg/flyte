// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	v1alpha1 "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

// NodeLookup is an autogenerated mock type for the NodeLookup type
type NodeLookup struct {
	mock.Mock
}

type NodeLookup_Expecter struct {
	mock *mock.Mock
}

func (_m *NodeLookup) EXPECT() *NodeLookup_Expecter {
	return &NodeLookup_Expecter{mock: &_m.Mock}
}

// FromNode provides a mock function with given fields: id
func (_m *NodeLookup) FromNode(id string) ([]string, error) {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for FromNode")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]string, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NodeLookup_FromNode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FromNode'
type NodeLookup_FromNode_Call struct {
	*mock.Call
}

// FromNode is a helper method to define mock.On call
//   - id string
func (_e *NodeLookup_Expecter) FromNode(id interface{}) *NodeLookup_FromNode_Call {
	return &NodeLookup_FromNode_Call{Call: _e.mock.On("FromNode", id)}
}

func (_c *NodeLookup_FromNode_Call) Run(run func(id string)) *NodeLookup_FromNode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *NodeLookup_FromNode_Call) Return(_a0 []string, _a1 error) *NodeLookup_FromNode_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NodeLookup_FromNode_Call) RunAndReturn(run func(string) ([]string, error)) *NodeLookup_FromNode_Call {
	_c.Call.Return(run)
	return _c
}

// GetNode provides a mock function with given fields: nodeID
func (_m *NodeLookup) GetNode(nodeID string) (v1alpha1.ExecutableNode, bool) {
	ret := _m.Called(nodeID)

	if len(ret) == 0 {
		panic("no return value specified for GetNode")
	}

	var r0 v1alpha1.ExecutableNode
	var r1 bool
	if rf, ok := ret.Get(0).(func(string) (v1alpha1.ExecutableNode, bool)); ok {
		return rf(nodeID)
	}
	if rf, ok := ret.Get(0).(func(string) v1alpha1.ExecutableNode); ok {
		r0 = rf(nodeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1alpha1.ExecutableNode)
		}
	}

	if rf, ok := ret.Get(1).(func(string) bool); ok {
		r1 = rf(nodeID)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// NodeLookup_GetNode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNode'
type NodeLookup_GetNode_Call struct {
	*mock.Call
}

// GetNode is a helper method to define mock.On call
//   - nodeID string
func (_e *NodeLookup_Expecter) GetNode(nodeID interface{}) *NodeLookup_GetNode_Call {
	return &NodeLookup_GetNode_Call{Call: _e.mock.On("GetNode", nodeID)}
}

func (_c *NodeLookup_GetNode_Call) Run(run func(nodeID string)) *NodeLookup_GetNode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *NodeLookup_GetNode_Call) Return(_a0 v1alpha1.ExecutableNode, _a1 bool) *NodeLookup_GetNode_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NodeLookup_GetNode_Call) RunAndReturn(run func(string) (v1alpha1.ExecutableNode, bool)) *NodeLookup_GetNode_Call {
	_c.Call.Return(run)
	return _c
}

// GetNodeExecutionStatus provides a mock function with given fields: ctx, id
func (_m *NodeLookup) GetNodeExecutionStatus(ctx context.Context, id string) v1alpha1.ExecutableNodeStatus {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetNodeExecutionStatus")
	}

	var r0 v1alpha1.ExecutableNodeStatus
	if rf, ok := ret.Get(0).(func(context.Context, string) v1alpha1.ExecutableNodeStatus); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1alpha1.ExecutableNodeStatus)
		}
	}

	return r0
}

// NodeLookup_GetNodeExecutionStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNodeExecutionStatus'
type NodeLookup_GetNodeExecutionStatus_Call struct {
	*mock.Call
}

// GetNodeExecutionStatus is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *NodeLookup_Expecter) GetNodeExecutionStatus(ctx interface{}, id interface{}) *NodeLookup_GetNodeExecutionStatus_Call {
	return &NodeLookup_GetNodeExecutionStatus_Call{Call: _e.mock.On("GetNodeExecutionStatus", ctx, id)}
}

func (_c *NodeLookup_GetNodeExecutionStatus_Call) Run(run func(ctx context.Context, id string)) *NodeLookup_GetNodeExecutionStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *NodeLookup_GetNodeExecutionStatus_Call) Return(_a0 v1alpha1.ExecutableNodeStatus) *NodeLookup_GetNodeExecutionStatus_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeLookup_GetNodeExecutionStatus_Call) RunAndReturn(run func(context.Context, string) v1alpha1.ExecutableNodeStatus) *NodeLookup_GetNodeExecutionStatus_Call {
	_c.Call.Return(run)
	return _c
}

// ToNode provides a mock function with given fields: id
func (_m *NodeLookup) ToNode(id string) ([]string, error) {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for ToNode")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]string, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NodeLookup_ToNode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ToNode'
type NodeLookup_ToNode_Call struct {
	*mock.Call
}

// ToNode is a helper method to define mock.On call
//   - id string
func (_e *NodeLookup_Expecter) ToNode(id interface{}) *NodeLookup_ToNode_Call {
	return &NodeLookup_ToNode_Call{Call: _e.mock.On("ToNode", id)}
}

func (_c *NodeLookup_ToNode_Call) Run(run func(id string)) *NodeLookup_ToNode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *NodeLookup_ToNode_Call) Return(_a0 []string, _a1 error) *NodeLookup_ToNode_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NodeLookup_ToNode_Call) RunAndReturn(run func(string) ([]string, error)) *NodeLookup_ToNode_Call {
	_c.Call.Return(run)
	return _c
}

// NewNodeLookup creates a new instance of NodeLookup. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewNodeLookup(t interface {
	mock.TestingT
	Cleanup(func())
}) *NodeLookup {
	mock := &NodeLookup{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
