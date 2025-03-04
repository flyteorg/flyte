// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	context "context"

	service "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	mock "github.com/stretchr/testify/mock"
)

// IdentityServiceServer is an autogenerated mock type for the IdentityServiceServer type
type IdentityServiceServer struct {
	mock.Mock
}

type IdentityServiceServer_Expecter struct {
	mock *mock.Mock
}

func (_m *IdentityServiceServer) EXPECT() *IdentityServiceServer_Expecter {
	return &IdentityServiceServer_Expecter{mock: &_m.Mock}
}

// UserInfo provides a mock function with given fields: _a0, _a1
func (_m *IdentityServiceServer) UserInfo(_a0 context.Context, _a1 *service.UserInfoRequest) (*service.UserInfoResponse, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for UserInfo")
	}

	var r0 *service.UserInfoResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *service.UserInfoRequest) (*service.UserInfoResponse, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *service.UserInfoRequest) *service.UserInfoResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*service.UserInfoResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *service.UserInfoRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IdentityServiceServer_UserInfo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UserInfo'
type IdentityServiceServer_UserInfo_Call struct {
	*mock.Call
}

// UserInfo is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *service.UserInfoRequest
func (_e *IdentityServiceServer_Expecter) UserInfo(_a0 interface{}, _a1 interface{}) *IdentityServiceServer_UserInfo_Call {
	return &IdentityServiceServer_UserInfo_Call{Call: _e.mock.On("UserInfo", _a0, _a1)}
}

func (_c *IdentityServiceServer_UserInfo_Call) Run(run func(_a0 context.Context, _a1 *service.UserInfoRequest)) *IdentityServiceServer_UserInfo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*service.UserInfoRequest))
	})
	return _c
}

func (_c *IdentityServiceServer_UserInfo_Call) Return(_a0 *service.UserInfoResponse, _a1 error) *IdentityServiceServer_UserInfo_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *IdentityServiceServer_UserInfo_Call) RunAndReturn(run func(context.Context, *service.UserInfoRequest) (*service.UserInfoResponse, error)) *IdentityServiceServer_UserInfo_Call {
	_c.Call.Return(run)
	return _c
}

// NewIdentityServiceServer creates a new instance of IdentityServiceServer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIdentityServiceServer(t interface {
	mock.TestingT
	Cleanup(func())
}) *IdentityServiceServer {
	mock := &IdentityServiceServer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
