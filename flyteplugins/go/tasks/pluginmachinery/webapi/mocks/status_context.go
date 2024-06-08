// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	core "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	io "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"

	mock "github.com/stretchr/testify/mock"

	storage "github.com/flyteorg/flyte/flytestdlib/storage"
)

// StatusContext is an autogenerated mock type for the StatusContext type
type StatusContext struct {
	mock.Mock
}

type StatusContext_DataStore struct {
	*mock.Call
}

func (_m StatusContext_DataStore) Return(_a0 *storage.DataStore) *StatusContext_DataStore {
	return &StatusContext_DataStore{Call: _m.Call.Return(_a0)}
}

func (_m *StatusContext) OnDataStore() *StatusContext_DataStore {
	c_call := _m.On("DataStore")
	return &StatusContext_DataStore{Call: c_call}
}

func (_m *StatusContext) OnDataStoreMatch(matchers ...interface{}) *StatusContext_DataStore {
	c_call := _m.On("DataStore", matchers...)
	return &StatusContext_DataStore{Call: c_call}
}

// DataStore provides a mock function with given fields:
func (_m *StatusContext) DataStore() *storage.DataStore {
	ret := _m.Called()

	var r0 *storage.DataStore
	if rf, ok := ret.Get(0).(func() *storage.DataStore); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*storage.DataStore)
		}
	}

	return r0
}

type StatusContext_InputReader struct {
	*mock.Call
}

func (_m StatusContext_InputReader) Return(_a0 io.InputReader) *StatusContext_InputReader {
	return &StatusContext_InputReader{Call: _m.Call.Return(_a0)}
}

func (_m *StatusContext) OnInputReader() *StatusContext_InputReader {
	c_call := _m.On("InputReader")
	return &StatusContext_InputReader{Call: c_call}
}

func (_m *StatusContext) OnInputReaderMatch(matchers ...interface{}) *StatusContext_InputReader {
	c_call := _m.On("InputReader", matchers...)
	return &StatusContext_InputReader{Call: c_call}
}

// InputReader provides a mock function with given fields:
func (_m *StatusContext) InputReader() io.InputReader {
	ret := _m.Called()

	var r0 io.InputReader
	if rf, ok := ret.Get(0).(func() io.InputReader); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.InputReader)
		}
	}

	return r0
}

type StatusContext_OutputWriter struct {
	*mock.Call
}

func (_m StatusContext_OutputWriter) Return(_a0 io.OutputWriter) *StatusContext_OutputWriter {
	return &StatusContext_OutputWriter{Call: _m.Call.Return(_a0)}
}

func (_m *StatusContext) OnOutputWriter() *StatusContext_OutputWriter {
	c_call := _m.On("OutputWriter")
	return &StatusContext_OutputWriter{Call: c_call}
}

func (_m *StatusContext) OnOutputWriterMatch(matchers ...interface{}) *StatusContext_OutputWriter {
	c_call := _m.On("OutputWriter", matchers...)
	return &StatusContext_OutputWriter{Call: c_call}
}

// OutputWriter provides a mock function with given fields:
func (_m *StatusContext) OutputWriter() io.OutputWriter {
	ret := _m.Called()

	var r0 io.OutputWriter
	if rf, ok := ret.Get(0).(func() io.OutputWriter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.OutputWriter)
		}
	}

	return r0
}

type StatusContext_Resource struct {
	*mock.Call
}

func (_m StatusContext_Resource) Return(_a0 interface{}) *StatusContext_Resource {
	return &StatusContext_Resource{Call: _m.Call.Return(_a0)}
}

func (_m *StatusContext) OnResource() *StatusContext_Resource {
	c_call := _m.On("Resource")
	return &StatusContext_Resource{Call: c_call}
}

func (_m *StatusContext) OnResourceMatch(matchers ...interface{}) *StatusContext_Resource {
	c_call := _m.On("Resource", matchers...)
	return &StatusContext_Resource{Call: c_call}
}

// Resource provides a mock function with given fields:
func (_m *StatusContext) Resource() interface{} {
	ret := _m.Called()

	var r0 interface{}
	if rf, ok := ret.Get(0).(func() interface{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	return r0
}

type StatusContext_ResourceMeta struct {
	*mock.Call
}

func (_m StatusContext_ResourceMeta) Return(_a0 interface{}) *StatusContext_ResourceMeta {
	return &StatusContext_ResourceMeta{Call: _m.Call.Return(_a0)}
}

func (_m *StatusContext) OnResourceMeta() *StatusContext_ResourceMeta {
	c_call := _m.On("ResourceMeta")
	return &StatusContext_ResourceMeta{Call: c_call}
}

func (_m *StatusContext) OnResourceMetaMatch(matchers ...interface{}) *StatusContext_ResourceMeta {
	c_call := _m.On("ResourceMeta", matchers...)
	return &StatusContext_ResourceMeta{Call: c_call}
}

// ResourceMeta provides a mock function with given fields:
func (_m *StatusContext) ResourceMeta() interface{} {
	ret := _m.Called()

	var r0 interface{}
	if rf, ok := ret.Get(0).(func() interface{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	return r0
}

type StatusContext_SecretManager struct {
	*mock.Call
}

func (_m StatusContext_SecretManager) Return(_a0 core.SecretManager) *StatusContext_SecretManager {
	return &StatusContext_SecretManager{Call: _m.Call.Return(_a0)}
}

func (_m *StatusContext) OnSecretManager() *StatusContext_SecretManager {
	c_call := _m.On("SecretManager")
	return &StatusContext_SecretManager{Call: c_call}
}

func (_m *StatusContext) OnSecretManagerMatch(matchers ...interface{}) *StatusContext_SecretManager {
	c_call := _m.On("SecretManager", matchers...)
	return &StatusContext_SecretManager{Call: c_call}
}

// SecretManager provides a mock function with given fields:
func (_m *StatusContext) SecretManager() core.SecretManager {
	ret := _m.Called()

	var r0 core.SecretManager
	if rf, ok := ret.Get(0).(func() core.SecretManager); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(core.SecretManager)
		}
	}

	return r0
}

type StatusContext_TaskExecutionMetadata struct {
	*mock.Call
}

func (_m StatusContext_TaskExecutionMetadata) Return(_a0 core.TaskExecutionMetadata) *StatusContext_TaskExecutionMetadata {
	return &StatusContext_TaskExecutionMetadata{Call: _m.Call.Return(_a0)}
}

func (_m *StatusContext) OnTaskExecutionMetadata() *StatusContext_TaskExecutionMetadata {
	c_call := _m.On("TaskExecutionMetadata")
	return &StatusContext_TaskExecutionMetadata{Call: c_call}
}

func (_m *StatusContext) OnTaskExecutionMetadataMatch(matchers ...interface{}) *StatusContext_TaskExecutionMetadata {
	c_call := _m.On("TaskExecutionMetadata", matchers...)
	return &StatusContext_TaskExecutionMetadata{Call: c_call}
}

// TaskExecutionMetadata provides a mock function with given fields:
func (_m *StatusContext) TaskExecutionMetadata() core.TaskExecutionMetadata {
	ret := _m.Called()

	var r0 core.TaskExecutionMetadata
	if rf, ok := ret.Get(0).(func() core.TaskExecutionMetadata); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(core.TaskExecutionMetadata)
		}
	}

	return r0
}

type StatusContext_TaskReader struct {
	*mock.Call
}

func (_m StatusContext_TaskReader) Return(_a0 core.TaskReader) *StatusContext_TaskReader {
	return &StatusContext_TaskReader{Call: _m.Call.Return(_a0)}
}

func (_m *StatusContext) OnTaskReader() *StatusContext_TaskReader {
	c_call := _m.On("TaskReader")
	return &StatusContext_TaskReader{Call: c_call}
}

func (_m *StatusContext) OnTaskReaderMatch(matchers ...interface{}) *StatusContext_TaskReader {
	c_call := _m.On("TaskReader", matchers...)
	return &StatusContext_TaskReader{Call: c_call}
}

// TaskReader provides a mock function with given fields:
func (_m *StatusContext) TaskReader() core.TaskReader {
	ret := _m.Called()

	var r0 core.TaskReader
	if rf, ok := ret.Get(0).(func() core.TaskReader); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(core.TaskReader)
		}
	}

	return r0
}
