// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	metadata "google.golang.org/grpc/metadata"

	watch "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/watch"
)

// WatchService_WatchExecutionStatusUpdatesServer is an autogenerated mock type for the WatchService_WatchExecutionStatusUpdatesServer type
type WatchService_WatchExecutionStatusUpdatesServer struct {
	mock.Mock
}

type WatchService_WatchExecutionStatusUpdatesServer_Context struct {
	*mock.Call
}

func (_m WatchService_WatchExecutionStatusUpdatesServer_Context) Return(_a0 context.Context) *WatchService_WatchExecutionStatusUpdatesServer_Context {
	return &WatchService_WatchExecutionStatusUpdatesServer_Context{Call: _m.Call.Return(_a0)}
}

func (_m *WatchService_WatchExecutionStatusUpdatesServer) OnContext() *WatchService_WatchExecutionStatusUpdatesServer_Context {
	c_call := _m.On("Context")
	return &WatchService_WatchExecutionStatusUpdatesServer_Context{Call: c_call}
}

func (_m *WatchService_WatchExecutionStatusUpdatesServer) OnContextMatch(matchers ...interface{}) *WatchService_WatchExecutionStatusUpdatesServer_Context {
	c_call := _m.On("Context", matchers...)
	return &WatchService_WatchExecutionStatusUpdatesServer_Context{Call: c_call}
}

// Context provides a mock function with given fields:
func (_m *WatchService_WatchExecutionStatusUpdatesServer) Context() context.Context {
	ret := _m.Called()

	var r0 context.Context
	if rf, ok := ret.Get(0).(func() context.Context); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(context.Context)
		}
	}

	return r0
}

type WatchService_WatchExecutionStatusUpdatesServer_RecvMsg struct {
	*mock.Call
}

func (_m WatchService_WatchExecutionStatusUpdatesServer_RecvMsg) Return(_a0 error) *WatchService_WatchExecutionStatusUpdatesServer_RecvMsg {
	return &WatchService_WatchExecutionStatusUpdatesServer_RecvMsg{Call: _m.Call.Return(_a0)}
}

func (_m *WatchService_WatchExecutionStatusUpdatesServer) OnRecvMsg(m interface{}) *WatchService_WatchExecutionStatusUpdatesServer_RecvMsg {
	c_call := _m.On("RecvMsg", m)
	return &WatchService_WatchExecutionStatusUpdatesServer_RecvMsg{Call: c_call}
}

func (_m *WatchService_WatchExecutionStatusUpdatesServer) OnRecvMsgMatch(matchers ...interface{}) *WatchService_WatchExecutionStatusUpdatesServer_RecvMsg {
	c_call := _m.On("RecvMsg", matchers...)
	return &WatchService_WatchExecutionStatusUpdatesServer_RecvMsg{Call: c_call}
}

// RecvMsg provides a mock function with given fields: m
func (_m *WatchService_WatchExecutionStatusUpdatesServer) RecvMsg(m interface{}) error {
	ret := _m.Called(m)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type WatchService_WatchExecutionStatusUpdatesServer_Send struct {
	*mock.Call
}

func (_m WatchService_WatchExecutionStatusUpdatesServer_Send) Return(_a0 error) *WatchService_WatchExecutionStatusUpdatesServer_Send {
	return &WatchService_WatchExecutionStatusUpdatesServer_Send{Call: _m.Call.Return(_a0)}
}

func (_m *WatchService_WatchExecutionStatusUpdatesServer) OnSend(_a0 *watch.WatchExecutionStatusUpdatesResponse) *WatchService_WatchExecutionStatusUpdatesServer_Send {
	c_call := _m.On("Send", _a0)
	return &WatchService_WatchExecutionStatusUpdatesServer_Send{Call: c_call}
}

func (_m *WatchService_WatchExecutionStatusUpdatesServer) OnSendMatch(matchers ...interface{}) *WatchService_WatchExecutionStatusUpdatesServer_Send {
	c_call := _m.On("Send", matchers...)
	return &WatchService_WatchExecutionStatusUpdatesServer_Send{Call: c_call}
}

// Send provides a mock function with given fields: _a0
func (_m *WatchService_WatchExecutionStatusUpdatesServer) Send(_a0 *watch.WatchExecutionStatusUpdatesResponse) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*watch.WatchExecutionStatusUpdatesResponse) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type WatchService_WatchExecutionStatusUpdatesServer_SendHeader struct {
	*mock.Call
}

func (_m WatchService_WatchExecutionStatusUpdatesServer_SendHeader) Return(_a0 error) *WatchService_WatchExecutionStatusUpdatesServer_SendHeader {
	return &WatchService_WatchExecutionStatusUpdatesServer_SendHeader{Call: _m.Call.Return(_a0)}
}

func (_m *WatchService_WatchExecutionStatusUpdatesServer) OnSendHeader(_a0 metadata.MD) *WatchService_WatchExecutionStatusUpdatesServer_SendHeader {
	c_call := _m.On("SendHeader", _a0)
	return &WatchService_WatchExecutionStatusUpdatesServer_SendHeader{Call: c_call}
}

func (_m *WatchService_WatchExecutionStatusUpdatesServer) OnSendHeaderMatch(matchers ...interface{}) *WatchService_WatchExecutionStatusUpdatesServer_SendHeader {
	c_call := _m.On("SendHeader", matchers...)
	return &WatchService_WatchExecutionStatusUpdatesServer_SendHeader{Call: c_call}
}

// SendHeader provides a mock function with given fields: _a0
func (_m *WatchService_WatchExecutionStatusUpdatesServer) SendHeader(_a0 metadata.MD) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(metadata.MD) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type WatchService_WatchExecutionStatusUpdatesServer_SendMsg struct {
	*mock.Call
}

func (_m WatchService_WatchExecutionStatusUpdatesServer_SendMsg) Return(_a0 error) *WatchService_WatchExecutionStatusUpdatesServer_SendMsg {
	return &WatchService_WatchExecutionStatusUpdatesServer_SendMsg{Call: _m.Call.Return(_a0)}
}

func (_m *WatchService_WatchExecutionStatusUpdatesServer) OnSendMsg(m interface{}) *WatchService_WatchExecutionStatusUpdatesServer_SendMsg {
	c_call := _m.On("SendMsg", m)
	return &WatchService_WatchExecutionStatusUpdatesServer_SendMsg{Call: c_call}
}

func (_m *WatchService_WatchExecutionStatusUpdatesServer) OnSendMsgMatch(matchers ...interface{}) *WatchService_WatchExecutionStatusUpdatesServer_SendMsg {
	c_call := _m.On("SendMsg", matchers...)
	return &WatchService_WatchExecutionStatusUpdatesServer_SendMsg{Call: c_call}
}

// SendMsg provides a mock function with given fields: m
func (_m *WatchService_WatchExecutionStatusUpdatesServer) SendMsg(m interface{}) error {
	ret := _m.Called(m)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type WatchService_WatchExecutionStatusUpdatesServer_SetHeader struct {
	*mock.Call
}

func (_m WatchService_WatchExecutionStatusUpdatesServer_SetHeader) Return(_a0 error) *WatchService_WatchExecutionStatusUpdatesServer_SetHeader {
	return &WatchService_WatchExecutionStatusUpdatesServer_SetHeader{Call: _m.Call.Return(_a0)}
}

func (_m *WatchService_WatchExecutionStatusUpdatesServer) OnSetHeader(_a0 metadata.MD) *WatchService_WatchExecutionStatusUpdatesServer_SetHeader {
	c_call := _m.On("SetHeader", _a0)
	return &WatchService_WatchExecutionStatusUpdatesServer_SetHeader{Call: c_call}
}

func (_m *WatchService_WatchExecutionStatusUpdatesServer) OnSetHeaderMatch(matchers ...interface{}) *WatchService_WatchExecutionStatusUpdatesServer_SetHeader {
	c_call := _m.On("SetHeader", matchers...)
	return &WatchService_WatchExecutionStatusUpdatesServer_SetHeader{Call: c_call}
}

// SetHeader provides a mock function with given fields: _a0
func (_m *WatchService_WatchExecutionStatusUpdatesServer) SetHeader(_a0 metadata.MD) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(metadata.MD) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetTrailer provides a mock function with given fields: _a0
func (_m *WatchService_WatchExecutionStatusUpdatesServer) SetTrailer(_a0 metadata.MD) {
	_m.Called(_a0)
}
