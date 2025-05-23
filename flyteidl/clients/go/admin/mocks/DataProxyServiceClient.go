// Code generated by mockery v2.40.3. DO NOT EDIT.

package mocks

import (
	context "context"

	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"

	service "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

// DataProxyServiceClient is an autogenerated mock type for the DataProxyServiceClient type
type DataProxyServiceClient struct {
	mock.Mock
}

type DataProxyServiceClient_Expecter struct {
	mock *mock.Mock
}

func (_m *DataProxyServiceClient) EXPECT() *DataProxyServiceClient_Expecter {
	return &DataProxyServiceClient_Expecter{mock: &_m.Mock}
}

// CreateDownloadLink provides a mock function with given fields: ctx, in, opts
func (_m *DataProxyServiceClient) CreateDownloadLink(ctx context.Context, in *service.CreateDownloadLinkRequest, opts ...grpc.CallOption) (*service.CreateDownloadLinkResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreateDownloadLink")
	}

	var r0 *service.CreateDownloadLinkResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *service.CreateDownloadLinkRequest, ...grpc.CallOption) (*service.CreateDownloadLinkResponse, error)); ok {
		return rf(ctx, in, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *service.CreateDownloadLinkRequest, ...grpc.CallOption) *service.CreateDownloadLinkResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*service.CreateDownloadLinkResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *service.CreateDownloadLinkRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DataProxyServiceClient_CreateDownloadLink_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateDownloadLink'
type DataProxyServiceClient_CreateDownloadLink_Call struct {
	*mock.Call
}

// CreateDownloadLink is a helper method to define mock.On call
//   - ctx context.Context
//   - in *service.CreateDownloadLinkRequest
//   - opts ...grpc.CallOption
func (_e *DataProxyServiceClient_Expecter) CreateDownloadLink(ctx interface{}, in interface{}, opts ...interface{}) *DataProxyServiceClient_CreateDownloadLink_Call {
	return &DataProxyServiceClient_CreateDownloadLink_Call{Call: _e.mock.On("CreateDownloadLink",
		append([]interface{}{ctx, in}, opts...)...)}
}

func (_c *DataProxyServiceClient_CreateDownloadLink_Call) Run(run func(ctx context.Context, in *service.CreateDownloadLinkRequest, opts ...grpc.CallOption)) *DataProxyServiceClient_CreateDownloadLink_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]grpc.CallOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(grpc.CallOption)
			}
		}
		run(args[0].(context.Context), args[1].(*service.CreateDownloadLinkRequest), variadicArgs...)
	})
	return _c
}

func (_c *DataProxyServiceClient_CreateDownloadLink_Call) Return(_a0 *service.CreateDownloadLinkResponse, _a1 error) *DataProxyServiceClient_CreateDownloadLink_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DataProxyServiceClient_CreateDownloadLink_Call) RunAndReturn(run func(context.Context, *service.CreateDownloadLinkRequest, ...grpc.CallOption) (*service.CreateDownloadLinkResponse, error)) *DataProxyServiceClient_CreateDownloadLink_Call {
	_c.Call.Return(run)
	return _c
}

// CreateDownloadLocation provides a mock function with given fields: ctx, in, opts
func (_m *DataProxyServiceClient) CreateDownloadLocation(ctx context.Context, in *service.CreateDownloadLocationRequest, opts ...grpc.CallOption) (*service.CreateDownloadLocationResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreateDownloadLocation")
	}

	var r0 *service.CreateDownloadLocationResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *service.CreateDownloadLocationRequest, ...grpc.CallOption) (*service.CreateDownloadLocationResponse, error)); ok {
		return rf(ctx, in, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *service.CreateDownloadLocationRequest, ...grpc.CallOption) *service.CreateDownloadLocationResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*service.CreateDownloadLocationResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *service.CreateDownloadLocationRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DataProxyServiceClient_CreateDownloadLocation_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateDownloadLocation'
type DataProxyServiceClient_CreateDownloadLocation_Call struct {
	*mock.Call
}

// CreateDownloadLocation is a helper method to define mock.On call
//   - ctx context.Context
//   - in *service.CreateDownloadLocationRequest
//   - opts ...grpc.CallOption
func (_e *DataProxyServiceClient_Expecter) CreateDownloadLocation(ctx interface{}, in interface{}, opts ...interface{}) *DataProxyServiceClient_CreateDownloadLocation_Call {
	return &DataProxyServiceClient_CreateDownloadLocation_Call{Call: _e.mock.On("CreateDownloadLocation",
		append([]interface{}{ctx, in}, opts...)...)}
}

func (_c *DataProxyServiceClient_CreateDownloadLocation_Call) Run(run func(ctx context.Context, in *service.CreateDownloadLocationRequest, opts ...grpc.CallOption)) *DataProxyServiceClient_CreateDownloadLocation_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]grpc.CallOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(grpc.CallOption)
			}
		}
		run(args[0].(context.Context), args[1].(*service.CreateDownloadLocationRequest), variadicArgs...)
	})
	return _c
}

func (_c *DataProxyServiceClient_CreateDownloadLocation_Call) Return(_a0 *service.CreateDownloadLocationResponse, _a1 error) *DataProxyServiceClient_CreateDownloadLocation_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DataProxyServiceClient_CreateDownloadLocation_Call) RunAndReturn(run func(context.Context, *service.CreateDownloadLocationRequest, ...grpc.CallOption) (*service.CreateDownloadLocationResponse, error)) *DataProxyServiceClient_CreateDownloadLocation_Call {
	_c.Call.Return(run)
	return _c
}

// CreateUploadLocation provides a mock function with given fields: ctx, in, opts
func (_m *DataProxyServiceClient) CreateUploadLocation(ctx context.Context, in *service.CreateUploadLocationRequest, opts ...grpc.CallOption) (*service.CreateUploadLocationResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreateUploadLocation")
	}

	var r0 *service.CreateUploadLocationResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *service.CreateUploadLocationRequest, ...grpc.CallOption) (*service.CreateUploadLocationResponse, error)); ok {
		return rf(ctx, in, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *service.CreateUploadLocationRequest, ...grpc.CallOption) *service.CreateUploadLocationResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*service.CreateUploadLocationResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *service.CreateUploadLocationRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DataProxyServiceClient_CreateUploadLocation_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateUploadLocation'
type DataProxyServiceClient_CreateUploadLocation_Call struct {
	*mock.Call
}

// CreateUploadLocation is a helper method to define mock.On call
//   - ctx context.Context
//   - in *service.CreateUploadLocationRequest
//   - opts ...grpc.CallOption
func (_e *DataProxyServiceClient_Expecter) CreateUploadLocation(ctx interface{}, in interface{}, opts ...interface{}) *DataProxyServiceClient_CreateUploadLocation_Call {
	return &DataProxyServiceClient_CreateUploadLocation_Call{Call: _e.mock.On("CreateUploadLocation",
		append([]interface{}{ctx, in}, opts...)...)}
}

func (_c *DataProxyServiceClient_CreateUploadLocation_Call) Run(run func(ctx context.Context, in *service.CreateUploadLocationRequest, opts ...grpc.CallOption)) *DataProxyServiceClient_CreateUploadLocation_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]grpc.CallOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(grpc.CallOption)
			}
		}
		run(args[0].(context.Context), args[1].(*service.CreateUploadLocationRequest), variadicArgs...)
	})
	return _c
}

func (_c *DataProxyServiceClient_CreateUploadLocation_Call) Return(_a0 *service.CreateUploadLocationResponse, _a1 error) *DataProxyServiceClient_CreateUploadLocation_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DataProxyServiceClient_CreateUploadLocation_Call) RunAndReturn(run func(context.Context, *service.CreateUploadLocationRequest, ...grpc.CallOption) (*service.CreateUploadLocationResponse, error)) *DataProxyServiceClient_CreateUploadLocation_Call {
	_c.Call.Return(run)
	return _c
}

// GetData provides a mock function with given fields: ctx, in, opts
func (_m *DataProxyServiceClient) GetData(ctx context.Context, in *service.GetDataRequest, opts ...grpc.CallOption) (*service.GetDataResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetData")
	}

	var r0 *service.GetDataResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *service.GetDataRequest, ...grpc.CallOption) (*service.GetDataResponse, error)); ok {
		return rf(ctx, in, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *service.GetDataRequest, ...grpc.CallOption) *service.GetDataResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*service.GetDataResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *service.GetDataRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DataProxyServiceClient_GetData_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetData'
type DataProxyServiceClient_GetData_Call struct {
	*mock.Call
}

// GetData is a helper method to define mock.On call
//   - ctx context.Context
//   - in *service.GetDataRequest
//   - opts ...grpc.CallOption
func (_e *DataProxyServiceClient_Expecter) GetData(ctx interface{}, in interface{}, opts ...interface{}) *DataProxyServiceClient_GetData_Call {
	return &DataProxyServiceClient_GetData_Call{Call: _e.mock.On("GetData",
		append([]interface{}{ctx, in}, opts...)...)}
}

func (_c *DataProxyServiceClient_GetData_Call) Run(run func(ctx context.Context, in *service.GetDataRequest, opts ...grpc.CallOption)) *DataProxyServiceClient_GetData_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]grpc.CallOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(grpc.CallOption)
			}
		}
		run(args[0].(context.Context), args[1].(*service.GetDataRequest), variadicArgs...)
	})
	return _c
}

func (_c *DataProxyServiceClient_GetData_Call) Return(_a0 *service.GetDataResponse, _a1 error) *DataProxyServiceClient_GetData_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DataProxyServiceClient_GetData_Call) RunAndReturn(run func(context.Context, *service.GetDataRequest, ...grpc.CallOption) (*service.GetDataResponse, error)) *DataProxyServiceClient_GetData_Call {
	_c.Call.Return(run)
	return _c
}

// NewDataProxyServiceClient creates a new instance of DataProxyServiceClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDataProxyServiceClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *DataProxyServiceClient {
	mock := &DataProxyServiceClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
