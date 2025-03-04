// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	context "context"

	service "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	mock "github.com/stretchr/testify/mock"
)

// AuthMetadataServiceServer is an autogenerated mock type for the AuthMetadataServiceServer type
type AuthMetadataServiceServer struct {
	mock.Mock
}

type AuthMetadataServiceServer_Expecter struct {
	mock *mock.Mock
}

func (_m *AuthMetadataServiceServer) EXPECT() *AuthMetadataServiceServer_Expecter {
	return &AuthMetadataServiceServer_Expecter{mock: &_m.Mock}
}

// GetOAuth2Metadata provides a mock function with given fields: _a0, _a1
func (_m *AuthMetadataServiceServer) GetOAuth2Metadata(_a0 context.Context, _a1 *service.OAuth2MetadataRequest) (*service.OAuth2MetadataResponse, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for GetOAuth2Metadata")
	}

	var r0 *service.OAuth2MetadataResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *service.OAuth2MetadataRequest) (*service.OAuth2MetadataResponse, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *service.OAuth2MetadataRequest) *service.OAuth2MetadataResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*service.OAuth2MetadataResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *service.OAuth2MetadataRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AuthMetadataServiceServer_GetOAuth2Metadata_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetOAuth2Metadata'
type AuthMetadataServiceServer_GetOAuth2Metadata_Call struct {
	*mock.Call
}

// GetOAuth2Metadata is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *service.OAuth2MetadataRequest
func (_e *AuthMetadataServiceServer_Expecter) GetOAuth2Metadata(_a0 interface{}, _a1 interface{}) *AuthMetadataServiceServer_GetOAuth2Metadata_Call {
	return &AuthMetadataServiceServer_GetOAuth2Metadata_Call{Call: _e.mock.On("GetOAuth2Metadata", _a0, _a1)}
}

func (_c *AuthMetadataServiceServer_GetOAuth2Metadata_Call) Run(run func(_a0 context.Context, _a1 *service.OAuth2MetadataRequest)) *AuthMetadataServiceServer_GetOAuth2Metadata_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*service.OAuth2MetadataRequest))
	})
	return _c
}

func (_c *AuthMetadataServiceServer_GetOAuth2Metadata_Call) Return(_a0 *service.OAuth2MetadataResponse, _a1 error) *AuthMetadataServiceServer_GetOAuth2Metadata_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AuthMetadataServiceServer_GetOAuth2Metadata_Call) RunAndReturn(run func(context.Context, *service.OAuth2MetadataRequest) (*service.OAuth2MetadataResponse, error)) *AuthMetadataServiceServer_GetOAuth2Metadata_Call {
	_c.Call.Return(run)
	return _c
}

// GetPublicClientConfig provides a mock function with given fields: _a0, _a1
func (_m *AuthMetadataServiceServer) GetPublicClientConfig(_a0 context.Context, _a1 *service.PublicClientAuthConfigRequest) (*service.PublicClientAuthConfigResponse, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for GetPublicClientConfig")
	}

	var r0 *service.PublicClientAuthConfigResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *service.PublicClientAuthConfigRequest) (*service.PublicClientAuthConfigResponse, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *service.PublicClientAuthConfigRequest) *service.PublicClientAuthConfigResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*service.PublicClientAuthConfigResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *service.PublicClientAuthConfigRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AuthMetadataServiceServer_GetPublicClientConfig_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPublicClientConfig'
type AuthMetadataServiceServer_GetPublicClientConfig_Call struct {
	*mock.Call
}

// GetPublicClientConfig is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *service.PublicClientAuthConfigRequest
func (_e *AuthMetadataServiceServer_Expecter) GetPublicClientConfig(_a0 interface{}, _a1 interface{}) *AuthMetadataServiceServer_GetPublicClientConfig_Call {
	return &AuthMetadataServiceServer_GetPublicClientConfig_Call{Call: _e.mock.On("GetPublicClientConfig", _a0, _a1)}
}

func (_c *AuthMetadataServiceServer_GetPublicClientConfig_Call) Run(run func(_a0 context.Context, _a1 *service.PublicClientAuthConfigRequest)) *AuthMetadataServiceServer_GetPublicClientConfig_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*service.PublicClientAuthConfigRequest))
	})
	return _c
}

func (_c *AuthMetadataServiceServer_GetPublicClientConfig_Call) Return(_a0 *service.PublicClientAuthConfigResponse, _a1 error) *AuthMetadataServiceServer_GetPublicClientConfig_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AuthMetadataServiceServer_GetPublicClientConfig_Call) RunAndReturn(run func(context.Context, *service.PublicClientAuthConfigRequest) (*service.PublicClientAuthConfigResponse, error)) *AuthMetadataServiceServer_GetPublicClientConfig_Call {
	_c.Call.Return(run)
	return _c
}

// NewAuthMetadataServiceServer creates a new instance of AuthMetadataServiceServer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAuthMetadataServiceServer(t interface {
	mock.TestingT
	Cleanup(func())
}) *AuthMetadataServiceServer {
	mock := &AuthMetadataServiceServer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
