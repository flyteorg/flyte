// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	core "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	bitarray "github.com/flyteorg/flyte/flytestdlib/bitarray"

	mock "github.com/stretchr/testify/mock"

	v1alpha1 "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

// ExecutableArrayNodeStatus is an autogenerated mock type for the ExecutableArrayNodeStatus type
type ExecutableArrayNodeStatus struct {
	mock.Mock
}

type ExecutableArrayNodeStatus_GetArrayNodePhase struct {
	*mock.Call
}

func (_m ExecutableArrayNodeStatus_GetArrayNodePhase) Return(_a0 v1alpha1.ArrayNodePhase) *ExecutableArrayNodeStatus_GetArrayNodePhase {
	return &ExecutableArrayNodeStatus_GetArrayNodePhase{Call: _m.Call.Return(_a0)}
}

func (_m *ExecutableArrayNodeStatus) OnGetArrayNodePhase() *ExecutableArrayNodeStatus_GetArrayNodePhase {
	c_call := _m.On("GetArrayNodePhase")
	return &ExecutableArrayNodeStatus_GetArrayNodePhase{Call: c_call}
}

func (_m *ExecutableArrayNodeStatus) OnGetArrayNodePhaseMatch(matchers ...interface{}) *ExecutableArrayNodeStatus_GetArrayNodePhase {
	c_call := _m.On("GetArrayNodePhase", matchers...)
	return &ExecutableArrayNodeStatus_GetArrayNodePhase{Call: c_call}
}

// GetArrayNodePhase provides a mock function with given fields:
func (_m *ExecutableArrayNodeStatus) GetArrayNodePhase() v1alpha1.ArrayNodePhase {
	ret := _m.Called()

	var r0 v1alpha1.ArrayNodePhase
	if rf, ok := ret.Get(0).(func() v1alpha1.ArrayNodePhase); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(v1alpha1.ArrayNodePhase)
	}

	return r0
}

type ExecutableArrayNodeStatus_GetExecutionError struct {
	*mock.Call
}

func (_m ExecutableArrayNodeStatus_GetExecutionError) Return(_a0 *core.ExecutionError) *ExecutableArrayNodeStatus_GetExecutionError {
	return &ExecutableArrayNodeStatus_GetExecutionError{Call: _m.Call.Return(_a0)}
}

func (_m *ExecutableArrayNodeStatus) OnGetExecutionError() *ExecutableArrayNodeStatus_GetExecutionError {
	c_call := _m.On("GetExecutionError")
	return &ExecutableArrayNodeStatus_GetExecutionError{Call: c_call}
}

func (_m *ExecutableArrayNodeStatus) OnGetExecutionErrorMatch(matchers ...interface{}) *ExecutableArrayNodeStatus_GetExecutionError {
	c_call := _m.On("GetExecutionError", matchers...)
	return &ExecutableArrayNodeStatus_GetExecutionError{Call: c_call}
}

// GetExecutionError provides a mock function with given fields:
func (_m *ExecutableArrayNodeStatus) GetExecutionError() *core.ExecutionError {
	ret := _m.Called()

	var r0 *core.ExecutionError
	if rf, ok := ret.Get(0).(func() *core.ExecutionError); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.ExecutionError)
		}
	}

	return r0
}

type ExecutableArrayNodeStatus_GetSubNodePhases struct {
	*mock.Call
}

func (_m ExecutableArrayNodeStatus_GetSubNodePhases) Return(_a0 bitarray.CompactArray) *ExecutableArrayNodeStatus_GetSubNodePhases {
	return &ExecutableArrayNodeStatus_GetSubNodePhases{Call: _m.Call.Return(_a0)}
}

func (_m *ExecutableArrayNodeStatus) OnGetSubNodePhases() *ExecutableArrayNodeStatus_GetSubNodePhases {
	c_call := _m.On("GetSubNodePhases")
	return &ExecutableArrayNodeStatus_GetSubNodePhases{Call: c_call}
}

func (_m *ExecutableArrayNodeStatus) OnGetSubNodePhasesMatch(matchers ...interface{}) *ExecutableArrayNodeStatus_GetSubNodePhases {
	c_call := _m.On("GetSubNodePhases", matchers...)
	return &ExecutableArrayNodeStatus_GetSubNodePhases{Call: c_call}
}

// GetSubNodePhases provides a mock function with given fields:
func (_m *ExecutableArrayNodeStatus) GetSubNodePhases() bitarray.CompactArray {
	ret := _m.Called()

	var r0 bitarray.CompactArray
	if rf, ok := ret.Get(0).(func() bitarray.CompactArray); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bitarray.CompactArray)
	}

	return r0
}

type ExecutableArrayNodeStatus_GetSubNodeRetryAttempts struct {
	*mock.Call
}

func (_m ExecutableArrayNodeStatus_GetSubNodeRetryAttempts) Return(_a0 bitarray.CompactArray) *ExecutableArrayNodeStatus_GetSubNodeRetryAttempts {
	return &ExecutableArrayNodeStatus_GetSubNodeRetryAttempts{Call: _m.Call.Return(_a0)}
}

func (_m *ExecutableArrayNodeStatus) OnGetSubNodeRetryAttempts() *ExecutableArrayNodeStatus_GetSubNodeRetryAttempts {
	c_call := _m.On("GetSubNodeRetryAttempts")
	return &ExecutableArrayNodeStatus_GetSubNodeRetryAttempts{Call: c_call}
}

func (_m *ExecutableArrayNodeStatus) OnGetSubNodeRetryAttemptsMatch(matchers ...interface{}) *ExecutableArrayNodeStatus_GetSubNodeRetryAttempts {
	c_call := _m.On("GetSubNodeRetryAttempts", matchers...)
	return &ExecutableArrayNodeStatus_GetSubNodeRetryAttempts{Call: c_call}
}

// GetSubNodeRetryAttempts provides a mock function with given fields:
func (_m *ExecutableArrayNodeStatus) GetSubNodeRetryAttempts() bitarray.CompactArray {
	ret := _m.Called()

	var r0 bitarray.CompactArray
	if rf, ok := ret.Get(0).(func() bitarray.CompactArray); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bitarray.CompactArray)
	}

	return r0
}

type ExecutableArrayNodeStatus_GetSubNodeSystemFailures struct {
	*mock.Call
}

func (_m ExecutableArrayNodeStatus_GetSubNodeSystemFailures) Return(_a0 bitarray.CompactArray) *ExecutableArrayNodeStatus_GetSubNodeSystemFailures {
	return &ExecutableArrayNodeStatus_GetSubNodeSystemFailures{Call: _m.Call.Return(_a0)}
}

func (_m *ExecutableArrayNodeStatus) OnGetSubNodeSystemFailures() *ExecutableArrayNodeStatus_GetSubNodeSystemFailures {
	c_call := _m.On("GetSubNodeSystemFailures")
	return &ExecutableArrayNodeStatus_GetSubNodeSystemFailures{Call: c_call}
}

func (_m *ExecutableArrayNodeStatus) OnGetSubNodeSystemFailuresMatch(matchers ...interface{}) *ExecutableArrayNodeStatus_GetSubNodeSystemFailures {
	c_call := _m.On("GetSubNodeSystemFailures", matchers...)
	return &ExecutableArrayNodeStatus_GetSubNodeSystemFailures{Call: c_call}
}

// GetSubNodeSystemFailures provides a mock function with given fields:
func (_m *ExecutableArrayNodeStatus) GetSubNodeSystemFailures() bitarray.CompactArray {
	ret := _m.Called()

	var r0 bitarray.CompactArray
	if rf, ok := ret.Get(0).(func() bitarray.CompactArray); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bitarray.CompactArray)
	}

	return r0
}

type ExecutableArrayNodeStatus_GetSubNodeTaskPhases struct {
	*mock.Call
}

func (_m ExecutableArrayNodeStatus_GetSubNodeTaskPhases) Return(_a0 bitarray.CompactArray) *ExecutableArrayNodeStatus_GetSubNodeTaskPhases {
	return &ExecutableArrayNodeStatus_GetSubNodeTaskPhases{Call: _m.Call.Return(_a0)}
}

func (_m *ExecutableArrayNodeStatus) OnGetSubNodeTaskPhases() *ExecutableArrayNodeStatus_GetSubNodeTaskPhases {
	c_call := _m.On("GetSubNodeTaskPhases")
	return &ExecutableArrayNodeStatus_GetSubNodeTaskPhases{Call: c_call}
}

func (_m *ExecutableArrayNodeStatus) OnGetSubNodeTaskPhasesMatch(matchers ...interface{}) *ExecutableArrayNodeStatus_GetSubNodeTaskPhases {
	c_call := _m.On("GetSubNodeTaskPhases", matchers...)
	return &ExecutableArrayNodeStatus_GetSubNodeTaskPhases{Call: c_call}
}

// GetSubNodeTaskPhases provides a mock function with given fields:
func (_m *ExecutableArrayNodeStatus) GetSubNodeTaskPhases() bitarray.CompactArray {
	ret := _m.Called()

	var r0 bitarray.CompactArray
	if rf, ok := ret.Get(0).(func() bitarray.CompactArray); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bitarray.CompactArray)
	}

	return r0
}

type ExecutableArrayNodeStatus_GetTaskPhaseVersion struct {
	*mock.Call
}

func (_m ExecutableArrayNodeStatus_GetTaskPhaseVersion) Return(_a0 uint32) *ExecutableArrayNodeStatus_GetTaskPhaseVersion {
	return &ExecutableArrayNodeStatus_GetTaskPhaseVersion{Call: _m.Call.Return(_a0)}
}

func (_m *ExecutableArrayNodeStatus) OnGetTaskPhaseVersion() *ExecutableArrayNodeStatus_GetTaskPhaseVersion {
	c_call := _m.On("GetTaskPhaseVersion")
	return &ExecutableArrayNodeStatus_GetTaskPhaseVersion{Call: c_call}
}

func (_m *ExecutableArrayNodeStatus) OnGetTaskPhaseVersionMatch(matchers ...interface{}) *ExecutableArrayNodeStatus_GetTaskPhaseVersion {
	c_call := _m.On("GetTaskPhaseVersion", matchers...)
	return &ExecutableArrayNodeStatus_GetTaskPhaseVersion{Call: c_call}
}

// GetTaskPhaseVersion provides a mock function with given fields:
func (_m *ExecutableArrayNodeStatus) GetTaskPhaseVersion() uint32 {
	ret := _m.Called()

	var r0 uint32
	if rf, ok := ret.Get(0).(func() uint32); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint32)
	}

	return r0
}
