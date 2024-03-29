// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	context "context"

	admin "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

	metadata "google.golang.org/grpc/metadata"

	mock "github.com/stretchr/testify/mock"
)

// SyncAgentService_ExecuteTaskSyncClient is an autogenerated mock type for the SyncAgentService_ExecuteTaskSyncClient type
type SyncAgentService_ExecuteTaskSyncClient struct {
	mock.Mock
}

type SyncAgentService_ExecuteTaskSyncClient_CloseSend struct {
	*mock.Call
}

func (_m SyncAgentService_ExecuteTaskSyncClient_CloseSend) Return(_a0 error) *SyncAgentService_ExecuteTaskSyncClient_CloseSend {
	return &SyncAgentService_ExecuteTaskSyncClient_CloseSend{Call: _m.Call.Return(_a0)}
}

func (_m *SyncAgentService_ExecuteTaskSyncClient) OnCloseSend() *SyncAgentService_ExecuteTaskSyncClient_CloseSend {
	c_call := _m.On("CloseSend")
	return &SyncAgentService_ExecuteTaskSyncClient_CloseSend{Call: c_call}
}

func (_m *SyncAgentService_ExecuteTaskSyncClient) OnCloseSendMatch(matchers ...interface{}) *SyncAgentService_ExecuteTaskSyncClient_CloseSend {
	c_call := _m.On("CloseSend", matchers...)
	return &SyncAgentService_ExecuteTaskSyncClient_CloseSend{Call: c_call}
}

// CloseSend provides a mock function with given fields:
func (_m *SyncAgentService_ExecuteTaskSyncClient) CloseSend() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type SyncAgentService_ExecuteTaskSyncClient_Context struct {
	*mock.Call
}

func (_m SyncAgentService_ExecuteTaskSyncClient_Context) Return(_a0 context.Context) *SyncAgentService_ExecuteTaskSyncClient_Context {
	return &SyncAgentService_ExecuteTaskSyncClient_Context{Call: _m.Call.Return(_a0)}
}

func (_m *SyncAgentService_ExecuteTaskSyncClient) OnContext() *SyncAgentService_ExecuteTaskSyncClient_Context {
	c_call := _m.On("Context")
	return &SyncAgentService_ExecuteTaskSyncClient_Context{Call: c_call}
}

func (_m *SyncAgentService_ExecuteTaskSyncClient) OnContextMatch(matchers ...interface{}) *SyncAgentService_ExecuteTaskSyncClient_Context {
	c_call := _m.On("Context", matchers...)
	return &SyncAgentService_ExecuteTaskSyncClient_Context{Call: c_call}
}

// Context provides a mock function with given fields:
func (_m *SyncAgentService_ExecuteTaskSyncClient) Context() context.Context {
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

type SyncAgentService_ExecuteTaskSyncClient_Header struct {
	*mock.Call
}

func (_m SyncAgentService_ExecuteTaskSyncClient_Header) Return(_a0 metadata.MD, _a1 error) *SyncAgentService_ExecuteTaskSyncClient_Header {
	return &SyncAgentService_ExecuteTaskSyncClient_Header{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *SyncAgentService_ExecuteTaskSyncClient) OnHeader() *SyncAgentService_ExecuteTaskSyncClient_Header {
	c_call := _m.On("Header")
	return &SyncAgentService_ExecuteTaskSyncClient_Header{Call: c_call}
}

func (_m *SyncAgentService_ExecuteTaskSyncClient) OnHeaderMatch(matchers ...interface{}) *SyncAgentService_ExecuteTaskSyncClient_Header {
	c_call := _m.On("Header", matchers...)
	return &SyncAgentService_ExecuteTaskSyncClient_Header{Call: c_call}
}

// Header provides a mock function with given fields:
func (_m *SyncAgentService_ExecuteTaskSyncClient) Header() (metadata.MD, error) {
	ret := _m.Called()

	var r0 metadata.MD
	if rf, ok := ret.Get(0).(func() metadata.MD); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(metadata.MD)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type SyncAgentService_ExecuteTaskSyncClient_Recv struct {
	*mock.Call
}

func (_m SyncAgentService_ExecuteTaskSyncClient_Recv) Return(_a0 *admin.ExecuteTaskSyncResponse, _a1 error) *SyncAgentService_ExecuteTaskSyncClient_Recv {
	return &SyncAgentService_ExecuteTaskSyncClient_Recv{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *SyncAgentService_ExecuteTaskSyncClient) OnRecv() *SyncAgentService_ExecuteTaskSyncClient_Recv {
	c_call := _m.On("Recv")
	return &SyncAgentService_ExecuteTaskSyncClient_Recv{Call: c_call}
}

func (_m *SyncAgentService_ExecuteTaskSyncClient) OnRecvMatch(matchers ...interface{}) *SyncAgentService_ExecuteTaskSyncClient_Recv {
	c_call := _m.On("Recv", matchers...)
	return &SyncAgentService_ExecuteTaskSyncClient_Recv{Call: c_call}
}

// Recv provides a mock function with given fields:
func (_m *SyncAgentService_ExecuteTaskSyncClient) Recv() (*admin.ExecuteTaskSyncResponse, error) {
	ret := _m.Called()

	var r0 *admin.ExecuteTaskSyncResponse
	if rf, ok := ret.Get(0).(func() *admin.ExecuteTaskSyncResponse); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.ExecuteTaskSyncResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type SyncAgentService_ExecuteTaskSyncClient_RecvMsg struct {
	*mock.Call
}

func (_m SyncAgentService_ExecuteTaskSyncClient_RecvMsg) Return(_a0 error) *SyncAgentService_ExecuteTaskSyncClient_RecvMsg {
	return &SyncAgentService_ExecuteTaskSyncClient_RecvMsg{Call: _m.Call.Return(_a0)}
}

func (_m *SyncAgentService_ExecuteTaskSyncClient) OnRecvMsg(m interface{}) *SyncAgentService_ExecuteTaskSyncClient_RecvMsg {
	c_call := _m.On("RecvMsg", m)
	return &SyncAgentService_ExecuteTaskSyncClient_RecvMsg{Call: c_call}
}

func (_m *SyncAgentService_ExecuteTaskSyncClient) OnRecvMsgMatch(matchers ...interface{}) *SyncAgentService_ExecuteTaskSyncClient_RecvMsg {
	c_call := _m.On("RecvMsg", matchers...)
	return &SyncAgentService_ExecuteTaskSyncClient_RecvMsg{Call: c_call}
}

// RecvMsg provides a mock function with given fields: m
func (_m *SyncAgentService_ExecuteTaskSyncClient) RecvMsg(m interface{}) error {
	ret := _m.Called(m)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type SyncAgentService_ExecuteTaskSyncClient_Send struct {
	*mock.Call
}

func (_m SyncAgentService_ExecuteTaskSyncClient_Send) Return(_a0 error) *SyncAgentService_ExecuteTaskSyncClient_Send {
	return &SyncAgentService_ExecuteTaskSyncClient_Send{Call: _m.Call.Return(_a0)}
}

func (_m *SyncAgentService_ExecuteTaskSyncClient) OnSend(_a0 *admin.ExecuteTaskSyncRequest) *SyncAgentService_ExecuteTaskSyncClient_Send {
	c_call := _m.On("Send", _a0)
	return &SyncAgentService_ExecuteTaskSyncClient_Send{Call: c_call}
}

func (_m *SyncAgentService_ExecuteTaskSyncClient) OnSendMatch(matchers ...interface{}) *SyncAgentService_ExecuteTaskSyncClient_Send {
	c_call := _m.On("Send", matchers...)
	return &SyncAgentService_ExecuteTaskSyncClient_Send{Call: c_call}
}

// Send provides a mock function with given fields: _a0
func (_m *SyncAgentService_ExecuteTaskSyncClient) Send(_a0 *admin.ExecuteTaskSyncRequest) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*admin.ExecuteTaskSyncRequest) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type SyncAgentService_ExecuteTaskSyncClient_SendMsg struct {
	*mock.Call
}

func (_m SyncAgentService_ExecuteTaskSyncClient_SendMsg) Return(_a0 error) *SyncAgentService_ExecuteTaskSyncClient_SendMsg {
	return &SyncAgentService_ExecuteTaskSyncClient_SendMsg{Call: _m.Call.Return(_a0)}
}

func (_m *SyncAgentService_ExecuteTaskSyncClient) OnSendMsg(m interface{}) *SyncAgentService_ExecuteTaskSyncClient_SendMsg {
	c_call := _m.On("SendMsg", m)
	return &SyncAgentService_ExecuteTaskSyncClient_SendMsg{Call: c_call}
}

func (_m *SyncAgentService_ExecuteTaskSyncClient) OnSendMsgMatch(matchers ...interface{}) *SyncAgentService_ExecuteTaskSyncClient_SendMsg {
	c_call := _m.On("SendMsg", matchers...)
	return &SyncAgentService_ExecuteTaskSyncClient_SendMsg{Call: c_call}
}

// SendMsg provides a mock function with given fields: m
func (_m *SyncAgentService_ExecuteTaskSyncClient) SendMsg(m interface{}) error {
	ret := _m.Called(m)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type SyncAgentService_ExecuteTaskSyncClient_Trailer struct {
	*mock.Call
}

func (_m SyncAgentService_ExecuteTaskSyncClient_Trailer) Return(_a0 metadata.MD) *SyncAgentService_ExecuteTaskSyncClient_Trailer {
	return &SyncAgentService_ExecuteTaskSyncClient_Trailer{Call: _m.Call.Return(_a0)}
}

func (_m *SyncAgentService_ExecuteTaskSyncClient) OnTrailer() *SyncAgentService_ExecuteTaskSyncClient_Trailer {
	c_call := _m.On("Trailer")
	return &SyncAgentService_ExecuteTaskSyncClient_Trailer{Call: c_call}
}

func (_m *SyncAgentService_ExecuteTaskSyncClient) OnTrailerMatch(matchers ...interface{}) *SyncAgentService_ExecuteTaskSyncClient_Trailer {
	c_call := _m.On("Trailer", matchers...)
	return &SyncAgentService_ExecuteTaskSyncClient_Trailer{Call: c_call}
}

// Trailer provides a mock function with given fields:
func (_m *SyncAgentService_ExecuteTaskSyncClient) Trailer() metadata.MD {
	ret := _m.Called()

	var r0 metadata.MD
	if rf, ok := ret.Get(0).(func() metadata.MD); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(metadata.MD)
		}
	}

	return r0
}
