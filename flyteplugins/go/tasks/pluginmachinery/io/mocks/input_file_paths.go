// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	storage "github.com/flyteorg/flyte/flytestdlib/storage"
	mock "github.com/stretchr/testify/mock"
)

// InputFilePaths is an autogenerated mock type for the InputFilePaths type
type InputFilePaths struct {
	mock.Mock
}

type InputFilePaths_GetInputPath struct {
	*mock.Call
}

func (_m InputFilePaths_GetInputPath) Return(_a0 storage.DataReference) *InputFilePaths_GetInputPath {
	return &InputFilePaths_GetInputPath{Call: _m.Call.Return(_a0)}
}

func (_m *InputFilePaths) OnGetInputPath() *InputFilePaths_GetInputPath {
	c_call := _m.On("GetInputPath")
	return &InputFilePaths_GetInputPath{Call: c_call}
}

func (_m *InputFilePaths) OnGetInputPathMatch(matchers ...interface{}) *InputFilePaths_GetInputPath {
	c_call := _m.On("GetInputPath", matchers...)
	return &InputFilePaths_GetInputPath{Call: c_call}
}

// GetInputPath provides a mock function with given fields:
func (_m *InputFilePaths) GetInputPath() storage.DataReference {
	ret := _m.Called()

	var r0 storage.DataReference
	if rf, ok := ret.Get(0).(func() storage.DataReference); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(storage.DataReference)
	}

	return r0
}

type InputFilePaths_GetInputPrefixPath struct {
	*mock.Call
}

func (_m InputFilePaths_GetInputPrefixPath) Return(_a0 storage.DataReference) *InputFilePaths_GetInputPrefixPath {
	return &InputFilePaths_GetInputPrefixPath{Call: _m.Call.Return(_a0)}
}

func (_m *InputFilePaths) OnGetInputPrefixPath() *InputFilePaths_GetInputPrefixPath {
	c_call := _m.On("GetInputPrefixPath")
	return &InputFilePaths_GetInputPrefixPath{Call: c_call}
}

func (_m *InputFilePaths) OnGetInputPrefixPathMatch(matchers ...interface{}) *InputFilePaths_GetInputPrefixPath {
	c_call := _m.On("GetInputPrefixPath", matchers...)
	return &InputFilePaths_GetInputPrefixPath{Call: c_call}
}

// GetInputPrefixPath provides a mock function with given fields:
func (_m *InputFilePaths) GetInputPrefixPath() storage.DataReference {
	ret := _m.Called()

	var r0 storage.DataReference
	if rf, ok := ret.Get(0).(func() storage.DataReference); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(storage.DataReference)
	}

	return r0
}
