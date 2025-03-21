// Code generated by mockery v2.40.3. DO NOT EDIT.

package mocks

import (
	context "context"

	v1alpha1 "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	mock "github.com/stretchr/testify/mock"
)

// BaseWorkflowWithStatus is an autogenerated mock type for the BaseWorkflowWithStatus type
type BaseWorkflowWithStatus struct {
	mock.Mock
}

type BaseWorkflowWithStatus_Expecter struct {
	mock *mock.Mock
}

func (_m *BaseWorkflowWithStatus) EXPECT() *BaseWorkflowWithStatus_Expecter {
	return &BaseWorkflowWithStatus_Expecter{mock: &_m.Mock}
}

// FromNode provides a mock function with given fields: name
func (_m *BaseWorkflowWithStatus) FromNode(name string) ([]string, error) {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for FromNode")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]string, error)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BaseWorkflowWithStatus_FromNode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FromNode'
type BaseWorkflowWithStatus_FromNode_Call struct {
	*mock.Call
}

// FromNode is a helper method to define mock.On call
//   - name string
func (_e *BaseWorkflowWithStatus_Expecter) FromNode(name interface{}) *BaseWorkflowWithStatus_FromNode_Call {
	return &BaseWorkflowWithStatus_FromNode_Call{Call: _e.mock.On("FromNode", name)}
}

func (_c *BaseWorkflowWithStatus_FromNode_Call) Run(run func(name string)) *BaseWorkflowWithStatus_FromNode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *BaseWorkflowWithStatus_FromNode_Call) Return(_a0 []string, _a1 error) *BaseWorkflowWithStatus_FromNode_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *BaseWorkflowWithStatus_FromNode_Call) RunAndReturn(run func(string) ([]string, error)) *BaseWorkflowWithStatus_FromNode_Call {
	_c.Call.Return(run)
	return _c
}

// GetID provides a mock function with given fields:
func (_m *BaseWorkflowWithStatus) GetID() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetID")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// BaseWorkflowWithStatus_GetID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetID'
type BaseWorkflowWithStatus_GetID_Call struct {
	*mock.Call
}

// GetID is a helper method to define mock.On call
func (_e *BaseWorkflowWithStatus_Expecter) GetID() *BaseWorkflowWithStatus_GetID_Call {
	return &BaseWorkflowWithStatus_GetID_Call{Call: _e.mock.On("GetID")}
}

func (_c *BaseWorkflowWithStatus_GetID_Call) Run(run func()) *BaseWorkflowWithStatus_GetID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *BaseWorkflowWithStatus_GetID_Call) Return(_a0 string) *BaseWorkflowWithStatus_GetID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BaseWorkflowWithStatus_GetID_Call) RunAndReturn(run func() string) *BaseWorkflowWithStatus_GetID_Call {
	_c.Call.Return(run)
	return _c
}

// GetNode provides a mock function with given fields: nodeID
func (_m *BaseWorkflowWithStatus) GetNode(nodeID string) (v1alpha1.ExecutableNode, bool) {
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

// BaseWorkflowWithStatus_GetNode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNode'
type BaseWorkflowWithStatus_GetNode_Call struct {
	*mock.Call
}

// GetNode is a helper method to define mock.On call
//   - nodeID string
func (_e *BaseWorkflowWithStatus_Expecter) GetNode(nodeID interface{}) *BaseWorkflowWithStatus_GetNode_Call {
	return &BaseWorkflowWithStatus_GetNode_Call{Call: _e.mock.On("GetNode", nodeID)}
}

func (_c *BaseWorkflowWithStatus_GetNode_Call) Run(run func(nodeID string)) *BaseWorkflowWithStatus_GetNode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *BaseWorkflowWithStatus_GetNode_Call) Return(_a0 v1alpha1.ExecutableNode, _a1 bool) *BaseWorkflowWithStatus_GetNode_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *BaseWorkflowWithStatus_GetNode_Call) RunAndReturn(run func(string) (v1alpha1.ExecutableNode, bool)) *BaseWorkflowWithStatus_GetNode_Call {
	_c.Call.Return(run)
	return _c
}

// GetNodeExecutionStatus provides a mock function with given fields: ctx, id
func (_m *BaseWorkflowWithStatus) GetNodeExecutionStatus(ctx context.Context, id string) v1alpha1.ExecutableNodeStatus {
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

// BaseWorkflowWithStatus_GetNodeExecutionStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNodeExecutionStatus'
type BaseWorkflowWithStatus_GetNodeExecutionStatus_Call struct {
	*mock.Call
}

// GetNodeExecutionStatus is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *BaseWorkflowWithStatus_Expecter) GetNodeExecutionStatus(ctx interface{}, id interface{}) *BaseWorkflowWithStatus_GetNodeExecutionStatus_Call {
	return &BaseWorkflowWithStatus_GetNodeExecutionStatus_Call{Call: _e.mock.On("GetNodeExecutionStatus", ctx, id)}
}

func (_c *BaseWorkflowWithStatus_GetNodeExecutionStatus_Call) Run(run func(ctx context.Context, id string)) *BaseWorkflowWithStatus_GetNodeExecutionStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *BaseWorkflowWithStatus_GetNodeExecutionStatus_Call) Return(_a0 v1alpha1.ExecutableNodeStatus) *BaseWorkflowWithStatus_GetNodeExecutionStatus_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BaseWorkflowWithStatus_GetNodeExecutionStatus_Call) RunAndReturn(run func(context.Context, string) v1alpha1.ExecutableNodeStatus) *BaseWorkflowWithStatus_GetNodeExecutionStatus_Call {
	_c.Call.Return(run)
	return _c
}

// StartNode provides a mock function with given fields:
func (_m *BaseWorkflowWithStatus) StartNode() v1alpha1.ExecutableNode {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for StartNode")
	}

	var r0 v1alpha1.ExecutableNode
	if rf, ok := ret.Get(0).(func() v1alpha1.ExecutableNode); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1alpha1.ExecutableNode)
		}
	}

	return r0
}

// BaseWorkflowWithStatus_StartNode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'StartNode'
type BaseWorkflowWithStatus_StartNode_Call struct {
	*mock.Call
}

// StartNode is a helper method to define mock.On call
func (_e *BaseWorkflowWithStatus_Expecter) StartNode() *BaseWorkflowWithStatus_StartNode_Call {
	return &BaseWorkflowWithStatus_StartNode_Call{Call: _e.mock.On("StartNode")}
}

func (_c *BaseWorkflowWithStatus_StartNode_Call) Run(run func()) *BaseWorkflowWithStatus_StartNode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *BaseWorkflowWithStatus_StartNode_Call) Return(_a0 v1alpha1.ExecutableNode) *BaseWorkflowWithStatus_StartNode_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BaseWorkflowWithStatus_StartNode_Call) RunAndReturn(run func() v1alpha1.ExecutableNode) *BaseWorkflowWithStatus_StartNode_Call {
	_c.Call.Return(run)
	return _c
}

// ToNode provides a mock function with given fields: name
func (_m *BaseWorkflowWithStatus) ToNode(name string) ([]string, error) {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for ToNode")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]string, error)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BaseWorkflowWithStatus_ToNode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ToNode'
type BaseWorkflowWithStatus_ToNode_Call struct {
	*mock.Call
}

// ToNode is a helper method to define mock.On call
//   - name string
func (_e *BaseWorkflowWithStatus_Expecter) ToNode(name interface{}) *BaseWorkflowWithStatus_ToNode_Call {
	return &BaseWorkflowWithStatus_ToNode_Call{Call: _e.mock.On("ToNode", name)}
}

func (_c *BaseWorkflowWithStatus_ToNode_Call) Run(run func(name string)) *BaseWorkflowWithStatus_ToNode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *BaseWorkflowWithStatus_ToNode_Call) Return(_a0 []string, _a1 error) *BaseWorkflowWithStatus_ToNode_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *BaseWorkflowWithStatus_ToNode_Call) RunAndReturn(run func(string) ([]string, error)) *BaseWorkflowWithStatus_ToNode_Call {
	_c.Call.Return(run)
	return _c
}

// NewBaseWorkflowWithStatus creates a new instance of BaseWorkflowWithStatus. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBaseWorkflowWithStatus(t interface {
	mock.TestingT
	Cleanup(func())
}) *BaseWorkflowWithStatus {
	mock := &BaseWorkflowWithStatus{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
