// Code generated by mockery v2.40.3. DO NOT EDIT.

package mocks

import (
	v1alpha1 "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	mock "github.com/stretchr/testify/mock"
)

// ExecutableBranchNodeStatus is an autogenerated mock type for the ExecutableBranchNodeStatus type
type ExecutableBranchNodeStatus struct {
	mock.Mock
}

type ExecutableBranchNodeStatus_Expecter struct {
	mock *mock.Mock
}

func (_m *ExecutableBranchNodeStatus) EXPECT() *ExecutableBranchNodeStatus_Expecter {
	return &ExecutableBranchNodeStatus_Expecter{mock: &_m.Mock}
}

// GetFinalizedNode provides a mock function with given fields:
func (_m *ExecutableBranchNodeStatus) GetFinalizedNode() *string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetFinalizedNode")
	}

	var r0 *string
	if rf, ok := ret.Get(0).(func() *string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*string)
		}
	}

	return r0
}

// ExecutableBranchNodeStatus_GetFinalizedNode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetFinalizedNode'
type ExecutableBranchNodeStatus_GetFinalizedNode_Call struct {
	*mock.Call
}

// GetFinalizedNode is a helper method to define mock.On call
func (_e *ExecutableBranchNodeStatus_Expecter) GetFinalizedNode() *ExecutableBranchNodeStatus_GetFinalizedNode_Call {
	return &ExecutableBranchNodeStatus_GetFinalizedNode_Call{Call: _e.mock.On("GetFinalizedNode")}
}

func (_c *ExecutableBranchNodeStatus_GetFinalizedNode_Call) Run(run func()) *ExecutableBranchNodeStatus_GetFinalizedNode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ExecutableBranchNodeStatus_GetFinalizedNode_Call) Return(_a0 *string) *ExecutableBranchNodeStatus_GetFinalizedNode_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ExecutableBranchNodeStatus_GetFinalizedNode_Call) RunAndReturn(run func() *string) *ExecutableBranchNodeStatus_GetFinalizedNode_Call {
	_c.Call.Return(run)
	return _c
}

// GetPhase provides a mock function with given fields:
func (_m *ExecutableBranchNodeStatus) GetPhase() v1alpha1.BranchNodePhase {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetPhase")
	}

	var r0 v1alpha1.BranchNodePhase
	if rf, ok := ret.Get(0).(func() v1alpha1.BranchNodePhase); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(v1alpha1.BranchNodePhase)
	}

	return r0
}

// ExecutableBranchNodeStatus_GetPhase_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPhase'
type ExecutableBranchNodeStatus_GetPhase_Call struct {
	*mock.Call
}

// GetPhase is a helper method to define mock.On call
func (_e *ExecutableBranchNodeStatus_Expecter) GetPhase() *ExecutableBranchNodeStatus_GetPhase_Call {
	return &ExecutableBranchNodeStatus_GetPhase_Call{Call: _e.mock.On("GetPhase")}
}

func (_c *ExecutableBranchNodeStatus_GetPhase_Call) Run(run func()) *ExecutableBranchNodeStatus_GetPhase_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ExecutableBranchNodeStatus_GetPhase_Call) Return(_a0 v1alpha1.BranchNodePhase) *ExecutableBranchNodeStatus_GetPhase_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ExecutableBranchNodeStatus_GetPhase_Call) RunAndReturn(run func() v1alpha1.BranchNodePhase) *ExecutableBranchNodeStatus_GetPhase_Call {
	_c.Call.Return(run)
	return _c
}

// NewExecutableBranchNodeStatus creates a new instance of ExecutableBranchNodeStatus. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewExecutableBranchNodeStatus(t interface {
	mock.TestingT
	Cleanup(func())
}) *ExecutableBranchNodeStatus {
	mock := &ExecutableBranchNodeStatus{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
