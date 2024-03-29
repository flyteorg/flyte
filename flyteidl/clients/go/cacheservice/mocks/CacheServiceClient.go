// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	context "context"

	cacheservice "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/cacheservice"

	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"
)

// CacheServiceClient is an autogenerated mock type for the CacheServiceClient type
type CacheServiceClient struct {
	mock.Mock
}

type CacheServiceClient_Delete struct {
	*mock.Call
}

func (_m CacheServiceClient_Delete) Return(_a0 *cacheservice.DeleteCacheResponse, _a1 error) *CacheServiceClient_Delete {
	return &CacheServiceClient_Delete{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *CacheServiceClient) OnDelete(ctx context.Context, in *cacheservice.DeleteCacheRequest, opts ...grpc.CallOption) *CacheServiceClient_Delete {
	c_call := _m.On("Delete", ctx, in, opts)
	return &CacheServiceClient_Delete{Call: c_call}
}

func (_m *CacheServiceClient) OnDeleteMatch(matchers ...interface{}) *CacheServiceClient_Delete {
	c_call := _m.On("Delete", matchers...)
	return &CacheServiceClient_Delete{Call: c_call}
}

// Delete provides a mock function with given fields: ctx, in, opts
func (_m *CacheServiceClient) Delete(ctx context.Context, in *cacheservice.DeleteCacheRequest, opts ...grpc.CallOption) (*cacheservice.DeleteCacheResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *cacheservice.DeleteCacheResponse
	if rf, ok := ret.Get(0).(func(context.Context, *cacheservice.DeleteCacheRequest, ...grpc.CallOption) *cacheservice.DeleteCacheResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*cacheservice.DeleteCacheResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *cacheservice.DeleteCacheRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type CacheServiceClient_Get struct {
	*mock.Call
}

func (_m CacheServiceClient_Get) Return(_a0 *cacheservice.GetCacheResponse, _a1 error) *CacheServiceClient_Get {
	return &CacheServiceClient_Get{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *CacheServiceClient) OnGet(ctx context.Context, in *cacheservice.GetCacheRequest, opts ...grpc.CallOption) *CacheServiceClient_Get {
	c_call := _m.On("Get", ctx, in, opts)
	return &CacheServiceClient_Get{Call: c_call}
}

func (_m *CacheServiceClient) OnGetMatch(matchers ...interface{}) *CacheServiceClient_Get {
	c_call := _m.On("Get", matchers...)
	return &CacheServiceClient_Get{Call: c_call}
}

// Get provides a mock function with given fields: ctx, in, opts
func (_m *CacheServiceClient) Get(ctx context.Context, in *cacheservice.GetCacheRequest, opts ...grpc.CallOption) (*cacheservice.GetCacheResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *cacheservice.GetCacheResponse
	if rf, ok := ret.Get(0).(func(context.Context, *cacheservice.GetCacheRequest, ...grpc.CallOption) *cacheservice.GetCacheResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*cacheservice.GetCacheResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *cacheservice.GetCacheRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type CacheServiceClient_GetOrExtendReservation struct {
	*mock.Call
}

func (_m CacheServiceClient_GetOrExtendReservation) Return(_a0 *cacheservice.GetOrExtendReservationResponse, _a1 error) *CacheServiceClient_GetOrExtendReservation {
	return &CacheServiceClient_GetOrExtendReservation{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *CacheServiceClient) OnGetOrExtendReservation(ctx context.Context, in *cacheservice.GetOrExtendReservationRequest, opts ...grpc.CallOption) *CacheServiceClient_GetOrExtendReservation {
	c_call := _m.On("GetOrExtendReservation", ctx, in, opts)
	return &CacheServiceClient_GetOrExtendReservation{Call: c_call}
}

func (_m *CacheServiceClient) OnGetOrExtendReservationMatch(matchers ...interface{}) *CacheServiceClient_GetOrExtendReservation {
	c_call := _m.On("GetOrExtendReservation", matchers...)
	return &CacheServiceClient_GetOrExtendReservation{Call: c_call}
}

// GetOrExtendReservation provides a mock function with given fields: ctx, in, opts
func (_m *CacheServiceClient) GetOrExtendReservation(ctx context.Context, in *cacheservice.GetOrExtendReservationRequest, opts ...grpc.CallOption) (*cacheservice.GetOrExtendReservationResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *cacheservice.GetOrExtendReservationResponse
	if rf, ok := ret.Get(0).(func(context.Context, *cacheservice.GetOrExtendReservationRequest, ...grpc.CallOption) *cacheservice.GetOrExtendReservationResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*cacheservice.GetOrExtendReservationResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *cacheservice.GetOrExtendReservationRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type CacheServiceClient_Put struct {
	*mock.Call
}

func (_m CacheServiceClient_Put) Return(_a0 *cacheservice.PutCacheResponse, _a1 error) *CacheServiceClient_Put {
	return &CacheServiceClient_Put{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *CacheServiceClient) OnPut(ctx context.Context, in *cacheservice.PutCacheRequest, opts ...grpc.CallOption) *CacheServiceClient_Put {
	c_call := _m.On("Put", ctx, in, opts)
	return &CacheServiceClient_Put{Call: c_call}
}

func (_m *CacheServiceClient) OnPutMatch(matchers ...interface{}) *CacheServiceClient_Put {
	c_call := _m.On("Put", matchers...)
	return &CacheServiceClient_Put{Call: c_call}
}

// Put provides a mock function with given fields: ctx, in, opts
func (_m *CacheServiceClient) Put(ctx context.Context, in *cacheservice.PutCacheRequest, opts ...grpc.CallOption) (*cacheservice.PutCacheResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *cacheservice.PutCacheResponse
	if rf, ok := ret.Get(0).(func(context.Context, *cacheservice.PutCacheRequest, ...grpc.CallOption) *cacheservice.PutCacheResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*cacheservice.PutCacheResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *cacheservice.PutCacheRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type CacheServiceClient_ReleaseReservation struct {
	*mock.Call
}

func (_m CacheServiceClient_ReleaseReservation) Return(_a0 *cacheservice.ReleaseReservationResponse, _a1 error) *CacheServiceClient_ReleaseReservation {
	return &CacheServiceClient_ReleaseReservation{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *CacheServiceClient) OnReleaseReservation(ctx context.Context, in *cacheservice.ReleaseReservationRequest, opts ...grpc.CallOption) *CacheServiceClient_ReleaseReservation {
	c_call := _m.On("ReleaseReservation", ctx, in, opts)
	return &CacheServiceClient_ReleaseReservation{Call: c_call}
}

func (_m *CacheServiceClient) OnReleaseReservationMatch(matchers ...interface{}) *CacheServiceClient_ReleaseReservation {
	c_call := _m.On("ReleaseReservation", matchers...)
	return &CacheServiceClient_ReleaseReservation{Call: c_call}
}

// ReleaseReservation provides a mock function with given fields: ctx, in, opts
func (_m *CacheServiceClient) ReleaseReservation(ctx context.Context, in *cacheservice.ReleaseReservationRequest, opts ...grpc.CallOption) (*cacheservice.ReleaseReservationResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *cacheservice.ReleaseReservationResponse
	if rf, ok := ret.Get(0).(func(context.Context, *cacheservice.ReleaseReservationRequest, ...grpc.CallOption) *cacheservice.ReleaseReservationResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*cacheservice.ReleaseReservationResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *cacheservice.ReleaseReservationRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
