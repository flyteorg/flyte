// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	v1alpha1 "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	mock "github.com/stretchr/testify/mock"
)

// ExecutableSubWorkflowNodeStatus is an autogenerated mock type for the ExecutableSubWorkflowNodeStatus type
type ExecutableSubWorkflowNodeStatus struct {
	mock.Mock
}

type ExecutableSubWorkflowNodeStatus_Expecter struct {
	mock *mock.Mock
}

func (_m *ExecutableSubWorkflowNodeStatus) EXPECT() *ExecutableSubWorkflowNodeStatus_Expecter {
	return &ExecutableSubWorkflowNodeStatus_Expecter{mock: &_m.Mock}
}

// GetPhase provides a mock function with no fields
func (_m *ExecutableSubWorkflowNodeStatus) GetPhase() v1alpha1.WorkflowPhase {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetPhase")
	}

	var r0 v1alpha1.WorkflowPhase
	if rf, ok := ret.Get(0).(func() v1alpha1.WorkflowPhase); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(v1alpha1.WorkflowPhase)
	}

	return r0
}

// ExecutableSubWorkflowNodeStatus_GetPhase_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPhase'
type ExecutableSubWorkflowNodeStatus_GetPhase_Call struct {
	*mock.Call
}

// GetPhase is a helper method to define mock.On call
func (_e *ExecutableSubWorkflowNodeStatus_Expecter) GetPhase() *ExecutableSubWorkflowNodeStatus_GetPhase_Call {
	return &ExecutableSubWorkflowNodeStatus_GetPhase_Call{Call: _e.mock.On("GetPhase")}
}

func (_c *ExecutableSubWorkflowNodeStatus_GetPhase_Call) Run(run func()) *ExecutableSubWorkflowNodeStatus_GetPhase_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ExecutableSubWorkflowNodeStatus_GetPhase_Call) Return(_a0 v1alpha1.WorkflowPhase) *ExecutableSubWorkflowNodeStatus_GetPhase_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ExecutableSubWorkflowNodeStatus_GetPhase_Call) RunAndReturn(run func() v1alpha1.WorkflowPhase) *ExecutableSubWorkflowNodeStatus_GetPhase_Call {
	_c.Call.Return(run)
	return _c
}

// NewExecutableSubWorkflowNodeStatus creates a new instance of ExecutableSubWorkflowNodeStatus. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewExecutableSubWorkflowNodeStatus(t interface {
	mock.TestingT
	Cleanup(func())
}) *ExecutableSubWorkflowNodeStatus {
	mock := &ExecutableSubWorkflowNodeStatus{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
