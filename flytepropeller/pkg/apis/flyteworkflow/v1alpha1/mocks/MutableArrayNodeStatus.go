// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	core "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	bitarray "github.com/flyteorg/flyte/flytestdlib/bitarray"

	mock "github.com/stretchr/testify/mock"

	v1alpha1 "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

// MutableArrayNodeStatus is an autogenerated mock type for the MutableArrayNodeStatus type
type MutableArrayNodeStatus struct {
	mock.Mock
}

type MutableArrayNodeStatus_Expecter struct {
	mock *mock.Mock
}

func (_m *MutableArrayNodeStatus) EXPECT() *MutableArrayNodeStatus_Expecter {
	return &MutableArrayNodeStatus_Expecter{mock: &_m.Mock}
}

// GetArrayNodePhase provides a mock function with no fields
func (_m *MutableArrayNodeStatus) GetArrayNodePhase() v1alpha1.ArrayNodePhase {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetArrayNodePhase")
	}

	var r0 v1alpha1.ArrayNodePhase
	if rf, ok := ret.Get(0).(func() v1alpha1.ArrayNodePhase); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(v1alpha1.ArrayNodePhase)
	}

	return r0
}

// MutableArrayNodeStatus_GetArrayNodePhase_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetArrayNodePhase'
type MutableArrayNodeStatus_GetArrayNodePhase_Call struct {
	*mock.Call
}

// GetArrayNodePhase is a helper method to define mock.On call
func (_e *MutableArrayNodeStatus_Expecter) GetArrayNodePhase() *MutableArrayNodeStatus_GetArrayNodePhase_Call {
	return &MutableArrayNodeStatus_GetArrayNodePhase_Call{Call: _e.mock.On("GetArrayNodePhase")}
}

func (_c *MutableArrayNodeStatus_GetArrayNodePhase_Call) Run(run func()) *MutableArrayNodeStatus_GetArrayNodePhase_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MutableArrayNodeStatus_GetArrayNodePhase_Call) Return(_a0 v1alpha1.ArrayNodePhase) *MutableArrayNodeStatus_GetArrayNodePhase_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MutableArrayNodeStatus_GetArrayNodePhase_Call) RunAndReturn(run func() v1alpha1.ArrayNodePhase) *MutableArrayNodeStatus_GetArrayNodePhase_Call {
	_c.Call.Return(run)
	return _c
}

// GetExecutionError provides a mock function with no fields
func (_m *MutableArrayNodeStatus) GetExecutionError() *core.ExecutionError {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetExecutionError")
	}

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

// MutableArrayNodeStatus_GetExecutionError_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetExecutionError'
type MutableArrayNodeStatus_GetExecutionError_Call struct {
	*mock.Call
}

// GetExecutionError is a helper method to define mock.On call
func (_e *MutableArrayNodeStatus_Expecter) GetExecutionError() *MutableArrayNodeStatus_GetExecutionError_Call {
	return &MutableArrayNodeStatus_GetExecutionError_Call{Call: _e.mock.On("GetExecutionError")}
}

func (_c *MutableArrayNodeStatus_GetExecutionError_Call) Run(run func()) *MutableArrayNodeStatus_GetExecutionError_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MutableArrayNodeStatus_GetExecutionError_Call) Return(_a0 *core.ExecutionError) *MutableArrayNodeStatus_GetExecutionError_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MutableArrayNodeStatus_GetExecutionError_Call) RunAndReturn(run func() *core.ExecutionError) *MutableArrayNodeStatus_GetExecutionError_Call {
	_c.Call.Return(run)
	return _c
}

// GetSubNodeDeltaTimestamps provides a mock function with no fields
func (_m *MutableArrayNodeStatus) GetSubNodeDeltaTimestamps() bitarray.CompactArray {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetSubNodeDeltaTimestamps")
	}

	var r0 bitarray.CompactArray
	if rf, ok := ret.Get(0).(func() bitarray.CompactArray); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bitarray.CompactArray)
	}

	return r0
}

// MutableArrayNodeStatus_GetSubNodeDeltaTimestamps_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSubNodeDeltaTimestamps'
type MutableArrayNodeStatus_GetSubNodeDeltaTimestamps_Call struct {
	*mock.Call
}

// GetSubNodeDeltaTimestamps is a helper method to define mock.On call
func (_e *MutableArrayNodeStatus_Expecter) GetSubNodeDeltaTimestamps() *MutableArrayNodeStatus_GetSubNodeDeltaTimestamps_Call {
	return &MutableArrayNodeStatus_GetSubNodeDeltaTimestamps_Call{Call: _e.mock.On("GetSubNodeDeltaTimestamps")}
}

func (_c *MutableArrayNodeStatus_GetSubNodeDeltaTimestamps_Call) Run(run func()) *MutableArrayNodeStatus_GetSubNodeDeltaTimestamps_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MutableArrayNodeStatus_GetSubNodeDeltaTimestamps_Call) Return(_a0 bitarray.CompactArray) *MutableArrayNodeStatus_GetSubNodeDeltaTimestamps_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MutableArrayNodeStatus_GetSubNodeDeltaTimestamps_Call) RunAndReturn(run func() bitarray.CompactArray) *MutableArrayNodeStatus_GetSubNodeDeltaTimestamps_Call {
	_c.Call.Return(run)
	return _c
}

// GetSubNodePhases provides a mock function with no fields
func (_m *MutableArrayNodeStatus) GetSubNodePhases() bitarray.CompactArray {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetSubNodePhases")
	}

	var r0 bitarray.CompactArray
	if rf, ok := ret.Get(0).(func() bitarray.CompactArray); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bitarray.CompactArray)
	}

	return r0
}

// MutableArrayNodeStatus_GetSubNodePhases_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSubNodePhases'
type MutableArrayNodeStatus_GetSubNodePhases_Call struct {
	*mock.Call
}

// GetSubNodePhases is a helper method to define mock.On call
func (_e *MutableArrayNodeStatus_Expecter) GetSubNodePhases() *MutableArrayNodeStatus_GetSubNodePhases_Call {
	return &MutableArrayNodeStatus_GetSubNodePhases_Call{Call: _e.mock.On("GetSubNodePhases")}
}

func (_c *MutableArrayNodeStatus_GetSubNodePhases_Call) Run(run func()) *MutableArrayNodeStatus_GetSubNodePhases_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MutableArrayNodeStatus_GetSubNodePhases_Call) Return(_a0 bitarray.CompactArray) *MutableArrayNodeStatus_GetSubNodePhases_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MutableArrayNodeStatus_GetSubNodePhases_Call) RunAndReturn(run func() bitarray.CompactArray) *MutableArrayNodeStatus_GetSubNodePhases_Call {
	_c.Call.Return(run)
	return _c
}

// GetSubNodeRetryAttempts provides a mock function with no fields
func (_m *MutableArrayNodeStatus) GetSubNodeRetryAttempts() bitarray.CompactArray {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetSubNodeRetryAttempts")
	}

	var r0 bitarray.CompactArray
	if rf, ok := ret.Get(0).(func() bitarray.CompactArray); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bitarray.CompactArray)
	}

	return r0
}

// MutableArrayNodeStatus_GetSubNodeRetryAttempts_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSubNodeRetryAttempts'
type MutableArrayNodeStatus_GetSubNodeRetryAttempts_Call struct {
	*mock.Call
}

// GetSubNodeRetryAttempts is a helper method to define mock.On call
func (_e *MutableArrayNodeStatus_Expecter) GetSubNodeRetryAttempts() *MutableArrayNodeStatus_GetSubNodeRetryAttempts_Call {
	return &MutableArrayNodeStatus_GetSubNodeRetryAttempts_Call{Call: _e.mock.On("GetSubNodeRetryAttempts")}
}

func (_c *MutableArrayNodeStatus_GetSubNodeRetryAttempts_Call) Run(run func()) *MutableArrayNodeStatus_GetSubNodeRetryAttempts_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MutableArrayNodeStatus_GetSubNodeRetryAttempts_Call) Return(_a0 bitarray.CompactArray) *MutableArrayNodeStatus_GetSubNodeRetryAttempts_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MutableArrayNodeStatus_GetSubNodeRetryAttempts_Call) RunAndReturn(run func() bitarray.CompactArray) *MutableArrayNodeStatus_GetSubNodeRetryAttempts_Call {
	_c.Call.Return(run)
	return _c
}

// GetSubNodeSystemFailures provides a mock function with no fields
func (_m *MutableArrayNodeStatus) GetSubNodeSystemFailures() bitarray.CompactArray {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetSubNodeSystemFailures")
	}

	var r0 bitarray.CompactArray
	if rf, ok := ret.Get(0).(func() bitarray.CompactArray); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bitarray.CompactArray)
	}

	return r0
}

// MutableArrayNodeStatus_GetSubNodeSystemFailures_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSubNodeSystemFailures'
type MutableArrayNodeStatus_GetSubNodeSystemFailures_Call struct {
	*mock.Call
}

// GetSubNodeSystemFailures is a helper method to define mock.On call
func (_e *MutableArrayNodeStatus_Expecter) GetSubNodeSystemFailures() *MutableArrayNodeStatus_GetSubNodeSystemFailures_Call {
	return &MutableArrayNodeStatus_GetSubNodeSystemFailures_Call{Call: _e.mock.On("GetSubNodeSystemFailures")}
}

func (_c *MutableArrayNodeStatus_GetSubNodeSystemFailures_Call) Run(run func()) *MutableArrayNodeStatus_GetSubNodeSystemFailures_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MutableArrayNodeStatus_GetSubNodeSystemFailures_Call) Return(_a0 bitarray.CompactArray) *MutableArrayNodeStatus_GetSubNodeSystemFailures_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MutableArrayNodeStatus_GetSubNodeSystemFailures_Call) RunAndReturn(run func() bitarray.CompactArray) *MutableArrayNodeStatus_GetSubNodeSystemFailures_Call {
	_c.Call.Return(run)
	return _c
}

// GetSubNodeTaskPhases provides a mock function with no fields
func (_m *MutableArrayNodeStatus) GetSubNodeTaskPhases() bitarray.CompactArray {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetSubNodeTaskPhases")
	}

	var r0 bitarray.CompactArray
	if rf, ok := ret.Get(0).(func() bitarray.CompactArray); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bitarray.CompactArray)
	}

	return r0
}

// MutableArrayNodeStatus_GetSubNodeTaskPhases_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSubNodeTaskPhases'
type MutableArrayNodeStatus_GetSubNodeTaskPhases_Call struct {
	*mock.Call
}

// GetSubNodeTaskPhases is a helper method to define mock.On call
func (_e *MutableArrayNodeStatus_Expecter) GetSubNodeTaskPhases() *MutableArrayNodeStatus_GetSubNodeTaskPhases_Call {
	return &MutableArrayNodeStatus_GetSubNodeTaskPhases_Call{Call: _e.mock.On("GetSubNodeTaskPhases")}
}

func (_c *MutableArrayNodeStatus_GetSubNodeTaskPhases_Call) Run(run func()) *MutableArrayNodeStatus_GetSubNodeTaskPhases_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MutableArrayNodeStatus_GetSubNodeTaskPhases_Call) Return(_a0 bitarray.CompactArray) *MutableArrayNodeStatus_GetSubNodeTaskPhases_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MutableArrayNodeStatus_GetSubNodeTaskPhases_Call) RunAndReturn(run func() bitarray.CompactArray) *MutableArrayNodeStatus_GetSubNodeTaskPhases_Call {
	_c.Call.Return(run)
	return _c
}

// GetTaskPhaseVersion provides a mock function with no fields
func (_m *MutableArrayNodeStatus) GetTaskPhaseVersion() uint32 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetTaskPhaseVersion")
	}

	var r0 uint32
	if rf, ok := ret.Get(0).(func() uint32); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint32)
	}

	return r0
}

// MutableArrayNodeStatus_GetTaskPhaseVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTaskPhaseVersion'
type MutableArrayNodeStatus_GetTaskPhaseVersion_Call struct {
	*mock.Call
}

// GetTaskPhaseVersion is a helper method to define mock.On call
func (_e *MutableArrayNodeStatus_Expecter) GetTaskPhaseVersion() *MutableArrayNodeStatus_GetTaskPhaseVersion_Call {
	return &MutableArrayNodeStatus_GetTaskPhaseVersion_Call{Call: _e.mock.On("GetTaskPhaseVersion")}
}

func (_c *MutableArrayNodeStatus_GetTaskPhaseVersion_Call) Run(run func()) *MutableArrayNodeStatus_GetTaskPhaseVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MutableArrayNodeStatus_GetTaskPhaseVersion_Call) Return(_a0 uint32) *MutableArrayNodeStatus_GetTaskPhaseVersion_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MutableArrayNodeStatus_GetTaskPhaseVersion_Call) RunAndReturn(run func() uint32) *MutableArrayNodeStatus_GetTaskPhaseVersion_Call {
	_c.Call.Return(run)
	return _c
}

// IsDirty provides a mock function with no fields
func (_m *MutableArrayNodeStatus) IsDirty() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsDirty")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// MutableArrayNodeStatus_IsDirty_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsDirty'
type MutableArrayNodeStatus_IsDirty_Call struct {
	*mock.Call
}

// IsDirty is a helper method to define mock.On call
func (_e *MutableArrayNodeStatus_Expecter) IsDirty() *MutableArrayNodeStatus_IsDirty_Call {
	return &MutableArrayNodeStatus_IsDirty_Call{Call: _e.mock.On("IsDirty")}
}

func (_c *MutableArrayNodeStatus_IsDirty_Call) Run(run func()) *MutableArrayNodeStatus_IsDirty_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MutableArrayNodeStatus_IsDirty_Call) Return(_a0 bool) *MutableArrayNodeStatus_IsDirty_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MutableArrayNodeStatus_IsDirty_Call) RunAndReturn(run func() bool) *MutableArrayNodeStatus_IsDirty_Call {
	_c.Call.Return(run)
	return _c
}

// SetArrayNodePhase provides a mock function with given fields: phase
func (_m *MutableArrayNodeStatus) SetArrayNodePhase(phase v1alpha1.ArrayNodePhase) {
	_m.Called(phase)
}

// MutableArrayNodeStatus_SetArrayNodePhase_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetArrayNodePhase'
type MutableArrayNodeStatus_SetArrayNodePhase_Call struct {
	*mock.Call
}

// SetArrayNodePhase is a helper method to define mock.On call
//   - phase v1alpha1.ArrayNodePhase
func (_e *MutableArrayNodeStatus_Expecter) SetArrayNodePhase(phase interface{}) *MutableArrayNodeStatus_SetArrayNodePhase_Call {
	return &MutableArrayNodeStatus_SetArrayNodePhase_Call{Call: _e.mock.On("SetArrayNodePhase", phase)}
}

func (_c *MutableArrayNodeStatus_SetArrayNodePhase_Call) Run(run func(phase v1alpha1.ArrayNodePhase)) *MutableArrayNodeStatus_SetArrayNodePhase_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(v1alpha1.ArrayNodePhase))
	})
	return _c
}

func (_c *MutableArrayNodeStatus_SetArrayNodePhase_Call) Return() *MutableArrayNodeStatus_SetArrayNodePhase_Call {
	_c.Call.Return()
	return _c
}

func (_c *MutableArrayNodeStatus_SetArrayNodePhase_Call) RunAndReturn(run func(v1alpha1.ArrayNodePhase)) *MutableArrayNodeStatus_SetArrayNodePhase_Call {
	_c.Run(run)
	return _c
}

// SetExecutionError provides a mock function with given fields: executionError
func (_m *MutableArrayNodeStatus) SetExecutionError(executionError *core.ExecutionError) {
	_m.Called(executionError)
}

// MutableArrayNodeStatus_SetExecutionError_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetExecutionError'
type MutableArrayNodeStatus_SetExecutionError_Call struct {
	*mock.Call
}

// SetExecutionError is a helper method to define mock.On call
//   - executionError *core.ExecutionError
func (_e *MutableArrayNodeStatus_Expecter) SetExecutionError(executionError interface{}) *MutableArrayNodeStatus_SetExecutionError_Call {
	return &MutableArrayNodeStatus_SetExecutionError_Call{Call: _e.mock.On("SetExecutionError", executionError)}
}

func (_c *MutableArrayNodeStatus_SetExecutionError_Call) Run(run func(executionError *core.ExecutionError)) *MutableArrayNodeStatus_SetExecutionError_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*core.ExecutionError))
	})
	return _c
}

func (_c *MutableArrayNodeStatus_SetExecutionError_Call) Return() *MutableArrayNodeStatus_SetExecutionError_Call {
	_c.Call.Return()
	return _c
}

func (_c *MutableArrayNodeStatus_SetExecutionError_Call) RunAndReturn(run func(*core.ExecutionError)) *MutableArrayNodeStatus_SetExecutionError_Call {
	_c.Run(run)
	return _c
}

// SetSubNodeDeltaTimestamps provides a mock function with given fields: subNodeDeltaTimestamps
func (_m *MutableArrayNodeStatus) SetSubNodeDeltaTimestamps(subNodeDeltaTimestamps bitarray.CompactArray) {
	_m.Called(subNodeDeltaTimestamps)
}

// MutableArrayNodeStatus_SetSubNodeDeltaTimestamps_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetSubNodeDeltaTimestamps'
type MutableArrayNodeStatus_SetSubNodeDeltaTimestamps_Call struct {
	*mock.Call
}

// SetSubNodeDeltaTimestamps is a helper method to define mock.On call
//   - subNodeDeltaTimestamps bitarray.CompactArray
func (_e *MutableArrayNodeStatus_Expecter) SetSubNodeDeltaTimestamps(subNodeDeltaTimestamps interface{}) *MutableArrayNodeStatus_SetSubNodeDeltaTimestamps_Call {
	return &MutableArrayNodeStatus_SetSubNodeDeltaTimestamps_Call{Call: _e.mock.On("SetSubNodeDeltaTimestamps", subNodeDeltaTimestamps)}
}

func (_c *MutableArrayNodeStatus_SetSubNodeDeltaTimestamps_Call) Run(run func(subNodeDeltaTimestamps bitarray.CompactArray)) *MutableArrayNodeStatus_SetSubNodeDeltaTimestamps_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(bitarray.CompactArray))
	})
	return _c
}

func (_c *MutableArrayNodeStatus_SetSubNodeDeltaTimestamps_Call) Return() *MutableArrayNodeStatus_SetSubNodeDeltaTimestamps_Call {
	_c.Call.Return()
	return _c
}

func (_c *MutableArrayNodeStatus_SetSubNodeDeltaTimestamps_Call) RunAndReturn(run func(bitarray.CompactArray)) *MutableArrayNodeStatus_SetSubNodeDeltaTimestamps_Call {
	_c.Run(run)
	return _c
}

// SetSubNodePhases provides a mock function with given fields: subNodePhases
func (_m *MutableArrayNodeStatus) SetSubNodePhases(subNodePhases bitarray.CompactArray) {
	_m.Called(subNodePhases)
}

// MutableArrayNodeStatus_SetSubNodePhases_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetSubNodePhases'
type MutableArrayNodeStatus_SetSubNodePhases_Call struct {
	*mock.Call
}

// SetSubNodePhases is a helper method to define mock.On call
//   - subNodePhases bitarray.CompactArray
func (_e *MutableArrayNodeStatus_Expecter) SetSubNodePhases(subNodePhases interface{}) *MutableArrayNodeStatus_SetSubNodePhases_Call {
	return &MutableArrayNodeStatus_SetSubNodePhases_Call{Call: _e.mock.On("SetSubNodePhases", subNodePhases)}
}

func (_c *MutableArrayNodeStatus_SetSubNodePhases_Call) Run(run func(subNodePhases bitarray.CompactArray)) *MutableArrayNodeStatus_SetSubNodePhases_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(bitarray.CompactArray))
	})
	return _c
}

func (_c *MutableArrayNodeStatus_SetSubNodePhases_Call) Return() *MutableArrayNodeStatus_SetSubNodePhases_Call {
	_c.Call.Return()
	return _c
}

func (_c *MutableArrayNodeStatus_SetSubNodePhases_Call) RunAndReturn(run func(bitarray.CompactArray)) *MutableArrayNodeStatus_SetSubNodePhases_Call {
	_c.Run(run)
	return _c
}

// SetSubNodeRetryAttempts provides a mock function with given fields: subNodeRetryAttempts
func (_m *MutableArrayNodeStatus) SetSubNodeRetryAttempts(subNodeRetryAttempts bitarray.CompactArray) {
	_m.Called(subNodeRetryAttempts)
}

// MutableArrayNodeStatus_SetSubNodeRetryAttempts_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetSubNodeRetryAttempts'
type MutableArrayNodeStatus_SetSubNodeRetryAttempts_Call struct {
	*mock.Call
}

// SetSubNodeRetryAttempts is a helper method to define mock.On call
//   - subNodeRetryAttempts bitarray.CompactArray
func (_e *MutableArrayNodeStatus_Expecter) SetSubNodeRetryAttempts(subNodeRetryAttempts interface{}) *MutableArrayNodeStatus_SetSubNodeRetryAttempts_Call {
	return &MutableArrayNodeStatus_SetSubNodeRetryAttempts_Call{Call: _e.mock.On("SetSubNodeRetryAttempts", subNodeRetryAttempts)}
}

func (_c *MutableArrayNodeStatus_SetSubNodeRetryAttempts_Call) Run(run func(subNodeRetryAttempts bitarray.CompactArray)) *MutableArrayNodeStatus_SetSubNodeRetryAttempts_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(bitarray.CompactArray))
	})
	return _c
}

func (_c *MutableArrayNodeStatus_SetSubNodeRetryAttempts_Call) Return() *MutableArrayNodeStatus_SetSubNodeRetryAttempts_Call {
	_c.Call.Return()
	return _c
}

func (_c *MutableArrayNodeStatus_SetSubNodeRetryAttempts_Call) RunAndReturn(run func(bitarray.CompactArray)) *MutableArrayNodeStatus_SetSubNodeRetryAttempts_Call {
	_c.Run(run)
	return _c
}

// SetSubNodeSystemFailures provides a mock function with given fields: subNodeSystemFailures
func (_m *MutableArrayNodeStatus) SetSubNodeSystemFailures(subNodeSystemFailures bitarray.CompactArray) {
	_m.Called(subNodeSystemFailures)
}

// MutableArrayNodeStatus_SetSubNodeSystemFailures_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetSubNodeSystemFailures'
type MutableArrayNodeStatus_SetSubNodeSystemFailures_Call struct {
	*mock.Call
}

// SetSubNodeSystemFailures is a helper method to define mock.On call
//   - subNodeSystemFailures bitarray.CompactArray
func (_e *MutableArrayNodeStatus_Expecter) SetSubNodeSystemFailures(subNodeSystemFailures interface{}) *MutableArrayNodeStatus_SetSubNodeSystemFailures_Call {
	return &MutableArrayNodeStatus_SetSubNodeSystemFailures_Call{Call: _e.mock.On("SetSubNodeSystemFailures", subNodeSystemFailures)}
}

func (_c *MutableArrayNodeStatus_SetSubNodeSystemFailures_Call) Run(run func(subNodeSystemFailures bitarray.CompactArray)) *MutableArrayNodeStatus_SetSubNodeSystemFailures_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(bitarray.CompactArray))
	})
	return _c
}

func (_c *MutableArrayNodeStatus_SetSubNodeSystemFailures_Call) Return() *MutableArrayNodeStatus_SetSubNodeSystemFailures_Call {
	_c.Call.Return()
	return _c
}

func (_c *MutableArrayNodeStatus_SetSubNodeSystemFailures_Call) RunAndReturn(run func(bitarray.CompactArray)) *MutableArrayNodeStatus_SetSubNodeSystemFailures_Call {
	_c.Run(run)
	return _c
}

// SetSubNodeTaskPhases provides a mock function with given fields: subNodeTaskPhases
func (_m *MutableArrayNodeStatus) SetSubNodeTaskPhases(subNodeTaskPhases bitarray.CompactArray) {
	_m.Called(subNodeTaskPhases)
}

// MutableArrayNodeStatus_SetSubNodeTaskPhases_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetSubNodeTaskPhases'
type MutableArrayNodeStatus_SetSubNodeTaskPhases_Call struct {
	*mock.Call
}

// SetSubNodeTaskPhases is a helper method to define mock.On call
//   - subNodeTaskPhases bitarray.CompactArray
func (_e *MutableArrayNodeStatus_Expecter) SetSubNodeTaskPhases(subNodeTaskPhases interface{}) *MutableArrayNodeStatus_SetSubNodeTaskPhases_Call {
	return &MutableArrayNodeStatus_SetSubNodeTaskPhases_Call{Call: _e.mock.On("SetSubNodeTaskPhases", subNodeTaskPhases)}
}

func (_c *MutableArrayNodeStatus_SetSubNodeTaskPhases_Call) Run(run func(subNodeTaskPhases bitarray.CompactArray)) *MutableArrayNodeStatus_SetSubNodeTaskPhases_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(bitarray.CompactArray))
	})
	return _c
}

func (_c *MutableArrayNodeStatus_SetSubNodeTaskPhases_Call) Return() *MutableArrayNodeStatus_SetSubNodeTaskPhases_Call {
	_c.Call.Return()
	return _c
}

func (_c *MutableArrayNodeStatus_SetSubNodeTaskPhases_Call) RunAndReturn(run func(bitarray.CompactArray)) *MutableArrayNodeStatus_SetSubNodeTaskPhases_Call {
	_c.Run(run)
	return _c
}

// SetTaskPhaseVersion provides a mock function with given fields: taskPhaseVersion
func (_m *MutableArrayNodeStatus) SetTaskPhaseVersion(taskPhaseVersion uint32) {
	_m.Called(taskPhaseVersion)
}

// MutableArrayNodeStatus_SetTaskPhaseVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetTaskPhaseVersion'
type MutableArrayNodeStatus_SetTaskPhaseVersion_Call struct {
	*mock.Call
}

// SetTaskPhaseVersion is a helper method to define mock.On call
//   - taskPhaseVersion uint32
func (_e *MutableArrayNodeStatus_Expecter) SetTaskPhaseVersion(taskPhaseVersion interface{}) *MutableArrayNodeStatus_SetTaskPhaseVersion_Call {
	return &MutableArrayNodeStatus_SetTaskPhaseVersion_Call{Call: _e.mock.On("SetTaskPhaseVersion", taskPhaseVersion)}
}

func (_c *MutableArrayNodeStatus_SetTaskPhaseVersion_Call) Run(run func(taskPhaseVersion uint32)) *MutableArrayNodeStatus_SetTaskPhaseVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(uint32))
	})
	return _c
}

func (_c *MutableArrayNodeStatus_SetTaskPhaseVersion_Call) Return() *MutableArrayNodeStatus_SetTaskPhaseVersion_Call {
	_c.Call.Return()
	return _c
}

func (_c *MutableArrayNodeStatus_SetTaskPhaseVersion_Call) RunAndReturn(run func(uint32)) *MutableArrayNodeStatus_SetTaskPhaseVersion_Call {
	_c.Run(run)
	return _c
}

// NewMutableArrayNodeStatus creates a new instance of MutableArrayNodeStatus. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMutableArrayNodeStatus(t interface {
	mock.TestingT
	Cleanup(func())
}) *MutableArrayNodeStatus {
	mock := &MutableArrayNodeStatus{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
