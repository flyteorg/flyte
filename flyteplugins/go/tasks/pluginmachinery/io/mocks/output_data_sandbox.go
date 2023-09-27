// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	storage "github.com/flyteorg/flyte/flytestdlib/storage"
	mock "github.com/stretchr/testify/mock"
)

// RawOutputPaths is an autogenerated mock type for the RawOutputPaths type
type OutputDataSandbox struct {
	mock.Mock
}

type OutputDataSandbox_GetOutputDataSandboxPath struct {
	*mock.Call
}

func (_m OutputDataSandbox_GetOutputDataSandboxPath) Return(_a0 storage.DataReference) *OutputDataSandbox_GetOutputDataSandboxPath {
	return &OutputDataSandbox_GetOutputDataSandboxPath{Call: _m.Call.Return(_a0)}
}

func (_m *OutputDataSandbox) OnGetOutputDataSandboxPath() *OutputDataSandbox_GetOutputDataSandboxPath {
	c := _m.On("GetRawOutputPrefix")
	return &OutputDataSandbox_GetOutputDataSandboxPath{Call: c}
}

func (_m *OutputDataSandbox) OnGetOutputDataSandboxPathMatch(matchers ...interface{}) *OutputDataSandbox_GetOutputDataSandboxPath {
	c := _m.On("GetRawOutputPrefix", matchers...)
	return &OutputDataSandbox_GetOutputDataSandboxPath{Call: c}
}

// GetRawOutputPrefix provides a mock function with given fields:
func (_m *OutputDataSandbox) GetRawOutputPrefix() storage.DataReference {
	ret := _m.Called()

	var r0 storage.DataReference
	if rf, ok := ret.Get(0).(func() storage.DataReference); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(storage.DataReference)
	}

	return r0
}
