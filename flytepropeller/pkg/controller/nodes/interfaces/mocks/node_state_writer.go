// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	handler "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"

	mock "github.com/stretchr/testify/mock"
)

// NodeStateWriter is an autogenerated mock type for the NodeStateWriter type
type NodeStateWriter struct {
	mock.Mock
}

// ClearNodeStatus provides a mock function with given fields:
func (_m *NodeStateWriter) ClearNodeStatus() {
	_m.Called()
}

type NodeStateWriter_PutArrayNodeState struct {
	*mock.Call
}

func (_m NodeStateWriter_PutArrayNodeState) Return(_a0 error) *NodeStateWriter_PutArrayNodeState {
	return &NodeStateWriter_PutArrayNodeState{Call: _m.Call.Return(_a0)}
}

func (_m *NodeStateWriter) OnPutArrayNodeState(s handler.ArrayNodeState) *NodeStateWriter_PutArrayNodeState {
	c_call := _m.On("PutArrayNodeState", s)
	return &NodeStateWriter_PutArrayNodeState{Call: c_call}
}

func (_m *NodeStateWriter) OnPutArrayNodeStateMatch(matchers ...interface{}) *NodeStateWriter_PutArrayNodeState {
	c_call := _m.On("PutArrayNodeState", matchers...)
	return &NodeStateWriter_PutArrayNodeState{Call: c_call}
}

// PutArrayNodeState provides a mock function with given fields: s
func (_m *NodeStateWriter) PutArrayNodeState(s handler.ArrayNodeState) error {
	ret := _m.Called(s)

	var r0 error
	if rf, ok := ret.Get(0).(func(handler.ArrayNodeState) error); ok {
		r0 = rf(s)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type NodeStateWriter_PutBranchNode struct {
	*mock.Call
}

func (_m NodeStateWriter_PutBranchNode) Return(_a0 error) *NodeStateWriter_PutBranchNode {
	return &NodeStateWriter_PutBranchNode{Call: _m.Call.Return(_a0)}
}

func (_m *NodeStateWriter) OnPutBranchNode(s handler.BranchNodeState) *NodeStateWriter_PutBranchNode {
	c_call := _m.On("PutBranchNode", s)
	return &NodeStateWriter_PutBranchNode{Call: c_call}
}

func (_m *NodeStateWriter) OnPutBranchNodeMatch(matchers ...interface{}) *NodeStateWriter_PutBranchNode {
	c_call := _m.On("PutBranchNode", matchers...)
	return &NodeStateWriter_PutBranchNode{Call: c_call}
}

// PutBranchNode provides a mock function with given fields: s
func (_m *NodeStateWriter) PutBranchNode(s handler.BranchNodeState) error {
	ret := _m.Called(s)

	var r0 error
	if rf, ok := ret.Get(0).(func(handler.BranchNodeState) error); ok {
		r0 = rf(s)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type NodeStateWriter_PutDynamicNodeState struct {
	*mock.Call
}

func (_m NodeStateWriter_PutDynamicNodeState) Return(_a0 error) *NodeStateWriter_PutDynamicNodeState {
	return &NodeStateWriter_PutDynamicNodeState{Call: _m.Call.Return(_a0)}
}

func (_m *NodeStateWriter) OnPutDynamicNodeState(s handler.DynamicNodeState) *NodeStateWriter_PutDynamicNodeState {
	c_call := _m.On("PutDynamicNodeState", s)
	return &NodeStateWriter_PutDynamicNodeState{Call: c_call}
}

func (_m *NodeStateWriter) OnPutDynamicNodeStateMatch(matchers ...interface{}) *NodeStateWriter_PutDynamicNodeState {
	c_call := _m.On("PutDynamicNodeState", matchers...)
	return &NodeStateWriter_PutDynamicNodeState{Call: c_call}
}

// PutDynamicNodeState provides a mock function with given fields: s
func (_m *NodeStateWriter) PutDynamicNodeState(s handler.DynamicNodeState) error {
	ret := _m.Called(s)

	var r0 error
	if rf, ok := ret.Get(0).(func(handler.DynamicNodeState) error); ok {
		r0 = rf(s)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type NodeStateWriter_PutGateNodeState struct {
	*mock.Call
}

func (_m NodeStateWriter_PutGateNodeState) Return(_a0 error) *NodeStateWriter_PutGateNodeState {
	return &NodeStateWriter_PutGateNodeState{Call: _m.Call.Return(_a0)}
}

func (_m *NodeStateWriter) OnPutGateNodeState(s handler.GateNodeState) *NodeStateWriter_PutGateNodeState {
	c_call := _m.On("PutGateNodeState", s)
	return &NodeStateWriter_PutGateNodeState{Call: c_call}
}

func (_m *NodeStateWriter) OnPutGateNodeStateMatch(matchers ...interface{}) *NodeStateWriter_PutGateNodeState {
	c_call := _m.On("PutGateNodeState", matchers...)
	return &NodeStateWriter_PutGateNodeState{Call: c_call}
}

// PutGateNodeState provides a mock function with given fields: s
func (_m *NodeStateWriter) PutGateNodeState(s handler.GateNodeState) error {
	ret := _m.Called(s)

	var r0 error
	if rf, ok := ret.Get(0).(func(handler.GateNodeState) error); ok {
		r0 = rf(s)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type NodeStateWriter_PutTaskNodeState struct {
	*mock.Call
}

func (_m NodeStateWriter_PutTaskNodeState) Return(_a0 error) *NodeStateWriter_PutTaskNodeState {
	return &NodeStateWriter_PutTaskNodeState{Call: _m.Call.Return(_a0)}
}

func (_m *NodeStateWriter) OnPutTaskNodeState(s handler.TaskNodeState) *NodeStateWriter_PutTaskNodeState {
	c_call := _m.On("PutTaskNodeState", s)
	return &NodeStateWriter_PutTaskNodeState{Call: c_call}
}

func (_m *NodeStateWriter) OnPutTaskNodeStateMatch(matchers ...interface{}) *NodeStateWriter_PutTaskNodeState {
	c_call := _m.On("PutTaskNodeState", matchers...)
	return &NodeStateWriter_PutTaskNodeState{Call: c_call}
}

// PutTaskNodeState provides a mock function with given fields: s
func (_m *NodeStateWriter) PutTaskNodeState(s handler.TaskNodeState) error {
	ret := _m.Called(s)

	var r0 error
	if rf, ok := ret.Get(0).(func(handler.TaskNodeState) error); ok {
		r0 = rf(s)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type NodeStateWriter_PutWorkflowNodeState struct {
	*mock.Call
}

func (_m NodeStateWriter_PutWorkflowNodeState) Return(_a0 error) *NodeStateWriter_PutWorkflowNodeState {
	return &NodeStateWriter_PutWorkflowNodeState{Call: _m.Call.Return(_a0)}
}

func (_m *NodeStateWriter) OnPutWorkflowNodeState(s handler.WorkflowNodeState) *NodeStateWriter_PutWorkflowNodeState {
	c_call := _m.On("PutWorkflowNodeState", s)
	return &NodeStateWriter_PutWorkflowNodeState{Call: c_call}
}

func (_m *NodeStateWriter) OnPutWorkflowNodeStateMatch(matchers ...interface{}) *NodeStateWriter_PutWorkflowNodeState {
	c_call := _m.On("PutWorkflowNodeState", matchers...)
	return &NodeStateWriter_PutWorkflowNodeState{Call: c_call}
}

// PutWorkflowNodeState provides a mock function with given fields: s
func (_m *NodeStateWriter) PutWorkflowNodeState(s handler.WorkflowNodeState) error {
	ret := _m.Called(s)

	var r0 error
	if rf, ok := ret.Get(0).(func(handler.WorkflowNodeState) error); ok {
		r0 = rf(s)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
