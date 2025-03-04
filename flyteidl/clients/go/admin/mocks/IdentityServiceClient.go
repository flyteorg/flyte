// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	context "context"

	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"

	service "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

// IdentityServiceClient is an autogenerated mock type for the IdentityServiceClient type
type IdentityServiceClient struct {
	mock.Mock
}

type IdentityServiceClient_Expecter struct {
	mock *mock.Mock
}

func (_m *IdentityServiceClient) EXPECT() *IdentityServiceClient_Expecter {
	return &IdentityServiceClient_Expecter{mock: &_m.Mock}
}

// UserInfo provides a mock function with given fields: ctx, in, opts
func (_m *IdentityServiceClient) UserInfo(ctx context.Context, in *service.UserInfoRequest, opts ...grpc.CallOption) (*service.UserInfoResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UserInfo")
	}

	var r0 *service.UserInfoResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *service.UserInfoRequest, ...grpc.CallOption) (*service.UserInfoResponse, error)); ok {
		return rf(ctx, in, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *service.UserInfoRequest, ...grpc.CallOption) *service.UserInfoResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*service.UserInfoResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *service.UserInfoRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IdentityServiceClient_UserInfo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UserInfo'
type IdentityServiceClient_UserInfo_Call struct {
	*mock.Call
}

// UserInfo is a helper method to define mock.On call
//   - ctx context.Context
//   - in *service.UserInfoRequest
//   - opts ...grpc.CallOption
func (_e *IdentityServiceClient_Expecter) UserInfo(ctx interface{}, in interface{}, opts ...interface{}) *IdentityServiceClient_UserInfo_Call {
	return &IdentityServiceClient_UserInfo_Call{Call: _e.mock.On("UserInfo",
		append([]interface{}{ctx, in}, opts...)...)}
}

func (_c *IdentityServiceClient_UserInfo_Call) Run(run func(ctx context.Context, in *service.UserInfoRequest, opts ...grpc.CallOption)) *IdentityServiceClient_UserInfo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]grpc.CallOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(grpc.CallOption)
			}
		}
		run(args[0].(context.Context), args[1].(*service.UserInfoRequest), variadicArgs...)
	})
	return _c
}

func (_c *IdentityServiceClient_UserInfo_Call) Return(_a0 *service.UserInfoResponse, _a1 error) *IdentityServiceClient_UserInfo_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *IdentityServiceClient_UserInfo_Call) RunAndReturn(run func(context.Context, *service.UserInfoRequest, ...grpc.CallOption) (*service.UserInfoResponse, error)) *IdentityServiceClient_UserInfo_Call {
	_c.Call.Return(run)
	return _c
}

// NewIdentityServiceClient creates a new instance of IdentityServiceClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIdentityServiceClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *IdentityServiceClient {
	mock := &IdentityServiceClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
