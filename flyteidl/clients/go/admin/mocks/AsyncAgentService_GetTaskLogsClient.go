// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	context "context"

	admin "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

	metadata "google.golang.org/grpc/metadata"

	mock "github.com/stretchr/testify/mock"
)

// AsyncAgentService_GetTaskLogsClient is an autogenerated mock type for the AsyncAgentService_GetTaskLogsClient type
type AsyncAgentService_GetTaskLogsClient struct {
	mock.Mock
}

type AsyncAgentService_GetTaskLogsClient_Expecter struct {
	mock *mock.Mock
}

func (_m *AsyncAgentService_GetTaskLogsClient) EXPECT() *AsyncAgentService_GetTaskLogsClient_Expecter {
	return &AsyncAgentService_GetTaskLogsClient_Expecter{mock: &_m.Mock}
}

// CloseSend provides a mock function with no fields
func (_m *AsyncAgentService_GetTaskLogsClient) CloseSend() error {
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

// AsyncAgentService_GetTaskLogsClient_CloseSend_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CloseSend'
type AsyncAgentService_GetTaskLogsClient_CloseSend_Call struct {
	*mock.Call
}

// CloseSend is a helper method to define mock.On call
func (_e *AsyncAgentService_GetTaskLogsClient_Expecter) CloseSend() *AsyncAgentService_GetTaskLogsClient_CloseSend_Call {
	return &AsyncAgentService_GetTaskLogsClient_CloseSend_Call{Call: _e.mock.On("CloseSend")}
}

func (_c *AsyncAgentService_GetTaskLogsClient_CloseSend_Call) Run(run func()) *AsyncAgentService_GetTaskLogsClient_CloseSend_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AsyncAgentService_GetTaskLogsClient_CloseSend_Call) Return(_a0 error) *AsyncAgentService_GetTaskLogsClient_CloseSend_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AsyncAgentService_GetTaskLogsClient_CloseSend_Call) RunAndReturn(run func() error) *AsyncAgentService_GetTaskLogsClient_CloseSend_Call {
	_c.Call.Return(run)
	return _c
}

// Context provides a mock function with no fields
func (_m *AsyncAgentService_GetTaskLogsClient) Context() context.Context {
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

// AsyncAgentService_GetTaskLogsClient_Context_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Context'
type AsyncAgentService_GetTaskLogsClient_Context_Call struct {
	*mock.Call
}

// Context is a helper method to define mock.On call
func (_e *AsyncAgentService_GetTaskLogsClient_Expecter) Context() *AsyncAgentService_GetTaskLogsClient_Context_Call {
	return &AsyncAgentService_GetTaskLogsClient_Context_Call{Call: _e.mock.On("Context")}
}

func (_c *AsyncAgentService_GetTaskLogsClient_Context_Call) Run(run func()) *AsyncAgentService_GetTaskLogsClient_Context_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AsyncAgentService_GetTaskLogsClient_Context_Call) Return(_a0 context.Context) *AsyncAgentService_GetTaskLogsClient_Context_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AsyncAgentService_GetTaskLogsClient_Context_Call) RunAndReturn(run func() context.Context) *AsyncAgentService_GetTaskLogsClient_Context_Call {
	_c.Call.Return(run)
	return _c
}

// Header provides a mock function with no fields
func (_m *AsyncAgentService_GetTaskLogsClient) Header() (metadata.MD, error) {
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

// AsyncAgentService_GetTaskLogsClient_Header_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Header'
type AsyncAgentService_GetTaskLogsClient_Header_Call struct {
	*mock.Call
}

// Header is a helper method to define mock.On call
func (_e *AsyncAgentService_GetTaskLogsClient_Expecter) Header() *AsyncAgentService_GetTaskLogsClient_Header_Call {
	return &AsyncAgentService_GetTaskLogsClient_Header_Call{Call: _e.mock.On("Header")}
}

func (_c *AsyncAgentService_GetTaskLogsClient_Header_Call) Run(run func()) *AsyncAgentService_GetTaskLogsClient_Header_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AsyncAgentService_GetTaskLogsClient_Header_Call) Return(_a0 metadata.MD, _a1 error) *AsyncAgentService_GetTaskLogsClient_Header_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AsyncAgentService_GetTaskLogsClient_Header_Call) RunAndReturn(run func() (metadata.MD, error)) *AsyncAgentService_GetTaskLogsClient_Header_Call {
	_c.Call.Return(run)
	return _c
}

// Recv provides a mock function with no fields
func (_m *AsyncAgentService_GetTaskLogsClient) Recv() (*admin.GetTaskLogsResponse, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Recv")
	}

	var r0 *admin.GetTaskLogsResponse
	var r1 error
	if rf, ok := ret.Get(0).(func() (*admin.GetTaskLogsResponse, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *admin.GetTaskLogsResponse); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.GetTaskLogsResponse)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AsyncAgentService_GetTaskLogsClient_Recv_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Recv'
type AsyncAgentService_GetTaskLogsClient_Recv_Call struct {
	*mock.Call
}

// Recv is a helper method to define mock.On call
func (_e *AsyncAgentService_GetTaskLogsClient_Expecter) Recv() *AsyncAgentService_GetTaskLogsClient_Recv_Call {
	return &AsyncAgentService_GetTaskLogsClient_Recv_Call{Call: _e.mock.On("Recv")}
}

func (_c *AsyncAgentService_GetTaskLogsClient_Recv_Call) Run(run func()) *AsyncAgentService_GetTaskLogsClient_Recv_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AsyncAgentService_GetTaskLogsClient_Recv_Call) Return(_a0 *admin.GetTaskLogsResponse, _a1 error) *AsyncAgentService_GetTaskLogsClient_Recv_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AsyncAgentService_GetTaskLogsClient_Recv_Call) RunAndReturn(run func() (*admin.GetTaskLogsResponse, error)) *AsyncAgentService_GetTaskLogsClient_Recv_Call {
	_c.Call.Return(run)
	return _c
}

// RecvMsg provides a mock function with given fields: m
func (_m *AsyncAgentService_GetTaskLogsClient) RecvMsg(m interface{}) error {
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

// AsyncAgentService_GetTaskLogsClient_RecvMsg_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RecvMsg'
type AsyncAgentService_GetTaskLogsClient_RecvMsg_Call struct {
	*mock.Call
}

// RecvMsg is a helper method to define mock.On call
//   - m interface{}
func (_e *AsyncAgentService_GetTaskLogsClient_Expecter) RecvMsg(m interface{}) *AsyncAgentService_GetTaskLogsClient_RecvMsg_Call {
	return &AsyncAgentService_GetTaskLogsClient_RecvMsg_Call{Call: _e.mock.On("RecvMsg", m)}
}

func (_c *AsyncAgentService_GetTaskLogsClient_RecvMsg_Call) Run(run func(m interface{})) *AsyncAgentService_GetTaskLogsClient_RecvMsg_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *AsyncAgentService_GetTaskLogsClient_RecvMsg_Call) Return(_a0 error) *AsyncAgentService_GetTaskLogsClient_RecvMsg_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AsyncAgentService_GetTaskLogsClient_RecvMsg_Call) RunAndReturn(run func(interface{}) error) *AsyncAgentService_GetTaskLogsClient_RecvMsg_Call {
	_c.Call.Return(run)
	return _c
}

// SendMsg provides a mock function with given fields: m
func (_m *AsyncAgentService_GetTaskLogsClient) SendMsg(m interface{}) error {
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

// AsyncAgentService_GetTaskLogsClient_SendMsg_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendMsg'
type AsyncAgentService_GetTaskLogsClient_SendMsg_Call struct {
	*mock.Call
}

// SendMsg is a helper method to define mock.On call
//   - m interface{}
func (_e *AsyncAgentService_GetTaskLogsClient_Expecter) SendMsg(m interface{}) *AsyncAgentService_GetTaskLogsClient_SendMsg_Call {
	return &AsyncAgentService_GetTaskLogsClient_SendMsg_Call{Call: _e.mock.On("SendMsg", m)}
}

func (_c *AsyncAgentService_GetTaskLogsClient_SendMsg_Call) Run(run func(m interface{})) *AsyncAgentService_GetTaskLogsClient_SendMsg_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *AsyncAgentService_GetTaskLogsClient_SendMsg_Call) Return(_a0 error) *AsyncAgentService_GetTaskLogsClient_SendMsg_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AsyncAgentService_GetTaskLogsClient_SendMsg_Call) RunAndReturn(run func(interface{}) error) *AsyncAgentService_GetTaskLogsClient_SendMsg_Call {
	_c.Call.Return(run)
	return _c
}

// Trailer provides a mock function with no fields
func (_m *AsyncAgentService_GetTaskLogsClient) Trailer() metadata.MD {
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

// AsyncAgentService_GetTaskLogsClient_Trailer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Trailer'
type AsyncAgentService_GetTaskLogsClient_Trailer_Call struct {
	*mock.Call
}

// Trailer is a helper method to define mock.On call
func (_e *AsyncAgentService_GetTaskLogsClient_Expecter) Trailer() *AsyncAgentService_GetTaskLogsClient_Trailer_Call {
	return &AsyncAgentService_GetTaskLogsClient_Trailer_Call{Call: _e.mock.On("Trailer")}
}

func (_c *AsyncAgentService_GetTaskLogsClient_Trailer_Call) Run(run func()) *AsyncAgentService_GetTaskLogsClient_Trailer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AsyncAgentService_GetTaskLogsClient_Trailer_Call) Return(_a0 metadata.MD) *AsyncAgentService_GetTaskLogsClient_Trailer_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AsyncAgentService_GetTaskLogsClient_Trailer_Call) RunAndReturn(run func() metadata.MD) *AsyncAgentService_GetTaskLogsClient_Trailer_Call {
	_c.Call.Return(run)
	return _c
}

// NewAsyncAgentService_GetTaskLogsClient creates a new instance of AsyncAgentService_GetTaskLogsClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAsyncAgentService_GetTaskLogsClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *AsyncAgentService_GetTaskLogsClient {
	mock := &AsyncAgentService_GetTaskLogsClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
