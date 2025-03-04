// Code generated by mockery v2.53.0. DO NOT EDIT.

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

type SyncAgentService_ExecuteTaskSyncClient_Expecter struct {
	mock *mock.Mock
}

func (_m *SyncAgentService_ExecuteTaskSyncClient) EXPECT() *SyncAgentService_ExecuteTaskSyncClient_Expecter {
	return &SyncAgentService_ExecuteTaskSyncClient_Expecter{mock: &_m.Mock}
}

// CloseSend provides a mock function with no fields
func (_m *SyncAgentService_ExecuteTaskSyncClient) CloseSend() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for CloseSend")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SyncAgentService_ExecuteTaskSyncClient_CloseSend_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CloseSend'
type SyncAgentService_ExecuteTaskSyncClient_CloseSend_Call struct {
	*mock.Call
}

// CloseSend is a helper method to define mock.On call
func (_e *SyncAgentService_ExecuteTaskSyncClient_Expecter) CloseSend() *SyncAgentService_ExecuteTaskSyncClient_CloseSend_Call {
	return &SyncAgentService_ExecuteTaskSyncClient_CloseSend_Call{Call: _e.mock.On("CloseSend")}
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_CloseSend_Call) Run(run func()) *SyncAgentService_ExecuteTaskSyncClient_CloseSend_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_CloseSend_Call) Return(_a0 error) *SyncAgentService_ExecuteTaskSyncClient_CloseSend_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_CloseSend_Call) RunAndReturn(run func() error) *SyncAgentService_ExecuteTaskSyncClient_CloseSend_Call {
	_c.Call.Return(run)
	return _c
}

// Context provides a mock function with no fields
func (_m *SyncAgentService_ExecuteTaskSyncClient) Context() context.Context {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Context")
	}

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

// SyncAgentService_ExecuteTaskSyncClient_Context_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Context'
type SyncAgentService_ExecuteTaskSyncClient_Context_Call struct {
	*mock.Call
}

// Context is a helper method to define mock.On call
func (_e *SyncAgentService_ExecuteTaskSyncClient_Expecter) Context() *SyncAgentService_ExecuteTaskSyncClient_Context_Call {
	return &SyncAgentService_ExecuteTaskSyncClient_Context_Call{Call: _e.mock.On("Context")}
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_Context_Call) Run(run func()) *SyncAgentService_ExecuteTaskSyncClient_Context_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_Context_Call) Return(_a0 context.Context) *SyncAgentService_ExecuteTaskSyncClient_Context_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_Context_Call) RunAndReturn(run func() context.Context) *SyncAgentService_ExecuteTaskSyncClient_Context_Call {
	_c.Call.Return(run)
	return _c
}

// Header provides a mock function with no fields
func (_m *SyncAgentService_ExecuteTaskSyncClient) Header() (metadata.MD, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Header")
	}

	var r0 metadata.MD
	var r1 error
	if rf, ok := ret.Get(0).(func() (metadata.MD, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() metadata.MD); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(metadata.MD)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SyncAgentService_ExecuteTaskSyncClient_Header_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Header'
type SyncAgentService_ExecuteTaskSyncClient_Header_Call struct {
	*mock.Call
}

// Header is a helper method to define mock.On call
func (_e *SyncAgentService_ExecuteTaskSyncClient_Expecter) Header() *SyncAgentService_ExecuteTaskSyncClient_Header_Call {
	return &SyncAgentService_ExecuteTaskSyncClient_Header_Call{Call: _e.mock.On("Header")}
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_Header_Call) Run(run func()) *SyncAgentService_ExecuteTaskSyncClient_Header_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_Header_Call) Return(_a0 metadata.MD, _a1 error) *SyncAgentService_ExecuteTaskSyncClient_Header_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_Header_Call) RunAndReturn(run func() (metadata.MD, error)) *SyncAgentService_ExecuteTaskSyncClient_Header_Call {
	_c.Call.Return(run)
	return _c
}

// Recv provides a mock function with no fields
func (_m *SyncAgentService_ExecuteTaskSyncClient) Recv() (*admin.ExecuteTaskSyncResponse, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Recv")
	}

	var r0 *admin.ExecuteTaskSyncResponse
	var r1 error
	if rf, ok := ret.Get(0).(func() (*admin.ExecuteTaskSyncResponse, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *admin.ExecuteTaskSyncResponse); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.ExecuteTaskSyncResponse)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SyncAgentService_ExecuteTaskSyncClient_Recv_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Recv'
type SyncAgentService_ExecuteTaskSyncClient_Recv_Call struct {
	*mock.Call
}

// Recv is a helper method to define mock.On call
func (_e *SyncAgentService_ExecuteTaskSyncClient_Expecter) Recv() *SyncAgentService_ExecuteTaskSyncClient_Recv_Call {
	return &SyncAgentService_ExecuteTaskSyncClient_Recv_Call{Call: _e.mock.On("Recv")}
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_Recv_Call) Run(run func()) *SyncAgentService_ExecuteTaskSyncClient_Recv_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_Recv_Call) Return(_a0 *admin.ExecuteTaskSyncResponse, _a1 error) *SyncAgentService_ExecuteTaskSyncClient_Recv_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_Recv_Call) RunAndReturn(run func() (*admin.ExecuteTaskSyncResponse, error)) *SyncAgentService_ExecuteTaskSyncClient_Recv_Call {
	_c.Call.Return(run)
	return _c
}

// RecvMsg provides a mock function with given fields: m
func (_m *SyncAgentService_ExecuteTaskSyncClient) RecvMsg(m interface{}) error {
	ret := _m.Called(m)

	if len(ret) == 0 {
		panic("no return value specified for RecvMsg")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SyncAgentService_ExecuteTaskSyncClient_RecvMsg_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RecvMsg'
type SyncAgentService_ExecuteTaskSyncClient_RecvMsg_Call struct {
	*mock.Call
}

// RecvMsg is a helper method to define mock.On call
//   - m interface{}
func (_e *SyncAgentService_ExecuteTaskSyncClient_Expecter) RecvMsg(m interface{}) *SyncAgentService_ExecuteTaskSyncClient_RecvMsg_Call {
	return &SyncAgentService_ExecuteTaskSyncClient_RecvMsg_Call{Call: _e.mock.On("RecvMsg", m)}
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_RecvMsg_Call) Run(run func(m interface{})) *SyncAgentService_ExecuteTaskSyncClient_RecvMsg_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_RecvMsg_Call) Return(_a0 error) *SyncAgentService_ExecuteTaskSyncClient_RecvMsg_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_RecvMsg_Call) RunAndReturn(run func(interface{}) error) *SyncAgentService_ExecuteTaskSyncClient_RecvMsg_Call {
	_c.Call.Return(run)
	return _c
}

// Send provides a mock function with given fields: _a0
func (_m *SyncAgentService_ExecuteTaskSyncClient) Send(_a0 *admin.ExecuteTaskSyncRequest) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Send")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*admin.ExecuteTaskSyncRequest) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SyncAgentService_ExecuteTaskSyncClient_Send_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Send'
type SyncAgentService_ExecuteTaskSyncClient_Send_Call struct {
	*mock.Call
}

// Send is a helper method to define mock.On call
//   - _a0 *admin.ExecuteTaskSyncRequest
func (_e *SyncAgentService_ExecuteTaskSyncClient_Expecter) Send(_a0 interface{}) *SyncAgentService_ExecuteTaskSyncClient_Send_Call {
	return &SyncAgentService_ExecuteTaskSyncClient_Send_Call{Call: _e.mock.On("Send", _a0)}
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_Send_Call) Run(run func(_a0 *admin.ExecuteTaskSyncRequest)) *SyncAgentService_ExecuteTaskSyncClient_Send_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*admin.ExecuteTaskSyncRequest))
	})
	return _c
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_Send_Call) Return(_a0 error) *SyncAgentService_ExecuteTaskSyncClient_Send_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_Send_Call) RunAndReturn(run func(*admin.ExecuteTaskSyncRequest) error) *SyncAgentService_ExecuteTaskSyncClient_Send_Call {
	_c.Call.Return(run)
	return _c
}

// SendMsg provides a mock function with given fields: m
func (_m *SyncAgentService_ExecuteTaskSyncClient) SendMsg(m interface{}) error {
	ret := _m.Called(m)

	if len(ret) == 0 {
		panic("no return value specified for SendMsg")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SyncAgentService_ExecuteTaskSyncClient_SendMsg_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendMsg'
type SyncAgentService_ExecuteTaskSyncClient_SendMsg_Call struct {
	*mock.Call
}

// SendMsg is a helper method to define mock.On call
//   - m interface{}
func (_e *SyncAgentService_ExecuteTaskSyncClient_Expecter) SendMsg(m interface{}) *SyncAgentService_ExecuteTaskSyncClient_SendMsg_Call {
	return &SyncAgentService_ExecuteTaskSyncClient_SendMsg_Call{Call: _e.mock.On("SendMsg", m)}
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_SendMsg_Call) Run(run func(m interface{})) *SyncAgentService_ExecuteTaskSyncClient_SendMsg_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_SendMsg_Call) Return(_a0 error) *SyncAgentService_ExecuteTaskSyncClient_SendMsg_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_SendMsg_Call) RunAndReturn(run func(interface{}) error) *SyncAgentService_ExecuteTaskSyncClient_SendMsg_Call {
	_c.Call.Return(run)
	return _c
}

// Trailer provides a mock function with no fields
func (_m *SyncAgentService_ExecuteTaskSyncClient) Trailer() metadata.MD {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Trailer")
	}

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

// SyncAgentService_ExecuteTaskSyncClient_Trailer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Trailer'
type SyncAgentService_ExecuteTaskSyncClient_Trailer_Call struct {
	*mock.Call
}

// Trailer is a helper method to define mock.On call
func (_e *SyncAgentService_ExecuteTaskSyncClient_Expecter) Trailer() *SyncAgentService_ExecuteTaskSyncClient_Trailer_Call {
	return &SyncAgentService_ExecuteTaskSyncClient_Trailer_Call{Call: _e.mock.On("Trailer")}
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_Trailer_Call) Run(run func()) *SyncAgentService_ExecuteTaskSyncClient_Trailer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_Trailer_Call) Return(_a0 metadata.MD) *SyncAgentService_ExecuteTaskSyncClient_Trailer_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SyncAgentService_ExecuteTaskSyncClient_Trailer_Call) RunAndReturn(run func() metadata.MD) *SyncAgentService_ExecuteTaskSyncClient_Trailer_Call {
	_c.Call.Return(run)
	return _c
}

// NewSyncAgentService_ExecuteTaskSyncClient creates a new instance of SyncAgentService_ExecuteTaskSyncClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSyncAgentService_ExecuteTaskSyncClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *SyncAgentService_ExecuteTaskSyncClient {
	mock := &SyncAgentService_ExecuteTaskSyncClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
