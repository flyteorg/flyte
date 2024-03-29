// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	handler "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"

	mock "github.com/stretchr/testify/mock"
)

// NodeStateReader is an autogenerated mock type for the NodeStateReader type
type NodeStateReader struct {
	mock.Mock
}

type NodeStateReader_GetArrayNodeState struct {
	*mock.Call
}

func (_m NodeStateReader_GetArrayNodeState) Return(_a0 handler.ArrayNodeState) *NodeStateReader_GetArrayNodeState {
	return &NodeStateReader_GetArrayNodeState{Call: _m.Call.Return(_a0)}
}

func (_m *NodeStateReader) OnGetArrayNodeState() *NodeStateReader_GetArrayNodeState {
	c_call := _m.On("GetArrayNodeState")
	return &NodeStateReader_GetArrayNodeState{Call: c_call}
}

func (_m *NodeStateReader) OnGetArrayNodeStateMatch(matchers ...interface{}) *NodeStateReader_GetArrayNodeState {
	c_call := _m.On("GetArrayNodeState", matchers...)
	return &NodeStateReader_GetArrayNodeState{Call: c_call}
}

// GetArrayNodeState provides a mock function with given fields:
func (_m *NodeStateReader) GetArrayNodeState() handler.ArrayNodeState {
	ret := _m.Called()

	var r0 handler.ArrayNodeState
	if rf, ok := ret.Get(0).(func() handler.ArrayNodeState); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(handler.ArrayNodeState)
	}

	return r0
}

type NodeStateReader_GetBranchNodeState struct {
	*mock.Call
}

func (_m NodeStateReader_GetBranchNodeState) Return(_a0 handler.BranchNodeState) *NodeStateReader_GetBranchNodeState {
	return &NodeStateReader_GetBranchNodeState{Call: _m.Call.Return(_a0)}
}

func (_m *NodeStateReader) OnGetBranchNodeState() *NodeStateReader_GetBranchNodeState {
	c_call := _m.On("GetBranchNodeState")
	return &NodeStateReader_GetBranchNodeState{Call: c_call}
}

func (_m *NodeStateReader) OnGetBranchNodeStateMatch(matchers ...interface{}) *NodeStateReader_GetBranchNodeState {
	c_call := _m.On("GetBranchNodeState", matchers...)
	return &NodeStateReader_GetBranchNodeState{Call: c_call}
}

// GetBranchNodeState provides a mock function with given fields:
func (_m *NodeStateReader) GetBranchNodeState() handler.BranchNodeState {
	ret := _m.Called()

	var r0 handler.BranchNodeState
	if rf, ok := ret.Get(0).(func() handler.BranchNodeState); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(handler.BranchNodeState)
	}

	return r0
}

type NodeStateReader_GetDynamicNodeState struct {
	*mock.Call
}

func (_m NodeStateReader_GetDynamicNodeState) Return(_a0 handler.DynamicNodeState) *NodeStateReader_GetDynamicNodeState {
	return &NodeStateReader_GetDynamicNodeState{Call: _m.Call.Return(_a0)}
}

func (_m *NodeStateReader) OnGetDynamicNodeState() *NodeStateReader_GetDynamicNodeState {
	c_call := _m.On("GetDynamicNodeState")
	return &NodeStateReader_GetDynamicNodeState{Call: c_call}
}

func (_m *NodeStateReader) OnGetDynamicNodeStateMatch(matchers ...interface{}) *NodeStateReader_GetDynamicNodeState {
	c_call := _m.On("GetDynamicNodeState", matchers...)
	return &NodeStateReader_GetDynamicNodeState{Call: c_call}
}

// GetDynamicNodeState provides a mock function with given fields:
func (_m *NodeStateReader) GetDynamicNodeState() handler.DynamicNodeState {
	ret := _m.Called()

	var r0 handler.DynamicNodeState
	if rf, ok := ret.Get(0).(func() handler.DynamicNodeState); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(handler.DynamicNodeState)
	}

	return r0
}

type NodeStateReader_GetGateNodeState struct {
	*mock.Call
}

func (_m NodeStateReader_GetGateNodeState) Return(_a0 handler.GateNodeState) *NodeStateReader_GetGateNodeState {
	return &NodeStateReader_GetGateNodeState{Call: _m.Call.Return(_a0)}
}

func (_m *NodeStateReader) OnGetGateNodeState() *NodeStateReader_GetGateNodeState {
	c_call := _m.On("GetGateNodeState")
	return &NodeStateReader_GetGateNodeState{Call: c_call}
}

func (_m *NodeStateReader) OnGetGateNodeStateMatch(matchers ...interface{}) *NodeStateReader_GetGateNodeState {
	c_call := _m.On("GetGateNodeState", matchers...)
	return &NodeStateReader_GetGateNodeState{Call: c_call}
}

// GetGateNodeState provides a mock function with given fields:
func (_m *NodeStateReader) GetGateNodeState() handler.GateNodeState {
	ret := _m.Called()

	var r0 handler.GateNodeState
	if rf, ok := ret.Get(0).(func() handler.GateNodeState); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(handler.GateNodeState)
	}

	return r0
}

type NodeStateReader_GetTaskNodeState struct {
	*mock.Call
}

func (_m NodeStateReader_GetTaskNodeState) Return(_a0 handler.TaskNodeState) *NodeStateReader_GetTaskNodeState {
	return &NodeStateReader_GetTaskNodeState{Call: _m.Call.Return(_a0)}
}

func (_m *NodeStateReader) OnGetTaskNodeState() *NodeStateReader_GetTaskNodeState {
	c_call := _m.On("GetTaskNodeState")
	return &NodeStateReader_GetTaskNodeState{Call: c_call}
}

func (_m *NodeStateReader) OnGetTaskNodeStateMatch(matchers ...interface{}) *NodeStateReader_GetTaskNodeState {
	c_call := _m.On("GetTaskNodeState", matchers...)
	return &NodeStateReader_GetTaskNodeState{Call: c_call}
}

// GetTaskNodeState provides a mock function with given fields:
func (_m *NodeStateReader) GetTaskNodeState() handler.TaskNodeState {
	ret := _m.Called()

	var r0 handler.TaskNodeState
	if rf, ok := ret.Get(0).(func() handler.TaskNodeState); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(handler.TaskNodeState)
	}

	return r0
}

type NodeStateReader_GetWorkflowNodeState struct {
	*mock.Call
}

func (_m NodeStateReader_GetWorkflowNodeState) Return(_a0 handler.WorkflowNodeState) *NodeStateReader_GetWorkflowNodeState {
	return &NodeStateReader_GetWorkflowNodeState{Call: _m.Call.Return(_a0)}
}

func (_m *NodeStateReader) OnGetWorkflowNodeState() *NodeStateReader_GetWorkflowNodeState {
	c_call := _m.On("GetWorkflowNodeState")
	return &NodeStateReader_GetWorkflowNodeState{Call: c_call}
}

func (_m *NodeStateReader) OnGetWorkflowNodeStateMatch(matchers ...interface{}) *NodeStateReader_GetWorkflowNodeState {
	c_call := _m.On("GetWorkflowNodeState", matchers...)
	return &NodeStateReader_GetWorkflowNodeState{Call: c_call}
}

// GetWorkflowNodeState provides a mock function with given fields:
func (_m *NodeStateReader) GetWorkflowNodeState() handler.WorkflowNodeState {
	ret := _m.Called()

	var r0 handler.WorkflowNodeState
	if rf, ok := ret.Get(0).(func() handler.WorkflowNodeState); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(handler.WorkflowNodeState)
	}

	return r0
}

type NodeStateReader_HasArrayNodeState struct {
	*mock.Call
}

func (_m NodeStateReader_HasArrayNodeState) Return(_a0 bool) *NodeStateReader_HasArrayNodeState {
	return &NodeStateReader_HasArrayNodeState{Call: _m.Call.Return(_a0)}
}

func (_m *NodeStateReader) OnHasArrayNodeState() *NodeStateReader_HasArrayNodeState {
	c_call := _m.On("HasArrayNodeState")
	return &NodeStateReader_HasArrayNodeState{Call: c_call}
}

func (_m *NodeStateReader) OnHasArrayNodeStateMatch(matchers ...interface{}) *NodeStateReader_HasArrayNodeState {
	c_call := _m.On("HasArrayNodeState", matchers...)
	return &NodeStateReader_HasArrayNodeState{Call: c_call}
}

// HasArrayNodeState provides a mock function with given fields:
func (_m *NodeStateReader) HasArrayNodeState() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

type NodeStateReader_HasBranchNodeState struct {
	*mock.Call
}

func (_m NodeStateReader_HasBranchNodeState) Return(_a0 bool) *NodeStateReader_HasBranchNodeState {
	return &NodeStateReader_HasBranchNodeState{Call: _m.Call.Return(_a0)}
}

func (_m *NodeStateReader) OnHasBranchNodeState() *NodeStateReader_HasBranchNodeState {
	c_call := _m.On("HasBranchNodeState")
	return &NodeStateReader_HasBranchNodeState{Call: c_call}
}

func (_m *NodeStateReader) OnHasBranchNodeStateMatch(matchers ...interface{}) *NodeStateReader_HasBranchNodeState {
	c_call := _m.On("HasBranchNodeState", matchers...)
	return &NodeStateReader_HasBranchNodeState{Call: c_call}
}

// HasBranchNodeState provides a mock function with given fields:
func (_m *NodeStateReader) HasBranchNodeState() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

type NodeStateReader_HasDynamicNodeState struct {
	*mock.Call
}

func (_m NodeStateReader_HasDynamicNodeState) Return(_a0 bool) *NodeStateReader_HasDynamicNodeState {
	return &NodeStateReader_HasDynamicNodeState{Call: _m.Call.Return(_a0)}
}

func (_m *NodeStateReader) OnHasDynamicNodeState() *NodeStateReader_HasDynamicNodeState {
	c_call := _m.On("HasDynamicNodeState")
	return &NodeStateReader_HasDynamicNodeState{Call: c_call}
}

func (_m *NodeStateReader) OnHasDynamicNodeStateMatch(matchers ...interface{}) *NodeStateReader_HasDynamicNodeState {
	c_call := _m.On("HasDynamicNodeState", matchers...)
	return &NodeStateReader_HasDynamicNodeState{Call: c_call}
}

// HasDynamicNodeState provides a mock function with given fields:
func (_m *NodeStateReader) HasDynamicNodeState() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

type NodeStateReader_HasGateNodeState struct {
	*mock.Call
}

func (_m NodeStateReader_HasGateNodeState) Return(_a0 bool) *NodeStateReader_HasGateNodeState {
	return &NodeStateReader_HasGateNodeState{Call: _m.Call.Return(_a0)}
}

func (_m *NodeStateReader) OnHasGateNodeState() *NodeStateReader_HasGateNodeState {
	c_call := _m.On("HasGateNodeState")
	return &NodeStateReader_HasGateNodeState{Call: c_call}
}

func (_m *NodeStateReader) OnHasGateNodeStateMatch(matchers ...interface{}) *NodeStateReader_HasGateNodeState {
	c_call := _m.On("HasGateNodeState", matchers...)
	return &NodeStateReader_HasGateNodeState{Call: c_call}
}

// HasGateNodeState provides a mock function with given fields:
func (_m *NodeStateReader) HasGateNodeState() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

type NodeStateReader_HasTaskNodeState struct {
	*mock.Call
}

func (_m NodeStateReader_HasTaskNodeState) Return(_a0 bool) *NodeStateReader_HasTaskNodeState {
	return &NodeStateReader_HasTaskNodeState{Call: _m.Call.Return(_a0)}
}

func (_m *NodeStateReader) OnHasTaskNodeState() *NodeStateReader_HasTaskNodeState {
	c_call := _m.On("HasTaskNodeState")
	return &NodeStateReader_HasTaskNodeState{Call: c_call}
}

func (_m *NodeStateReader) OnHasTaskNodeStateMatch(matchers ...interface{}) *NodeStateReader_HasTaskNodeState {
	c_call := _m.On("HasTaskNodeState", matchers...)
	return &NodeStateReader_HasTaskNodeState{Call: c_call}
}

// HasTaskNodeState provides a mock function with given fields:
func (_m *NodeStateReader) HasTaskNodeState() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

type NodeStateReader_HasWorkflowNodeState struct {
	*mock.Call
}

func (_m NodeStateReader_HasWorkflowNodeState) Return(_a0 bool) *NodeStateReader_HasWorkflowNodeState {
	return &NodeStateReader_HasWorkflowNodeState{Call: _m.Call.Return(_a0)}
}

func (_m *NodeStateReader) OnHasWorkflowNodeState() *NodeStateReader_HasWorkflowNodeState {
	c_call := _m.On("HasWorkflowNodeState")
	return &NodeStateReader_HasWorkflowNodeState{Call: c_call}
}

func (_m *NodeStateReader) OnHasWorkflowNodeStateMatch(matchers ...interface{}) *NodeStateReader_HasWorkflowNodeState {
	c_call := _m.On("HasWorkflowNodeState", matchers...)
	return &NodeStateReader_HasWorkflowNodeState{Call: c_call}
}

// HasWorkflowNodeState provides a mock function with given fields:
func (_m *NodeStateReader) HasWorkflowNodeState() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
