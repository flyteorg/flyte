// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	storage "github.com/flyteorg/flyte/flytestdlib/storage"
	mock "github.com/stretchr/testify/mock"
)

// OutputFilePaths is an autogenerated mock type for the OutputFilePaths type
type OutputFilePaths struct {
	mock.Mock
}

type OutputFilePaths_GetCheckpointPrefix struct {
	*mock.Call
}

func (_m OutputFilePaths_GetCheckpointPrefix) Return(_a0 storage.DataReference) *OutputFilePaths_GetCheckpointPrefix {
	return &OutputFilePaths_GetCheckpointPrefix{Call: _m.Call.Return(_a0)}
}

func (_m *OutputFilePaths) OnGetCheckpointPrefix() *OutputFilePaths_GetCheckpointPrefix {
	c_call := _m.On("GetCheckpointPrefix")
	return &OutputFilePaths_GetCheckpointPrefix{Call: c_call}
}

func (_m *OutputFilePaths) OnGetCheckpointPrefixMatch(matchers ...interface{}) *OutputFilePaths_GetCheckpointPrefix {
	c_call := _m.On("GetCheckpointPrefix", matchers...)
	return &OutputFilePaths_GetCheckpointPrefix{Call: c_call}
}

// GetCheckpointPrefix provides a mock function with given fields:
func (_m *OutputFilePaths) GetCheckpointPrefix() storage.DataReference {
	ret := _m.Called()

	var r0 storage.DataReference
	if rf, ok := ret.Get(0).(func() storage.DataReference); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(storage.DataReference)
	}

	return r0
}

type OutputFilePaths_GetDeckPath struct {
	*mock.Call
}

func (_m OutputFilePaths_GetDeckPath) Return(_a0 storage.DataReference) *OutputFilePaths_GetDeckPath {
	return &OutputFilePaths_GetDeckPath{Call: _m.Call.Return(_a0)}
}

func (_m *OutputFilePaths) OnGetDeckPath() *OutputFilePaths_GetDeckPath {
	c_call := _m.On("GetDeckPath")
	return &OutputFilePaths_GetDeckPath{Call: c_call}
}

func (_m *OutputFilePaths) OnGetDeckPathMatch(matchers ...interface{}) *OutputFilePaths_GetDeckPath {
	c_call := _m.On("GetDeckPath", matchers...)
	return &OutputFilePaths_GetDeckPath{Call: c_call}
}

// GetDeckPath provides a mock function with given fields:
func (_m *OutputFilePaths) GetDeckPath() storage.DataReference {
	ret := _m.Called()

	var r0 storage.DataReference
	if rf, ok := ret.Get(0).(func() storage.DataReference); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(storage.DataReference)
	}

	return r0
}

type OutputFilePaths_GetErrorPath struct {
	*mock.Call
}

func (_m OutputFilePaths_GetErrorPath) Return(_a0 storage.DataReference) *OutputFilePaths_GetErrorPath {
	return &OutputFilePaths_GetErrorPath{Call: _m.Call.Return(_a0)}
}

func (_m *OutputFilePaths) OnGetErrorPath() *OutputFilePaths_GetErrorPath {
	c_call := _m.On("GetErrorPath")
	return &OutputFilePaths_GetErrorPath{Call: c_call}
}

func (_m *OutputFilePaths) OnGetErrorPathMatch(matchers ...interface{}) *OutputFilePaths_GetErrorPath {
	c_call := _m.On("GetErrorPath", matchers...)
	return &OutputFilePaths_GetErrorPath{Call: c_call}
}

// GetErrorPath provides a mock function with given fields:
func (_m *OutputFilePaths) GetErrorPath() storage.DataReference {
	ret := _m.Called()

	var r0 storage.DataReference
	if rf, ok := ret.Get(0).(func() storage.DataReference); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(storage.DataReference)
	}

	return r0
}

type OutputFilePaths_GetOutputPath struct {
	*mock.Call
}

func (_m OutputFilePaths_GetOutputPath) Return(_a0 storage.DataReference) *OutputFilePaths_GetOutputPath {
	return &OutputFilePaths_GetOutputPath{Call: _m.Call.Return(_a0)}
}

func (_m *OutputFilePaths) OnGetOutputPath() *OutputFilePaths_GetOutputPath {
	c_call := _m.On("GetOutputPath")
	return &OutputFilePaths_GetOutputPath{Call: c_call}
}

func (_m *OutputFilePaths) OnGetOutputPathMatch(matchers ...interface{}) *OutputFilePaths_GetOutputPath {
	c_call := _m.On("GetOutputPath", matchers...)
	return &OutputFilePaths_GetOutputPath{Call: c_call}
}

// GetOutputPath provides a mock function with given fields:
func (_m *OutputFilePaths) GetOutputPath() storage.DataReference {
	ret := _m.Called()

	var r0 storage.DataReference
	if rf, ok := ret.Get(0).(func() storage.DataReference); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(storage.DataReference)
	}

	return r0
}

type OutputFilePaths_GetOutputPrefixPath struct {
	*mock.Call
}

func (_m OutputFilePaths_GetOutputPrefixPath) Return(_a0 storage.DataReference) *OutputFilePaths_GetOutputPrefixPath {
	return &OutputFilePaths_GetOutputPrefixPath{Call: _m.Call.Return(_a0)}
}

func (_m *OutputFilePaths) OnGetOutputPrefixPath() *OutputFilePaths_GetOutputPrefixPath {
	c_call := _m.On("GetOutputPrefixPath")
	return &OutputFilePaths_GetOutputPrefixPath{Call: c_call}
}

func (_m *OutputFilePaths) OnGetOutputPrefixPathMatch(matchers ...interface{}) *OutputFilePaths_GetOutputPrefixPath {
	c_call := _m.On("GetOutputPrefixPath", matchers...)
	return &OutputFilePaths_GetOutputPrefixPath{Call: c_call}
}

// GetOutputPrefixPath provides a mock function with given fields:
func (_m *OutputFilePaths) GetOutputPrefixPath() storage.DataReference {
	ret := _m.Called()

	var r0 storage.DataReference
	if rf, ok := ret.Get(0).(func() storage.DataReference); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(storage.DataReference)
	}

	return r0
}

type OutputFilePaths_GetPreviousCheckpointsPrefix struct {
	*mock.Call
}

func (_m OutputFilePaths_GetPreviousCheckpointsPrefix) Return(_a0 storage.DataReference) *OutputFilePaths_GetPreviousCheckpointsPrefix {
	return &OutputFilePaths_GetPreviousCheckpointsPrefix{Call: _m.Call.Return(_a0)}
}

func (_m *OutputFilePaths) OnGetPreviousCheckpointsPrefix() *OutputFilePaths_GetPreviousCheckpointsPrefix {
	c_call := _m.On("GetPreviousCheckpointsPrefix")
	return &OutputFilePaths_GetPreviousCheckpointsPrefix{Call: c_call}
}

func (_m *OutputFilePaths) OnGetPreviousCheckpointsPrefixMatch(matchers ...interface{}) *OutputFilePaths_GetPreviousCheckpointsPrefix {
	c_call := _m.On("GetPreviousCheckpointsPrefix", matchers...)
	return &OutputFilePaths_GetPreviousCheckpointsPrefix{Call: c_call}
}

// GetPreviousCheckpointsPrefix provides a mock function with given fields:
func (_m *OutputFilePaths) GetPreviousCheckpointsPrefix() storage.DataReference {
	ret := _m.Called()

	var r0 storage.DataReference
	if rf, ok := ret.Get(0).(func() storage.DataReference); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(storage.DataReference)
	}

	return r0
}

type OutputFilePaths_GetRawOutputPrefix struct {
	*mock.Call
}

func (_m OutputFilePaths_GetRawOutputPrefix) Return(_a0 storage.DataReference) *OutputFilePaths_GetRawOutputPrefix {
	return &OutputFilePaths_GetRawOutputPrefix{Call: _m.Call.Return(_a0)}
}

func (_m *OutputFilePaths) OnGetRawOutputPrefix() *OutputFilePaths_GetRawOutputPrefix {
	c_call := _m.On("GetRawOutputPrefix")
	return &OutputFilePaths_GetRawOutputPrefix{Call: c_call}
}

func (_m *OutputFilePaths) OnGetRawOutputPrefixMatch(matchers ...interface{}) *OutputFilePaths_GetRawOutputPrefix {
	c_call := _m.On("GetRawOutputPrefix", matchers...)
	return &OutputFilePaths_GetRawOutputPrefix{Call: c_call}
}

// GetRawOutputPrefix provides a mock function with given fields:
func (_m *OutputFilePaths) GetRawOutputPrefix() storage.DataReference {
	ret := _m.Called()

	var r0 storage.DataReference
	if rf, ok := ret.Get(0).(func() storage.DataReference); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(storage.DataReference)
	}

	return r0
}
