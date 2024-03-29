// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	v1alpha1 "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	mock "github.com/stretchr/testify/mock"
)

// ExecutableBranchNodeStatus is an autogenerated mock type for the ExecutableBranchNodeStatus type
type ExecutableBranchNodeStatus struct {
	mock.Mock
}

type ExecutableBranchNodeStatus_GetFinalizedNode struct {
	*mock.Call
}

func (_m ExecutableBranchNodeStatus_GetFinalizedNode) Return(_a0 *string) *ExecutableBranchNodeStatus_GetFinalizedNode {
	return &ExecutableBranchNodeStatus_GetFinalizedNode{Call: _m.Call.Return(_a0)}
}

func (_m *ExecutableBranchNodeStatus) OnGetFinalizedNode() *ExecutableBranchNodeStatus_GetFinalizedNode {
	c_call := _m.On("GetFinalizedNode")
	return &ExecutableBranchNodeStatus_GetFinalizedNode{Call: c_call}
}

func (_m *ExecutableBranchNodeStatus) OnGetFinalizedNodeMatch(matchers ...interface{}) *ExecutableBranchNodeStatus_GetFinalizedNode {
	c_call := _m.On("GetFinalizedNode", matchers...)
	return &ExecutableBranchNodeStatus_GetFinalizedNode{Call: c_call}
}

// GetFinalizedNode provides a mock function with given fields:
func (_m *ExecutableBranchNodeStatus) GetFinalizedNode() *string {
	ret := _m.Called()

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

type ExecutableBranchNodeStatus_GetPhase struct {
	*mock.Call
}

func (_m ExecutableBranchNodeStatus_GetPhase) Return(_a0 v1alpha1.BranchNodePhase) *ExecutableBranchNodeStatus_GetPhase {
	return &ExecutableBranchNodeStatus_GetPhase{Call: _m.Call.Return(_a0)}
}

func (_m *ExecutableBranchNodeStatus) OnGetPhase() *ExecutableBranchNodeStatus_GetPhase {
	c_call := _m.On("GetPhase")
	return &ExecutableBranchNodeStatus_GetPhase{Call: c_call}
}

func (_m *ExecutableBranchNodeStatus) OnGetPhaseMatch(matchers ...interface{}) *ExecutableBranchNodeStatus_GetPhase {
	c_call := _m.On("GetPhase", matchers...)
	return &ExecutableBranchNodeStatus_GetPhase{Call: c_call}
}

// GetPhase provides a mock function with given fields:
func (_m *ExecutableBranchNodeStatus) GetPhase() v1alpha1.BranchNodePhase {
	ret := _m.Called()

	var r0 v1alpha1.BranchNodePhase
	if rf, ok := ret.Get(0).(func() v1alpha1.BranchNodePhase); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(v1alpha1.BranchNodePhase)
	}

	return r0
}
